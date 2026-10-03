package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/engine"
	"github.com/octoplorer/octopulse/internal/monitoring"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/store"
)

type MonitorWrite struct {
	domain.Monitor
	Enabled        *bool `json:"enabled,omitempty"`
	Retries        *int  `json:"retries,omitempty"`
	NotifyRecovery *bool `json:"notifyRecovery,omitempty"`
}
type HistoryInput struct {
	ID    string `path:"id"`
	From  int64  `query:"from"`
	To    int64  `query:"to"`
	Limit int    `query:"limit" default:"200" minimum:"1" maximum:"1000"`
}
type MonitorHistory struct {
	Availability domain.Availability   `json:"availability"`
	Rounds       []store.Round         `json:"rounds"`
	Latency      []domain.LatencyPoint `json:"latency"`
}
type HeartbeatToken struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}
type HeartbeatReport struct {
	Status      string `json:"status,omitempty" enum:"up,down"`
	Description string `json:"description,omitempty" maxLength:"1000"`
}

func (s *Server) monitor(ctx context.Context, id string) (domain.Monitor, error) {
	return s.Monitors.Get(ctx, id)
}

func (s *Server) registerMonitors() {
	huma.Register(s.API, huma.Operation{OperationID: "listMonitors", Method: "GET", Path: "/api/v1/monitors"}, func(ctx context.Context, _ *struct{}) (*Output[Items[domain.Monitor]], error) {
		items, e := s.Monitors.List(ctx)
		if e != nil {
			return nil, apiError(ctx, e)
		}
		return &Output[Items[domain.Monitor]]{Body: Items[domain.Monitor]{items}}, nil
	})
	huma.Register(s.API, huma.Operation{OperationID: "getMonitor", Method: "GET", Path: "/api/v1/monitors/{id}"}, func(ctx context.Context, in *IDInput) (*Output[domain.Monitor], error) {
		m, e := s.monitor(ctx, in.ID)
		return &Output[domain.Monitor]{Body: m}, apiError(ctx, e)
	})
	huma.Register(s.API, huma.Operation{OperationID: "createMonitor", Method: "POST", Path: "/api/v1/monitors", MaxBodyBytes: 8 << 20}, func(ctx context.Context, in *CreateInput[MonitorWrite]) (*Output[domain.Monitor], error) {
		return s.saveMonitor(ctx, "", in.Body)
	})
	huma.Register(s.API, huma.Operation{OperationID: "updateMonitor", Method: "PATCH", Path: "/api/v1/monitors/{id}", MaxBodyBytes: 8 << 20}, func(ctx context.Context, in *WriteInput[MonitorWrite]) (*Output[domain.Monitor], error) {
		return s.saveMonitor(ctx, in.ID, in.Body)
	})
	huma.Register(s.API, huma.Operation{OperationID: "deleteMonitor", Method: "DELETE", Path: "/api/v1/monitors/{id}"}, func(ctx context.Context, in *IDInput) (*Output[Ack], error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		e := s.Store.WithTx(ctx, func(t *store.Tx) error {
			pages, e := t.List(ctx, "pages")
			if e != nil {
				return e
			}
			for _, b := range pages {
				var p domain.Page
				if e = decode(b, &p); e != nil {
					return e
				}
				configs := []domain.PageConfig{p.Draft}
				if p.Published != nil {
					configs = append(configs, *p.Published)
				}
				for _, c := range configs {
					for _, g := range c.Groups {
						for _, m := range g.Monitors {
							if m.MonitorID == in.ID {
								return huma.Error409Conflict("Monitor is referenced by a status page")
							}
						}
					}
				}
			}
			if e = t.DeleteMonitor(ctx, in.ID); e != nil {
				return e
			}
			if e = t.Delete(ctx, "monitors", in.ID); e != nil && !errors.Is(e, store.ErrNotFound) {
				return e
			}
			return audit(ctx, t, "delete", "monitors", in.ID)
		})
		if e == nil && s.Changed != nil {
			e = s.Changed(ctx, in.ID)
			if errors.Is(e, store.ErrNotFound) {
				e = nil
			}
		}
		return &Output[Ack]{Body: Ack{true}}, statusOrAPIError(ctx, e)
	})
	huma.Register(s.API, huma.Operation{OperationID: "checkMonitor", Method: "POST", Path: "/api/v1/monitors/{id}/check", DefaultStatus: http.StatusAccepted}, func(ctx context.Context, in *IDInput) (*Output[Ack], error) {
		if s.Check == nil {
			return nil, huma.Error503ServiceUnavailable("Scheduler is not running")
		}
		if e := s.Check(ctx, in.ID); e != nil && !errors.Is(e, engine.ErrCheckAccepted) {
			return nil, huma.Error409Conflict("Monitor check could not complete")
		}
		return &Output[Ack]{Body: Ack{true}}, nil
	})
	huma.Register(s.API, huma.Operation{OperationID: "getMonitorHistory", Method: "GET", Path: "/api/v1/monitors/{id}/history"}, func(ctx context.Context, in *HistoryInput) (*Output[MonitorHistory], error) {
		if _, e := s.Store.GetMonitor(ctx, in.ID); e != nil {
			return nil, apiError(ctx, e)
		}
		to := in.To
		if to == 0 {
			to = domain.Now()
		}
		if to > domain.Now() {
			to = domain.Now()
		}
		from := in.From
		if from == 0 {
			from = to - 24*60*60*1000
		}
		if from < 0 || to <= from || to-from > 410*24*60*60*1000 {
			return nil, huma.Error422UnprocessableEntity("Invalid history window")
		}
		rounds, e := s.Store.ListRoundsBetween(ctx, in.ID, from, to, in.Limit)
		if e != nil {
			return nil, apiError(ctx, e)
		}
		filtered := []store.Round{}
		latency := []domain.LatencyPoint{}
		for _, r := range rounds {
			if r.FinishedAt > to {
				continue
			}
			if CurrentUser(ctx).Role == domain.RoleViewer {
				for i := range r.Attempts {
					r.Attempts[i].Detail = nil
					r.Attempts[i].Error = ""
				}
			}
			filtered = append(filtered, r)
			latency = append(latency, domain.LatencyPoint{At: r.FinishedAt, LatencyMs: float64(r.LatencyMS), Success: r.Success})
		}
		stats := domain.Availability{From: from, To: to, UnknownMs: to - from}
		if s.Latency != nil {
			latency, e = s.Latency(ctx, in.ID, from, to)
			if e != nil {
				return nil, apiError(ctx, e)
			}
		}
		if s.Stats != nil {
			stats, e = s.Stats(ctx, in.ID, from, to)
			if e != nil {
				return nil, apiError(ctx, e)
			}
		}
		return &Output[MonitorHistory]{Body: MonitorHistory{stats, filtered, latency}}, nil
	})
	huma.Register(s.API, huma.Operation{OperationID: "rotateHeartbeat", Method: "POST", Path: "/api/v1/monitors/{id}/heartbeat/rotate"}, func(ctx context.Context, in *IDInput) (*Output[HeartbeatToken], error) {
		m, e := s.monitor(ctx, in.ID)
		if e != nil {
			return nil, apiError(ctx, e)
		}
		if m.Type != domain.MonitorHeartbeat {
			return nil, huma.Error422UnprocessableEntity("Monitor is not a heartbeat")
		}
		token := security.Token()
		e = s.Store.WithTx(ctx, func(t *store.Tx) error {
			if e := t.Put(ctx, "heartbeatSecrets", in.ID, struct {
				Hash string `json:"hash"`
			}{security.HashToken(token)}); e != nil {
				return e
			}
			return audit(ctx, t, "rotate", "heartbeats", in.ID)
		})
		return &Output[HeartbeatToken]{Body: HeartbeatToken{Token: token, URL: "/api/heartbeat/" + in.ID + "/" + token}}, apiError(ctx, e)
	})
	huma.Register(s.API, huma.Operation{OperationID: "reportHeartbeat", Method: "POST", Path: "/api/heartbeat/{id}/{token}"}, func(ctx context.Context, in *struct {
		ID    string `path:"id"`
		Token string `path:"token"`
		Body  HeartbeatReport
	}) (*Output[Ack], error) {
		var secret struct {
			Hash string `json:"hash"`
		}
		if e := s.Store.Get(ctx, "heartbeatSecrets", in.ID, &secret); e != nil || !security.Equal(secret.Hash, security.HashToken(in.Token)) {
			return nil, huma.Error404NotFound("Heartbeat not found")
		}
		if s.Heartbeat == nil {
			return nil, huma.Error503ServiceUnavailable("Scheduler is not running")
		}
		if e := s.Heartbeat(ctx, in.ID, in.Body.Status != "down", in.Body.Description); e != nil {
			return nil, huma.Error409Conflict("Heartbeat report could not be recorded")
		}
		return &Output[Ack]{Body: Ack{true}}, nil
	})
}

func (s *Server) saveMonitor(ctx context.Context, id string, in MonitorWrite) (*Output[domain.Monitor], error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.Monitors.Save(ctx, id, monitoring.Write{Monitor: in.Monitor, Enabled: in.Enabled, Retries: in.Retries, NotifyRecovery: in.NotifyRecovery}, CurrentUser(ctx))
	return &Output[domain.Monitor]{Body: m}, apiError(ctx, err)
}
