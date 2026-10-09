package engine

import (
	"context"
	"encoding/json"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

// A live scheduler cannot extend a known state forever after persistence or
// scheduling failures. The next cadence plus one full round is the latest
// expected result deadline. Past it, record an explicit collection gap.
func (e *Engine) expireCollection(
	ctx context.Context,
	record store.Monitor,
	m domain.Monitor,
	runtime store.Runtime,
	now int64,
) (store.Runtime, error) {
	passiveMonitor := m.Type == domain.MonitorHeartbeat || m.Type == domain.MonitorCertificate
	knownState := runtime.State == domain.StateUp || runtime.State == domain.StateDown
	hasCollection := runtime.LastCollectedAt != 0 && knownState
	if !m.Enabled || passiveMonitor || !hasCollection {
		return runtime, nil
	}
	interval := int64(m.IntervalSeconds) * 1000
	boundary := runtime.LastCollectedAt + 2*interval
	if now <= boundary {
		return runtime, nil
	}
	var next store.Runtime
	err := e.Store.WithTx(ctx, func(tx *store.Tx) error {
		current, err := tx.GetMonitor(ctx, m.ID)
		if err != nil {
			return err
		}
		latest, err := tx.GetRuntime(ctx, m.ID)
		if err != nil {
			return err
		}
		configurationChanged := current.ConfigVersion != record.ConfigVersion
		runtimeChanged := latest.LastCollectedAt != runtime.LastCollectedAt || latest.Generation != runtime.Generation
		if !current.Enabled || configurationChanged || runtimeChanged {
			next = latest
			return nil
		}
		current.Generation++
		next = latest
		next.Generation = current.Generation
		next.State = domain.StateUnknown
		next.Failures = 0
		next.Successes = 0
		if err = tx.PutMonitor(ctx, current); err != nil {
			return err
		}
		if err = tx.CompareRuntime(ctx, latest, next); err != nil {
			return err
		}
		if err = tx.ReplaceInterval(
			ctx,
			store.Interval{
				ID:        domain.ID(),
				MonitorID: m.ID,
				State:     domain.StateUnknown,
				StartedAt: boundary,
			},
		); err != nil {
			return err
		}
		payload, _ := json.Marshal(map[string]int64{"since": boundary, "observedAt": now})
		return tx.PutEvent(
			ctx,
			store.Event{
				ID:         domain.ID(),
				MonitorID:  m.ID,
				Generation: next.Generation,
				Kind:       "collection_gap",
				CreatedAt:  now,
				Payload:    payload,
			},
		)
	})
	return next, err
}
