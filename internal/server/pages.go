package server

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
var colorPattern = regexp.MustCompile(`^#[a-fA-F0-9]{6}$`)
var reservedSlugs = map[string]bool{"app": true, "api": true, "assets": true, "healthz": true, "incidents": true, "favicon": true, "robots": true}

func safePublicURL(value string, asset bool) bool {
	if asset && strings.HasPrefix(value, "/assets/uploads/") && !strings.Contains(value, "..") && !strings.ContainsAny(value, "?#\\") {
		return true
	}
	u, e := url.Parse(value)
	return e == nil && (u.Scheme == "https" || (!asset && u.Scheme == "http")) && u.Host != "" && u.User == nil
}

func (s *Server) validatePage(ctx context.Context, t *store.Tx, p *domain.Page) error {
	if strings.TrimSpace(p.Name) == "" || !slugPattern.MatchString(p.Slug) || reservedSlugs[p.Slug] {
		return huma.Error422UnprocessableEntity("Page name and a non-reserved lowercase slug are required")
	}
	if strings.TrimSpace(p.Draft.Title) == "" || !colorPattern.MatchString(p.Draft.BrandColor) {
		return huma.Error422UnprocessableEntity("Page title and a six-digit brand color are required")
	}
	if p.Draft.LogoURL != "" && !safePublicURL(p.Draft.LogoURL, true) {
		return huma.Error422UnprocessableEntity("Logo must use HTTPS or an uploaded image")
	}
	if p.Draft.ColorScheme == "" {
		p.Draft.ColorScheme = "system"
	}
	for _, l := range p.Draft.Links {
		if strings.TrimSpace(l.Label) == "" || !safePublicURL(l.URL, false) {
			return huma.Error422UnprocessableEntity("Public links require a label and HTTP(S) URL")
		}
	}
	seen := map[string]bool{}
	groups := map[string]bool{}
	for _, g := range p.Draft.Groups {
		if g.ID == "" || groups[g.ID] || strings.TrimSpace(g.Name) == "" {
			return huma.Error422UnprocessableEntity("Groups require unique IDs and names")
		}
		groups[g.ID] = true
		for _, m := range g.Monitors {
			if seen[m.MonitorID] {
				return huma.Error422UnprocessableEntity("Each monitor may appear once on a page")
			}
			seen[m.MonitorID] = true
			if _, e := t.GetMonitor(ctx, m.MonitorID); e != nil {
				return huma.Error422UnprocessableEntity("Page monitor is unavailable")
			}
		}
	}
	p.Domain = hostname(strings.TrimSpace(p.Domain))
	if p.Domain != "" {
		settings := domain.DefaultSettings()
		if e := t.Get(ctx, "settings", "organization", &settings); e != nil && !errors.Is(e, store.ErrNotFound) {
			return e
		}
		allowed := false
		for _, h := range settings.AllowedDomains {
			if h == p.Domain {
				allowed = true
			}
		}
		if !allowed || s.adminHost(p.Domain) {
			return huma.Error422UnprocessableEntity("Select an administrator-configured public domain")
		}
	}
	return nil
}

func bindPage(ctx context.Context, t *store.Tx, p *domain.Page) error {
	domains := []string{}
	if p.Domain != "" {
		domains = append(domains, p.Domain)
	}
	return t.BindPage(ctx, store.PageBinding{PageID: p.ID, Slug: p.Slug, Domains: domains})
}

