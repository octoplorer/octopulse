package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/octoplorer/octopulse/internal/config"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/store"
	"github.com/octoplorer/octopulse/internal/testutil"
)

func testServer(t *testing.T) (*Server, *httptest.Server, *http.Client) {
	t.Helper()
	dir := t.TempDir()
	st, e := store.Open(context.Background(), testutil.Database(t))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { st.Close() })
	v, e := security.Open(dir, "")
	if e != nil {
		t.Fatal(e)
	}
	s := New(st, v, config.Config{DataDir: dir, StaticDir: filepath.Join(dir, "static"), AdminHosts: []string{"127.0.0.1", "localhost"}})
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	jar, _ := cookiejar.New(nil)
	return s, ts, &http.Client{Jar: jar}
}
func request(t *testing.T, c *http.Client, method, url, csrf string, body any) (int, []byte) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, e := http.NewRequest(method, url, bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	req.Header.Set("Content-Type", "application/json")
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	res, e := c.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	return res.StatusCode, out
}
func bootstrap(t *testing.T, c *http.Client, base string) string {
	t.Helper()
	status, b := request(t, c, "POST", base+"/api/v1/setup", "", map[string]any{"username": "admin", "password": "test passphrase 1234", "organizationName": "Test", "timezone": "UTC"})
	if status != 200 {
		t.Fatalf("setup %d: %s", status, b)
	}
	var session SessionBody
	if e := json.Unmarshal(b, &session); e != nil {
		t.Fatal(e)
	}
	if session.CSRFToken == "" || session.User == nil {
		t.Fatal("missing setup session")
	}
	if strings.Contains(string(b), "passwordHash") {
		t.Fatal("password hash leaked")
	}
	return session.CSRFToken
}
func TestBootstrapSessionCSRFAndSecrets(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	if status, b := request(t, c, "POST", ts.URL+"/api/v1/setup", "", map[string]any{"username": "another", "password": "test passphrase 1234", "organizationName": "Test", "timezone": "UTC"}); status != 409 {
		t.Fatalf("repeat setup %d %s", status, b)
	}
	if status, _ := request(t, c, "POST", ts.URL+"/api/v1/secrets", "", map[string]any{"name": "private", "value": "super-secret-token"}); status != 403 {
		t.Fatalf("missing csrf accepted: %d", status)
	}
	status, b := request(t, c, "POST", ts.URL+"/api/v1/secrets", csrf, map[string]any{"name": "private", "value": "super-secret-token"})
	if status != 200 {
		t.Fatalf("secret %d %s", status, b)
	}
	if strings.Contains(string(b), "super-secret-token") || strings.Contains(string(b), "ciphertext") {
		t.Fatal("secret readback")
	}
	var secret domain.Secret
	json.Unmarshal(b, &secret)
	var persisted domain.SecretRecord
	if e := s.Store.Get(context.Background(), "secrets", secret.ID, &persisted); e != nil {
		t.Fatal(e)
	}
	if strings.Contains(persisted.Ciphertext, "super-secret-token") {
		t.Fatal("plaintext persistence")
	}
	plain, e := s.ResolveSecret(context.Background(), secret.ID)
	if e != nil || plain != "super-secret-token" {
		t.Fatal("runtime secret resolution")
	}
	status, b = request(t, c, "GET", ts.URL+"/api/v1/secrets", "", nil)
	if status != 200 || strings.Contains(string(b), "ciphertext") || strings.Contains(string(b), "super-secret-token") {
		t.Fatalf("list privacy %d %s", status, b)
	}
	status, b = request(t, c, "DELETE", ts.URL+"/api/v1/session", csrf, nil)
	if status != 200 {
		t.Fatalf("logout %d %s", status, b)
	}
	if status, _ = request(t, c, "GET", ts.URL+"/api/v1/secrets", "", nil); status != 401 {
		t.Fatalf("logout auth %d", status)
	}
}
func TestReadOnlyAndLastAdmin(t *testing.T) {
	_, ts, admin := testServer(t)
	csrf := bootstrap(t, admin, ts.URL)
	status, b := request(t, admin, "POST", ts.URL+"/api/v1/users", csrf, map[string]any{"username": "viewer", "name": "Observer", "role": "viewer", "locale": "en", "timezone": "UTC", "enabled": true, "password": "observer password123"})
	if status != 200 {
		t.Fatalf("create user %d %s", status, b)
	}
	jar, _ := cookiejar.New(nil)
	viewer := &http.Client{Jar: jar}
	status, b = request(t, viewer, "POST", ts.URL+"/api/v1/session", "", map[string]any{"username": "viewer", "password": "observer password123"})
	if status != 200 {
		t.Fatalf("viewer login %d %s", status, b)
	}
	var session SessionBody
	json.Unmarshal(b, &session)
	if status, _ = request(t, viewer, "POST", ts.URL+"/api/v1/secrets", session.CSRFToken, map[string]any{"name": "bad", "value": "bad"}); status != 403 {
		t.Fatal("viewer can mutate")
	}
	if status, _ = request(t, viewer, "GET", ts.URL+"/api/v1/users", "", nil); status != 403 {
		t.Fatal("viewer can enumerate users")
	}
	status, b = request(t, admin, "GET", ts.URL+"/api/v1/session", "", nil)
	var own SessionBody
	json.Unmarshal(b, &own)
	status, b = request(t, admin, "DELETE", ts.URL+"/api/v1/users/"+own.User.ID, csrf, nil)
	if status != 409 {
		t.Fatalf("last admin delete %d %s", status, b)
	}
}

func TestProfilePasswordChangeKeepsCurrentSessionAndRevokesOthers(t *testing.T) {
	_, ts, primary := testServer(t)
	csrf := bootstrap(t, primary, ts.URL)
	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	if status, _ := request(t, other, "POST", ts.URL+"/api/v1/session", "", Credentials{Username: "admin", Password: "test passphrase 1234"}); status != 200 {
		t.Fatal("second login failed")
	}
	input := ProfileWrite{Name: "Administrator", Locale: "en", Timezone: "UTC", OldPassword: "wrong", Password: "replacement password 1234"}
	if status, _ := request(t, primary, "PATCH", ts.URL+"/api/v1/profile", csrf, input); status != 403 {
		t.Fatal("incorrect current password accepted")
	}
	input.OldPassword = "test passphrase 1234"
	if status, b := request(t, primary, "PATCH", ts.URL+"/api/v1/profile", csrf, input); status != 200 {
		t.Fatalf("password change: %d %s", status, b)
	}
	if status, _ := request(t, primary, "GET", ts.URL+"/api/v1/monitors", "", nil); status != 200 {
		t.Fatal("current session was revoked")
	}
	if status, _ := request(t, other, "GET", ts.URL+"/api/v1/monitors", "", nil); status != 401 {
		t.Fatal("other session remained active")
	}
	if status, _ := request(t, other, "POST", ts.URL+"/api/v1/session", "", Credentials{Username: "admin", Password: input.Password}); status != 200 {
		t.Fatal("new password cannot log in")
	}
}
