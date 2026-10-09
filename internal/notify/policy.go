package notify

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type awaitingDown struct{ until int64 }

func (*awaitingDown) Error() string { return "waiting for in-flight fault delivery completion" }

func (w *Worker) relevant(ctx context.Context, d store.Delivery, p domain.NotificationPayload) (bool, error) {
	monitor, err := w.Store.GetMonitor(ctx, p.MonitorID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if !monitor.Enabled {
		return false, nil
	}
	runtime, err := w.Store.GetRuntime(ctx, p.MonitorID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if runtime.Generation != p.Generation || d.Generation != p.Generation {
		return false, nil
	}
	maintenance, err := w.Store.List(ctx, "maintenance")
	if err != nil {
		return false, err
	}
	now := w.Now().UnixMilli()
	for _, raw := range maintenance {
		var window domain.Maintenance
		if err = json.Unmarshal(raw, &window); err != nil {
			return false, err
		}
		availabilityNotification := !strings.HasPrefix(p.Kind, "certificate_")
		activeWindow := window.StartsAt <= now && now < window.EndsAt
		if !availabilityNotification || !activeWindow {
			continue
		}
		for _, id := range window.MonitorIDs {
			if id == p.MonitorID {
				return false, nil
			}
		}
	}
	switch p.Kind {
	case "down", "reminder":
		return runtime.State == domain.StateDown, nil
	case "up":
		if runtime.State != domain.StateUp {
			return false, nil
		}
		flights, err := w.Store.InFlightDeliveries(ctx, p.MonitorID, d.ChannelID)
		if err != nil {
			return false, err
		}
		for _, flight := range flights {
			var fault domain.NotificationPayload
			if err := json.Unmarshal(flight.Payload, &fault); err != nil {
				continue
			}
			sameFaultCycle := fault.Kind == "down" && fault.CycleID == p.CycleID
			if sameFaultCycle && flight.LeaseUntil > now {
				return false, &awaitingDown{until: flight.LeaseUntil}
			}
		}
		var marker domain.DeliveryMarker
		err = w.Store.Get(
			ctx,
			"deliveryMarkers",
			domain.DeliveryMarkerID(p.MonitorID, d.ChannelID, p.CycleID),
			&marker,
		)
		if errors.Is(err, store.ErrNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return marker.DownSentAt > 0 && marker.RecoverySentAt == 0, nil
	default:
		if !strings.HasPrefix(p.Kind, "certificate_") {
			return false, nil
		}
		var metadata struct {
			Certificate *domain.CertificateConfig `json:"certificate"`
		}
		err = w.Store.Get(
			ctx,
			"engineMonitor",
			p.MonitorID,
			&metadata,
		)
		if errors.Is(err, store.ErrNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if metadata.Certificate == nil {
			return false, nil
		}
		if p.CycleID != "" && metadata.Certificate.Fingerprint != p.CycleID {
			return false, nil
		}
		return true, nil
	}
}
