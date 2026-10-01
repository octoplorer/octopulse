package engine

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/store"
)

func (e *Engine) Heartbeat(ctx context.Context, id string, success bool, description string) error {
	if len(description) > 1024 {
		return errors.New("heartbeat description exceeds 1024 bytes")
	}
	description = strings.TrimSpace(description)
	// Reports do not acquire the active-check mutex: an arrival can race an
	// expiry evaluation. CompareRuntime's heartbeat-version CAS decides which
	// observation is still current; the report retries the fresh runtime.
	for attempt := 0; attempt < 4; attempt++ {
		record, err := e.Store.GetMonitor(ctx, id)
		if err != nil {
			return err
		}
		m, err := decodeMonitor(record)
		if err != nil {
			return err
		}
		if m.Type != domain.MonitorHeartbeat {
			return errors.New("monitor does not accept heartbeat reports")
		}
		if !record.Enabled {
			return ErrPaused
		}
		runtime, err := e.ensureRuntime(ctx, record, m)
		if errors.Is(err, ErrSuperseded) {
			continue
		}
		if err != nil {
			return err
		}
		record.Generation = runtime.Generation
		now := e.now()
		detail, _ := json.Marshal(map[string]string{"description": description})
		round := store.Round{ID: domain.ID(), MonitorID: id, ConfigVersion: record.ConfigVersion, Generation: record.Generation, StartedAt: now, FinishedAt: now, Success: success, Attempts: []store.Attempt{{Number: 1, StartedAt: now, FinishedAt: now, Success: success, Detail: detail}}}
		err = e.commit(ctx, record, m, runtime, round, probe.Result{Success: success}, true)
		if errors.Is(err, ErrSuperseded) {
			continue
		}
		return err
	}
	return ErrSuperseded
}

func (e *Engine) evaluateHeartbeat(ctx context.Context, record store.Monitor, m domain.Monitor, runtime store.Runtime) error {
	now := e.now()
	meta := Metadata{}
	err := e.Store.Get(ctx, "engineMonitor", m.ID, &meta)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}
	base := runtime.HeartbeatAt
	if base == 0 {
		base = m.CreatedAt
		if base == 0 {
			base = meta.InitializedAt
		}
		if base == 0 {
			base = now
		}
	}
	deadline := base + int64(m.Heartbeat.PeriodSeconds+m.Heartbeat.GraceSeconds)*1000
	rows, err := e.Store.List(ctx, "maintenance")
	if err != nil {
		return err
	}
	exited := meta.MaintenanceActive && !inMaintenance(rows, m.ID, now)
	if now < deadline {
		// No report has ever arrived: preserve Unknown even during a manual
		// check, and never treat waiting as an implicit successful heartbeat.
		if runtime.HeartbeatAt == 0 {
			if exited {
				if err = e.finishUnknownMaintenanceEvaluation(ctx, m.ID); err != nil {
					return err
				}
				e.clearPendingEvaluation(ctx, m.ID)
			}
			return nil
		}
		if runtime.State == domain.StateDown && !exited {
			return nil
		}
		// After a restart, a valid last report is evidence for the new current
		// state; its earlier period remains the explicit collection gap.
		if runtime.State != domain.StateUnknown && !exited {
			return nil
		}
	}
	if now >= deadline && runtime.State == domain.StateDown && !exited && (m.ReminderSeconds == 0 || now-meta.LastReminderAt < int64(m.ReminderSeconds)*1000) {
		return nil
	}
	success := now < deadline && runtime.HeartbeatAt > 0
	if meta.HasHeartbeatReport && !meta.HeartbeatSuccess {
		success = false
	}
	round := store.Round{ID: domain.ID(), MonitorID: m.ID, ConfigVersion: record.ConfigVersion, Generation: record.Generation, StartedAt: now, FinishedAt: now, Success: success, Attempts: []store.Attempt{{Number: 1, StartedAt: now, FinishedAt: now, Success: success}}}
	err = e.commit(ctx, record, m, runtime, round, probe.Result{Success: success}, false)
	if err == nil {
		e.clearPendingEvaluation(ctx, m.ID)
	}
	return err
}

func (e *Engine) finishUnknownMaintenanceEvaluation(ctx context.Context, id string) error {
	return e.Store.WithTx(ctx, func(tx *store.Tx) error {
		var meta Metadata
		if err := tx.Get(ctx, "engineMonitor", id, &meta); err != nil {
			return err
		}
		meta.MaintenanceActive = false
		meta.EvaluationAfter = 0
		return tx.Put(ctx, "engineMonitor", id, meta)
	})
}
