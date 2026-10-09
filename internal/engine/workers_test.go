package engine

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/probe"
)

func awaitEngineStats(t *testing.T, e *Engine, want EngineStats) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if e.Stats() == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("engine stats = %+v, want %+v", e.Stats(), want)
}

func TestSchedulerQueuesOneRoundPerMonitorWithoutWaitingGoroutines(t *testing.T) {
	h := newHarness(t, activeHTTP())
	const monitors = 48
	for i := 1; i < monitors; i++ {
		m := h.m
		m.ID = fmt.Sprintf("queued-%02d", i)
		h.save(t, m)
	}
	h.e.Concurrency = 2
	h.e.PollInterval = time.Hour
	release := make(chan struct{})
	var attempts atomic.Int32
	h.e.Attempt = func(ctx context.Context, _ domain.Monitor) probe.Result {
		attempts.Add(1)
		select {
		case <-release:
			return probe.Result{Success: true}
		case <-ctx.Done():
			return probe.Result{Error: "cancelled"}
		}
	}
	baseline := runtime.NumGoroutine()
	if err := h.e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.e.Stop)
	awaitEngineStats(t, h.e, EngineStats{Active: 2, Queued: monitors - 2})
	if growth := runtime.NumGoroutine() - baseline; growth > 12 {
		t.Fatalf("queued checks created waiting goroutines: growth=%d", growth)
	}
	for i := 0; i < 3; i++ {
		if err := h.e.poll(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := h.e.Check(context.Background(), h.m.ID); !errors.Is(err, ErrBusy) {
			t.Fatalf("duplicate queued manual request = %v", err)
		}
	}
	awaitEngineStats(t, h.e, EngineStats{Active: 2, Queued: monitors - 2})
	close(release)
	awaitEngineStats(t, h.e, EngineStats{})
	if attempts.Load() != monitors {
		t.Fatalf("scheduled rounds lost or duplicated: attempts=%d", attempts.Load())
	}
}

func TestQueuedManualCheckOutlivesCallerCancellation(t *testing.T) {
	h := newHarness(t, activeHTTP())
	blockedID := h.m.ID
	m := h.m
	m.ID = "manual"
	h.save(t, m)
	h.e.schedules[m.ID] = schedule{version: m.ConfigVersion, enabled: true, next: h.clock.Load() + 30000}
	h.e.Concurrency = 1
	h.e.PollInterval = time.Hour
	release := make(chan struct{})
	h.e.Attempt = func(ctx context.Context, m domain.Monitor) probe.Result {
		if m.ID == blockedID {
			select {
			case <-release:
			case <-ctx.Done():
				return probe.Result{Error: "cancelled"}
			}
		}
		return probe.Result{Success: true}
	}
	if err := h.e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.e.Stop)
	awaitEngineStats(t, h.e, EngineStats{Active: 1})
	caller, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- h.e.Check(caller, m.ID) }()
	awaitEngineStats(t, h.e, EngineStats{Active: 1, Queued: 1})
	cancel()
	if err := <-result; !errors.Is(err, ErrCheckAccepted) || !errors.Is(err, context.Canceled) {
		t.Fatalf("queued manual cancellation = %v", err)
	}
	if err := h.e.Check(context.Background(), m.ID); !errors.Is(err, ErrBusy) {
		t.Fatalf("caller cancellation released queued ownership: %v", err)
	}
	close(release)
	awaitEngineStats(t, h.e, EngineStats{})
	rows, err := h.s.ListRounds(
		context.Background(),
		m.ID,
		0,
		10,
	)
	if err != nil || len(rows) != 1 || !rows[0].Success {
		t.Fatalf("accepted queued round was not committed: rows=%+v err=%v", rows, err)
	}
}

func TestStopCompletesQueuedManualRequestsWithoutProbing(t *testing.T) {
	h := newHarness(t, activeHTTP())
	blockedID := h.m.ID
	m := h.m
	m.ID = "manual"
	h.save(t, m)
	h.e.schedules[m.ID] = schedule{version: m.ConfigVersion, enabled: true, next: h.clock.Load() + 30000}
	h.e.Concurrency = 1
	h.e.PollInterval = time.Hour
	var attempts atomic.Int32
	h.e.Attempt = func(ctx context.Context, m domain.Monitor) probe.Result {
		attempts.Add(1)
		if m.ID == blockedID {
			<-ctx.Done()
		}
		return probe.Result{Error: "cancelled"}
	}
	if err := h.e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.e.Stop)
	awaitEngineStats(t, h.e, EngineStats{Active: 1})
	result := make(chan error, 1)
	go func() { result <- h.e.Check(context.Background(), m.ID) }()
	awaitEngineStats(t, h.e, EngineStats{Active: 1, Queued: 1})
	h.e.Stop()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("queued request shutdown = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown left a queued manual caller waiting")
	}
	if attempts.Load() != 1 || h.e.Stats() != (EngineStats{}) {
		t.Fatalf("shutdown executed queued work: attempts=%d stats=%+v", attempts.Load(), h.e.Stats())
	}
}

func TestConfigurationChangeReplacesQueuedRoundBeforeCapacityReturns(t *testing.T) {
	h := newHarness(t, activeHTTP())
	blockedID := h.m.ID
	m := h.m
	m.ID = "queued"
	h.save(t, m)
	h.e.Concurrency = 1
	h.e.PollInterval = time.Hour
	release := make(chan struct{})
	checkedVersion := make(chan int64, 2)
	h.e.Attempt = func(ctx context.Context, m domain.Monitor) probe.Result {
		if m.ID == blockedID {
			select {
			case <-release:
			case <-ctx.Done():
				return probe.Result{Error: "cancelled"}
			}
		} else {
			checkedVersion <- m.ConfigVersion
		}
		return probe.Result{Success: true}
	}
	if err := h.e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.e.Stop)
	awaitEngineStats(t, h.e, EngineStats{Active: 1, Queued: 1})
	// Removing a queued cancelled task must not wait for a free worker, and
	// the next poll must accept the replacement even with a one-hour cadence.
	m.ConfigVersion++
	h.save(t, m)
	if err := h.e.NotifyConfigurationChanged(context.Background(), m.ID); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		h.e.mu.Lock()
		active := h.e.running[m.ID]
		h.e.mu.Unlock()
		if active.version == m.ConfigVersion {
			break
		}
		time.Sleep(time.Millisecond)
	}
	h.e.mu.Lock()
	version := h.e.running[m.ID].version
	h.e.mu.Unlock()
	if version != m.ConfigVersion {
		t.Fatalf("cancelled queued round retained admission: version=%d", version)
	}
	close(release)
	select {
	case version := <-checkedVersion:
		if version != m.ConfigVersion {
			t.Fatalf("stale queued configuration executed: version=%d", version)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("replacement queued round was not executed")
	}
	awaitEngineStats(t, h.e, EngineStats{})
	select {
	case version := <-checkedVersion:
		t.Fatalf("configuration evaluated more than once: version=%d", version)
	default:
	}
}
