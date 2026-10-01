package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/store"
)

func (e *Engine) commit(ctx context.Context, record store.Monitor, m domain.Monitor, previous store.Runtime, round store.Round, result probe.Result, heartbeatReport bool) error {
	return reconcileCommit(ctx, func() error { return e.commitOnce(ctx, record, m, previous, round, result, heartbeatReport) }, func() (bool, error) { return e.Store.RoundExists(ctx, round.ID) })
}

// Reconcile an uncertain commit by its stable round ID before retrying the
// same observation. Database retries never repeat the external probe.
func reconcileCommit(ctx context.Context, write func() error, exists func() (bool, error)) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = write()
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		committed, lookupErr := exists()
		if lookupErr == nil && committed {
			return nil
		}
		if errors.Is(err, ErrSuperseded) || errors.Is(err, store.ErrConflict) || errors.Is(err, store.ErrNotFound) {
			return err
		}
		if attempt < 2 {
			timer := time.NewTimer(time.Duration(attempt+1) * 25 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return err
}

func (e *Engine) commitOnce(ctx context.Context, record store.Monitor, m domain.Monitor, previous store.Runtime, round store.Round, result probe.Result, heartbeatReport bool) error {
	return e.Store.WithTx(ctx, func(tx *store.Tx) error {
		current, err := tx.GetMonitor(ctx, m.ID)
		if err != nil {
			return err
		}
		if !current.Enabled || current.ConfigVersion != record.ConfigVersion || current.Generation != record.Generation {
			return ErrSuperseded
		}
		meta, err := getMetadata(ctx, tx, m, current, round.FinishedAt)
		if err != nil {
			return err
		}
		maintenanceRows, err := tx.List(ctx, "maintenance")
		if err != nil {
			return err
		}
		maintenance := inMaintenance(maintenanceRows, m.ID, round.FinishedAt)
		if meta.MaintenanceActive && !maintenance && meta.EvaluationAfter == 0 {
			meta.EvaluationAfter = latestMaintenanceEnd(maintenanceRows, m.ID, round.FinishedAt)
		}
		// A probe begun within maintenance is not the required fresh evaluation
		// after it ends. Keep suppression until a later round starts after exit.
		if !maintenance && meta.EvaluationAfter > round.StartedAt {
			maintenance = true
		}
		exitedMaintenance := meta.MaintenanceActive && !maintenance
		next := previous
		next.LastRoundID = round.ID
		next.LastCollectedAt = round.FinishedAt
		if heartbeatReport {
			next.HeartbeatVersion++
			next.HeartbeatAt = round.FinishedAt
			meta.HasHeartbeatReport = true
			meta.HeartbeatSuccess = round.Success
		}
		if m.Type == domain.MonitorCertificate {
			if result.Certificate == nil {
				result.Certificate = &probe.CertificateResult{State: domain.CertificateCheckFailed}
			}
			next.State = result.Certificate.State
		} else if m.Type == domain.MonitorHeartbeat {
			if round.Success {
				next.State = domain.StateUp
				next.Failures = 0
				next.Successes = 1
			} else {
				next.State = domain.StateDown
				next.Failures = 1
				next.Successes = 0
			}
		} else {
			if round.Success {
				next.Failures = 0
				next.Successes++
				if next.State == domain.StateUp || next.Successes >= int64(m.RecoveryThreshold) {
					next.State = domain.StateUp
				}
			} else {
				next.Successes = 0
				next.Failures++
				if next.State == domain.StateDown || next.Failures >= int64(m.FailureThreshold) {
					next.State = domain.StateDown
				}
			}
		}
		changed := next.State != previous.State
		var payloads []NotificationPayload
		if m.Type == domain.MonitorCertificate {
			// Certificate risk notifications are independent of availability
			// maintenance, which only suppresses Down/Up/reminders.
			payloads = certificateNotifications(m, previous.State, &meta, result.Certificate, false, false, round.FinishedAt)
		} else {
			payloads = availabilityNotifications(m, previous.State, next.State, &meta, maintenance, exitedMaintenance, round.FinishedAt)
		}
		// A new risk/maintenance-exit event supersedes older queued faults even
		// when the confirmed state itself stays the same.
		if changed || len(payloads) > 0 && payloads[0].Kind != "reminder" {
			current.Generation++
			next.Generation = current.Generation
			if err = tx.PutMonitor(ctx, current); err != nil {
				return err
			}
		}
		if err = tx.CompareRuntime(ctx, previous, next); err != nil {
			if errors.Is(err, store.ErrConflict) {
				return ErrSuperseded
			}
			return err
		}
		if err = tx.PutRound(ctx, round); err != nil {
			return err
		}
		if changed && m.IsAvailability() {
			if err = tx.ReplaceInterval(ctx, store.Interval{ID: domain.ID(), MonitorID: m.ID, State: next.State, StartedAt: round.FinishedAt}); err != nil {
				return err
			}
		}
		if changed {
			payload, _ := json.Marshal(map[string]any{"from": previous.State, "to": next.State, "roundId": round.ID})
			if err = tx.PutEvent(ctx, store.Event{ID: domain.ID(), MonitorID: m.ID, Generation: next.Generation, Kind: "state_changed", CreatedAt: round.FinishedAt, Payload: payload}); err != nil {
				return err
			}
		}
		meta.MaintenanceActive = maintenance
		if exitedMaintenance {
			meta.EvaluationAfter = 0
		}
		if m.Type == domain.MonitorCertificate {
			meta.Certificate = result.Certificate
		}
		if err = tx.Put(ctx, "engineMonitor", m.ID, meta); err != nil {
			return err
		}
		for _, payload := range payloads {
			payload.Generation = next.Generation
			payload.State = next.State
			if err = enqueue(ctx, tx, m, payload); err != nil {
				return err
			}
		}
		return putCollectionWatermark(ctx, tx, round.FinishedAt)
	})
}

func putCollectionWatermark(ctx context.Context, tx *store.Tx, at int64) error {
	previous, err := tx.Watermark(ctx, collectionWatermark)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}
	if previous > at {
		at = previous
	}
	return tx.PutWatermark(ctx, collectionWatermark, at)
}

