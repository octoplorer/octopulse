package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

func TestPublicPageDailyAvailabilityProjection(t *testing.T) {
	s, ts, client := testServer(t)
	csrf := bootstrap(t, client, ts.URL)
	first := createTestMonitor(t, client, ts.URL, csrf)
	second := createTestMonitor(t, client, ts.URL, csrf)
	const dayMS int64 = 86400000
	calls := 0
	s.DailyStatsBatch = func(_ context.Context, ids []string, from, to int64) (map[string][]domain.Availability, error) {
		calls++
		if len(ids) != 1 || ids[0] != first.ID {
			t.Fatalf("unexpected daily monitor IDs: %v", ids)
		}
		if from%dayMS != 0 || from != to/dayMS*dayMS-89*dayMS {
			t.Fatalf("expected 90 UTC days including today, got [%d,%d)", from, to)
		}
		return map[string][]domain.Availability{first.ID: {{From: from, To: from + dayMS, UnknownMs: dayMS}}}, nil
	}
	config := domain.PageConfig{Groups: []domain.PageGroup{
		{ID: "a", Monitors: []domain.PageMonitor{{MonitorID: first.ID, Alias: "Visible", ShowUptime: true}, {MonitorID: second.ID, ShowLatency: true}}},
		{ID: "b", Monitors: []domain.PageMonitor{{MonitorID: first.ID, Alias: "Hidden"}, {MonitorID: first.ID, Alias: "Repeated", ShowUptime: true}}},
	}}
	page, err := s.projectPage(context.Background(), domain.Page{ID: "daily-page"}, config)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("daily batch calls=%d", calls)
	}
	visible, hidden, repeated := page.Groups[0].Monitors[0], page.Groups[1].Monitors[0], page.Groups[1].Monitors[1]
	if visible.Name != "Visible" || len(visible.DailyAvailability) != 1 || visible.DailyAvailability[0].Uptime != nil {
		t.Fatalf("visible daily projection=%+v", visible)
	}
	if hidden.Name != "Hidden" || hidden.DailyAvailability == nil || len(hidden.DailyAvailability) != 0 || len(page.Groups[0].Monitors[1].DailyAvailability) != 0 {
		t.Fatalf("hidden daily history exposed: %+v", page.Groups)
	}
	if repeated.Name != "Repeated" || len(repeated.DailyAvailability) != 1 {
		t.Fatalf("repeated daily projection=%+v", repeated)
	}
	s.DailyStatsBatch = func(context.Context, []string, int64, int64) (map[string][]domain.Availability, error) {
		return map[string][]domain.Availability{}, nil
	}
	if _, err := s.projectPage(context.Background(), domain.Page{}, config); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("missing daily result must fail: %v", err)
	}
}

func TestPublicCertificateHasNoDailyAvailability(t *testing.T) {
	s, ts, client := testServer(t)
	csrf := bootstrap(t, client, ts.URL)
	status, body := request(t, client, "POST", ts.URL+"/api/v1/monitors", csrf, map[string]any{
		"name": "Certificate", "type": "certificate", "certificate": map[string]any{"host": "example.test", "port": 443},
	})
	if status != 200 {
		t.Fatalf("certificate create: %d %s", status, body)
	}
	var monitor domain.Monitor
	if err := json.Unmarshal(body, &monitor); err != nil {
		t.Fatal(err)
	}
	s.DailyStatsBatch = func(context.Context, []string, int64, int64) (map[string][]domain.Availability, error) {
		t.Fatal("certificate requested daily availability")
		return nil, nil
	}
	config := domain.PageConfig{Groups: []domain.PageGroup{{ID: "a", Monitors: []domain.PageMonitor{{MonitorID: monitor.ID, ShowUptime: true}}}}}
	page, err := s.projectPage(context.Background(), domain.Page{}, config)
	if err != nil {
		t.Fatal(err)
	}
	days := page.Groups[0].Monitors[0].DailyAvailability
	if days == nil || len(days) != 0 {
		t.Fatalf("certificate daily availability must be empty: %+v", days)
	}
}

func TestPublicPageDailyAvailabilityFallback(t *testing.T) {
	s, ts, client := testServer(t)
	csrf := bootstrap(t, client, ts.URL)
	m := createTestMonitor(t, client, ts.URL, csrf)
	config := domain.PageConfig{Groups: []domain.PageGroup{{ID: "a", Monitors: []domain.PageMonitor{{MonitorID: m.ID, ShowUptime: true}}}}}
	page, err := s.projectPage(context.Background(), domain.Page{}, config)
	if err != nil {
		t.Fatal(err)
	}
	days := page.Groups[0].Monitors[0].DailyAvailability
	if len(days) != 90 || days[89].To != page.UpdatedAt {
		t.Fatalf("expected 90 days through update time: %+v", days)
	}
	for i, day := range days {
		if day.From%86400000 != 0 || day.Uptime != nil || day.UnknownMs != day.To-day.From {
			t.Fatalf("unobserved day %d was misrepresented: %+v", i, day)
		}
		if i > 0 && days[i-1].To != day.From {
			t.Fatalf("daily slots are not consecutive: %+v", days)
		}
	}
}

func TestAuthenticatedPreviewPreloadsHiddenDailyHistory(t *testing.T) {
	s, ts, client := testServer(t)
	csrf := bootstrap(t, client, ts.URL)
	m := createTestMonitor(t, client, ts.URL, csrf)
	p := domain.Page{ID: "preview-history", Draft: domain.PageConfig{Groups: []domain.PageGroup{{ID: "a", Monitors: []domain.PageMonitor{{MonitorID: m.ID, ShowUptime: false}}}}}}
	if err := s.Store.WithTx(context.Background(), func(tx *store.Tx) error { return tx.Put(context.Background(), "pages", p.ID, p) }); err != nil {
		t.Fatal(err)
	}
	status, body := request(t, client, "GET", ts.URL+"/api/v1/pages/"+p.ID+"/preview", "", nil)
	if status != 200 {
		t.Fatalf("preview: %d %s", status, body)
	}
	var preview domain.PublicPage
	if err := json.Unmarshal(body, &preview); err != nil {
		t.Fatal(err)
	}
	if len(preview.Groups[0].Monitors[0].DailyAvailability) != 90 || preview.Config.Groups[0].Monitors[0].ShowUptime {
		t.Fatalf("preview must preload history without changing display settings: %+v", preview)
	}
	public, err := s.projectPage(context.Background(), p, p.Draft)
	if err != nil || len(public.Groups[0].Monitors[0].DailyAvailability) != 0 {
		t.Fatalf("public response must keep hidden history private: %+v %v", public, err)
	}
}
