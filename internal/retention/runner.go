// Package retention bounds operational data maintenance independently from the
// monitor statistics retention policy. Historical cleanup is explicitly opt-in.
package retention

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
	"github.com/octoplorer/octopulse/internal/telemetry"
)

type Runner struct {
	Store                *store.Store
	OperationHistoryDays int
	BatchSize            int
	Interval             time.Duration
	Now                  func() time.Time
	OnError              func(error)

	mu      sync.Mutex
	cursors map[string]string
}

func New(s *store.Store) *Runner {
	return &Runner{Store: s, BatchSize: 100, Interval: time.Minute, Now: time.Now, cursors: map[string]string{}}
}

// Start blocks until cancellation. Each interval advances at most one page per
// document kind and one batch of deliveries/events, keeping transactions small.
func (r *Runner) Start(ctx context.Context) {
	interval := r.Interval
	if interval <= 0 {
		interval = time.Minute
	}
	for {
		if err := r.RunOnce(ctx); err != nil && ctx.Err() == nil {
			if r.OnError != nil {
				r.OnError(err)
			} else {
				telemetry.LogError(ctx, "retention.cleanup", err)
			}
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	invalidBatchSize := r.BatchSize <= 0 || r.BatchSize > 1000
	invalidHistoryDays := r.OperationHistoryDays < 0 || r.OperationHistoryDays > 3650
	if invalidBatchSize || invalidHistoryDays {
		return fmt.Errorf("invalid retention batch size or operation history days")
	}
	now := r.Now().UnixMilli()
	failures := []error{}
	if err := r.sweep(ctx, "sessions", func(raw json.RawMessage) (bool, error) {
		var session domain.Session
		if err := json.Unmarshal(raw, &session); err != nil {
			return false, err
		}
		return session.ExpiresAt > 0 && session.ExpiresAt <= now, nil
	}); err != nil {
		failures = append(failures, err)
	}
	if r.OperationHistoryDays == 0 {
		return errors.Join(failures...)
	}
	before := now - int64(r.OperationHistoryDays)*24*time.Hour.Milliseconds()
	if err := r.sweep(ctx, "audit", func(raw json.RawMessage) (bool, error) {
		var audit domain.Audit
		if err := json.Unmarshal(raw, &audit); err != nil {
			return false, err
		}
		return audit.CreatedAt > 0 && audit.CreatedAt < before, nil
	}); err != nil {
		failures = append(failures, err)
	}
	if err := r.sweep(ctx, "channelTests", func(raw json.RawMessage) (bool, error) {
		var test struct {
			Status    string `json:"status"`
			CreatedAt int64  `json:"createdAt"`
		}
		if err := json.Unmarshal(raw, &test); err != nil {
			return false, err
		}
		return (test.Status == "sent" || test.Status == "failed") && test.CreatedAt > 0 && test.CreatedAt < before, nil
	}); err != nil {
		failures = append(failures, err)
	}
	if before > 0 {
		if err := r.Store.PruneOperationHistory(ctx, before, r.BatchSize); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (r *Runner) sweep(ctx context.Context, kind string, expired func(json.RawMessage) (bool, error)) error {
	documents, err := r.Store.DocumentPage(
		ctx,
		kind,
		r.cursors[kind],
		r.BatchSize,
	)
	if err != nil {
		return err
	}
	candidates := make([]store.Document, 0, len(documents))
	failures := []error{}
	for _, document := range documents {
		remove, err := expired(document.Payload)
		if err != nil {
			// Preserve malformed records but continue the cursor so one record
			// cannot prevent all subsequent valid records from being cleaned up.
			failures = append(failures, fmt.Errorf(
				"decode retention document %s/%s: %w",
				kind,
				document.ID,
				err,
			))
		} else if remove {
			candidates = append(candidates, document)
		}
	}
	if err := r.Store.DeleteDocumentsIfUnchanged(ctx, kind, candidates); err != nil {
		return errors.Join(append(failures, err)...)
	}
	if len(documents) < r.BatchSize {
		r.cursors[kind] = ""
	} else {
		r.cursors[kind] = documents[len(documents)-1].ID
	}
	return errors.Join(failures...)
}
