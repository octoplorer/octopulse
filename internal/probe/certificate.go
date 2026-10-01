package probe

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
)

func (x *execution) certificate(ctx context.Context, c domain.CertificateConfig) (Result, error) {
	result := Result{Certificate: &CertificateResult{State: domain.CertificateCheckFailed}}
	start := time.Now()
	conn, err := x.dial(ctx, c.Connection, net.JoinHostPort(c.Host, strconv.Itoa(c.Port)))
	if err != nil {
		return result, err
	}
	defer conn.Close()
	stop := bindCancellation(ctx, conn)
	defer stop()
	result.Diagnostics.ConnectMs = time.Since(start).Milliseconds()
	config, err := x.tlsConfig(ctx, c.TLS, c.Host)
	if err != nil {
		return result, err
	}
	verify := !config.InsecureSkipVerify
	config.InsecureSkipVerify = true // Inspect expired leaf; verify manually below.
	secured := tls.Client(conn, config)
	start = time.Now()
	if err = secured.HandshakeContext(ctx); err != nil {
		return result, errors.New("certificate TLS handshake failed")
	}
	result.Diagnostics.TLSMs = time.Since(start).Milliseconds()
	state := secured.ConnectionState()
	result.Diagnostics.TLSVersion = tlsVersion(state.Version)
	if len(state.PeerCertificates) == 0 {
		return result, errors.New("server did not provide a certificate")
	}
	leaf := state.PeerCertificates[0]
	now := time.Now()
	if x.runner.Now != nil {
		now = x.runner.Now()
	}
	details := result.Certificate
	details.ExpiresAt = leaf.NotAfter.UnixMilli()
	details.Fingerprint = fingerprint(leaf.Raw)
	details.DaysRemaining = leaf.NotAfter.Sub(now).Hours() / 24
	if verify {
		if err = leaf.VerifyHostname(config.ServerName); err != nil {
			return result, errors.New("certificate hostname validation failed")
		}
		intermediates := x509.NewCertPool()
		for _, certificate := range state.PeerCertificates[1:] {
			intermediates.AddCert(certificate)
		}
		verifyAt := now
		if now.After(leaf.NotAfter) {
			verifyAt = leaf.NotAfter.Add(-time.Second)
		}
		if _, err = leaf.Verify(x509.VerifyOptions{DNSName: config.ServerName, Roots: config.RootCAs, Intermediates: intermediates, CurrentTime: verifyAt}); err != nil {
			return result, errors.New("certificate trust or validity validation failed")
		}
	}
	if !now.Before(leaf.NotAfter) {
		details.State = domain.CertificateExpired
	} else {
		details.State = domain.CertificateHealthy
		for _, days := range c.WarningDays {
			if details.DaysRemaining <= float64(days) {
				details.State = domain.CertificateExpiring
				break
			}
		}
	}
	// Success reports a completed certificate observation. Expiry is a certificate
	// risk, not an availability failure; the scheduler uses Certificate.State.
	result.Success = true
	return result, nil
}
