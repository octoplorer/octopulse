package server

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/octoplorer/octopulse/internal/beszel"
	"github.com/octoplorer/octopulse/internal/config"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/monitoring"
	"github.com/octoplorer/octopulse/internal/secrets"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/statistics"
	"github.com/octoplorer/octopulse/internal/store"
	"github.com/octoplorer/octopulse/internal/telemetry"
)

type Server struct {
	Store           *store.Store
	Vault           *security.Vault
	Secrets         *secrets.Resolver
	Monitors        *monitoring.Service
	Config          config.Config
	API             huma.API
	Metrics         *telemetry.Metrics
	Mux             *http.ServeMux
	mu              sync.Mutex
	dummyPassword   string
	Check           func(context.Context, string) error
	NextCheck       func(string) int64
	Stats           func(context.Context, string, int64, int64) (domain.Availability, error)
	Latency         func(context.Context, string, int64, int64) ([]domain.LatencyPoint, error)
	StatsBatch      func(context.Context, []string, int64, int64) (map[string]domain.Availability, error)
	LatencyBatch    func(context.Context, []string, int64, int64) (map[string][]domain.LatencyPoint, error)
	DailyStatsBatch func(context.Context, []string, int64, int64) (map[string][]domain.Availability, error)
	Changed         func(context.Context, string) error
	Heartbeat       func(context.Context, string, bool, string) error
	Wake            func()
	TestChannel     func(context.Context, string) error
	Beszel          *beszel.Client
}
type contextKey struct{}
type hostContextKey struct{}

func requestHost(ctx context.Context) string { v, _ := ctx.Value(hostContextKey{}).(string); return v }

type identity struct {
	User    domain.User
	Session domain.Session
}
type Items[T any] struct {
	Items []T `json:"items"`
}
type Output[T any] struct{ Body T }
type IDInput struct {
	ID string `path:"id"`
}
type WriteInput[T any] struct {
	ID   string `path:"id"`
	Body T
}
type CreateInput[T any] struct{ Body T }
type Ack struct {
	OK bool `json:"ok"`
}

