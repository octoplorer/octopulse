package engine

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/store"
)

type harness struct {
	e     *Engine
	s     *store.Store
	clock atomic.Int64
	m     domain.Monitor
}

func newHarness(t *testing.T, m domain.Monitor) *harness {
	t.Helper()
	s, err := store.Open(context.Background(), store.Config{Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "engine.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := &harness{s: s, m: m}
	h.clock.Store(1700000000000)
	h.e = New(s, nil)
	h.e.Now = func() time.Time { return time.UnixMilli(h.clock.Load()) }
	m.ID = "monitor"
	m.Name = "test"
	m.Enabled = true
	m.ConfigVersion = 1
	m.CreatedAt = h.clock.Load()
	m.NotifyRecovery = true
	m.NotificationChannelIDs = []string{"one", "two"}
	m.Defaults()
	h.m = m
	if err = m.Validate(); err != nil {
		t.Fatal(err)
	}
	h.save(t, m)
	for _, id := range m.NotificationChannelIDs {
		if err = s.Put(context.Background(), "channels", id, domain.Channel{ID: id, Name: id, Enabled: true}); err != nil {
			t.Fatal(err)
		}
	}
	if err = h.e.NotifyConfigurationChanged(context.Background(), m.ID); err != nil {
		t.Fatal(err)
	}
	return h
}
func (h *harness) save(t *testing.T, m domain.Monitor) {
	t.Helper()
	data, _ := json.Marshal(m)
	record, err := h.s.GetMonitor(context.Background(), m.ID)
	generation := int64(1)
	if err == nil {
		generation = record.Generation
	}
	if err = h.s.WithTx(context.Background(), func(tx *store.Tx) error {
		if err := tx.Put(context.Background(), "monitors", m.ID, m); err != nil {
			return err
		}
		return tx.PutMonitor(context.Background(), store.Monitor{ID: m.ID, Kind: m.Type, Enabled: m.Enabled, ConfigVersion: m.ConfigVersion, Generation: generation, IntervalMS: int64(m.IntervalSeconds) * 1000, ConfigJSON: data})
	}); err != nil {
		t.Fatal(err)
	}
	h.m = m
}
func (h *harness) state(t *testing.T) store.Runtime {
	t.Helper()
	runtime, err := h.s.GetRuntime(context.Background(), h.m.ID)
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}
func (h *harness) check(t *testing.T) {
	t.Helper()
	if err := h.e.Check(context.Background(), h.m.ID); err != nil {
		t.Fatal(err)
	}
}
func (h *harness) events(t *testing.T, kind string) []store.Event {
	t.Helper()
	events, err := h.s.ListEvents(context.Background(), h.m.ID, 0, 1000)
	if err != nil {
		t.Fatal(err)
	}
	var selected []store.Event
	for _, event := range events {
		if event.Kind == kind {
			selected = append(selected, event)
		}
	}
	return selected
}
func activeHTTP() domain.Monitor {
	return domain.Monitor{Type: domain.MonitorHTTP, IntervalSeconds: 30, TimeoutSeconds: 2, Retries: 2, HTTP: &domain.HTTPConfig{URL: "https://example.test"}}
}

func TestAttemptsThresholdsStateIntervalsAndAtomicOutbox(t *testing.T) {
	m := activeHTTP()
	m.FailureThreshold = 2
	m.RecoveryThreshold = 2
	h := newHarness(t, m)
	var outcomes []bool
	var calls int
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		h.clock.Add(10)
		success := outcomes[calls]
		calls++
		return probe.Result{Success: success, LatencyMs: 10}
	}
	run := func(results []bool, wantState string, attempts int) {
		t.Helper()
		outcomes = results
		calls = 0
		h.check(t)
		if got := h.state(t).State; got != wantState {
			t.Fatalf("state=%s want=%s", got, wantState)
		}
		rounds, err := h.s.ListRounds(context.Background(), h.m.ID, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(rounds) != 1 || len(rounds[0].Attempts) != attempts || rounds[0].LatencyMS != int64(attempts*10) {
			t.Fatalf("rounds=%+v", rounds)
		}
		h.clock.Add(30000)
	}
	run([]bool{false, false, false}, domain.StateUnknown, 3)
	run([]bool{false, false, true}, domain.StateUnknown, 3)
	run([]bool{true}, domain.StateUp, 1)
	if len(h.events(t, "up")) != 0 {
		t.Fatal("initial Up notified")
	}
	run([]bool{false, false, false}, domain.StateUp, 3)
	run([]bool{false, false, false}, domain.StateDown, 3)
	run([]bool{true}, domain.StateDown, 1)
	run([]bool{true}, domain.StateUp, 1)
	if len(h.events(t, "down")) != 1 || len(h.events(t, "up")) != 1 {
		t.Fatal("confirmed transitions must notify exactly once")
	}
	deliveries, err := h.s.ListDeliveries(context.Background(), 100)
	if err != nil || len(deliveries) != 4 {
		t.Fatalf("deliveries=%+v err=%v", deliveries, err)
	}
	var downCycle, upCycle string
	for _, delivery := range deliveries {
		var payload NotificationPayload
		_ = json.Unmarshal(delivery.Payload, &payload)
		if payload.Kind == "down" {
			downCycle = payload.CycleID
		} else if payload.Kind == "up" {
			upCycle = payload.CycleID
		}
	}
	if downCycle == "" || downCycle != upCycle {
		t.Fatal("recovery lost its fault cycle")
	}
	intervals, err := h.s.Intervals(context.Background(), h.m.ID, h.m.CreatedAt, h.clock.Load()+1)
	if err != nil {
		t.Fatal(err)
	}
	states := []string{}
	for _, interval := range intervals {
		states = append(states, interval.State)
	}
	want := []string{"unknown", "up", "down", "up"}
	if len(states) != len(want) {
		t.Fatal(states)
	}
	for i, state := range want {
		if states[i] != state {
			t.Fatal(states)
		}
	}
}

func TestNoOverlapPauseAndObsoleteConfigurationRejection(t *testing.T) {
	h := newHarness(t, activeHTTP())
	started := make(chan struct{})
	release := make(chan struct{})
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		close(started)
		<-release
		return probe.Result{Success: true}
	}
	result := make(chan error, 1)
	go func() { result <- h.e.Check(context.Background(), h.m.ID) }()
	<-started
	if err := h.e.Check(context.Background(), h.m.ID); !errors.Is(err, ErrBusy) {
		t.Fatalf("overlap err=%v", err)
	}
	h.clock.Add(100)
	h.m.Enabled = false
	h.m.ConfigVersion++
	h.save(t, h.m)
	if err := h.e.NotifyConfigurationChanged(context.Background(), h.m.ID); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-result; err == nil {
		t.Fatal("obsolete result committed")
	}
	if h.state(t).State != StatePaused {
		t.Fatal(h.state(t))
	}
	rounds, _ := h.s.ListRounds(context.Background(), h.m.ID, 0, 100)
	if len(rounds) != 0 {
		t.Fatal("paused obsolete result stored")
	}
	if err := h.e.Check(context.Background(), h.m.ID); !errors.Is(err, ErrPaused) {
		t.Fatal(err)
	}
	h.m.Enabled = true
	h.m.ConfigVersion++
	h.save(t, h.m)
	if err := h.e.NotifyConfigurationChanged(context.Background(), h.m.ID); err != nil {
		t.Fatal(err)
	}
	if h.state(t).State != domain.StateUnknown {
		t.Fatal("resume carried an old state")
	}
}

func TestParentDeadlineBoundsRetryWaitAndExplicitCancellationDiscardsObservation(t *testing.T) {
	m := activeHTTP()
	m.RetryDelaySeconds = 5
	h := newHarness(t, m)
	var attempts atomic.Int32
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		attempts.Add(1)
		return probe.Result{Error: "failed"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := h.e.Check(ctx, h.m.ID)
	if err != nil || attempts.Load() != 1 || time.Since(start) > time.Second {
		t.Fatalf("err=%v attempts=%d elapsed=%s", err, attempts.Load(), time.Since(start))
	}
	if h.state(t).State != domain.StateDown {
		t.Fatal("failed attempt must confirm the round when the remaining retry budget expires")
	}
	h2 := newHarness(t, activeHTTP())
	began := make(chan struct{})
	h2.e.Attempt = func(ctx context.Context, _ domain.Monitor) probe.Result {
		close(began)
		<-ctx.Done()
		return probe.Result{Error: "cancelled"}
	}
	cancelledCtx, cancelNow := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- h2.e.Check(cancelledCtx, h2.m.ID) }()
	<-began
	cancelNow()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if h2.state(t).State != domain.StateUnknown {
		t.Fatal("cancelled collection caused target Down")
	}
}

func TestStartupCutsKnownStateAtWatermarkAndChecksImmediately(t *testing.T) {
	h := newHarness(t, activeHTTP())
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: true} }
	h.check(t)
	oldEnd := h.clock.Add(5000)
	if err := h.s.WithTx(context.Background(), func(tx *store.Tx) error { return tx.PutWatermark(context.Background(), collectionWatermark, oldEnd) }); err != nil {
		t.Fatal(err)
	}
	h.clock.Add(20000)
	restarted := New(h.s, nil)
	restarted.Now = h.e.Now
	observed := make(chan struct{}, 1)
	restarted.Attempt = func(context.Context, domain.Monitor) probe.Result {
		select {
		case observed <- struct{}{}:
		default:
		}
		return probe.Result{Success: true}
	}
	restarted.PollInterval = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := restarted.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer restarted.Stop()
	select {
	case <-observed:
	case <-time.After(time.Second):
		t.Fatal("restart did not immediately probe")
	}
	deadline := time.Now().Add(time.Second)
	for h.state(t).State != domain.StateUp && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	intervals, err := h.s.Intervals(context.Background(), h.m.ID, h.m.CreatedAt, h.clock.Load()+1)
	if err != nil {
		t.Fatal(err)
	}
	gapFound := false
	for _, interval := range intervals {
		if interval.State == domain.StateUp && interval.StartedAt < oldEnd && (interval.EndedAt == nil || *interval.EndedAt > oldEnd) {
			t.Fatal("Up carried over a collection gap")
		}
		if interval.State == domain.StateUnknown && interval.StartedAt == oldEnd && interval.EndedAt != nil && *interval.EndedAt == h.clock.Load() {
			gapFound = true
		}
	}
	if !gapFound {
		t.Fatalf("missing explicit gap: %+v", intervals)
	}
}

