package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeWaitsForActiveHTTPHandlers(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := l.Addr().String()
	l.Close()
	started := make(chan struct{})
	release := make(chan struct{})
	srv := &http.Server{
		Addr: address,
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				close(started)
				<-release
				w.Write([]byte("drained"))
			},
		),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- serveHTTP(
			ctx,
			cancel,
			srv,
			nil,
		)
	}()
	response := make(chan error, 1)
	go func() {
		var res *http.Response
		var err error
		for i := 0; i < 100; i++ {
			res, err = http.Get("http://" + address)
			if err == nil {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if err != nil {
			response <- err
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		response <- err
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not start")
	}
	cancel()
	select {
	case err := <-done:
		t.Fatalf("returned before handler drained: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown failed to finish")
	}
	if err := <-response; err != nil {
		t.Fatal(err)
	}
}

func TestAdditionalListenerFailureClosesApplicationListener(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	available, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := available.Addr().String()
	available.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = serveHTTP(
		ctx,
		cancel,
		&http.Server{Addr: address},
		nil,
		&http.Server{Addr: occupied.Addr().String()},
	)
	if err == nil {
		t.Fatal("metrics bind failure was ignored")
	}
	if ctx.Err() == nil {
		t.Fatal("runtime was not cancelled after bind failure")
	}
	reopened, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("application listener leaked: %v", err)
	}
	reopened.Close()
}