func New(st *store.Store, v *security.Vault, c config.Config) *Server {
	s := &Server{Store: st, Vault: v, Secrets: secrets.New(st, v), Config: c, Mux: http.NewServeMux()}
	s.Monitors = &monitoring.Service{Store: st, NextCheck: func(id string) int64 {
		if s.NextCheck != nil {
			return s.NextCheck(id)
		}
		return 0
	}, Changed: func(ctx context.Context, id string) error {
		if s.Changed != nil {
			return s.Changed(ctx, id)
		}
		return nil
	}}
	s.dummyPassword, _ = security.HashPassword(security.Token())
	hc := huma.DefaultConfig("Octopulse API", "0.1.0")
	hc.OpenAPIPath = "/api/openapi"
	hc.DocsPath = "/api/docs"
	hc.SchemasPath = "/api/schemas"
	hc.Formats = apiJSONFormats(hc.Formats)
	s.API = humago.New(s.Mux, hc)
	s.registerIdentity()
	s.registerConfiguration()
	s.registerMonitors()
	s.registerPages()
	s.registerAssets()
	s.registerNotifications()
	s.Beszel = s.registerBeszel()
	s.Mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if s.Store == nil || s.Store.Ping(r.Context()) != nil {
			writeProblem(w, http.StatusServiceUnavailable, "Database is unavailable")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Ack{OK: true})
	})
	s.Mux.HandleFunc("GET /", s.spa)
	return s
}
func (s *Server) Handler() http.Handler { return http.HandlerFunc(s.dispatch) }
func CurrentUser(ctx context.Context) domain.User {
	v, _ := ctx.Value(contextKey{}).(identity)
	return v.User
}
func currentIdentity(ctx context.Context) identity {
	v, _ := ctx.Value(contextKey{}).(identity)
	return v
}
func (s *Server) authenticate(ctx context.Context, token string) (identity, error) {
	var session domain.Session
	if token == "" {
		return identity{}, store.ErrNotFound
	}
	if e := s.Store.Get(ctx, "sessions", security.HashToken(token), &session); e != nil {
		return identity{}, e
	}
	if session.ExpiresAt <= domain.Now() {
		return identity{}, store.ErrNotFound
	}
	var user domain.UserRecord
	if e := s.Store.Get(ctx, "users", session.UserID, &user); e != nil {
		return identity{}, e
	}
	if !user.Enabled {
		return identity{}, store.ErrNotFound
	}
	return identity{User: user.User, Session: session}, nil
}
func (s *Server) dispatch(w http.ResponseWriter, r *http.Request) {
	requestID := security.Token()
	w.Header().Set("X-Request-ID", requestID)
	r = r.WithContext(telemetry.WithRequestID(r.Context(), requestID))
	started := time.Now()
	response := &responseStatus{ResponseWriter: w, status: http.StatusOK}
	w = response
	if s.Metrics != nil {
		s.Metrics.BeginRequest()
		defer func() { s.Metrics.EndRequest(r.Pattern, r.Method, response.status, time.Since(started)) }()
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(context.WithValue(ctx, hostContextKey{}, r.Host))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("X-Frame-Options", "DENY")
	path := r.URL.Path
	managed := strings.HasPrefix(path, "/api/v1/") || path == "/app" || strings.HasPrefix(path, "/app/") || strings.HasPrefix(path, "/api/docs") || strings.HasPrefix(path, "/api/openapi") || strings.HasPrefix(path, "/api/schemas")
	if managed && !s.adminHost(r.Host) {
		writeProblem(w, 404, "Unknown management host")
		return
	}
	if strings.HasPrefix(path, "/api/v1/") {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" && r.Method != "HEAD" {
			if origin := r.Header.Get("Origin"); origin != "" {
				u, e := url.Parse(origin)
				if e != nil || u.Host != r.Host || (s.Config.CookieSecure && u.Scheme != "https") {
					writeProblem(w, 403, "Untrusted request origin")
					return
				}
			}
			r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
		}
		var ident identity
		if cookie, e := r.Cookie("octopulse_session"); e == nil {
			var authErr error
			ident, authErr = s.authenticate(r.Context(), cookie.Value)
			if authErr != nil && !errors.Is(authErr, store.ErrNotFound) {
				telemetry.LogError(r.Context(), "api.authenticate", authErr)
				writeProblem(w, http.StatusServiceUnavailable, "Authentication is unavailable")
				return
			}
		}
		open := path == "/api/v1/setup" || path == "/api/v1/session"
		if !open {
			if ident.User.ID == "" {
				writeProblem(w, 401, "Authentication required")
				return
			}
			if r.Method != "GET" && r.Method != "HEAD" {
				if !security.Equal(r.Header.Get("X-CSRF-Token"), ident.Session.CSRFToken) {
					writeProblem(w, 403, "Invalid CSRF token")
					return
				}
				if ident.User.Role == domain.RoleViewer && path != "/api/v1/profile" {
					writeProblem(w, 403, "Read-only account")
					return
				}
				if adminResource(path) && ident.User.Role != domain.RoleAdmin {
					writeProblem(w, 403, "Administrator permission required")
					return
				}
			}
		} else if r.Method == "DELETE" && path == "/api/v1/session" {
			if ident.User.ID == "" || !security.Equal(r.Header.Get("X-CSRF-Token"), ident.Session.CSRFToken) {
				writeProblem(w, 403, "Invalid CSRF token")
				return
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), contextKey{}, ident))
	}
	s.Mux.ServeHTTP(w, r)
}
func adminResource(path string) bool {
	for _, p := range []string{"/users", "/secrets", "/channels", "/settings", "/beszel/config"} {
		if strings.HasPrefix(path, "/api/v1"+p) {
			return true
		}
	}
	return false
}
func (s *Server) adminHost(host string) bool {
	h := host
	if v, _, e := net.SplitHostPort(h); e == nil {
		h = v
	}
	h = strings.ToLower(strings.TrimSuffix(h, "."))
	for _, allowed := range s.Config.AdminHosts {
		if h == allowed {
			return true
		}
	}
	return false
}
func writeProblem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"status": status, "title": http.StatusText(status), "detail": detail})
}
func apiError(ctx context.Context, e error) error {
	if e == nil {
		return nil
	}
	var status huma.StatusError
	if errors.As(e, &status) {
		return status
	}
	var invalid *monitoring.ValidationError
	if errors.As(e, &invalid) {
		return huma.Error422UnprocessableEntity(invalid.Message)
	}
	if errors.Is(e, statistics.ErrInvalidWindow) {
		return huma.Error422UnprocessableEntity("Invalid history window")
	}
	if errors.Is(e, store.ErrNotFound) {
		return huma.Error404NotFound("Record not found")
	}
	if errors.Is(e, store.ErrConflict) {
		return huma.Error409Conflict("Configuration conflicts with current data")
	}
	telemetry.LogError(ctx, "api.persistence", e)
	return huma.Error500InternalServerError("Persistence operation failed")
}
func list[T any](ctx context.Context, s *store.Store, kind string) ([]T, error) {
	raw, e := s.List(ctx, kind)
	if e != nil {
		return nil, e
	}
	items := make([]T, 0, len(raw))
	for _, b := range raw {
		var v T
		if e = json.Unmarshal(b, &v); e != nil {
			return nil, e
		}
		items = append(items, v)
	}
	return items, nil
}
func audit(ctx context.Context, t *store.Tx, action, kind, id string) error {
	u := CurrentUser(ctx)
	a := domain.Audit{ID: domain.ID(), UserID: u.ID, Username: u.Username, Action: action, ResourceType: kind, ResourceID: id, CreatedAt: domain.Now()}
	return t.Put(ctx, "audit", a.ID, a)
}
func (s *Server) ResolveSecret(ctx context.Context, id string) (string, error) {
	return s.Secrets.ResolveSecret(ctx, id)
}

