package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/store"
)

func TestMaintenanceSuppressesAndRecoversOnlyAfterValidEvaluation(t *testing.T) {
	h := newHarness(t, activeHTTP())
	success := false
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: success} }
	h.check(t)
	if len(h.events(t, "down")) != 1 {
		t.Fatal("fault missing")
	}
	maintenance := domain.Maintenance{
		ID:         "maintenance",
		MonitorIDs: []string{h.m.ID},
		StartsAt:   h.clock.Load(),
		EndsAt:     h.clock.Load() + 60000,
	}
	if err := h.s.Put(
		context.Background(),
		"maintenance",
		maintenance.ID,
		maintenance,
	); err != nil {
		t.Fatal(err)
	}
	h.clock.Add(1000)
	success = true
	h.check(t)
	if h.state(t).State != domain.StateUp || len(h.events(t, "up")) != 0 {
		t.Fatal("maintenance must collect but suppress recovery")
	}
	h.clock.Store(maintenance.EndsAt)
	h.check(t)
	if len(h.events(t, "up")) != 1 {
		t.Fatal("recovery not supplemented after maintenance")
	}
}

func TestMaintenanceExitFreshDownAndPreExitAttemptDoesNotAlert(t *testing.T) {
	h := newHarness(t, activeHTTP())
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: true} }
	h.check(t)
	maintenance := domain.Maintenance{
		ID:         "maintenance",
		MonitorIDs: []string{h.m.ID},
		StartsAt:   h.clock.Load(),
		EndsAt:     h.clock.Load() + 60000,
	}
	if err := h.s.Put(
		context.Background(),
		"maintenance",
		maintenance.ID,
		maintenance,
	); err != nil {
		t.Fatal(err)
	}
	if err := h.e.markMaintenance(context.Background(), h.m.ID); err != nil {
		t.Fatal(err)
	}
	h.clock.Add(1000)
	started := make(chan struct{})
	release := make(chan struct{})
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		select {
		case <-started:
		default:
			close(started)
		}
		<-release
		return probe.Result{Success: false}
	}
	done := make(chan error, 1)
	go func() { done <- h.e.Check(context.Background(), h.m.ID) }()
	<-started
	h.clock.Store(maintenance.EndsAt)
	if err := h.e.markMaintenanceExit(context.Background(), h.m.ID, h.clock.Load()); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if h.state(t).State != domain.StateDown || len(h.events(t, "down")) != 0 {
		t.Fatal("pre-exit probe generated premature notification")
	}
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: false} }
	h.check(t)
	if len(h.events(t, "down")) != 1 {
		t.Fatal("fresh post-maintenance fault was not sent")
	}
	h.check(t)
	if len(h.events(t, "down")) != 1 {
		t.Fatal("fault repeated without a transition")
	}
}

func TestRemindersAreOptionalAndDoNotChangeConfirmedStateGeneration(t *testing.T) {
	m := activeHTTP()
	m.ReminderSeconds = 60
	h := newHarness(t, m)
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: false} }
	h.check(t)
	generation := h.state(t).Generation
	h.clock.Add(59000)
	h.check(t)
	if len(h.events(t, "reminder")) != 0 {
		t.Fatal("early reminder")
	}
	h.clock.Add(1000)
	h.check(t)
	if len(h.events(t, "reminder")) != 1 || h.state(t).Generation != generation {
		t.Fatal("reminder missing or invalidated current fault")
	}
}

func TestHeartbeatWaitingGraceReportsExplicitFailureAndExpiryCAS(t *testing.T) {
	h := newHarness(
		t,
		domain.Monitor{
			Type:            domain.MonitorHeartbeat,
			IntervalSeconds: 30,
			TimeoutSeconds:  1,
			Heartbeat: &domain.HeartbeatConfig{
				PeriodSeconds: 30,
				GraceSeconds:  10,
			},
		},
	)
	h.check(t)
	if h.state(t).State != domain.StateUnknown {
		t.Fatal("first heartbeat wait must be unknown")
	}
	h.clock.Add(39999)
	h.check(t)
	if h.state(t).State != domain.StateUnknown {
		t.Fatal("grace ignored")
	}
	h.clock.Add(1)
	h.check(t)
	if h.state(t).State != domain.StateDown || len(h.events(t, "down")) != 1 {
		t.Fatal("heartbeat deadline did not fault")
	}
	old := h.state(t)
	record, err := h.s.GetMonitor(context.Background(), h.m.ID)
	if err != nil {
		t.Fatal(err)
	}
	h.clock.Add(1)
	if err = h.e.Heartbeat(
		context.Background(),
		h.m.ID,
		true,
		"ready",
	); err != nil {
		t.Fatal(err)
	}
	if h.state(t).State != domain.StateUp || h.state(t).HeartbeatVersion <= old.HeartbeatVersion {
		t.Fatal("heartbeat did not recover/version report")
	}
	round := store.Round{
		ID:            domain.ID(),
		MonitorID:     h.m.ID,
		ConfigVersion: record.ConfigVersion,
		Generation:    record.Generation,
		StartedAt:     h.clock.Load(),
		FinishedAt:    h.clock.Load(),
		Success:       false,
	}
	err = h.e.commit(
		context.Background(),
		record,
		h.m,
		old,
		round,
		probe.Result{},
		false,
	)
	if !errors.Is(err, ErrSuperseded) || h.state(t).State != domain.StateUp {
		t.Fatal("stale expiry overwrote a newer heartbeat", err)
	}
	h.clock.Add(1)
	if err = h.e.Heartbeat(
		context.Background(),
		h.m.ID,
		false,
		"job failed",
	); err != nil {
		t.Fatal(err)
	}
	if h.state(t).State != domain.StateDown {
		t.Fatal("explicit failure ignored")
	}
	if err = h.e.initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	if h.state(t).State != domain.StateUnknown {
		t.Fatal("restart should reset current collection state")
	}
	h.check(t)
	if h.state(t).State != domain.StateDown {
		t.Fatal("last explicit failed report became implicit success after restart")
	}
	if len(h.events(t, "down")) != 2 {
		t.Fatal("restart repeated the same fault notification")
	}
}

