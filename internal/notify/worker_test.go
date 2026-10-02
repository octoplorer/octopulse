package notify

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nicholas-fedor/shoutrrr/pkg/router"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type secretMap map[string]string

func (s secretMap) ResolveSecret(_ context.Context, id string) (string, error) {
	if raw, ok := s[id]; ok {
		return raw, nil
	}
	return "", errors.New("provider-token-must-not-leak")
}

type fixture struct {
	s     *store.Store
	w     *Worker
	clock atomic.Int64
}

func setup(t *testing.T, raw string) *fixture {
	t.Helper()
	ctx := context.Background()
	cfg := store.Config{Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "notifications.db")}
	if dsn := os.Getenv("OCTOPULSE_NOTIFY_TEST_POSTGRES_DSN"); dsn != "" {
		admin, err := sql.Open("pgx", dsn)
		if err != nil {
			t.Fatal(err)
		}
		schema := "notify_" + domain.ID()
		if _, err = admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); admin.Close() })
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		q := u.Query()
		q.Set("search_path", schema)
		u.RawQuery = q.Encode()
		cfg = store.Config{Driver: "postgres", DSN: u.String()}
	}
	s, err := store.Open(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	f := &fixture{s: s}
	f.clock.Store(time.Now().UnixMilli())
	f.w = New(s, secretMap{"webhook": raw})
	f.w.Now = func() time.Time { return time.UnixMilli(f.clock.Load()) }
	err = s.WithTx(ctx, func(tx *store.Tx) error {
		if err := tx.PutMonitor(ctx, store.Monitor{ID: "monitor", ConfigVersion: 1, Generation: 1, Kind: "http", Enabled: true, IntervalMS: 30000}); err != nil {
			return err
		}
		if err := tx.PutRuntime(ctx, store.Runtime{MonitorID: "monitor", ConfigVersion: 1, Generation: 1, State: domain.StateDown}); err != nil {
			return err
		}
		return tx.Put(ctx, "channels", "channel", domain.Channel{ID: "channel", Name: "webhook", ServiceURLSecretID: "webhook", Enabled: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func (f *fixture) job(t *testing.T, id, kind string, generation int64, cycle string) {
	t.Helper()
	ctx := context.Background()
	p := domain.NotificationPayload{MonitorID: "monitor", Name: "website", Kind: kind, Generation: generation, CycleID: cycle, CreatedAt: f.clock.Load(), Message: "website " + kind}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	err = f.s.WithTx(ctx, func(tx *store.Tx) error {
		if err := tx.PutEvent(ctx, store.Event{ID: id + "-event", MonitorID: "monitor", Generation: generation, Kind: kind, CreatedAt: f.clock.Load(), Payload: raw}); err != nil {
			return err
		}
		return tx.PutDelivery(ctx, store.Delivery{ID: id, EventID: id + "-event", ChannelID: "channel", Generation: generation, DueAt: f.clock.Load(), Payload: raw})
	})
	if err != nil {
		t.Fatal(err)
	}
}
func (f *fixture) state(t *testing.T, state string, generation int64) {
	t.Helper()
	err := f.s.WithTx(context.Background(), func(tx *store.Tx) error {
		return tx.PutRuntime(context.Background(), store.Runtime{MonitorID: "monitor", ConfigVersion: 1, Generation: generation, State: state})
	})
	if err != nil {
		t.Fatal(err)
	}
}
func genericURL(server string) string {
	return "generic://" + strings.TrimPrefix(server, "http://") + "/hook?disabletls=yes&template=json"
}

func TestActualWebhookDeliveryAndRecoveryMarkers(t *testing.T) {
	var received atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "website") {
			t.Error("missing message", string(body))
		}
		received.Add(1)
		w.WriteHeader(204)
	}))
	defer srv.Close()
	f := setup(t, genericURL(srv.URL))
	ctx := context.Background()
	f.job(t, "down", "down", 1, "outage-cycle")
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	var marker domain.DeliveryMarker
	if err := f.s.Get(ctx, "deliveryMarkers", domain.DeliveryMarkerID("monitor", "channel", "outage-cycle"), &marker); err != nil || marker.DownSentAt == 0 {
		t.Fatal(marker, err)
	}
	f.state(t, domain.StateUp, 2)
	f.job(t, "up", "up", 2, "outage-cycle")
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if err := f.s.Get(ctx, "deliveryMarkers", domain.DeliveryMarkerID("monitor", "channel", "outage-cycle"), &marker); err != nil || marker.RecoverySentAt == 0 {
		t.Fatal(marker, err)
	}
	f.job(t, "duplicate-up", "up", 2, "outage-cycle")
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	d, err := f.s.GetDelivery(ctx, "duplicate-up")
	if err != nil || d.State != "cancelled" {
		t.Fatal(d, err)
	}
	if received.Load() != 2 {
		t.Fatal("unexpected sends", received.Load())
	}
}

func TestProviderFailureRetryAndRedaction(t *testing.T) {
	var received atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Add(1)
		w.WriteHeader(503)
		w.Write([]byte("private-provider-token"))
	}))
	defer srv.Close()
	f := setup(t, genericURL(srv.URL)+"&@Authorization=Bearer%20secret-token")
	f.w.MaxAttempts = 2
	f.job(t, "retry", "down", 1, "cycle")
	if err := f.w.DeliverOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	d, err := f.s.GetDelivery(context.Background(), "retry")
	if err != nil || d.State != "pending" || d.Attempts != 1 || d.DueAt <= f.clock.Load() {
		t.Fatal(d, err)
	}
	if strings.Contains(d.LastError, "private") || strings.Contains(d.LastError, "secret") || strings.Contains(d.LastError, srv.URL) {
		t.Fatal("diagnostic leaked", d.LastError)
	}
	if err := f.w.DeliverOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if received.Load() != 1 {
		t.Fatal("retried before due")
	}
	f.clock.Store(d.DueAt + 1)
	if err := f.w.DeliverOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	d, err = f.s.GetDelivery(context.Background(), "retry")
	if err != nil || d.State != "failed" || d.Attempts != 2 {
		t.Fatal(d, err)
	}
	if received.Load() != 2 {
		t.Fatal(received.Load())
	}
}