func (s *Server) spa(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeProblem(w, 404, "Route not found")
		return
	}
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		p := filepath.Join(s.Config.StaticDir, filepath.Clean(r.URL.Path))
		if strings.HasPrefix(r.URL.Path, "/assets/uploads/") {
			name := strings.TrimPrefix(r.URL.Path, "/assets/uploads/")
			if name == "" || strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
				writeProblem(w, 404, "Asset not found")
				return
			}
			p = filepath.Join(s.Config.DataDir, "uploads", name)
		}
		if info, e := os.Stat(p); e == nil && !info.IsDir() {
			w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
			http.ServeFile(w, r, p)
			return
		}
		writeProblem(w, 404, "Asset not found")
		return
	}
	if !s.adminHost(r.Host) {
		if s.Store == nil {
			writeProblem(w, 404, "Page not found")
			return
		}
		id, e := s.Store.PageIDByDomain(r.Context(), hostname(r.Host))
		if e != nil {
			writeProblem(w, 404, "Page not found")
			return
		}
		var p domain.Page
		if s.Store.Get(r.Context(), "pages", id, &p) != nil || p.Published == nil || !validPublicPath(r.URL.Path, "") {
			writeProblem(w, 404, "Page not found")
			return
		}
	} else if r.URL.Path != "/" && r.URL.Path != "/app" && !strings.HasPrefix(r.URL.Path, "/app/") {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		id, e := s.Store.PageIDBySlug(r.Context(), parts[0])
		var p domain.Page
		if e != nil || s.Store.Get(r.Context(), "pages", id, &p) != nil || p.Published == nil || !validPublicPath(r.URL.Path, "/"+parts[0]) {
			writeProblem(w, 404, "Page not found")
			return
		}
	}
	p := filepath.Join(s.Config.StaticDir, "index.html")
	if _, e := os.Stat(p); e != nil {
		writeProblem(w, 503, "Frontend build is unavailable; run mise run build:web")
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; font-src 'self'; frame-ancestors 'none'; base-uri 'self'; object-src 'none'")
	http.ServeFile(w, r, p)
}
func validPublicPath(path, base string) bool {
	if path == base || path == base+"/" {
		return true
	}
	rest := strings.TrimPrefix(path, base+"/")
	parts := strings.Split(rest, "/")
	return len(parts) == 2 && parts[0] == "incidents" && parts[1] != ""
}
func hostname(h string) string {
	if v, _, e := net.SplitHostPort(h); e == nil {
		h = v
	}
	return strings.ToLower(strings.TrimSuffix(h, "."))
}

// Keep time.Duration in this package for fixed cookie/session lifetime.
const sessionLifetime = time.Hour * 24 * 7

// Unwrap lets ResponseController retain support for the underlying writer.
type responseStatus struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *responseStatus) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *responseStatus) WriteHeader(status int) {
	if status >= 100 && status < 200 {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.wroteHeader {
		return
	}
	w.status, w.wroteHeader = status, true
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseStatus) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}
