// Package statuspage builds the public field whitelist from batched monitor reads.
package statuspage

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/monitoring"
	"github.com/octoplorer/octopulse/internal/store"
)

type Reader struct {
	Store        *store.Store
	Monitors     *monitoring.Service
	Stats        func(context.Context, string, int64, int64) (domain.Availability, error)
	Latency      func(context.Context, string, int64, int64) ([]domain.LatencyPoint, error)
	StatsBatch   func(context.Context, []string, int64, int64) (map[string]domain.Availability, error)
	LatencyBatch func(context.Context, []string, int64, int64) (map[string][]domain.LatencyPoint, error)
}

func hasID(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
func list[T any](ctx context.Context, st *store.Store, kind string) ([]T, error) {
	raw, err := st.List(ctx, kind)
	if err != nil {
		return nil, err
	}
	items := make([]T, 0, len(raw))
	for _, b := range raw {
		var item T
		if err := json.Unmarshal(b, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
func (s *Reader) Project(ctx context.Context, p domain.Page, c domain.PageConfig) (domain.PublicPage, error) {
	now := domain.Now()
	from := now - 86400000
	result := domain.PublicPage{ID: p.ID, Slug: p.Slug, Config: c, State: "unknown", Groups: []domain.PublicGroup{}, Incidents: []domain.PublicIncident{}, Maintenance: []domain.PublicMaintenance{}, UpdatedAt: now}
	maintenance, e := list[domain.Maintenance](ctx, s.Store, "maintenance")
	if e != nil {
		return result, e
	}
	var monitorIDs, availabilityIDs, latencyIDs []string
	seen := map[string]bool{}
	needsLatency := map[string]bool{}
	for _, group := range c.Groups {
		for _, ref := range group.Monitors {
			if !seen[ref.MonitorID] {
				monitorIDs = append(monitorIDs, ref.MonitorID)
				seen[ref.MonitorID] = true
			}
			needsLatency[ref.MonitorID] = needsLatency[ref.MonitorID] || ref.ShowLatency
		}
	}
	monitors, e := s.Monitors.Read(ctx, monitorIDs)
	if e != nil {
		return result, e
	}
	for _, id := range monitorIDs {
		m, ok := monitors[id]
		if !ok {
			return result, store.ErrNotFound
		}
		if m.IsAvailability() {
			availabilityIDs = append(availabilityIDs, id)
			if needsLatency[id] {
				latencyIDs = append(latencyIDs, id)
			}
		}
	}
	var availability map[string]domain.Availability
	var latency map[string][]domain.LatencyPoint
	if s.StatsBatch != nil && len(availabilityIDs) > 0 {
		availability, e = s.StatsBatch(ctx, availabilityIDs, from, now)
		if e != nil {
			return result, e
		}
	}
	if s.LatencyBatch != nil && len(latencyIDs) > 0 {
		latency, e = s.LatencyBatch(ctx, latencyIDs, from, now)
		if e != nil {
			return result, e
		}
	}
	ids := map[string]bool{}
	states := []domain.PublicMonitor{}
	for _, g := range c.Groups {
		group := domain.PublicGroup{ID: g.ID, Name: g.Name, Monitors: []domain.PublicMonitor{}}
		for _, ref := range g.Monitors {
			m, ok := monitors[ref.MonitorID]
			if !ok {
				return result, store.ErrNotFound
			}
			ids[m.ID] = true
			name := ref.Alias
			if strings.TrimSpace(name) == "" {
				name = m.Name
			}
			item := domain.PublicMonitor{ID: m.ID, Name: name, Type: m.Type, State: m.State, Paused: !m.Enabled, Availability: domain.Availability{From: from, To: now, UnknownMs: now - from}, Latency: []domain.LatencyPoint{}}
			for _, w := range maintenance {
				if hasID(w.MonitorIDs, m.ID) && w.StartsAt <= now && now < w.EndsAt {
					item.Maintenance = true
				}
			}
			if m.Certificate != nil {
				item.Certificate = &domain.PublicCertificate{State: m.Certificate.State, ExpiresAt: m.Certificate.ExpiresAt, DaysRemaining: m.Certificate.DaysRemaining}
				if item.Certificate.State == "" {
					item.Certificate.State = domain.CertificateCheckFailed
				}
			}
			if m.IsAvailability() {
				if availability != nil {
					var ok bool
					item.Availability, ok = availability[m.ID]
					if !ok {
						return result, store.ErrNotFound
					}
				} else if s.Stats != nil {
					item.Availability, e = s.Stats(ctx, m.ID, from, now)
					if e != nil {
						return result, e
					}
				}
				if ref.ShowLatency {
					if latency != nil {
						var ok bool
						item.Latency, ok = latency[m.ID]
						if !ok {
							return result, store.ErrNotFound
						}
					} else if s.Latency != nil {
						item.Latency, e = s.Latency(ctx, m.ID, from, now)
						if e != nil {
							return result, e
						}
					} else {
						rounds, e := s.Store.ListRounds(ctx, m.ID, from, 96)
						if e != nil {
							return result, e
						}
						for i := len(rounds) - 1; i >= 0; i-- {
							r := rounds[i]
							item.Latency = append(item.Latency, domain.LatencyPoint{At: r.FinishedAt, LatencyMs: float64(r.LatencyMS), Success: r.Success})
						}
					}
				}
			}
			group.Monitors = append(group.Monitors, item)
			states = append(states, item)
		}
		result.Groups = append(result.Groups, group)
	}
	result.State = State(states)
	for _, w := range maintenance {
		applies := hasID(w.PageIDs, p.ID)
		for _, id := range w.MonitorIDs {
			applies = applies || ids[id]
		}
		if applies && w.EndsAt > now && w.StartsAt < now+30*86400000 {
			result.Maintenance = append(result.Maintenance, domain.PublicMaintenance{ID: w.ID, Name: w.Name, Description: w.Description, StartsAt: w.StartsAt, EndsAt: w.EndsAt})
		}
	}
	incidents, e := list[domain.Incident](ctx, s.Store, "incidents")
	if e != nil {
		return result, e
	}
	sort.Slice(incidents, func(i, j int) bool { return incidents[i].CreatedAt > incidents[j].CreatedAt })
	for _, incident := range incidents {
		if !hasID(incident.PageIDs, p.ID) {
			continue
		}
		if incident.Status != "resolved" || len(result.Incidents) < 100 {
			result.Incidents = append(result.Incidents, domain.PublicIncident{ID: incident.ID, Title: incident.Title, Body: incident.Body, Status: incident.Status, Impact: incident.Impact, Updates: incident.Updates, CreatedAt: incident.CreatedAt, ResolvedAt: incident.ResolvedAt})
		}
		if incident.Status != "resolved" {
			if incident.Impact == "outage" {
				result.State = "outage"
			} else if incident.Impact == "partial" && result.State != "outage" {
				result.State = "partial"
			}
		}
	}
	return result, nil
}
func State(items []domain.PublicMonitor) string {
	eligible, down, unknown, maintenance := 0, 0, 0, 0
	for _, m := range items {
		if m.Paused || m.Type == domain.MonitorCertificate {
			continue
		}
		eligible++
		if m.Maintenance {
			maintenance++
			continue
		}
		switch m.State {
		case domain.StateDown:
			down++
		case domain.StateUp:
		default:
			unknown++
		}
	}
	switch {
	case eligible == 0:
		return "unknown"
	case down == eligible:
		return "outage"
	case down > 0:
		return "partial"
	case unknown > 0:
		return "unknown"
	case maintenance > 0:
		return "maintenance"
	default:
		return "operational"
	}
}
