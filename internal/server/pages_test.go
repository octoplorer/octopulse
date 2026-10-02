package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

func createTestMonitor(t *testing.T, c *http.Client, base, csrf string) domain.Monitor {
	t.Helper()
	status, b := request(t, c, "POST", base+"/api/v1/monitors", csrf, map[string]any{"name": "Private internal name", "type": "http", "http": map[string]any{"url": "http://internal.example.test/private?key=private-token"}})
	if status != 200 {
		t.Fatalf("monitor create %d %s", status, b)
	}
	var m domain.Monitor
	json.Unmarshal(b, &m)
	return m
}
func TestMonitorCreationDefaultsAndExplicitZero(t *testing.T) {
	_, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	m := createTestMonitor(t, c, ts.URL, csrf)
	if !m.Enabled || m.Retries != 2 || m.IntervalSeconds != 60 || !m.NotifyRecovery {
		t.Fatalf("wrong defaults %+v", m)
	}
	status, b := request(t, c, "POST", ts.URL+"/api/v1/monitors", csrf, map[string]any{"name": "Paused", "type": "http", "enabled": false, "retries": 0, "notifyRecovery": false, "http": map[string]any{"url": "https://example.test"}})
	if status != 200 {
		t.Fatalf("explicit zero %d %s", status, b)
	}
	json.Unmarshal(b, &m)
	if m.Enabled || m.Retries != 0 || m.NotifyRecovery {
		t.Fatalf("explicit zero replaced %+v", m)
	}
}
func TestPublicationPrivacyDraftAndHostBinding(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	m := createTestMonitor(t, c, ts.URL, csrf)
	settings := domain.DefaultSettings()
	settings.AllowedDomains = []string{"status.test"}
	status, b := request(t, c, "PATCH", ts.URL+"/api/v1/settings", csrf, settings)
	if status != 200 {
		t.Fatalf("settings %d %s", status, b)
	}
	p := domain.Page{Name: "Status", Slug: "status1", Domain: "status.test", Draft: domain.PageConfig{Title: "Public service status", BrandColor: "#008877", ColorScheme: "system", Links: []domain.Link{}, Groups: []domain.PageGroup{{ID: "services", Name: "Services", Monitors: []domain.PageMonitor{{MonitorID: m.ID, Alias: "Public API", ShowUptime: true, ShowLatency: true}}}}}}
	status, b = request(t, c, "POST", ts.URL+"/api/v1/pages", csrf, p)
	if status != 200 {
		t.Fatalf("createpage %d %s", status, b)
	}
	json.Unmarshal(b, &p)
	if status, _ = request(t, c, "GET", ts.URL+"/api/public/pages/status1", "", nil); status != 404 {
		t.Fatal("unpublished draft visible")
	}
	status, b = request(t, c, "POST", ts.URL+"/api/v1/pages/"+p.ID+"/publish", csrf, nil)
	if status != 200 {
		t.Fatalf("publish %d %s", status, b)
	}
	json.Unmarshal(b, &p)
	status, b = request(t, &http.Client{}, "GET", ts.URL+"/api/public/pages/status1", "", nil)
	if status != 200 || !strings.Contains(string(b), "Public API") || strings.Contains(string(b), "internal.example") || strings.Contains(string(b), "private-token") || strings.Contains(string(b), "Private internal name") {
		t.Fatalf("public privacy %d %s", status, b)
	}
	p.Draft.Title = "Draft replacement"
	status, b = request(t, c, "PATCH", ts.URL+"/api/v1/pages/"+p.ID, csrf, p)
	if status != 200 {
		t.Fatalf("draftsave %d %s", status, b)
	}
	_, b = request(t, c, "GET", ts.URL+"/api/public/pages/status1", "", nil)
	if strings.Contains(string(b), "Draft replacement") {
		t.Fatal("draft changed published page")
	}
	_, b = request(t, c, "GET", ts.URL+"/api/v1/pages/"+p.ID+"/preview", "", nil)
	if !strings.Contains(string(b), "Draft replacement") {
		t.Fatal("preview ignored draft")
	}
	for _, test := range []struct {
		host, path string
		want       int
	}{{"status.test", "/api/public/resolve?host=status.test", 200}, {"unknown.test", "/api/public/resolve?host=status.test", 404}, {"status.test", "/api/v1/monitors", 404}, {"status.test", "/api/public/pages/status1", 200}} {
		req, _ := http.NewRequest("GET", ts.URL+test.path, nil)
		req.Host = test.host
		res, e := http.DefaultClient.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		if res.StatusCode != test.want {
			t.Fatalf("host %s route %s: %d", test.host, test.path, res.StatusCode)
		}
	}
	if status, _ = request(t, c, "DELETE", ts.URL+"/api/v1/monitors/"+m.ID, csrf, nil); status != 409 {
		t.Fatal("deleted published monitor")
	}
	if status, _ = request(t, c, "DELETE", ts.URL+"/api/v1/pages/"+p.ID, csrf, nil); status != 200 {
		t.Fatal("page delete")
	}
	if _, e := s.Store.PageIDByDomain(context.Background(), "status.test"); e != store.ErrNotFound {
		t.Fatalf("domain not released: %v", e)
	}
}
func TestPageStateOrdering(t *testing.T) {
	for _, test := range []struct {
		items []domain.PublicMonitor
		want  string
	}{
		{nil, "unknown"}, {[]domain.PublicMonitor{{Type: "certificate", State: "expired"}}, "unknown"},
		{[]domain.PublicMonitor{{Type: "http", State: "down"}, {Type: "http", State: "down"}}, "outage"},
		{[]domain.PublicMonitor{{Type: "http", State: "down"}, {Type: "http", Maintenance: true, State: "down"}}, "partial"},
		{[]domain.PublicMonitor{{Type: "http", State: "unknown"}, {Type: "http", Maintenance: true, State: "up"}}, "unknown"},
		{[]domain.PublicMonitor{{Type: "http", State: "up"}, {Type: "http", Maintenance: true, State: "up"}}, "maintenance"},
		{[]domain.PublicMonitor{{Type: "http", State: "up"}, {Type: "http", Paused: true, State: "down"}}, "operational"},
	} {
		if got := pageState(test.items); got != test.want {
			t.Fatalf("%+v got %s want %s", test.items, got, test.want)
		}
	}
}
func TestImageUploadValidationAndPersistence(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	var imageBytes bytes.Buffer
	png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	status, b := request(t, c, "POST", ts.URL+"/api/v1/assets", csrf, AssetWrite{Filename: "../../bad.svg", ContentType: "image/svg+xml", Base64: base64.StdEncoding.EncodeToString(imageBytes.Bytes())})
	if status != 200 {
		t.Fatalf("image %d %s", status, b)
	}
	var asset AssetURL
	json.Unmarshal(b, &asset)
	if !strings.HasSuffix(asset.URL, ".png") {
		t.Fatal("trusted client image type")
	}
	res, e := http.Get(ts.URL + asset.URL)
	if e != nil {
		t.Fatal(e)
	}
	data, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || !bytes.Equal(data, imageBytes.Bytes()) {
		t.Fatal("uploaded image not served")
	}
	if _, e = os.Stat(filepath.Join(s.Config.DataDir, "uploads", filepath.Base(asset.URL))); e != nil {
		t.Fatal(e)
	}
	status, _ = request(t, c, "POST", ts.URL+"/api/v1/assets", csrf, AssetWrite{Filename: "bad.svg", ContentType: "image/svg+xml", Base64: base64.StdEncoding.EncodeToString([]byte(`<svg onload="alert(1)"></svg>`))})
	if status != 422 {
		t.Fatal("SVG accepted")
	}
}