func availabilityNotifications(m domain.Monitor, previous, state string, meta *Metadata, maintenance, exited bool, now int64) []NotificationPayload {
	makePayload := func(kind string) NotificationPayload {
		return NotificationPayload{MonitorID: m.ID, Name: m.Name, Kind: kind, State: state, CycleID: meta.CycleID, CreatedAt: now, Message: fmt.Sprintf("%s is %s", m.Name, state)}
	}
	if state == domain.StateDown {
		if meta.CycleID == "" {
			meta.CycleID = domain.ID()
			meta.FaultQueued = false
		}
		meta.PendingRecovery = false
		if maintenance {
			return nil
		}
		if previous == domain.StateUp || !meta.FaultQueued || exited {
			meta.FaultQueued = true
			meta.LastReminderAt = now
			return []NotificationPayload{makePayload("down")}
		}
		if m.ReminderSeconds > 0 && now-meta.LastReminderAt >= int64(m.ReminderSeconds)*1000 {
			meta.LastReminderAt = now
			payload := makePayload("reminder")
			payload.Message = fmt.Sprintf("%s remains down", m.Name)
			return []NotificationPayload{payload}
		}
		return nil
	}
	if state == domain.StateUp {
		if maintenance {
			if meta.CycleID != "" && meta.FaultQueued {
				meta.PendingRecovery = true
			}
			return nil
		}
		var result []NotificationPayload
		if meta.CycleID != "" && meta.FaultQueued && (previous == domain.StateDown || meta.PendingRecovery || exited || previous == domain.StateUnknown) && m.NotifyRecovery {
			result = []NotificationPayload{makePayload("up")}
		}
		meta.CycleID = ""
		meta.FaultQueued = false
		meta.PendingRecovery = false
		meta.LastReminderAt = 0
		return result
	}
	// Unknown does not generate a deferred fault or recovery. Keep the cycle so
	// a later valid observation can recover a fault sent before a collection gap.
	return nil
}

