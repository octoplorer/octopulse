package probe

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
)

func httpMonitor(target string) domain.Monitor {
	m := domain.Monitor{
		Name:            "test",
		Type:            domain.MonitorHTTP,
		Enabled:         true,
		IntervalSeconds: 30,
		TimeoutSeconds:  2,
		Retries:         2,
		HTTP: &domain.HTTPConfig{
			URL: target,
		},
	}
	m.Defaults()
	return m
}
func runMonitor(t *testing.T, runner *Runner, m domain.Monitor) Result {
	t.Helper()
	result := runner.Run(context.Background(), m)
	if !result.Success {
		t.Fatalf("probe failed: %+v", result)
	}
	return result
}

func TestHTTPQueryHeadersHostJSONAndAssertions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Host != "api.example.test" || strings.Join(
			r.URL.Query()["tag"],
			",",
		) != "original,a,b" || strings.Join(r.Header.Values("X-Repeated"), ",") != "one,two" {
			t.Errorf(
				"request settings lost: method=%s host=%s query=%v headers=%v",
				r.Method,
				r.Host,
				r.URL.Query(),
				r.Header,
			)
		}
		if r.Header.Get("Authorization") != "Bearer bearer-secret" {
			t.Error("bearer authentication missing")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"check":true}` {
			t.Errorf("body=%q", body)
		}
		w.Header().Add("X-Test", "first")
		w.Header().Add("X-Test", "second")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"ready":true,"nested":{"a/b":["ok",3]},"status":"available"}`))
	}))
	defer server.Close()
	m := httpMonitor(server.URL + "?tag=original")
	m.HTTP.Method = "POST"
	m.HTTP.Host = "api.example.test"
	m.HTTP.Query = []domain.NameValue{
		{
			Name:  "tag",
			Value: "a",
		},
		{
			Name:  "tag",
			Value: "b",
		},
		{
			Name:      "token",
			SecretRef: "token",
		},
	}
	m.HTTP.Headers = []domain.NameValue{{Name: "X-Repeated", Value: "one"}, {Name: "X-Repeated", Value: "two"}}
	m.HTTP.Body = domain.HTTPBody{Format: "json", Text: `{"check":true}`}
	m.HTTP.Auth = domain.HTTPAuth{Type: "bearer", SecretRef: "token"}
	m.HTTP.Assertions = domain.HTTPAssertions{
		StatusCodes: []int{201},
		Headers: []domain.ValueAssertion{
			{
				Name:     "X-Test",
				Operator: "equals",
				Value:    "second",
			},
		},
		TextContains:    []string{"available"},
		TextNotContains: []string{"unavailable"},
		Regex:           []string{`"ready":true`},
		JSON: []domain.JSONAssertion{
			{
				Pointer:  "/nested/a~1b/1",
				Operator: "equals",
				Value:    json.RawMessage("3"),
			},
			{
				Pointer:  "/ready",
				Operator: "equals",
				Value:    json.RawMessage("true"),
			},
			{
				Pointer:  "/status",
				Operator: "regex",
				Value:    json.RawMessage(`"^avail"`),
			},
		},
	}
	result := runMonitor(
		t,
		NewRunner(
			SecretResolverFunc(
				func(context.Context, string) (string, error) {
					return "bearer-secret", nil
				},
			),
		),
		m,
	)
	finalURL := result.Diagnostics.FinalURL
	if strings.Contains(finalURL, "token") || strings.Contains(finalURL, "bearer-secret") {
		t.Fatal("diagnostics leaked query credentials")
	}
	if result.Diagnostics.StatusCode != 201 {
		t.Fatal(result.Diagnostics)
	}
	m.HTTP.Assertions.JSON[0].Value = json.RawMessage("4")
	result = NewRunner(
		SecretResolverFunc(
			func(context.Context, string) (string, error) {
				return "bearer-secret", nil
			},
		),
	).Run(context.Background(), m)
	if result.Success || !strings.Contains(result.Error, "JSON assertion") {
		t.Fatalf("failed assertion accepted: %+v", result)
	}
}

