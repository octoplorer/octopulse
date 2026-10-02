package engine

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/store"
)

func startManagedEngine(t *testing.T, h *harness) {
	t.Helper()
	// The fixture starts with its first periodic check in the future so the
	// explicit manual request owns this round. Startup itself still runs normally.
	h.e.schedules[h.m.ID] = schedule{version: h.m.ConfigVersion, enabled: true, next: h.clock.Load() + int64(h.m.IntervalSeconds)*1000}
	h.e.PollInterval = time.Hour
	if err := h.e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.e.Stop)
}

func TestManualCheckCallerDeadlineDoesNotShortenProbeOrReleaseOverlap(t *testing.T) {
	began := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(began)
		}
		select {
		case <-release:
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(target.Close)
	m := activeHTTP()
	m.Retries = 0
	m.TimeoutSeconds = 5
	m.HTTP.URL = target.URL
	h := newHarness(t, m)
	runner := probe.NewRunner(nil)
	h.e.Attempt = runner.Run
	startManagedEngine(t, h)
	caller, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- h.e.Check(caller, h.m.ID) }()
	select {
	case <-began:
	case <-caller.Done():
		t.Fatal("caller expired before the manual round was accepted")
	}
	if err := <-result; !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, ErrCheckAccepted) {
		t.Fatalf("caller deadline result=%v", err)
	}
	if runtime := h.state(t); runtime.State != domain.StateUnknown {
		t.Fatalf("short caller deadline changed the target state: %+v", runtime)
	}
	if rows, err := h.s.ListRounds(context.Background(), h.m.ID, 0, 10); err != nil || len(rows) != 0 {
		t.Fatalf("caller deadline persisted a synthetic result: %+v %v", rows, err)
	}
	if err := h.e.Check(context.Background(), h.m.ID); !errors.Is(err, ErrBusy) {
		t.Fatalf("caller departure released overlap ownership: %v", err)
	}
	close(release)
	deadline := time.Now().Add(2 * time.Second)
	for h.state(t).State != domain.StateUp && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if h.state(t).State != domain.StateUp {
		t.Fatal("accepted round did not record the real successful response")
	}
	rows, err := h.s.ListRounds(context.Background(), h.m.ID, 0, 10)
	if err != nil || len(rows) != 1 || !rows[0].Success || calls.Load() != 1 {
		t.Fatalf("accepted round results=%+v calls=%d error=%v", rows, calls.Load(), err)
	}
}

func TestStopCancelsAcceptedManualRoundAndWaitsForProbe(t *testing.T) {
	h := newHarness(t, activeHTTP())
	began := make(chan struct{})
	finished := make(chan struct{})
	h.e.Attempt = func(ctx context.Context, _ domain.Monitor) probe.Result {
		close(began)
		<-ctx.Done()
		close(finished)
		return probe.Result{Error: "cancelled"}
	}
	startManagedEngine(t, h)
	result := make(chan error, 1)
	go func() { result <- h.e.Check(context.Background(), h.m.ID) }()
	select {
	case <-began:
	case <-time.After(2 * time.Second):
		t.Fatal("manual check did not begin")
	}
	h.e.Stop()
	select {
	case <-finished:
	default:
		t.Fatal("Stop returned while the accepted probe was still running")
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("stopped check result=%v", err)
	}
	if h.state(t).State != domain.StateUnknown {
		t.Fatal("shutdown cancellation was recorded as target Down")
	}
	rows, err := h.s.ListRounds(context.Background(), h.m.ID, 0, 10)
	if err != nil || len(rows) != 0 {
		t.Fatalf("shutdown recorded an incomplete round: %+v %v", rows, err)
	}
	if err := h.e.Check(context.Background(), h.m.ID); !errors.Is(err, context.Canceled) || errors.Is(err, ErrCheckAccepted) {
		t.Fatalf("stopped engine accepted a new check: %v", err)
	}
}

func TestHeartbeatDescriptionUsesCharacterLimitAndPersistsLastReport(t *testing.T) {
	h := newHarness(t, domain.Monitor{Type: domain.MonitorHeartbeat, Heartbeat: &domain.HeartbeatConfig{PeriodSeconds: 30, GraceSeconds: 10}})
	description := strings.Repeat("监", 999) + "🙂"
	if err := h.e.Heartbeat(context.Background(), h.m.ID, true, description); err != nil {
		t.Fatal(err)
	}
	var meta Metadata
	if err := h.s.Get(context.Background(), "engineMonitor", h.m.ID, &meta); err != nil {
		t.Fatal(err)
	}
	if !meta.HasHeartbeatReport || !meta.HeartbeatSuccess || meta.HeartbeatDescription != description {
		t.Fatalf("last report was not persisted: %+v", meta)
	}
	if err := h.e.Heartbeat(context.Background(), h.m.ID, false, description+"额"); err == nil {
		t.Fatal("accepted more than 1000 characters")
	}
	if h.state(t).State != domain.StateUp {
		t.Fatal("invalid description changed the report state")
	}
	h.clock.Add(40001)
	h.check(t)
	if err := h.s.Get(context.Background(), "engineMonitor", h.m.ID, &meta); err != nil || meta.HeartbeatDescription != description || !meta.HeartbeatSuccess {
		t.Fatal("expiry replaced the last actual report description/outcome", meta, err)
	}
	if err := h.e.Heartbeat(context.Background(), h.m.ID, false, "  服务故障  "); err != nil {
		t.Fatal(err)
	}
	if err := h.s.Get(context.Background(), "engineMonitor", h.m.ID, &meta); err != nil || meta.HeartbeatSuccess || meta.HeartbeatDescription != "服务故障" {
		t.Fatal("explicit failed report did not replace the last report", meta, err)
	}
}

