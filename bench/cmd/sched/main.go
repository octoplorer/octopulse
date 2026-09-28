// Command sched models the scheduler's memory footprint: one goroutine and one
// ticker per target, each issuing an HTTP check against an in-process server so
// the probe stays self-contained. It prints the Go runtime's own view; read
// container memory via `docker stats` for the number that matters.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func main() {
	targets := flag.Int("targets", 1000, "number of monitored targets (goroutines)")
	interval := flag.Duration("interval", 60*time.Second, "check interval per target")
	settle := flag.Duration("settle", 2*time.Second, "wait before printing stats")
	url := flag.String("url", "http://127.0.0.1:18080/", "URL each target checks")
	flag.Parse()

	target := &http.Server{Addr: "127.0.0.1:18080", Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}
	go func() { _ = target.ListenAndServe() }()

	client := &http.Client{Timeout: 5 * time.Second}
	for range *targets {
		go func() {
			ticker := time.NewTicker(*interval)
			defer ticker.Stop()
			for range ticker.C {
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, *url, nil)
				if err != nil {
					continue
				}
				if resp, err := client.Do(req); err == nil {
					_ = resp.Body.Close()
				}
			}
		}()
	}

	time.Sleep(*settle)
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("targets=%d interval=%s HeapAlloc=%.1fMB Sys=%.1fMB goroutines=%d\n",
		*targets, *interval, float64(m.HeapAlloc)/(1<<20), float64(m.Sys)/(1<<20), runtime.NumGoroutine())

	select {}
}
