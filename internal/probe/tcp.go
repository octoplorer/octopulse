package probe

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
)

func (x *execution) tcp(ctx context.Context, c domain.TCPConfig) (Result, error) {
	result := Result{}
	start := time.Now()
	conn, err := x.dial(ctx, c.Connection, net.JoinHostPort(c.Host, strconv.Itoa(c.Port)))
	if err != nil {
		return result, err
	}
	defer conn.Close()
	stop := bindCancellation(ctx, conn)
	defer stop()
	result.Diagnostics.ConnectMs = time.Since(start).Milliseconds()
	if c.TLS.Enabled {
		config, e := x.tlsConfig(ctx, c.TLS, c.Host)
		if e != nil {
			return result, e
		}
		secured := tls.Client(conn, config)
		start = time.Now()
		if e = secured.HandshakeContext(ctx); e != nil {
			return result, errors.New("target TLS handshake failed")
		}
		result.Diagnostics.TLSMs = time.Since(start).Milliseconds()
		result.Diagnostics.TLSVersion = tlsVersion(secured.ConnectionState().Version)
		conn = secured
	}
	var payload []byte
	if c.SendSecretRef != "" {
		value, e := x.secret(ctx, c.SendSecretRef)
		if e != nil {
			return result, e
		}
		payload, err = encodeText(value, c.Charset)
	} else if c.SendBase64 != "" {
		payload, err = base64.StdEncoding.DecodeString(c.SendBase64)
	} else {
		payload, err = encodeText(c.SendText, c.Charset)
	}
	if err != nil {
		return result, err
	}
	if len(payload) > 0 {
		for len(payload) > 0 {
			n, e := conn.Write(payload)
			if e != nil {
				return result, e
			}
			if n == 0 {
				return result, io.ErrNoProgress
			}
			payload = payload[n:]
		}
	}
	if c.ReceiveContains == "" && c.ReceiveRegex == "" {
		result.Success = true
		return result, nil
	}
	var expression *regexp.Regexp
	if c.ReceiveRegex != "" {
		expression, err = regexp.Compile(c.ReceiveRegex)
		if err != nil {
			return result, errors.New("invalid receive regex")
		}
	}
	buffer := make([]byte, 4096)
	data := make([]byte, 0, 4096)
	for {
		remaining := c.MaxReceiveBytes + 1 - int64(len(data))
		if remaining < 1 {
			return result, errors.New("TCP response exceeds size limit")
		}
		chunk := buffer
		if remaining < int64(len(chunk)) {
			chunk = chunk[:remaining]
		}
		n, e := conn.Read(chunk)
		data = append(data, chunk[:n]...)
		if int64(len(data)) > c.MaxReceiveBytes {
			return result, errors.New("TCP response exceeds size limit")
		}
		text, decodeErr := decodeText(data, c.Charset)
		if decodeErr != nil {
			return result, decodeErr
		}
		result.Diagnostics.ResponseBytes = int64(len(data))
		if (c.ReceiveContains == "" || strings.Contains(text, c.ReceiveContains)) && (expression == nil || expression.MatchString(text)) {
			result.Success = true
			return result, nil
		}
		if e != nil {
			if errors.Is(e, io.EOF) {
				return result, errors.New("TCP receive assertion failed")
			}
			return result, e
		}
		if n == 0 {
			return result, io.ErrNoProgress
		}
	}
}
