// Package notify owns durable per-channel Shoutrrr delivery. External sends
// occur outside transactions; successful sends can still be duplicated if a
// process fails before the database completion transaction commits.
package notify

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
	"github.com/octoplorer/octopulse/internal/telemetry"
)

type SecretResolver interface {
	ResolveSecret(context.Context, string) (string, error)
}

type Worker struct {
	Store       *store.Store
	Secrets     SecretResolver
	SendTimeout time.Duration
	MaxAttempts int64
	RetryBase   time.Duration
	Now         func() time.Time
	OnError     func(error)
	OnDelivery  func(string)
	slots       chan struct{}
}

func New(s *store.Store, secrets SecretResolver) *Worker {
	return &Worker{
		Store:       s,
		Secrets:     secrets,
		SendTimeout: 30 * time.Second,
		MaxAttempts: 8,
		RetryBase:   5 * time.Second,
		Now:         time.Now,
		slots:       make(chan struct{}, 4),
	}
}

// Start blocks until cancellation and uses four bounded delivery workers.
func (w *Worker) Start(ctx context.Context) {
	var workers sync.WaitGroup
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				processed, err := w.deliver(ctx)
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					if w.OnError != nil {
						w.OnError(err)
					} else {
						telemetry.LogError(ctx, "notification.persistence", err)
					}
				}
				if processed && err == nil {
					continue
				}
				timer := time.NewTimer(time.Second)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
		}()
	}
	workers.Wait()
}

func (w *Worker) DeliverOnce(ctx context.Context) error {
	_, err := w.deliver(ctx)
	return err
}

func (w *Worker) deliver(ctx context.Context) (bool, error) {
	select {
	case w.slots <- struct{}{}:
		defer func() { <-w.slots }()
	case <-ctx.Done():
		return false, ctx.Err()
	}
	ctx, cancel := context.WithTimeout(ctx, w.SendTimeout+15*time.Second)
	defer cancel()
	now := w.Now().UnixMilli()
	token := domain.ID()
	deliveries, err := w.Store.ClaimDeliveries(
		ctx,
		now,
		(w.SendTimeout + 60*time.Second).Milliseconds(),
		1,
		token,
	)
	if err != nil {
		return false, err
	}
	if len(deliveries) == 0 {
		return false, nil
	}
	d := deliveries[0]
	var payload domain.NotificationPayload
	if err = json.Unmarshal(d.Payload, &payload); err != nil {
		return true, w.finish(
			ctx,
			d,
			"failed",
			0,
			"invalid notification payload",
			nil,
		)
	}
	valid, err := w.relevant(ctx, d, payload)
	if err != nil {
		var waiting *awaitingDown
		if errors.As(err, &waiting) {
			return true, w.finish(
				ctx,
				d,
				"pending",
				waiting.until+1,
				waiting.Error(),
				nil,
			)
		}
		return true, err
	}
	if !valid {
		return true, w.finish(
			ctx,
			d,
			"cancelled",
			0,
			"notification is no longer applicable",
			nil,
		)
	}
	var channel domain.Channel
	if err = w.Store.Get(
		ctx,
		"channels",
		d.ChannelID,
		&channel,
	); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return true, w.finish(
				ctx,
				d,
				"cancelled",
				0,
				"notification channel was removed",
				nil,
			)
		}
		return true, err
	}
	if !channel.Enabled {
		return true, w.finish(
			ctx,
			d,
			"cancelled",
			0,
			"notification channel is disabled",
			nil,
		)
	}
	url, err := w.Secrets.ResolveSecret(ctx, channel.ServiceURLSecretID)
	if err != nil {
		return true, w.retry(ctx, d, "notification channel secret is unavailable")
	}
	sendCtx, sendCancel := context.WithTimeout(ctx, w.SendTimeout)
	err = send(sendCtx, url, payload.Message)
	sendCancel()
	if err != nil {
		return true, w.retry(ctx, d, err.Error())
	}
	return true, w.finish(
		ctx,
		d,
		"sent",
		0,
		"",
		&payload,
	)
}

func (w *Worker) retry(ctx context.Context, d store.Delivery, diagnostic string) error {
	if d.Attempts >= w.MaxAttempts {
		return w.finish(
			ctx,
			d,
			"failed",
			0,
			diagnostic,
			nil,
		)
	}
	exponent := min(max(d.Attempts-1, 0), 10)
	delay := min(w.RetryBase*time.Duration(int64(1)<<exponent), 15*time.Minute)
	// Equal jitter spreads retries while retaining a nonzero bounded delay.
	if delay > 1 {
		delay = delay/2 + rand.N(delay-delay/2+1)
	}
	return w.finish(
		ctx,
		d,
		"pending",
		w.Now().UnixMilli()+delay.Milliseconds(),
		diagnostic,
		nil,
	)
}

func (w *Worker) finish(
	ctx context.Context,
	d store.Delivery,
	state string,
	nextDue int64,
	diagnostic string,
	payload *domain.NotificationPayload,
) error {
	now := w.Now().UnixMilli()
	err := w.Store.WithTx(ctx, func(tx *store.Tx) error {
		if err := tx.CompleteDelivery(
			ctx,
			d.ID,
			d.LeaseToken,
			now,
			state,
			nextDue,
			diagnostic,
		); err != nil {
			return err
		}
		if payload == nil || (payload.Kind != "down" && payload.Kind != "up") {
			return nil
		}
		id := domain.DeliveryMarkerID(payload.MonitorID, d.ChannelID, payload.CycleID)
		marker := domain.DeliveryMarker{MonitorID: payload.MonitorID, ChannelID: d.ChannelID, CycleID: payload.CycleID}
		if err := tx.Get(
			ctx,
			"deliveryMarkers",
			id,
			&marker,
		); err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
		if payload.Kind == "down" {
			marker.DownSentAt = now
		} else {
			marker.RecoverySentAt = now
		}
		return tx.Put(
			ctx,
			"deliveryMarkers",
			id,
			marker,
		)
	})
	if err == nil && w.OnDelivery != nil {
		w.OnDelivery(state)
	}
	return err
}

func (w *Worker) TestChannel(ctx context.Context, id string) error {
	select {
	case w.slots <- struct{}{}:
		defer func() { <-w.slots }()
	case <-ctx.Done():
		return ctx.Err()
	}
	var channel domain.Channel
	if err := w.Store.Get(
		ctx,
		"channels",
		id,
		&channel,
	); err != nil {
		return err
	}
	raw, err := w.Secrets.ResolveSecret(ctx, channel.ServiceURLSecretID)
	if err != nil {
		return errors.New("notification channel secret is unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, w.SendTimeout)
	defer cancel()
	return send(ctx, raw, "Octopulse notification channel test")
}