func TestHeartbeatMaintenanceExitForUnknownAndFailedReports(t *testing.T) {
	h := newHarness(
		t,
		domain.Monitor{
			Type:            domain.MonitorHeartbeat,
			IntervalSeconds: 30,
			TimeoutSeconds:  1,
			Heartbeat: &domain.HeartbeatConfig{
				PeriodSeconds: 30,
				GraceSeconds:  10,
			},
		},
	)
	now := h.clock.Load()
	maintenance := domain.Maintenance{ID: "maintenance", MonitorIDs: []string{h.m.ID}, StartsAt: now, EndsAt: now + 1000}
	_ = h.s.Put(
		context.Background(),
		"maintenance",
		maintenance.ID,
		maintenance,
	)
	_ = h.e.markMaintenance(context.Background(), h.m.ID)
	h.clock.Add(1000)
	h.check(t)
	if len(h.events(t, "down")) != 0 || len(h.events(t, "up")) != 0 {
		t.Fatal("Unknown caused a deferred notification")
	}
	var meta Metadata
	_ = h.s.Get(
		context.Background(),
		"engineMonitor",
		h.m.ID,
		&meta,
	)
	if meta.MaintenanceActive || meta.EvaluationAfter != 0 {
		t.Fatal("unknown evaluation failed to finish maintenance exit")
	}
	maintenance.StartsAt = h.clock.Load()
	maintenance.EndsAt = h.clock.Load() + 1000
	_ = h.s.Put(
		context.Background(),
		"maintenance",
		maintenance.ID,
		maintenance,
	)
	if err := h.e.Heartbeat(
		context.Background(),
		h.m.ID,
		false,
		"failed during maintenance",
	); err != nil {
		t.Fatal(err)
	}
	if len(h.events(t, "down")) != 0 {
		t.Fatal("maintenance report notified")
	}
	h.clock.Add(1000)
	h.check(t)
	if len(h.events(t, "down")) != 1 {
		t.Fatal("explicit heartbeat fault did not notify after fresh deadline evaluation")
	}
}

func TestCertificateThresholdRenewalAndFailureIndependentOfAvailabilityMaintenance(t *testing.T) {
	h := newHarness(
		t,
		domain.Monitor{
			Type:            domain.MonitorCertificate,
			IntervalSeconds: 86400,
			TimeoutSeconds:  1,
			Certificate: &domain.CertificateConfig{
				Host:          "certificate.test",
				Port:          443,
				NotifyRenewal: true,
			},
		},
	)
	maintenance := domain.Maintenance{
		ID:         "maintenance",
		MonitorIDs: []string{h.m.ID},
		StartsAt:   h.clock.Load(),
		EndsAt:     h.clock.Load() + 1000000,
	}
	_ = h.s.Put(
		context.Background(),
		"maintenance",
		maintenance.ID,
		maintenance,
	)
	current := &probe.CertificateResult{
		State:         domain.CertificateHealthy,
		Fingerprint:   "first",
		ExpiresAt:     h.clock.Load() + 90*86400000,
		DaysRemaining: 90,
	}
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result {
		return probe.Result{Success: current.State != domain.CertificateCheckFailed, Certificate: current}
	}
	h.check(t)
	if len(h.events(t, "certificate_threshold")) != 0 {
		t.Fatal("healthy certificate notified")
	}
	current = &probe.CertificateResult{State: domain.CertificateExpiring, Fingerprint: "first", DaysRemaining: 21}
	h.clock.Add(1)
	h.check(t)
	h.check(t)
	if len(h.events(t, "certificate_threshold")) != 1 {
		t.Fatal("threshold absent or repeated")
	}
	current.DaysRemaining = 13
	h.clock.Add(1)
	h.check(t)
	if len(h.events(t, "certificate_threshold")) != 2 {
		t.Fatal("lower threshold missing")
	}
	current = &probe.CertificateResult{State: domain.CertificateExpired, Fingerprint: "first", DaysRemaining: -1}
	h.clock.Add(1)
	h.check(t)
	if len(h.events(t, "certificate_expired")) != 1 {
		t.Fatal("expiry missing")
	}
	current = &probe.CertificateResult{State: domain.CertificateCheckFailed}
	h.clock.Add(1)
	h.check(t)
	h.check(t)
	if len(h.events(t, "certificate_check_failed")) != 1 {
		t.Fatal("check failure absent or repeated")
	}
	current = &probe.CertificateResult{
		State:         domain.CertificateHealthy,
		Fingerprint:   "renewed",
		ExpiresAt:     h.clock.Load() + 90*86400000,
		DaysRemaining: 90,
	}
	h.clock.Add(1)
	h.check(t)
	if len(h.events(t, "certificate_renewed")) != 1 {
		t.Fatal("renewal missing")
	}
	intervals, err := h.s.Intervals(
		context.Background(),
		h.m.ID,
		h.m.CreatedAt,
		h.clock.Load()+1,
	)
	if err != nil || len(intervals) != 0 {
		t.Fatal("cert risk entered availability intervals", intervals, err)
	}
	var meta Metadata
	_ = h.s.Get(
		context.Background(),
		"engineMonitor",
		h.m.ID,
		&meta,
	)
	if meta.Certificate.Fingerprint != "renewed" {
		t.Fatal("certificate runtime snapshot missing")
	}
}

