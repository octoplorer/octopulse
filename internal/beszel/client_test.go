package beszel

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type testSecret string

func (s testSecret) ResolveSecret(context.Context, string) (string, error) { return string(s), nil }

func adapter(t *testing.T, address, email, password string) *Client {
	t.Helper()
	s, err := store.Open(
		context.Background(),
		store.Config{
			Driver: "sqlite",
			DSN:    filepath.Join(t.TempDir(), "beszel.db"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	cfg := domain.BeszelConfig{
		URL:              address,
		Email:            email,
		PasswordSecretID: "password-ref",
		Enabled:          true,
		PollSeconds:      30,
	}
	if err = s.Put(
		context.Background(),
		"beszel",
		"config",
		cfg,
	); err != nil {
		t.Fatal(err)
	}
	return New(s, testSecret(password))
}
func jwt(expires int64) string {
	payload, _ := json.Marshal(map[string]int64{"exp": expires})
	return "header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}

func TestPocketBaseCompatibleAuthRefreshProjectionAndStaleCache(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	var clock atomic.Int64
	clock.Store(now.UnixMilli())
	var authCalls, refreshCalls atomic.Int64
	var unavailable atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if unavailable.Load() {
			w.WriteHeader(503)
			w.Write([]byte(`{"message":"super-private-password"}`))
			return
		}
		requiresToken := r.URL.Path != "/hub/api/beszel/info" && !strings.Contains(r.URL.Path, "auth-with-password")
		if requiresToken && r.Header.Get("Authorization") == "" {
			t.Error("missing authorization token")
		}
		switch r.URL.Path {
		case "/hub/api/beszel/info":
			json.NewEncoder(w).Encode(map[string]any{"v": "0.20.0", "key": "public-key"})
		case "/hub/api/collections/users/auth-with-password":
			var credentials map[string]string
			json.NewDecoder(r.Body).Decode(&credentials)
			if credentials["identity"] != "reader@example.invalid" || credentials["password"] != "super-private-password" {
				t.Error(credentials)
			}
			authCalls.Add(1)
			json.NewEncoder(w).Encode(
				map[string]any{
					"token": jwt(time.UnixMilli(clock.Load()).Add(6 * time.Minute).Unix()),
					"record": map[string]any{
						"id":             "user",
						"collectionName": "users",
						"role":           "readonly",
					},
				},
			)
		case "/hub/api/collections/users/auth-refresh":
			refreshCalls.Add(1)
			json.NewEncoder(w).Encode(
				map[string]any{
					"token": jwt(time.UnixMilli(clock.Load()).Add(30 * time.Minute).Unix()),
					"record": map[string]any{
						"id":             "user",
						"collectionName": "users",
					},
				},
			)
		case "/hub/api/collections/systems/records":
			json.NewEncoder(w).Encode(
				map[string]any{
					"items": []any{
						map[string]any{
							"id":      "system1",
							"name":    "server",
							"host":    "server.internal",
							"status":  "up",
							"updated": time.UnixMilli(clock.Load()).UTC().Format(time.RFC3339Nano),
							"info": map[string]any{
								"cpu": 23.5,
								"mp":  42.0,
								"dp":  55.0,
								"h":   "host",
								"m":   "CPU",
								"c":   4,
								"t":   8,
								"u":   1234,
								"v":   "0.20.0",
							},
						},
					},
					"totalPages": 1,
				},
			)
		case "/hub/api/collections/system_stats/records":
			filter := r.URL.Query().Get("filter")
			if !strings.Contains(filter, `system="system1"`) || !strings.Contains(filter, `type="20m"`) {
				t.Error("wrong record filter", filter)
			}
			json.NewEncoder(w).Encode(
				map[string]any{
					"items": []any{
						map[string]any{
							"created": now.Format(time.RFC3339Nano),
							"stats": map[string]any{
								"cpu": 23.5,
								"mp":  42.0,
								"dp":  55.0,
								"b":   []int{1024, 2048},
							},
						},
						map[string]any{
							"created": now.Add(time.Minute).UnixMilli(),
							"stats": map[string]any{
								"cpu": 2,
								"mp":  3,
								"dp":  4,
								"ns":  1,
								"nr":  2,
							},
						},
					},
					"totalPages": 1,
				},
			)
		case "/hub/api/collections/containers/records":
			json.NewEncoder(w).Encode(
				map[string]any{
					"items": []any{
						map[string]any{
							"id":      "container1",
							"name":    "web",
							"image":   "image:v1",
							"status":  "running",
							"cpu":     4.5,
							"memory":  120,
							"updated": clock.Load(),
						},
					},
					"totalPages": 1,
				},
			)
		default:
			t.Error("unexpected endpoint", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	c := adapter(
		t,
		srv.URL+"/hub",
		"reader@example.invalid",
		"super-private-password",
	)
	c.Now = func() time.Time { return time.UnixMilli(clock.Load()) }
	// The integration must leave website monitoring state untouched.
	if err := c.Store.WithTx(ctx, func(tx *store.Tx) error {
		if err := tx.PutMonitor(
			ctx,
			store.Monitor{
				ID:            "website",
				ConfigVersion: 1,
				Generation:    1,
				Kind:          "http",
				Enabled:       true,
				IntervalMS:    30000,
			},
		); err != nil {
			return err
		}
		return tx.PutRuntime(ctx, store.Runtime{MonitorID: "website", ConfigVersion: 1, Generation: 1, State: domain.StateUp})
	}); err != nil {
		t.Fatal(err)
	}
	if err := c.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	systems, err := c.Systems(ctx)
	if err != nil || len(systems.Items) != 1 {
		t.Fatal(systems, err)
	}
	metricsMatch := systems.Items[0].CPU == 23.5 && systems.Items[0].Memory == 42
	metadataMatches := systems.Items[0].Info.Cores == 4 && !systems.Stale
	if !metricsMatch || !metadataMatches {
		t.Fatal(systems, err)
	}
	history, err := c.History(ctx, "system1", "24h")
	if err != nil || len(history.Items) != 2 {
		t.Fatal(history, err)
	}
	firstNetworkMetricsMatch := history.Items[0].NetworkIn == 2048 && history.Items[0].NetworkOut == 1024
	secondNetworkMetricMatches := history.Items[1].NetworkIn == 2*1_048_576
	windowMatches := secondNetworkMetricMatches && history.IntervalMS == 1200000
	if !firstNetworkMetricsMatch || !windowMatches {
		t.Fatal(history, err)
	}
	containers, err := c.Containers(ctx, "system1")
	if err != nil || len(containers.Items) != 1 {
		t.Fatal(containers, err)
	}
	if containers.Items[0].Memory != 120 {
		t.Fatal(containers, err)
	}
	clock.Add((2 * time.Minute).Milliseconds())
	if err = c.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	if authCalls.Load() != 1 || refreshCalls.Load() != 1 {
		t.Fatal("auth token refresh was not used", authCalls.Load(), refreshCalls.Load())
	}
	unavailable.Store(true)
	clock.Add((time.Minute).Milliseconds())
	if err = c.Poll(ctx); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	systems, err = c.Systems(ctx)
	if err != nil || len(systems.Items) != 1 {
		t.Fatal(systems, err)
	}
	hasStaleSnapshot := systems.Stale && systems.Items[0].Stale
	if !hasStaleSnapshot || systems.Error == "" {
		t.Fatal(systems, err)
	}
	history, err = c.History(ctx, "system1", "24h")
	if err != nil || len(history.Items) != 2 {
		t.Fatal(history, err)
	}
	if !history.Stale {
		t.Fatal(history, err)
	}
	raw, _ := json.Marshal(systems)
	encodedSnapshot := string(raw)
	hasSecret := strings.Contains(encodedSnapshot, "super-private")
	hasSignature := strings.Contains(encodedSnapshot, "signature")
	hasEmail := strings.Contains(encodedSnapshot, "reader@example")
	hasCredentials := hasSecret || hasEmail
	if hasCredentials || hasSignature {
		t.Fatal("credential material leaked", string(raw))
	}
	runtime, err := c.Store.GetRuntime(ctx, "website")
	if err != nil {
		t.Fatal("Beszel modified uptime state", runtime, err)
	}
	if runtime.State != domain.StateUp || runtime.Generation != 1 {
		t.Fatal("Beszel modified uptime state", runtime, err)
	}
}

func TestMFAAndVersionFailuresAreExplicit(t *testing.T) {
	for _, kind := range []string{"mfa", "version", "permission"} {
		t.Run(kind, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/beszel/info" {
					version := "0.20.0"
					if kind == "version" {
						version = "0.21.0"
					}
					json.NewEncoder(w).Encode(map[string]string{"v": version})
					return
				}
				if kind == "permission" && strings.Contains(r.URL.Path, "records") {
					w.WriteHeader(403)
					return
				}
				if kind == "mfa" {
					w.WriteHeader(401)
					w.Write([]byte(`{"mfaId":"challenge","password":"should-not-leak"}`))
					return
				}
				json.NewEncoder(w).Encode(
					map[string]any{
						"token": jwt(time.Now().Add(time.Hour).Unix()),
						"record": map[string]string{
							"collectionName": "users",
						},
					},
				)
			}))
			defer srv.Close()
			c := adapter(
				t,
				srv.URL,
				"reader@example.invalid",
				"password",
			)
			err := c.Poll(context.Background())
			expected := map[string]error{"mfa": ErrAuth, "version": ErrVersion, "permission": ErrPermission}[kind]
			if !errors.Is(err, expected) {
				t.Fatal(err)
			}
			result, _ := c.Systems(context.Background())
			hasStaleError := result.Stale && result.Error != ""
			if !hasStaleError || strings.Contains(result.Error, "should-not-leak") {
				t.Fatal(result)
			}
		})
	}
}

func TestActualBeszelHubReadonlyIntegration(t *testing.T) {
	path := os.Getenv("OCTOPULSE_TEST_BESZEL_FIXTURE")
	if path == "" {
		t.Skip("set fixture path for an isolated real Beszel 0.20 Hub")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		URL      string `json:"url"`
		Email    string `json:"email"`
		Password string `json:"password"`
		System   string `json:"system"`
	}
	if err = json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	c := adapter(
		t,
		cfg.URL,
		cfg.Email,
		cfg.Password,
	)
	ctx := context.Background()
	if err = c.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	systems, err := c.Systems(ctx)
	if err != nil || len(systems.Items) != 1 {
		t.Fatal(systems, err)
	}
	if systems.Items[0].ID != cfg.System || systems.Items[0].Name != "Adapter integration fixture" {
		t.Fatal(systems, err)
	}
	history, err := c.History(ctx, cfg.System, "24h")
	if err != nil || len(history.Items) != 1 {
		t.Fatal(history, err)
	}
	if history.Items[0].NetworkIn != 2048 {
		t.Fatal(history, err)
	}
	containers, err := c.Containers(ctx, cfg.System)
	if err != nil || len(containers.Items) != 1 {
		t.Fatal(containers, err)
	}
	if containers.Items[0].Name != "fixture-container" {
		t.Fatal(containers, err)
	}
	c.session.mu.Lock()
	c.session.tokenUntil = time.Now().Unix()
	c.session.mu.Unlock()
	if err = c.Poll(ctx); err != nil {
		t.Fatal("real token refresh", err)
	}
}
