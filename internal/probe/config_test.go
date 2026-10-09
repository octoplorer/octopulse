package probe

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
)

func TestMonitorValidationRejectsUnsafeOrInconsistentSettings(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*domain.Monitor)
	}{
		{name: "extra-config", mutate: func(m *domain.Monitor) { m.TCP = &domain.TCPConfig{Host: "localhost", Port: 80} }},
		{name: "method", mutate: func(m *domain.Monitor) { m.HTTP.Method = "GET\r\nX-Inject: yes" }},
		{name: "URL-credentials", mutate: func(m *domain.Monitor) { m.HTTP.URL = "https://user:password@example.test" }},
		{name: "header-injection", mutate: func(m *domain.Monitor) {
			m.HTTP.Headers = []domain.NameValue{{Name: "X-Test", Value: "a\r\nX-Secret: value"}}
		}},
		{name: "multipart-boundary", mutate: func(m *domain.Monitor) {
			m.HTTP.Body.Format = "multipart"
			m.HTTP.Headers = []domain.NameValue{{Name: "Content-Type", Value: "multipart/form-data; boundary=wrong"}}
		}},
		{
			name: "JSON-charset",
			mutate: func(m *domain.Monitor) {
				m.HTTP.Body = domain.HTTPBody{
					Format:  "json",
					Text:    `{}`,
					Charset: "gbk",
				}
			},
		},
		{name: "JSON-invalid", mutate: func(m *domain.Monitor) { m.HTTP.Body = domain.HTTPBody{Format: "json", Text: `{`} }},
		{
			name: "raw-invalid",
			mutate: func(m *domain.Monitor) {
				m.HTTP.Body = domain.HTTPBody{
					Format: "raw",
					Base64: "invalid:",
				}
			},
		},
		{name: "TLS-key-only", mutate: func(m *domain.Monitor) { m.HTTP.TLS.ClientKeySecretRef = "key" }},
		{
			name: "TLS-version-inverted",
			mutate: func(m *domain.Monitor) {
				m.HTTP.TLS.MinVersion = "1.3"
				m.HTTP.TLS.MaxVersion = "1.2"
			},
		},
		{name: "proxy-fixed-IP", mutate: func(m *domain.Monitor) {
			m.HTTP.Connection = domain.ConnectionConfig{ProxyURL: "http://proxy.test", FixedIP: "127.0.0.1"}
		}},
		{
			name: "proxy-plaintext",
			mutate: func(m *domain.Monitor) {
				m.HTTP.Connection.ProxyURL = "http://user:password@proxy.test"
			},
		},
		{name: "custom-DNS-port", mutate: func(m *domain.Monitor) { m.HTTP.Connection.DNSServer = "resolver.test" }},
		{name: "redirect-scope", mutate: func(m *domain.Monitor) { m.HTTP.Redirects.Scope = "subdomain" }},
		{name: "invalid-regex", mutate: func(m *domain.Monitor) { m.HTTP.Assertions.Regex = []string{"["} }},
		{
			name: "invalid-range",
			mutate: func(m *domain.Monitor) {
				m.HTTP.Assertions.StatusRanges = []domain.StatusRange{
					{
						Min: 300,
						Max: 200,
					},
				}
			},
		},
		{name: "retry-budget", mutate: func(m *domain.Monitor) { m.Retries = 11 }},
		{name: "short-interval", mutate: func(m *domain.Monitor) { m.IntervalSeconds = 5 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := httpMonitor("https://example.test")
			test.mutate(&m)
			if m.Validate() == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}

func TestSecretFailureHasNoPlaintextDiagnostics(t *testing.T) {
	m := httpMonitor("https://example.test")
	m.HTTP.Auth = domain.HTTPAuth{Type: "bearer", SecretRef: "token"}
	runner := NewRunner(SecretResolverFunc(func(context.Context, string) (string, error) {
		return "", errors.New("storage failed with password plaintext-secret")
	}))
	result := runner.Run(context.Background(), m)
	if result.Success || strings.Contains(result.Error, "plaintext-secret") || result.Error != "secret unavailable" {
		t.Fatal(result)
	}
	encoded, _ := json.Marshal(domain.Monitor{Heartbeat: &domain.HeartbeatConfig{SecretHash: "secret-hash"}})
	if strings.Contains(string(encoded), "secret-hash") {
		t.Fatal("heartbeat hash leaked in monitor DTO")
	}
}

func TestJSONPointerAndNumericValueSemantics(t *testing.T) {
	document := map[string]any{"a/b": []any{json.Number("3.00")}, "~key": true, "": nil}
	if value, ok := jsonPointer(document, "/a~1b/0"); !ok || !equalJSON(value, json.Number("3e0")) {
		t.Error("JSON number equality or pointer decoding failed")
	}
	if _, ok := jsonPointer(document, "/a~1b/+0"); ok {
		t.Error("noncanonical array index accepted")
	}
	if _, ok := jsonPointer(document, "/~2key"); ok {
		t.Error("invalid pointer escape accepted")
	}
	if value, ok := jsonPointer(document, "/"); !ok || value != nil {
		t.Error("empty field pointer failed")
	}
	if !equalJSON(json.Number("0e999999999999999999999"), json.Number("-0.0")) {
		t.Error("numeric zero should compare equal")
	}
	if equalJSON(json.Number("1e999999999999999999999"), json.Number("1")) {
		t.Error("large exponent numbers should differ")
	}
}