func TestNotificationCommitRollsBackRoundAndStateOnInvalidChannelDocument(t *testing.T) {
	h := newHarness(t, activeHTTP())
	if err := h.s.Put(
		context.Background(),
		"channels",
		"one",
		map[string]any{"enabled": "invalid boolean"},
	); err != nil {
		t.Fatal(err)
	}
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: false} }
	if err := h.e.Check(context.Background(), h.m.ID); err == nil {
		t.Fatal("invalid notification channel transaction succeeded")
	}
	if h.state(t).State != domain.StateUnknown {
		t.Fatal("state committed without outbox")
	}
	rounds, _ := h.s.ListRounds(
		context.Background(),
		h.m.ID,
		0,
		100,
	)
	deliveries, _ := h.s.ListDeliveries(context.Background(), 100)
	if len(rounds) != 0 || len(deliveries) != 0 {
		t.Fatal("part of failed transaction committed")
	}
}

func TestSchedulerChecksHeartbeatDeadlineWithoutWaitingForInterval(t *testing.T) {
	h := newHarness(
		t,
		domain.Monitor{
			Type:            domain.MonitorHeartbeat,
			IntervalSeconds: 3600,
			TimeoutSeconds:  1,
			Heartbeat: &domain.HeartbeatConfig{
				PeriodSeconds: 30,
			},
		},
	)
	if err := h.e.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.e.wg.Wait()
	h.clock.Add(30000)
	if err := h.e.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.e.wg.Wait()
	if h.state(t).State != domain.StateDown {
		t.Fatal("heartbeat expiry incorrectly waited for generic check interval")
	}
}

func TestLiveCollectionGapDoesNotCarryUpOrConfirmAnUnobservedFault(t *testing.T) {
	m := activeHTTP()
	m.FailureThreshold = 2
	h := newHarness(t, m)
	success := true
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: success} }
	h.check(t)
	boundary := h.clock.Load() + 60000
	h.clock.Add(60001)
	success = false
	if err := h.e.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.e.wg.Wait()
	if h.state(t).State != domain.StateUnknown || h.state(t).Failures != 1 {
		t.Fatal("collection gap carried an old state or inherited old counters", h.state(t))
	}
	if len(h.events(t, "collection_gap")) != 1 || len(h.events(t, "down")) != 0 {
		t.Fatal("missing collection gap or unconfirmed Down notified")
	}
	intervals, err := h.s.Intervals(
		context.Background(),
		h.m.ID,
		h.m.CreatedAt,
		h.clock.Load()+1,
	)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, interval := range intervals {
		if interval.State == domain.StateUp {
			if interval.EndedAt == nil || *interval.EndedAt != boundary {
				t.Fatal("Up extended past the latest expected round deadline")
			}
		}
		if interval.State == domain.StateUnknown && interval.StartedAt == boundary {
			found = true
		}
	}
	if !found {
		t.Fatal("explicit Unknown gap missing")
	}
}

func TestCollectionWatermarkCannotMoveBackwardsWhenPollAndRoundCommitRace(t *testing.T) {
	h := newHarness(t, activeHTTP())
	h.e.Attempt = func(context.Context, domain.Monitor) probe.Result { return probe.Result{Success: true} }
	h.check(t)
	newer := h.clock.Add(500)
	h.check(t)
	if err := h.s.WithTx(
		context.Background(),
		func(tx *store.Tx) error {
			return putCollectionWatermark(context.Background(), tx, newer-500)
		},
	); err != nil {
		t.Fatal(err)
	}
	at, err := h.s.Watermark(context.Background(), collectionWatermark)
	if err != nil || at != newer {
		t.Fatalf("watermark regressed: %d err=%v", at, err)
	}
}
