// Package probe performs one attempt. Scheduling, retry budgets, confirmed
// states and notifications deliberately live outside network execution.
package probe

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
)

type SecretResolver interface {
	ResolveSecret(context.Context, string) (string, error)
}
type SecretResolverFunc func(context.Context, string) (string, error)

func (f SecretResolverFunc) ResolveSecret(ctx context.Context, id string) (string, error) {
	return f(ctx, id)
}

type Diagnostics struct {
	StatusCode    int      `json:"statusCode,omitempty"`
	FinalURL      string   `json:"finalUrl,omitempty"`
	Redirects     int      `json:"redirects,omitempty"`
	ResponseBytes int64    `json:"responseBytes,omitempty"`
	DNSMs         int64    `json:"dnsMs,omitempty"`
	ConnectMs     int64    `json:"connectMs,omitempty"`
	TLSMs         int64    `json:"tlsMs,omitempty"`
	FirstByteMs   int64    `json:"firstByteMs,omitempty"`
	DNSRCode      string   `json:"dnsRCode,omitempty"`
	DNSValues     []string `json:"dnsValues,omitempty"`
	TLSVersion    string   `json:"tlsVersion,omitempty"`
}
type CertificateResult struct {
	State         string  `json:"state"`
	ExpiresAt     int64   `json:"expiresAt"`
	Fingerprint   string  `json:"fingerprint"`
	DaysRemaining float64 `json:"daysRemaining"`
}
type Result struct {
	Success     bool               `json:"success"`
	LatencyMs   int64              `json:"latencyMs"`
	Error       string             `json:"error,omitempty"`
	Diagnostics Diagnostics        `json:"diagnostics"`
	Certificate *CertificateResult `json:"certificate,omitempty"`
}
type ProbeResult = Result
type Runner struct {
	Resolver SecretResolver
	Now      func() time.Time
}

func NewRunner(resolver SecretResolver) *Runner { return &Runner{Resolver: resolver, Now: time.Now} }

type execution struct {
	runner *Runner
	values []string
}

func (x *execution) secret(ctx context.Context, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	if x.runner.Resolver == nil {
		return "", errors.New("secret resolver unavailable")
	}
	value, err := x.runner.Resolver.ResolveSecret(ctx, ref)
	if err != nil {
		return "", errors.New("secret unavailable")
	}
	x.values = append(x.values, value)
	return value, nil
}
func (x *execution) value(ctx context.Context, v domain.NameValue) (string, error) {
	if v.SecretRef != "" {
		return x.secret(ctx, v.SecretRef)
	}
	return v.Value, nil
}

func (r *Runner) Run(ctx context.Context, monitor domain.Monitor) Result {
	start := time.Now()
	monitor.Defaults()
	if err := monitor.Validate(); err != nil {
		return Result{Error: "invalid monitor configuration: " + err.Error()}
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(monitor.TimeoutSeconds)*time.Second)
	defer cancel()
	x := &execution{runner: r}
	var result Result
	var err error
	switch monitor.Type {
	case domain.MonitorHTTP:
		result, err = x.http(ctx, *monitor.HTTP)
	case domain.MonitorTCP:
		result, err = x.tcp(ctx, *monitor.TCP)
	case domain.MonitorDNS:
		result, err = x.dns(ctx, *monitor.DNS)
	case domain.MonitorCertificate:
		result, err = x.certificate(ctx, *monitor.Certificate)
	default:
		err = errors.New("passive monitors cannot be probed")
	}
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		result.Success = false
		result.Error = x.safeError(ctx, err)
	}
	if monitor.Type == domain.MonitorCertificate && result.Certificate == nil {
		result.Certificate = &CertificateResult{State: domain.CertificateCheckFailed}
	}
	return result
}

func (x *execution) safeError(ctx context.Context, err error) string {
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "probe timeout"
		}
		return "probe cancelled"
	}
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return "network timeout"
		}
		return "network connection failed"
	}
	message := err.Error()
	for _, value := range x.values {
		if value != "" {
			message = strings.ReplaceAll(message, value, "[redacted]")
		}
	}
	if len(message) > 512 {
		message = message[:512]
	}
	return message
}

func fingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:])
}
func assertionError(kind string, index int) error {
	return fmt.Errorf("%s assertion %d failed", kind, index+1)
}
