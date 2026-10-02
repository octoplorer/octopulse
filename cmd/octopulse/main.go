package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/octoplorer/octopulse/internal/config"
	"github.com/octoplorer/octopulse/internal/engine"
	"github.com/octoplorer/octopulse/internal/notify"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/server"
	"github.com/octoplorer/octopulse/internal/statistics"
	"github.com/octoplorer/octopulse/internal/store"
)

func main() {
	if e := run(); e != nil {
		slog.Error("octopulse stopped", "error", e.Error())
		os.Exit(1)
	}
}
func run() error {
	cfg, e := config.Load()
	if e != nil {
		return e
	}
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command == "openapi" {
		s := server.New(nil, nil, cfg)
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(s.API.OpenAPI())
	}
	if command != "serve" && command != "backup" {
		return fmt.Errorf("usage: octopulse [serve|openapi|backup DESTINATION]")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	vault, e := security.Open(cfg.DataDir, cfg.EncryptionKey)
	if e != nil {
		return fmt.Errorf("initialize secret vault: %w", e)
	}
	st, e := store.Open(ctx, store.Config{Driver: cfg.Driver, DSN: cfg.DSN, MaxConnections: cfg.DBMaxConnections})
	if e != nil {
		return fmt.Errorf("open database: %w", e)
	}
	defer st.Close()
	if command == "backup" {
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: octopulse backup DESTINATION")
		}
		return st.BackupSQLite(ctx, os.Args[2])
	}
	s := server.New(st, vault, cfg)
	runtimeCtx, cancelRuntime := context.WithCancel(ctx)
	scheduler := engine.New(st, probe.NewRunner(s))
	scheduler.OnError = func(error) { slog.Error("collection operation failed") }
	if e = scheduler.Start(runtimeCtx); e != nil {
		cancelRuntime()
		return fmt.Errorf("start monitoring engine: %w", e)
	}
	stats := statistics.New(st)
	stats.Interval = time.Duration(cfg.StatisticsIntervalSeconds) * time.Second
	stats.OnError = func(error) { slog.Error("statistics persistence operation failed") }
	if e = stats.Start(runtimeCtx); e != nil {
		cancelRuntime()
		scheduler.Stop()
		return fmt.Errorf("start statistics: %w", e)
	}
	deliveries := notify.New(st, s)
	var workers sync.WaitGroup
	workers.Add(2)
	go func() { defer workers.Done(); deliveries.Start(runtimeCtx) }()
	go func() { defer workers.Done(); s.Beszel.Start(runtimeCtx) }()
	defer func() { cancelRuntime(); scheduler.Stop(); stats.Stop(); workers.Wait() }()
	s.Check = scheduler.Check
	s.NextCheck = scheduler.NextCheckAt
	s.Changed = scheduler.NotifyConfigurationChanged
	s.Heartbeat = scheduler.Heartbeat
	s.Stats = stats.Availability
	s.Latency = stats.Latency
	s.TestChannel = deliveries.TestChannel
	s.Wake = func() { scheduler.Wake(); stats.Wake() }
	srv := &http.Server{Addr: cfg.Address, Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	slog.Info("octopulse ready", "address", cfg.Address, "database", cfg.Driver)
	return serveHTTP(ctx, stop, srv, st.LockLost())
}

// Keep the database and workers alive until active HTTP handlers finish draining.
func serveHTTP(ctx context.Context, cancel context.CancelFunc, srv *http.Server, lockLost <-chan struct{}) error {
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		select {
		case <-ctx.Done():
		case <-lockLost:
			slog.Error("database instance lock lost")
			cancel()
		}
		shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := srv.Shutdown(shutdown); err != nil {
			srv.Close()
		}
	}()
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		cancel()
		<-shutdownDone
		return err
	}
	<-shutdownDone
	return nil
}
