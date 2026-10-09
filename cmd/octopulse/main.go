package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/octoplorer/octopulse/internal/config"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/engine"
	"github.com/octoplorer/octopulse/internal/notify"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/retention"
	"github.com/octoplorer/octopulse/internal/secrets"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/server"
	"github.com/octoplorer/octopulse/internal/statistics"
	"github.com/octoplorer/octopulse/internal/store"
	"github.com/octoplorer/octopulse/internal/telemetry"
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
	metrics := telemetry.New()
	metrics.ObserveStore(st)
	s := server.New(st, vault, cfg)
	s.Metrics = metrics
	runtimeCtx, cancelRuntime := context.WithCancel(ctx)
	resolver := secrets.New(st, vault)
	scheduler := engine.New(st, probe.NewRunner(resolver))
	scheduler.Concurrency = cfg.ProbeConcurrency
	scheduler.OnRoundStart = metrics.QueueWait
	attempt := scheduler.Attempt
	scheduler.Attempt = func(ctx context.Context, m domain.Monitor) probe.Result {
		started := time.Now()
		result := attempt(ctx, m)
		metrics.Probe(m.Type, result.Success, time.Since(started))
		return result
	}
	scheduler.OnError = func(err error) {
		metrics.OperationError("collection")
		telemetry.LogError(runtimeCtx, "collection", err)
	}
	metrics.Gauge(
		"probe_active",
		"Currently executing probe rounds.",
		func() float64 {
			return float64(scheduler.Stats().Active)
		},
	)
	metrics.Gauge(
		"probe_queued",
		"Probe rounds waiting for execution.",
		func() float64 {
			return float64(scheduler.Stats().Queued)
		},
	)
	if e = scheduler.Start(runtimeCtx); e != nil {
		cancelRuntime()
		return fmt.Errorf("start monitoring engine: %w", e)
	}
	stats := statistics.New(st)
	stats.Interval = time.Duration(cfg.StatisticsIntervalSeconds) * time.Second
	stats.OnError = func(err error) {
		metrics.OperationError("statistics")
		telemetry.LogError(runtimeCtx, "statistics", err)
	}
	if e = stats.Start(runtimeCtx); e != nil {
		cancelRuntime()
		scheduler.Stop()
		return fmt.Errorf("start statistics: %w", e)
	}
	deliveries := notify.New(st, resolver)
	deliveries.OnError = func(err error) {
		metrics.OperationError("notification")
		telemetry.LogError(runtimeCtx, "notification.persistence", err)
	}
	deliveries.OnDelivery = metrics.Delivery
	cleanup := retention.New(st)
	cleanup.OperationHistoryDays = cfg.OperationHistoryDays
	cleanup.OnError = func(err error) {
		metrics.OperationError("retention")
		telemetry.LogError(runtimeCtx, "retention", err)
	}
	var workers sync.WaitGroup
	workers.Add(3)
	go func() {
		defer workers.Done()
		deliveries.Start(runtimeCtx)
	}()
	go func() {
		defer workers.Done()
		cleanup.Start(runtimeCtx)
	}()
	go func() {
		defer workers.Done()
		s.Beszel.Start(runtimeCtx)
	}()
	defer func() {
		cancelRuntime()
		scheduler.Stop()
		stats.Stop()
		workers.Wait()
	}()
	s.Check = scheduler.Check
	s.NextCheck = scheduler.NextCheckAt
	s.Changed = scheduler.NotifyConfigurationChanged
	s.Heartbeat = scheduler.Heartbeat
	s.Stats = stats.Availability
	s.StatsBatch = stats.AvailabilityBatch
	s.DailyStatsBatch = stats.DailyAvailabilityBatch
	s.LatencyBatch = stats.LatencyBatch
	s.Latency = stats.Latency
	s.TestChannel = deliveries.TestChannel
	s.Wake = func() {
		scheduler.Wake()
		stats.Wake()
	}
	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info(
		"octopulse starting",
		"address",
		cfg.Address,
		"database",
		cfg.Driver,
	)
	additional := []*http.Server{}
	if cfg.MetricsAddress != "" {
		mux := http.NewServeMux()
		mux.Handle("GET /metrics", metrics.Handler())
		additional = append(
			additional,
			&http.Server{
				Addr:              cfg.MetricsAddress,
				Handler:           mux,
				ReadHeaderTimeout: 5 * time.Second,
				ReadTimeout:       5 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       30 * time.Second,
			},
		)
	}
	return serveHTTP(
		ctx,
		stop,
		srv,
		st.LockLost(),
		additional...,
	)
}

// Keep the database and workers alive until active HTTP handlers finish draining.
func serveHTTP(
	ctx context.Context,
	cancel context.CancelFunc,
	srv *http.Server,
	lockLost <-chan struct{},
	additional ...*http.Server,
) error {
	servers := append([]*http.Server{srv}, additional...)
	listeners := make([]net.Listener, 0, len(servers))
	for _, server := range servers {
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			for _, opened := range listeners {
				opened.Close()
			}
			cancel()
			return err
		}
		listeners = append(listeners, listener)
	}
	slog.Info(
		"HTTP listeners ready",
		"address",
		srv.Addr,
		"additional_listeners",
		len(additional),
	)
	failures := make(chan error, len(servers))
	for i, server := range servers {
		go func() { failures <- server.Serve(listeners[i]) }()
	}
	var result error
	select {
	case <-ctx.Done():
	case <-lockLost:
		result = fmt.Errorf("database instance lock lost")
	case err := <-failures:
		if err != http.ErrServerClosed {
			result = err
		}
	}
	cancel()
	shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	var draining sync.WaitGroup
	for _, server := range servers {
		draining.Add(1)
		go func() {
			defer draining.Done()
			if err := server.Shutdown(shutdown); err != nil {
				server.Close()
			}
		}()
	}
	draining.Wait()
	return result
}