func TestHTTPBodyFormatsAndRequestCompression(t *testing.T) {
	for _, format := range []string{"none", "json", "form", "multipart", "text", "raw"} {
		t.Run(format, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reader, err := gzip.NewReader(r.Body)
				if err != nil {
					t.Error(err)
					w.WriteHeader(500)
					return
				}
				defer reader.Close()
				data, _ := io.ReadAll(reader)
				switch format {
				case "none":
					if len(data) != 0 {
						t.Error("expected empty body")
					}
				case "json":
					if string(data) != `{"ok":true}` {
						t.Errorf("JSON body=%q", data)
					}
				case "text":
					text, e := decodeText(data, "gbk")
					if e != nil || text != "服务在线" {
						t.Errorf("text=%q err=%v", text, e)
					}
				case "raw":
					if !bytes.Equal(data, []byte{0, 1, 2, 255}) {
						t.Errorf("bytes=%v", data)
					}
				case "form":
					values, e := url.ParseQuery(string(data))
					if e != nil {
						t.Error(e)
					}
					text, e := decodeText([]byte(values.Get("name")), "gbk")
					if e != nil || text != "服务在线" || len(values["name"]) != 2 {
						t.Errorf(
							"form=%v text=%q err=%v",
							values,
							text,
							e,
						)
					}
				case "multipart":
					r.Body = io.NopCloser(bytes.NewReader(data))
					if err := r.ParseMultipartForm(1 << 20); err != nil {
						t.Error(err)
					}
					if r.FormValue("note") != "check" {
						t.Error("missing multipart field")
					}
					file, header, e := r.FormFile("file")
					if e != nil {
						t.Error(e)
						break
					}
					defer file.Close()
					bytes, _ := io.ReadAll(file)
					if string(bytes) != "file-data" || header.Filename != "check.txt" {
						t.Error("invalid multipart file")
					}
				}
				w.WriteHeader(204)
			}))
			defer server.Close()
			m := httpMonitor(server.URL)
			m.HTTP.Method = "POST"
			m.HTTP.RequestGzip = true
			m.HTTP.Body.Format = format
			switch format {
			case "json":
				m.HTTP.Body.Text = `{"ok":true}`
			case "text":
				m.HTTP.Body.Text = "服务在线"
				m.HTTP.Body.Charset = "gbk"
			case "form":
				m.HTTP.Body.Charset = "gbk"
				m.HTTP.Body.Fields = []domain.NameValue{{Name: "name", Value: "服务在线"}, {Name: "name", Value: "another"}}
			case "raw":
				m.HTTP.Body.Base64 = base64.StdEncoding.EncodeToString([]byte{0, 1, 2, 255})
			case "multipart":
				m.HTTP.Body.Fields = []domain.NameValue{{Name: "note", Value: "check"}}
				m.HTTP.Body.Files = []domain.MultipartFile{
					{
						Field:    "file",
						Filename: "check.txt",
						Base64:   base64.StdEncoding.EncodeToString([]byte("file-data")),
					},
				}
			}
			runMonitor(t, NewRunner(nil), m)
		})
	}
}

func TestHTTPGzipAndCharsetLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") != "gzip" {
			t.Error("gzip not negotiated")
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/plain; charset=gbk")
		writer := gzip.NewWriter(w)
		data, _ := encodeText(strings.Repeat("服务在线", 30), "gbk")
		_, _ = writer.Write(data)
		_ = writer.Close()
	}))
	defer server.Close()
	m := httpMonitor(server.URL)
	m.HTTP.AcceptEncoding = "gzip"
	m.HTTP.Assertions.TextContains = []string{"服务在线"}
	runMonitor(t, NewRunner(nil), m)
	m.HTTP.MaxResponseBytes = 100
	result := NewRunner(nil).Run(context.Background(), m)
	if result.Success || !strings.Contains(result.Error, "size limit") {
		t.Fatalf("gzip bomb limit not enforced: %+v", result)
	}
}

