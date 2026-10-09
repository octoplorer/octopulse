package server

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type DeliveryView struct {
	ID        string `json:"id"`
	EventID   string `json:"eventId"`
	MonitorID string `json:"monitorId"`
	ChannelID string `json:"channelId"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
	DueAt     int64  `json:"dueAt"`
	Attempts  int64  `json:"attempts"`
	LastError string `json:"lastError"`
}

func (s *Server) registerNotifications() {
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "listDeliveries",
			Method:      "GET",
			Path:        "/api/v1/deliveries",
		},
		func(ctx context.Context, in *struct {
			Limit int `query:"limit" default:"200" minimum:"1" maximum:"1000"`
		}) (*Output[Items[DeliveryView]], error) {
			rows, e := s.Store.ListDeliveries(ctx, in.Limit)
			if e != nil {
				return nil, apiError(ctx, e)
			}
			items := []DeliveryView{}
			for _, d := range rows {
				var p domain.NotificationPayload
				if e = json.Unmarshal(d.Payload, &p); e != nil {
					return nil, apiError(ctx, e)
				}
				items = append(
					items,
					DeliveryView{
						ID:        d.ID,
						EventID:   d.EventID,
						MonitorID: p.MonitorID,
						ChannelID: d.ChannelID,
						Kind:      p.Kind,
						Status:    d.State,
						CreatedAt: p.CreatedAt,
						DueAt:     d.DueAt,
						Attempts:  d.Attempts,
						LastError: d.LastError,
					},
				)
			}
			tests, e := list[DeliveryView](ctx, s.Store, "channelTests")
			if e != nil {
				return nil, apiError(ctx, e)
			}
			items = append(items, tests...)
			sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
			if len(items) > in.Limit {
				items = items[:in.Limit]
			}
			return &Output[Items[DeliveryView]]{Body: Items[DeliveryView]{Items: items}}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "testChannel",
			Method:      "POST",
			Path:        "/api/v1/channels/{id}/test",
		},
		func(ctx context.Context, in *IDInput) (*Output[Ack], error) {
			if s.TestChannel == nil {
				return nil, huma.Error503ServiceUnavailable("Notification worker is unavailable")
			}
			var channel domain.Channel
			if e := s.Store.Get(
				ctx,
				"channels",
				in.ID,
				&channel,
			); e != nil {
				return nil, apiError(ctx, e)
			}
			record := DeliveryView{
				ID:        domain.ID(),
				ChannelID: in.ID,
				Kind:      "test",
				Status:    "sending",
				CreatedAt: domain.Now(),
				Attempts:  1,
			}
			if e := s.Store.WithTx(ctx, func(t *store.Tx) error {
				if e := t.Put(
					ctx,
					"channelTests",
					record.ID,
					record,
				); e != nil {
					return e
				}
				return audit(
					ctx,
					t,
					"test",
					"channels",
					in.ID,
				)
			}); e != nil {
				return nil, apiError(ctx, e)
			}
			sendError := s.TestChannel(ctx, in.ID)
			record.Status = "sent"
			if sendError != nil {
				record.Status = "failed"
				record.LastError = "Notification provider rejected the test or the send timed out"
			}
			saveCtx := ctx
			done, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if ctx.Err() != nil {
				saveCtx = done
			}
			if e := s.Store.Put(
				saveCtx,
				"channelTests",
				record.ID,
				record,
			); e != nil {
				return nil, apiError(ctx, e)
			}
			if sendError != nil {
				return nil, huma.Error502BadGateway(record.LastError)
			}
			return &Output[Ack]{Body: Ack{OK: true}}, nil
		},
	)
}