func TestStaleMaintenanceAndUnsentRecoveryCancellation(t *testing.T) {
	var received atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received.Add(1); w.WriteHeader(204) }))
	defer srv.Close()
	f := setup(t, genericURL(srv.URL))
	ctx := context.Background()
	f.job(t, "old-generation", "down", 0, "cycle")
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	f.state(t, domain.StateUp, 2)
	f.job(t, "unsent-up", "up", 2, "cycle")
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	f.state(t, domain.StateDown, 3)
	f.job(t, "maintenance", "down", 3, "cycle")
	if err := f.s.Put(ctx, "maintenance", "window", domain.Maintenance{ID: "window", MonitorIDs: []string{"monitor"}, StartsAt: f.clock.Load() - 1000, EndsAt: f.clock.Load() + 10000}); err != nil {
		t.Fatal(err)
	}
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"old-generation", "unsent-up", "maintenance"} {
		d, err := f.s.GetDelivery(ctx, id)
		if err != nil || d.State != "cancelled" {
			t.Fatal(id, d, err)
		}
	}
	if received.Load() != 0 {
		t.Fatal("stale notifications sent", received.Load())
	}
}

func TestAbandonedLeaseRecoverySendsAndRejectsOldToken(t *testing.T) {
	var received atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received.Add(1); w.WriteHeader(204) }))
	defer srv.Close()
	f := setup(t, genericURL(srv.URL))
	ctx := context.Background()
	f.job(t, "abandoned", "down", 1, "cycle")
	claimed, err := f.s.ClaimDeliveries(ctx, f.clock.Load(), 1000, 1, "dead-process")
	if err != nil || len(claimed) != 1 {
		t.Fatal(claimed, err)
	}
	f.clock.Add(1001)
	if err = f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	d, err := f.s.GetDelivery(ctx, "abandoned")
	if err != nil || d.State != "sent" || d.Attempts != 2 || received.Load() != 1 {
		t.Fatal(d, err, received.Load())
	}
	if err = f.s.CompleteDelivery(ctx, d.ID, "dead-process", f.clock.Load(), "sent", 0, ""); !errors.Is(err, store.ErrLeaseLost) {
		t.Fatal(err)
	}
}

func TestExplicitChannelTestAndDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	}))
	defer srv.Close()
	f := setup(t, genericURL(srv.URL))
	f.w.SendTimeout = 50 * time.Millisecond
	started := time.Now()
	err := f.w.TestChannel(context.Background(), "channel")
	if !errors.Is(err, ErrSendTimeout) {
		t.Fatal(err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("send timeout did not bound network wait")
	}
	if err = ValidateURL("unknown://super-secret-token"); !errors.Is(err, ErrInvalidURL) || strings.Contains(err.Error(), "super-secret") {
		t.Fatal(err)
	}
}

func TestCertificateFingerprintAndGeneration(t *testing.T) {
	var received atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received.Add(1); w.WriteHeader(204) }))
	defer srv.Close()
	f := setup(t, genericURL(srv.URL))
	ctx := context.Background()
	if err := f.s.Put(ctx, "maintenance", "certificate-independent", domain.Maintenance{ID: "certificate-independent", MonitorIDs: []string{"monitor"}, StartsAt: f.clock.Load() - 1000, EndsAt: f.clock.Load() + 10000}); err != nil {
		t.Fatal(err)
	}
	if err := f.s.Put(ctx, "engineMonitor", "monitor", map[string]any{"certificate": map[string]any{"fingerprint": "current-cert", "state": "expiring"}}); err != nil {
		t.Fatal(err)
	}
	f.job(t, "old-certificate", "certificate_threshold", 1, "old-cert")
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	f.job(t, "new-certificate", "certificate_threshold", 1, "current-cert")
	if err := f.w.DeliverOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if received.Load() != 1 {
		t.Fatal(received.Load())
	}
}

func TestInstalledProvidersAcceptBoundedTransportOrContext(t *testing.T) {
	r := router.ServiceRouter{}
	for _, scheme := range r.ListServices() {
		service, err := r.NewService(scheme)
		if err != nil {
			t.Fatal(scheme, err)
		}
		if !boundedSupport(service) {
			t.Errorf("%s cannot receive a request context or bounded network transport", scheme)
		}
	}
}

func TestRecoveryWaitsForConcurrentFaultSend(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "fault_sent", true: "fault_never_sent"}[fail], func(t *testing.T) {
			var count atomic.Int64
			started := make(chan struct{})
			release := make(chan struct{})
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if count.Add(1) == 1 {
					close(started)
					<-release
					if fail {
						w.WriteHeader(503)
						return
					}
				}
				w.WriteHeader(204)
			}))
			defer srv.Close()
			f := setup(t, genericURL(srv.URL))
			ctx := context.Background()
			f.job(t, "concurrent-down", "down", 1, "cycle")
			completed := make(chan error, 1)
			go func() { completed <- f.w.DeliverOnce(ctx) }()
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("fault send did not start")
			}
			f.state(t, domain.StateUp, 2)
			f.job(t, "concurrent-up", "up", 2, "cycle")
			if err := f.w.DeliverOnce(ctx); err != nil {
				t.Fatal(err)
			}
			up, err := f.s.GetDelivery(ctx, "concurrent-up")
			if err != nil || up.State != "pending" {
				t.Fatal("recovery lost while fault send in progress", up, err)
			}
			close(release)
			if err = <-completed; err != nil {
				t.Fatal(err)
			}
			f.clock.Store(up.DueAt + 1)
			if err = f.w.DeliverOnce(ctx); err != nil {
				t.Fatal(err)
			}
			if fail {
				if err = f.w.DeliverOnce(ctx); err != nil {
					t.Fatal(err)
				}
			}
			up, err = f.s.GetDelivery(ctx, "concurrent-up")
			expected := "sent"
			sends := int64(2)
			if fail {
				expected = "cancelled"
				sends = 1
			}
			if err != nil || up.State != expected || count.Load() != sends {
				t.Fatal(up, err, count.Load())
			}
		})
	}
}
