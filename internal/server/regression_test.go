package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/engine"
	"github.com/octoplorer/octopulse/internal/store"
)

func TestHistoryFiltersBeforeLimitAndUsesRetainedLatency(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	m := createTestMonitor(
		t,
		c,
		ts.URL,
		csrf,
	)
	from := domain.Now() - 7*86400000
	to := from + 2000
	err := s.Store.WithTx(context.Background(), func(tx *store.Tx) error {
		for i := 0; i < 250; i++ {
			at := from + int64(i)*1000
			if err := tx.PutRound(
				context.Background(),
				store.Round{
					ID:            fmt.Sprintf("round-%03d", i),
					MonitorID:     m.ID,
					ConfigVersion: 1,
					Generation:    1,
					StartedAt:     at,
					FinishedAt:    at + 100,
					Success:       true,
					LatencyMS:     100,
				},
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	s.Latency = func(_ context.Context, id string, a, b int64) ([]domain.LatencyPoint, error) {
		called = id == m.ID && a == from && b == to
		return []domain.LatencyPoint{{At: from, LatencyMs: 321, Success: true}}, nil
	}
	status, b := request(
		t,
		c,
		"GET",
		fmt.Sprintf(
			"%s/api/v1/monitors/%s/history?from=%d&to=%d&limit=2",
			ts.URL,
			m.ID,
			from,
			to,
		),
		"",
		nil,
	)
	var history MonitorHistory
	if err := json.Unmarshal(b, &history); err != nil {
		t.Fatal(err)
	}
	if status != 200 || len(history.Rounds) != 2 || history.Rounds[0].ID != "round-001" || !called || len(
		history.Latency,
	) != 1 || history.Latency[0].LatencyMs != 321 {
		t.Fatalf("past window lost or retained latency ignored: %d %s", status, b)
	}
	if status, _ := request(
		t,
		c,
		"GET",
		ts.URL+"/api/v1/monitors/"+m.ID+"/history?from=-1&to=1000",
		"",
		nil,
	); status != 422 {
		t.Fatal("negative history accepted")
	}
}

func TestConfigurationSnapshotRejectsPreviousRuntime(t *testing.T) {
	s, ts, c := testServer(t)
	m := createTestMonitor(
		t,
		c,
		ts.URL,
		bootstrap(t, c, ts.URL),
	)
	m.ConfigVersion = 2
	m.State = domain.StateUnknown
	payload, _ := json.Marshal(m)
	err := s.Store.WithTx(context.Background(), func(tx *store.Tx) error {
		if err := tx.PutMonitor(
			context.Background(),
			store.Monitor{
				ID:            m.ID,
				ConfigVersion: 2,
				Generation:    2,
				Kind:          m.Type,
				Enabled:       true,
				IntervalMS:    60000,
				ConfigJSON:    payload,
			},
		); err != nil {
			return err
		}
		return tx.PutRuntime(
			context.Background(),
			store.Runtime{
				MonitorID:       m.ID,
				ConfigVersion:   1,
				Generation:      1,
				State:           domain.StateUp,
				LastCollectedAt: 999,
			},
		)
	})
	if err != nil {
		t.Fatal(err)
	}
	actual, err := s.monitor(context.Background(), m.ID)
	if err != nil || actual.State != domain.StateUnknown || actual.LastCheckedAt != 0 {
		t.Fatalf("old runtime leaked into new config: %+v %v", actual, err)
	}
	s.Check = func(context.Context, string) error {
		return fmt.Errorf("%w: %w", engine.ErrCheckAccepted, context.DeadlineExceeded)
	}
	if status, b := request(
		t,
		c,
		"POST",
		ts.URL+"/api/v1/monitors/"+m.ID+"/check",
		currentCSRF(t, c, ts.URL),
		nil,
	); status != 202 {
		t.Fatalf("accepted check reported failure: %d %s", status, b)
	}
}

func currentCSRF(t *testing.T, c *http.Client, base string) string {
	_, b := request(
		t,
		c,
		"GET",
		base+"/api/v1/session",
		"",
		nil,
	)
	var session SessionBody
	if err := json.Unmarshal(b, &session); err != nil {
		t.Fatal(err)
	}
	return session.CSRFToken
}

func TestMaintenanceEditsPreserveRetainedLatency(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	m := createTestMonitor(
		t,
		c,
		ts.URL,
		csrf,
	)
	at := (domain.Now() - 20*86400000) / 3600000 * 3600000
	ctx := context.Background()
	if err := s.Store.WithTx(ctx, func(tx *store.Tx) error {
		return tx.PutAggregate(
			ctx,
			store.Aggregate{
				MonitorID:            m.ID,
				BucketAt:             at,
				WidthMS:              3600000,
				LatencyTotalMS:       1234,
				RoundCount:           3,
				SuccessfulRoundCount: 2,
			},
		)
	}); err != nil {
		t.Fatal(err)
	}
	v := domain.Maintenance{
		Name:       "Old maintenance",
		MonitorIDs: []string{m.ID},
		PageIDs:    []string{},
		StartsAt:   at,
		EndsAt:     at + 3600000,
		Timezone:   "UTC",
	}
	status, b := request(
		t,
		c,
		"POST",
		ts.URL+"/api/v1/maintenance",
		csrf,
		v,
	)
	if status != 200 {
		t.Fatalf("maintenance create: %d %s", status, b)
	}
	json.Unmarshal(b, &v)
	v.EndsAt = at + 1800000
	if status, b = request(
		t,
		c,
		"PATCH",
		ts.URL+"/api/v1/maintenance/"+v.ID,
		csrf,
		v,
	); status != 200 {
		t.Fatalf("maintenance edit: %d %s", status, b)
	}
	if status, _ = request(
		t,
		c,
		"DELETE",
		ts.URL+"/api/v1/maintenance/"+v.ID,
		csrf,
		nil,
	); status != 200 {
		t.Fatal("maintenance delete")
	}
	a, err := s.Store.Aggregates(
		ctx,
		m.ID,
		at,
		at+3600000,
		3600000,
	)
	if err != nil || len(a) != 1 || a[0].LatencyTotalMS != 1234 || a[0].RoundCount != 3 {
		t.Fatalf("only retained latency deleted: %+v %v", a, err)
	}
}

func TestActiveIncidentsSurviveHistoryLimitAndConcurrentProgress(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	p := domain.Page{
		Name: "Status",
		Slug: "status",
		Draft: domain.PageConfig{
			Title:       "Status",
			BrandColor:  "#008877",
			ColorScheme: "system",
			Links:       []domain.Link{},
			Groups:      []domain.PageGroup{},
		},
	}
	status, b := request(
		t,
		c,
		"POST",
		ts.URL+"/api/v1/pages",
		csrf,
		p,
	)
	if status != 200 {
		t.Fatalf("page create: %d %s", status, b)
	}
	json.Unmarshal(b, &p)
	ctx := context.Background()
	active := domain.Incident{
		ID:        "active",
		Title:     "Active outage",
		Body:      "Investigating",
		Status:    "investigating",
		Impact:    "outage",
		PageIDs:   []string{p.ID},
		CreatedAt: 1,
	}
	if err := s.Store.WithTx(ctx, func(tx *store.Tx) error {
		if err := tx.Put(
			ctx,
			"incidents",
			active.ID,
			active,
		); err != nil {
			return err
		}
		for i := 0; i < 101; i++ {
			v := domain.Incident{
				ID:         fmt.Sprintf("resolved-%d", i),
				Title:      "Resolved",
				Status:     "resolved",
				PageIDs:    []string{p.ID},
				CreatedAt:  int64(i + 2),
				ResolvedAt: domain.Now(),
			}
			if err := tx.Put(
				ctx,
				"incidents",
				v.ID,
				v,
			); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	public, err := s.projectPage(ctx, p, p.Draft)
	if err != nil || public.State != "outage" || public.Incidents[len(public.Incidents)-1].ID != "active" {
		t.Fatalf("old active outage hidden: %+v %v", public, err)
	}
	var wg sync.WaitGroup
	errors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body, _ := json.Marshal(IncidentProgress{Body: fmt.Sprintf("progress %d", i), Status: "identified"})
			req, _ := http.NewRequest("POST", ts.URL+"/api/v1/incidents/active/updates", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", csrf)
			res, err := c.Do(req)
			if err != nil {
				errors <- err
				return
			}
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			if res.StatusCode != 200 {
				errors <- fmt.Errorf("progress status %d", res.StatusCode)
			}
		}(i)
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
	if err := s.Store.Get(
		ctx,
		"incidents",
		"active",
		&active,
	); err != nil || len(active.Updates) != 2 {
		t.Fatalf("concurrent progress lost: %+v %v", active, err)
	}
}

func TestAdminPasswordResetRevokesExistingSessions(t *testing.T) {
	_, ts, admin := testServer(t)
	csrf := bootstrap(t, admin, ts.URL)
	v := UserWrite{
		Username: "operator",
		Name:     "Operator",
		Role:     domain.RoleOperator,
		Locale:   "en",
		Timezone: "UTC",
		Password: "old passphrase 1234",
	}
	status, b := request(
		t,
		admin,
		"POST",
		ts.URL+"/api/v1/users",
		csrf,
		v,
	)
	if status != 200 {
		t.Fatalf("user create: %d %s", status, b)
	}
	var u domain.User
	json.Unmarshal(b, &u)
	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	if status, _ = request(
		t,
		other,
		"POST",
		ts.URL+"/api/v1/session",
		"",
		map[string]any{
			"username": "operator",
			"password": v.Password,
		},
	); status != 200 {
		t.Fatal("operator login")
	}
	v.Password = "new passphrase 1234"
	if status, b = request(
		t,
		admin,
		"PATCH",
		ts.URL+"/api/v1/users/"+u.ID,
		csrf,
		v,
	); status != 200 {
		t.Fatalf("password reset: %d %s", status, b)
	}
	if status, _ = request(
		t,
		other,
		"GET",
		ts.URL+"/api/v1/monitors",
		"",
		nil,
	); status != 401 {
		t.Fatal("reset kept old session alive")
	}
}

func TestDomainAndLogoConfigurationMatchDispatch(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	settings := domain.DefaultSettings()
	settings.AllowedDomains = []string{"localhost"}
	if status, _ := request(
		t,
		c,
		"PATCH",
		ts.URL+"/api/v1/settings",
		csrf,
		settings,
	); status != 422 {
		t.Fatal("management domain offered as public domain")
	}
	p := domain.Page{
		Name: "Status",
		Slug: "status",
		Draft: domain.PageConfig{
			Title:       "Status",
			BrandColor:  "#008877",
			ColorScheme: "system",
			LogoURL:     "http://images.example.test/logo.png",
		},
	}
	if status, _ := request(
		t,
		c,
		"POST",
		ts.URL+"/api/v1/pages",
		csrf,
		p,
	); status != 422 {
		t.Fatal("CSP-blocked image allowed")
	}
	if status, _ := request(
		t,
		c,
		"GET",
		ts.URL+"/healthz",
		"",
		nil,
	); status != 200 {
		t.Fatal("healthy database failed")
	}
	s.Store.Close()
	if status, _ := request(
		t,
		c,
		"GET",
		ts.URL+"/healthz",
		"",
		nil,
	); status != 503 {
		t.Fatal("closed database reported ready")
	}
}

func TestPageRoutingChangesOnlyOnPublication(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	settings := domain.DefaultSettings()
	settings.AllowedDomains = []string{"status.test", "newstatus.test"}
	if status, b := request(
		t,
		c,
		"PATCH",
		ts.URL+"/api/v1/settings",
		csrf,
		settings,
	); status != 200 {
		t.Fatalf("domains: %d %s", status, b)
	}
	p := domain.Page{
		Name:   "Status",
		Slug:   "status1",
		Domain: "status.test",
		Draft: domain.PageConfig{
			Title:       "Published",
			BrandColor:  "#008877",
			ColorScheme: "system",
			Links:       []domain.Link{},
			Groups:      []domain.PageGroup{},
		},
	}
	status, b := request(
		t,
		c,
		"POST",
		ts.URL+"/api/v1/pages",
		csrf,
		p,
	)
	if status != 200 {
		t.Fatalf("create: %d %s", status, b)
	}
	json.Unmarshal(b, &p)
	status, b = request(
		t,
		c,
		"POST",
		ts.URL+"/api/v1/pages/"+p.ID+"/publish",
		csrf,
		nil,
	)
	if status != 200 {
		t.Fatalf("publish: %d %s", status, b)
	}
	json.Unmarshal(b, &p)
	p.Slug = "status2"
	p.Domain = "newstatus.test"
	p.Draft.Title = "New draft"
	status, b = request(
		t,
		c,
		"PATCH",
		ts.URL+"/api/v1/pages/"+p.ID,
		csrf,
		p,
	)
	if status != 200 {
		t.Fatalf("draft: %d %s", status, b)
	}
	status, b = request(
		t,
		c,
		"GET",
		ts.URL+"/api/public/pages/status1",
		"",
		nil,
	)
	if status != 200 || !bytes.Contains(b, []byte(`"slug":"status1"`)) || bytes.Contains(b, []byte("New draft")) {
		t.Fatalf("draft replaced published routing: %d %s", status, b)
	}
	if status, _ = request(
		t,
		c,
		"GET",
		ts.URL+"/api/public/pages/status2",
		"",
		nil,
	); status != 404 {
		t.Fatal("draft slug exposed")
	}
	if id, err := s.Store.PageIDByDomain(context.Background(), "status.test"); err != nil || id != p.ID {
		t.Fatal("draft domain replaced published binding")
	}
	status, b = request(
		t,
		c,
		"POST",
		ts.URL+"/api/v1/pages/"+p.ID+"/publish",
		csrf,
		nil,
	)
	if status != 200 {
		t.Fatalf("publish changed route: %d %s", status, b)
	}
	if status, _ = request(
		t,
		c,
		"GET",
		ts.URL+"/api/public/pages/status1",
		"",
		nil,
	); status != 404 {
		t.Fatal("old slug retained after publication")
	}
	if status, b = request(
		t,
		c,
		"GET",
		ts.URL+"/api/public/pages/status2",
		"",
		nil,
	); status != 200 || !bytes.Contains(b, []byte("New draft")) {
		t.Fatalf("new publication missing: %d %s", status, b)
	}
	if _, err := s.Store.PageIDByDomain(context.Background(), "status.test"); err != store.ErrNotFound {
		t.Fatal("old domain retained after publication")
	}
}