func TestNextCheckAtReturnsOnlyAnActivePlan(t *testing.T) {
	h := newHarness(t, activeHTTP())
	if h.e.NextCheckAt(h.m.ID) != 0 {
		t.Fatal("unplanned monitor has a next check")
	}
	at := h.clock.Load() + 30000
	h.e.schedules[h.m.ID] = schedule{enabled: true, next: at}
	if h.e.NextCheckAt(h.m.ID) != at {
		t.Fatal("active plan timestamp was not returned")
	}
	h.e.schedules[h.m.ID] = schedule{enabled: false, next: at}
	if h.e.NextCheckAt(h.m.ID) != 0 {
		t.Fatal("paused monitor still has a next check")
	}
	h.e.schedules[h.m.ID] = schedule{enabled: true, next: at}
	ctx, cancel := context.WithCancel(context.Background())
	h.e.ctx = ctx
	cancel()
	if h.e.NextCheckAt(h.m.ID) != 0 {
		t.Fatal("stopped engine returned a next check")
	}
}

func TestCommittedStartupRoundClearsItsPendingEvaluationAfterPollCancelsOldGeneration(t *testing.T) {
	h := newHarness(t, activeHTTP())
	h.e.ctx = context.Background()
	var calls atomic.Int32
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		calls.Add(1)
		return probe.Result{Success: true}
	}
	record, err := h.s.GetMonitor(context.Background(), h.m.ID)
	if err != nil {
		t.Fatal(err)
	}
	previous := h.state(t)
	roundContext, cancelRound := context.WithCancel(context.Background())
	defer cancelRound()
	h.e.evaluationEpoch = 1
	h.e.schedules[h.m.ID] = schedule{version: record.ConfigVersion, enabled: true, next: h.clock.Load() + 30000, pendingEvaluation: true, evaluationEpoch: 1}
	h.e.running[h.m.ID] = inFlight{cancel: cancelRound, version: record.ConfigVersion, generation: previous.Generation}
	round := store.Round{ID: domain.ID(), MonitorID: h.m.ID, ConfigVersion: record.ConfigVersion, Generation: previous.Generation, StartedAt: h.clock.Load(), FinishedAt: h.clock.Load(), Success: true}
	if err := h.e.commit(roundContext, record, h.m, previous, round, probe.Result{Success: true}, false); err != nil {
		t.Fatal(err)
	}
	// Force the precise interleaving from the capacity failure: commit advanced
	// Generation, then poll cancelled the still-present old inFlight generation.
	if err := h.e.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.e.wg.Wait()
	if !errors.Is(roundContext.Err(), context.Canceled) {
		t.Fatal("fixture did not cancel the already committed round")
	}
	h.e.clearPendingEvaluation(roundContext, h.m.ID, record.ConfigVersion, 1)
	h.e.mu.Lock()
	pending := h.e.schedules[h.m.ID].pendingEvaluation
	delete(h.e.running, h.m.ID)
	h.e.mu.Unlock()
	if pending {
		t.Fatal("own-generation cancellation left the startup evaluation pending")
	}
	if err := h.e.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.e.wg.Wait()
	rows, err := h.s.ListRounds(context.Background(), h.m.ID, 0, 10)
	if err != nil || len(rows) != 1 || calls.Load() != 0 {
		t.Fatalf("startup completion triggered an extra early round: rows=%+v calls=%d err=%v", rows, calls.Load(), err)
	}
}

func TestOldRoundCannotClearLaterMaintenanceOrConfigurationEvaluation(t *testing.T) {
	h := newHarness(t, activeHTTP())
	h.e.ctx = context.Background()
	// A maintenance-exit request is already present in the in-memory plan,
	// while its metadata write has not happened yet. EvaluationAfter is still
	// zero, so only ownership of the epoch can prevent the old round clearing it.
	h.e.schedules[h.m.ID] = schedule{version: h.m.ConfigVersion, enabled: true, next: h.clock.Load() + 30000, pendingEvaluation: true, evaluationEpoch: 2}
	h.e.clearPendingEvaluation(context.Background(), h.m.ID, h.m.ConfigVersion, 1)
	if !h.e.schedules[h.m.ID].pendingEvaluation {
		t.Fatal("old round erased a newer maintenance-exit evaluation")
	}
	h.e.clearPendingEvaluation(context.Background(), h.m.ID, h.m.ConfigVersion, 2)
	if h.e.schedules[h.m.ID].pendingEvaluation {
		t.Fatal("owning round could not clear its completed evaluation")
	}
	h.m.ConfigVersion++
	h.save(t, h.m)
	if err := h.e.NotifyConfigurationChanged(context.Background(), h.m.ID); err != nil {
		t.Fatal(err)
	}
	h.e.schedules[h.m.ID] = schedule{version: h.m.ConfigVersion, enabled: true, next: h.clock.Load() + 30000, pendingEvaluation: true, evaluationEpoch: 3}
	h.e.clearPendingEvaluation(context.Background(), h.m.ID, h.m.ConfigVersion-1, 3)
	if !h.e.schedules[h.m.ID].pendingEvaluation {
		t.Fatal("old configuration erased the new configuration's first evaluation")
	}
}