func TestFixedCadenceSkipsMissedChecksAndPreservesNoOverlap(t *testing.T) {
	h := newHarness(t, activeHTTP())
	var attempts atomic.Int32
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		attempts.Add(1)
		return probe.Result{Success: true}
	}
	h.e.ctx = context.Background()
	if err := h.e.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.e.wg.Wait()
	if attempts.Load() != 1 {
		t.Fatal(attempts.Load())
	}
	h.clock.Add(125000)
	if err := h.e.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.e.wg.Wait()
	if attempts.Load() != 2 {
		t.Fatalf("missed checks were replayed: %d", attempts.Load())
	}
	h.e.mu.Lock()
	next := h.e.schedules[h.m.ID].next
	h.e.mu.Unlock()
	if next != h.m.CreatedAt+150000 {
		t.Fatalf("cadence drifted: %d", next-h.m.CreatedAt)
	}
}

func TestBoundedGlobalConcurrency(t *testing.T) {
	h := newHarness(t, activeHTTP())
	h.e.semaphore = make(chan struct{}, 2)
	release := make(chan struct{})
	var active, maxActive atomic.Int32
	var wg sync.WaitGroup
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		n := active.Add(1)
		for current := maxActive.Load(); n > current && !maxActive.CompareAndSwap(current, n); current = maxActive.Load() {
		}
		<-release
		active.Add(-1)
		return probe.Result{Success: true}
	}
	for i := 0; i < 5; i++ {
		m := h.m
		m.ID = domain.ID()
		h.save(t, m)
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := h.e.Check(context.Background(), id); err != nil {
				t.Error(err)
			}
		}(m.ID)
	}
	deadline := time.Now().Add(time.Second)
	for maxActive.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if maxActive.Load() != 2 {
		t.Fatal("expected two simultaneous attempts")
	}
	close(release)
	wg.Wait()
	if maxActive.Load() != 2 {
		t.Fatalf("concurrency exceeded limit: %d", maxActive.Load())
	}
}
