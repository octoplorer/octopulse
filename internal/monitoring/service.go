// Package monitoring owns monitor configuration and runtime projections.
package monitoring

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type Service struct {
	Store     *store.Store
	NextCheck func(string) int64
	Changed   func(context.Context, string) error
}
type Write struct {
	Monitor        domain.Monitor
	Enabled        *bool
	Retries        *int
	NotifyRecovery *bool
}
type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }
func invalid(message string) error       { return &ValidationError{Message: message} }

func (s *Service) Save(ctx context.Context, id string, in Write, actor domain.User) (domain.Monitor, error) {
	m := in.Monitor
	var previous domain.Monitor
	if id != "" {
		record, e := s.Store.GetMonitor(ctx, id)
		if e != nil {
			return domain.Monitor{}, e
		}
		if e := json.Unmarshal(record.ConfigJSON, &previous); e != nil {
			return domain.Monitor{}, e
		}
		m.ID = id
		if m.Type != previous.Type {
			return domain.Monitor{}, invalid("Monitor type is fixed after creation; create a new monitor to change it")
		}
		m.CreatedAt = previous.CreatedAt
		m.ConfigVersion = previous.ConfigVersion + 1
	} else {
		m.ID = domain.ID()
		m.CreatedAt = domain.Now()
		m.ConfigVersion = 1
	}
	m.Enabled = previous.Enabled
	m.Retries = previous.Retries
	m.NotifyRecovery = previous.NotifyRecovery
	if id == "" {
		m.Enabled = true
		m.Retries = 2
		m.NotifyRecovery = true
	}
	if in.Enabled != nil {
		m.Enabled = *in.Enabled
	}
	if in.Retries != nil {
		m.Retries = *in.Retries
	}
	if in.NotifyRecovery != nil {
		m.NotifyRecovery = *in.NotifyRecovery
	}
	m.State = domain.StateUnknown
	m.FailureCount = 0
	m.SuccessCount = 0
	m.LastCheckedAt = 0
	m.NextCheckAt = 0
	m.UpdatedAt = domain.Now()
	if m.Certificate != nil {
		m.Certificate.State = ""
		m.Certificate.ExpiresAt = 0
		m.Certificate.DaysRemaining = 0
		m.Certificate.Fingerprint = ""
	}
	if m.Heartbeat != nil {
		m.Heartbeat.LastReceivedAt = 0
		m.Heartbeat.LastSuccess = false
		m.Heartbeat.Description = ""
	}
	m.Defaults()
	if e := m.Validate(); e != nil {
		return domain.Monitor{}, invalid(e.Error())
	}
	e := s.Store.WithTx(ctx, func(t *store.Tx) error {
		generation := int64(1)
		if id != "" {
			current, e := t.GetMonitor(ctx, id)
			if e != nil {
				return e
			}
			if current.ConfigVersion != previous.ConfigVersion {
				return store.ErrConflict
			}
			generation = current.Generation
		}
		for _, ref := range m.SecretReferences() {
			var v domain.SecretRecord
			if e := t.Get(
				ctx,
				"secrets",
				ref,
				&v,
			); e != nil {
				if errors.Is(e, store.ErrNotFound) {
					return invalid("A referenced secret is unavailable")
				}
				return e
			}
		}
		for _, channelID := range m.NotificationChannelIDs {
			var v domain.Channel
			if e := t.Get(
				ctx,
				"channels",
				channelID,
				&v,
			); e != nil {
				if errors.Is(e, store.ErrNotFound) {
					return invalid("A notification channel is unavailable")
				}
				return e
			}
		}
		b, e := json.Marshal(m)
		if e != nil {
			return e
		}
		row := store.Monitor{
			ID:            m.ID,
			ConfigVersion: m.ConfigVersion,
			Generation:    generation,
			Kind:          m.Type,
			Enabled:       m.Enabled,
			IntervalMS:    int64(m.IntervalSeconds) * 1000,
			ConfigJSON:    b,
		}
		if e = t.PutMonitor(ctx, row); e != nil {
			return e
		}
		if e = t.Put(
			ctx,
			"monitors",
			m.ID,
			m,
		); e != nil {
			return e
		}
		a := domain.Audit{
			ID:           domain.ID(),
			UserID:       actor.ID,
			Username:     actor.Username,
			Action:       "save",
			ResourceType: "monitors",
			ResourceID:   m.ID,
			CreatedAt:    domain.Now(),
		}
		return t.Put(
			ctx,
			"audit",
			a.ID,
			a,
		)
	})
	if e != nil {
		return domain.Monitor{}, e
	}
	if s.Changed != nil {
		if e = s.Changed(ctx, m.ID); e != nil {
			return domain.Monitor{}, e
		}
	}
	return m, nil
}