func (s *Server) registerPages() {
	collection[domain.Page, *domain.Page](s, "pages", resourceHooks[domain.Page]{Save: func(ctx context.Context, t *store.Tx, p, old *domain.Page) error {
		p.CreatedAt = domain.Now()
		p.UpdatedAt = p.CreatedAt
		p.Published = nil
		p.PublishedAt = 0
		p.Version = 0
		p.PublishedSlug = ""
		p.PublishedDomain = ""
		if old != nil {
			p.CreatedAt = old.CreatedAt
			p.Published = old.Published
			p.PublishedAt = old.PublishedAt
			p.Version = old.Version
			p.PublishedSlug = old.PublishedSlug
			p.PublishedDomain = old.PublishedDomain
			if old.Published != nil && p.PublishedSlug == "" {
				p.PublishedSlug = old.Slug
				p.PublishedDomain = old.Domain
			}
		}
		if err := s.validatePage(ctx, t, p); err != nil {
			return err
		}
		if p.Published == nil {
			return bindPage(ctx, t, p)
		}
		return nil
	}})
	huma.Register(s.API, huma.Operation{OperationID: "publishPage", Method: "POST", Path: "/api/v1/pages/{id}/publish"}, func(ctx context.Context, in *IDInput) (*Output[domain.Page], error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		var p domain.Page
		e := s.Store.WithTx(ctx, func(t *store.Tx) error {
			if e := t.Get(ctx, "pages", in.ID, &p); e != nil {
				return e
			}
			if e := s.validatePage(ctx, t, &p); e != nil {
				return e
			}
			if err := bindPage(ctx, t, &p); err != nil {
				return err
			}
			p.PublishedSlug = p.Slug
			p.PublishedDomain = p.Domain
			published := p.Draft
			p.Published = &published
			p.PublishedAt = domain.Now()
			p.UpdatedAt = p.PublishedAt
			p.Version++
			if e := t.Put(ctx, "pages", p.ID, p); e != nil {
				return e
			}
			return audit(ctx, t, "publish", "pages", p.ID)
		})
		return &Output[domain.Page]{Body: p}, statusOrAPIError(e)
	})
	huma.Register(s.API, huma.Operation{OperationID: "previewPage", Method: "GET", Path: "/api/v1/pages/{id}/preview"}, func(ctx context.Context, in *IDInput) (*Output[domain.PublicPage], error) {
		var p domain.Page
		if e := s.Store.Get(ctx, "pages", in.ID, &p); e != nil {
			return nil, apiError(e)
		}
		v, e := s.projectPage(ctx, p, p.Draft)
		return &Output[domain.PublicPage]{Body: v}, apiError(e)
	})
	huma.Register(s.API, huma.Operation{OperationID: "getPublicPage", Method: "GET", Path: "/api/public/pages/{slug}"}, func(ctx context.Context, in *struct {
		Slug string `path:"slug"`
	}) (*Output[domain.PublicPage], error) {
		id, e := s.Store.PageIDBySlug(ctx, in.Slug)
		if e != nil {
			return nil, apiError(e)
		}
		host := requestHost(ctx)
		if !s.adminHost(host) {
			bound, e := s.Store.PageIDByDomain(ctx, hostname(host))
			if e != nil || bound != id {
				return nil, huma.Error404NotFound("Page not found")
			}
		}
		return s.publishedPage(ctx, id)
	})
	huma.Register(s.API, huma.Operation{OperationID: "resolvePublicPage", Method: "GET", Path: "/api/public/resolve"}, func(ctx context.Context, in *struct {
		Host string `query:"host"`
	}) (*Output[domain.PublicPage], error) {
		host := hostname(requestHost(ctx))
		if in.Host != "" && hostname(in.Host) != host {
			return nil, huma.Error404NotFound("Page not found")
		}
		id, e := s.Store.PageIDByDomain(ctx, host)
		if e != nil {
			return nil, apiError(e)
		}
		return s.publishedPage(ctx, id)
	})
	huma.Register(s.API, huma.Operation{OperationID: "createIncidentUpdate", Method: "POST", Path: "/api/v1/incidents/{id}/updates"}, func(ctx context.Context, in *WriteInput[IncidentProgress]) (*Output[domain.Incident], error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		var incident domain.Incident
		e := s.Store.WithTx(ctx, func(t *store.Tx) error {
			if e := t.Get(ctx, "incidents", in.ID, &incident); e != nil {
				return e
			}
			incident.Updates = append(incident.Updates, domain.IncidentUpdate{ID: domain.ID(), Body: in.Body.Body, Status: in.Body.Status, CreatedAt: domain.Now()})
			incident.Status = in.Body.Status
			incident.UpdatedAt = domain.Now()
			incident.ResolvedAt = 0
			if incident.Status == "resolved" {
				incident.ResolvedAt = incident.UpdatedAt
			}
			if e := t.Put(ctx, "incidents", in.ID, incident); e != nil {
				return e
			}
			return audit(ctx, t, "update", "incidents", in.ID)
		})
		return &Output[domain.Incident]{Body: incident}, apiError(e)
	})
}

type IncidentProgress struct {
	Body   string `json:"body" minLength:"1" maxLength:"20000"`
	Status string `json:"status" enum:"investigating,identified,monitoring,resolved"`
}

func (s *Server) publishedPage(ctx context.Context, id string) (*Output[domain.PublicPage], error) {
	var p domain.Page
	if e := s.Store.Get(ctx, "pages", id, &p); e != nil {
		return nil, apiError(e)
	}
	if p.Published == nil {
		return nil, huma.Error404NotFound("Page is not published")
	}
	if p.PublishedSlug != "" {
		p.Slug = p.PublishedSlug
	}
	v, e := s.projectPage(ctx, p, *p.Published)
	return &Output[domain.PublicPage]{Body: v}, apiError(e)
}
func hasID(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
func (s *Server) projectPage(ctx context.Context, p domain.Page, c domain.PageConfig) (domain.PublicPage, error) {
	now := domain.Now()
	from := now - 86400000
	result := domain.PublicPage{ID: p.ID, Slug: p.Slug, Config: c, State: "unknown", Groups: []domain.PublicGroup{}, Incidents: []domain.PublicIncident{}, Maintenance: []domain.PublicMaintenance{}, UpdatedAt: now}
	maintenance, e := list[domain.Maintenance](ctx, s.Store, "maintenance")
	if e != nil {
		return result, e
	}
	ids := map[string]bool{}
	states := []domain.PublicMonitor{}
	for _, g := range c.Groups {
		group := domain.PublicGroup{ID: g.ID, Name: g.Name, Monitors: []domain.PublicMonitor{}}
		for _, ref := range g.Monitors {
			m, e := s.monitor(ctx, ref.MonitorID)
			if e != nil {
				return result, e
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
				if s.Stats != nil {
					item.Availability, e = s.Stats(ctx, m.ID, from, now)
					if e != nil {
						return result, e
					}
				}
				if ref.ShowLatency {
					if s.Latency != nil {
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
	result.State = pageState(states)
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
func pageState(items []domain.PublicMonitor) string {
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
