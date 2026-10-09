package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMutationOriginChecks(t *testing.T) {
	s, ts, c := testServer(t)
	for _, tc := range []struct {
		name       string
		origin     string
		forwarded  string
		secure     bool
		wantStatus int
	}{
		{name: "same origin", origin: ts.URL, wantStatus: 422},
		{name: "CLI without origin", wantStatus: 422},
		{name: "different host", origin: "http://untrusted.example", wantStatus: 403},
		{name: "different port", origin: "http://127.0.0.1:5173", wantStatus: 403},
		{name: "opaque origin", origin: "null", wantStatus: 403},
		{
			name:       "forwarded host cannot authorize origin",
			origin:     "http://untrusted.example",
			forwarded:  "untrusted.example",
			wantStatus: 403,
		},
		{name: "secure cookie rejects HTTP origin", origin: ts.URL, secure: true, wantStatus: 403},
		{
			name: "secure cookie accepts HTTPS origin",
			origin: strings.Replace(
				ts.URL,
				"http://",
				"https://",
				1,
			),
			secure:     true,
			wantStatus: 422,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s.Config.CookieSecure = tc.secure
			req, err := http.NewRequest(
				http.MethodPost,
				ts.URL+"/api/v1/session",
				strings.NewReader(`{"username":"","password":""}`),
			)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("X-Forwarded-Host", tc.forwarded)
			res, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != tc.wantStatus {
				t.Fatalf(
					"status %d, want %d: %s",
					res.StatusCode,
					tc.wantStatus,
					body,
				)
			}
			if tc.wantStatus == 403 {
				var problem struct{ Detail string }
				if err := json.Unmarshal(body, &problem); err != nil || problem.Detail != "Untrusted request origin" {
					t.Fatalf("wrong rejection: %s", body)
				}
			}
		})
	}
}