func certificateNotifications(m domain.Monitor, previous string, meta *Metadata, certificate *probe.CertificateResult, maintenance, exited bool, now int64) []NotificationPayload {
	if certificate == nil {
		return nil
	}
	makePayload := func(kind, message string) NotificationPayload {
		return NotificationPayload{MonitorID: m.ID, Name: m.Name, Kind: kind, State: certificate.State, CycleID: certificate.Fingerprint, CreatedAt: now, Message: message}
	}
	if certificate.State == domain.CertificateCheckFailed {
		if !maintenance && (previous != domain.CertificateCheckFailed || exited) {
			return []NotificationPayload{makePayload("certificate_check_failed", fmt.Sprintf("%s certificate check failed", m.Name))}
		}
		return nil
	}
	renewed := meta.CertificateFingerprint != "" && meta.CertificateFingerprint != certificate.Fingerprint
	if meta.CertificateFingerprint != certificate.Fingerprint {
		meta.CertificateFingerprint = certificate.Fingerprint
		meta.CertificateWarnings = nil
	}
	var payloads []NotificationPayload
	if renewed && m.Certificate.NotifyRenewal && !maintenance {
		payloads = append(payloads, makePayload("certificate_renewed", fmt.Sprintf("%s certificate was renewed; %.1f days remain", m.Name, certificate.DaysRemaining)))
	}
	if maintenance {
		return payloads
	}
	threshold := -1
	if certificate.State == domain.CertificateExpired {
		threshold = 0
	} else {
		days := append([]int{}, m.Certificate.WarningDays...)
		sort.Ints(days)
		for _, day := range days {
			if certificate.DaysRemaining <= float64(day) {
				threshold = day
				break
			}
		}
	}
	if threshold >= 0 {
		already := false
		for _, day := range meta.CertificateWarnings {
			if day == threshold {
				already = true
			}
		}
		if !already || exited {
			meta.CertificateWarnings = append(meta.CertificateWarnings, threshold)
			kind := "certificate_threshold"
			message := fmt.Sprintf("%s certificate crossed the %d day threshold; %.1f days remain", m.Name, threshold, certificate.DaysRemaining)
			if threshold == 0 {
				kind = "certificate_expired"
				message = fmt.Sprintf("%s certificate has expired", m.Name)
			}
			payloads = append(payloads, makePayload(kind, message))
		}
	}
	return payloads
}

func enqueue(ctx context.Context, tx *store.Tx, m domain.Monitor, payload NotificationPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	event := store.Event{ID: domain.ID(), MonitorID: m.ID, Generation: payload.Generation, Kind: payload.Kind, CreatedAt: payload.CreatedAt, Payload: data}
	if err = tx.PutEvent(ctx, event); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, channelID := range m.NotificationChannelIDs {
		if channelID == "" || seen[channelID] {
			continue
		}
		seen[channelID] = true
		var channel domain.Channel
		if err = tx.Get(ctx, "channels", channelID, &channel); errors.Is(err, store.ErrNotFound) {
			continue
		} else if err != nil {
			return err
		}
		if !channel.Enabled {
			continue
		}
		if err = tx.PutDelivery(ctx, store.Delivery{ID: domain.ID(), EventID: event.ID, ChannelID: channelID, Generation: payload.Generation, State: "pending", DueAt: payload.CreatedAt, Payload: data}); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) markMaintenance(ctx context.Context, id string) error {
	return e.Store.WithTx(ctx, func(tx *store.Tx) error {
		record, err := tx.GetMonitor(ctx, id)
		if err != nil {
			return err
		}
		m, err := decodeMonitor(record)
		if err != nil {
			return err
		}
		meta, err := getMetadata(ctx, tx, m, record, e.now())
		if err != nil {
			return err
		}
		meta.MaintenanceActive = true
		return tx.Put(ctx, "engineMonitor", id, meta)
	})
}

func (e *Engine) markMaintenanceExit(ctx context.Context, id string, at int64) error {
	return e.Store.WithTx(ctx, func(tx *store.Tx) error {
		record, err := tx.GetMonitor(ctx, id)
		if err != nil {
			return err
		}
		m, err := decodeMonitor(record)
		if err != nil {
			return err
		}
		meta, err := getMetadata(ctx, tx, m, record, at)
		if err != nil {
			return err
		}
		meta.MaintenanceActive = true
		meta.EvaluationAfter = at
		return tx.Put(ctx, "engineMonitor", id, meta)
	})
}
func latestMaintenanceEnd(raw []json.RawMessage, id string, now int64) int64 {
	var latest int64
	for _, data := range raw {
		var m domain.Maintenance
		if json.Unmarshal(data, &m) != nil || m.EndsAt > now {
			continue
		}
		for _, monitorID := range m.MonitorIDs {
			if monitorID == id && m.EndsAt > latest {
				latest = m.EndsAt
			}
		}
	}
	return latest
}
