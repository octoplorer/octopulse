package server

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/statuspage"
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
		return &Output[domain.Page]{Body: p}, statusOrAPIError(ctx, e)
	})
	huma.Register(s.API, huma.Operation{OperationID: "previewPage", Method: "GET", Path: "/api/v1/pages/{id}/preview"}, func(ctx context.Context, in *IDInput) (*Output[domain.PublicPage], error) {
		var p domain.Page
		if e := s.Store.Get(ctx, "pages", in.ID, &p); e != nil {
			return nil, apiError(ctx, e)
		}
		v, e := s.projectPageWithHistory(ctx, p, p.Draft, true)
		return &Output[domain.PublicPage]{Body: v}, apiError(ctx, e)
	})
	huma.Register(s.API, huma.Operation{OperationID: "getPublicPage", Method: "GET", Path: "/api/public/pages/{slug}"}, func(ctx context.Context, in *struct {
		Slug string `path:"slug"`
	}) (*Output[domain.PublicPage], error) {
		id, e := s.Store.PageIDBySlug(ctx, in.Slug)
		if e != nil {
			return nil, apiError(ctx, e)
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
			return nil, apiError(ctx, e)
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
		return &Output[domain.Incident]{Body: incident}, apiError(ctx, e)
	})
}

type IncidentProgress struct {
	Body   string `json:"body" minLength:"1" maxLength:"20000"`
	Status string `json:"status" enum:"investigating,identified,monitoring,resolved"`
}

func (s *Server) publishedPage(ctx context.Context, id string) (*Output[domain.PublicPage], error) {
	var p domain.Page
	if e := s.Store.Get(ctx, "pages", id, &p); e != nil {
		return nil, apiError(ctx, e)
	}
	if p.Published == nil {
		return nil, huma.Error404NotFound("Page is not published")
	}
	if p.PublishedSlug != "" {
		p.Slug = p.PublishedSlug
	}
	v, e := s.projectPage(ctx, p, *p.Published)
	return &Output[domain.PublicPage]{Body: v}, apiError(ctx, e)
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
	return s.projectPageWithHistory(ctx, p, c, false)
}
func (s *Server) projectPageWithHistory(ctx context.Context, p domain.Page, c domain.PageConfig, includeHidden bool) (domain.PublicPage, error) {
	reader := statuspage.Reader{IncludeHiddenDaily: includeHidden, Store: s.Store, Monitors: s.Monitors, Stats: s.Stats, Latency: s.Latency, StatsBatch: s.StatsBatch, LatencyBatch: s.LatencyBatch, DailyStatsBatch: s.DailyStatsBatch}
	return reader.Project(ctx, p, c)
}
func pageState(items []domain.PublicMonitor) string { return statuspage.State(items) }