func TestHTTPRedirectScopeAndCredentials(t *testing.T) {
	var receivedCredential atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Private") != "" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			receivedCredential.Store(true)
		}
		if r.URL.Path == "/" {
			http.Redirect(
				w,
				r,
				"/final",
				302,
			)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()
	origin := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(
					w,
					r,
					target.URL,
					302,
				)
			},
		),
	)
	defer origin.Close()
	m := httpMonitor(origin.URL)
	m.HTTP.Redirects.Enabled = true
	m.HTTP.Auth = domain.HTTPAuth{Type: "basic", Username: "user", SecretRef: "secret"}
	m.HTTP.Headers = []domain.NameValue{
		{
			Name:      "X-Private",
			SecretRef: "secret",
		},
		{
			Name:  "Cookie",
			Value: "session=secret",
		},
	}
	runner := NewRunner(SecretResolverFunc(func(context.Context, string) (string, error) { return "private-value", nil }))
	result := runner.Run(context.Background(), m)
	if result.Success {
		t.Fatal("cross-origin redirect accepted under same-origin scope")
	}
	m.HTTP.Redirects.Scope = "any"
	result = runMonitor(t, runner, m)
	if receivedCredential.Load() {
		t.Fatal("credentials leaked across redirect origin")
	}
	if result.Diagnostics.Redirects != 2 {
		t.Fatal(result.Diagnostics)
	}
	loop := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(
					w,
					r,
					"/again",
					302,
				)
			},
		),
	)
	defer loop.Close()
	m = httpMonitor(loop.URL)
	m.HTTP.Redirects.Enabled = true
	m.HTTP.Redirects.MaxHops = 2
	result = NewRunner(nil).Run(context.Background(), m)
	if result.Success {
		t.Fatal("redirect loop accepted")
	}
}

func TestHTTPRedirectMethodAndOneAttemptForEveryMethod(t *testing.T) {
	for _, status := range []int{301, 302, 303, 307, 308} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/start" {
					http.Redirect(
						w,
						r,
						"/end",
						status,
					)
					return
				}
				data, _ := io.ReadAll(r.Body)
				if status == 307 || status == 308 {
					if r.Method != "POST" || string(data) != "payload" {
						t.Error("body-preserving redirect failed")
					}
				} else if r.Method != "GET" {
					t.Error("POST redirect method not transformed")
				}
				w.WriteHeader(200)
			}))
			defer server.Close()
			m := httpMonitor(server.URL + "/start")
			m.HTTP.Method = "POST"
			m.HTTP.Body = domain.HTTPBody{Format: "text", Text: "payload"}
			m.HTTP.Redirects.Enabled = true
			runMonitor(t, NewRunner(nil), m)
		})
	}
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"} {
		t.Run(method, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						requests.Add(1)
						w.WriteHeader(503)
					},
				),
			)
			defer server.Close()
			m := httpMonitor(server.URL)
			m.HTTP.Method = method
			m.Retries = 2
			result := NewRunner(nil).Run(context.Background(), m)
			if result.Success || requests.Load() != 1 {
				t.Errorf(
					"probe must perform one logical attempt regardless method/retries: result=%+v requests=%d",
					result,
					requests.Load(),
				)
			}
		})
	}
}

func TestHTTPTimeoutAndLatencyAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Millisecond)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()
	m := httpMonitor(server.URL)
	m.HTTP.Assertions.MaxLatencyMs = 10
	result := NewRunner(nil).Run(context.Background(), m)
	if result.Success || result.Error != "response latency assertion failed" {
		t.Fatal(result)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	m.HTTP.Assertions.MaxLatencyMs = 0
	result = NewRunner(nil).Run(ctx, m)
	if result.Success || result.Error != "probe timeout" {
		t.Fatal(result)
	}
}
