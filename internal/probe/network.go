package probe

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
)

func (x *execution) tlsConfig(ctx context.Context, c domain.TLSConfig, host string) (*tls.Config, error) {
	config := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: c.InsecureSkipVerify, ServerName: c.ServerName}
	if config.ServerName == "" {
		config.ServerName = host
	}
	version := func(v string) uint16 {
		if v == "1.3" {
			return tls.VersionTLS13
		}
		if v == "1.2" {
			return tls.VersionTLS12
		}
		return 0
	}
	if c.MinVersion != "" {
		config.MinVersion = version(c.MinVersion)
	}
	config.MaxVersion = version(c.MaxVersion)
	if c.CASecretRef != "" {
		pem, err := x.secret(ctx, c.CASecretRef)
		if err != nil {
			return nil, err
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM([]byte(pem)) {
			return nil, errors.New("CA secret does not contain a PEM certificate")
		}
		config.RootCAs = roots
	}
	if c.ClientCertificateSecretRef != "" {
		certPEM, err := x.secret(ctx, c.ClientCertificateSecretRef)
		if err != nil {
			return nil, err
		}
		keyPEM, err := x.secret(ctx, c.ClientKeySecretRef)
		if err != nil {
			return nil, err
		}
		pair, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
		if err != nil {
			return nil, errors.New("invalid client certificate/key pair")
		}
		config.Certificates = []tls.Certificate{pair}
	}
	return config, nil
}

func makeDialer(c domain.ConnectionConfig) *net.Dialer {
	d := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	if c.DNSServer != "" {
		d.Resolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var nd net.Dialer
			return nd.DialContext(ctx, network, c.DNSServer)
		}}
	}
	return d
}

func (x *execution) proxyURL(ctx context.Context, c domain.ConnectionConfig) (*url.URL, error) {
	if c.ProxyURL == "" {
		return nil, nil
	}
	u, err := url.Parse(c.ProxyURL)
	if err != nil {
		return nil, errors.New("invalid proxy URL")
	}
	if c.ProxyUsername != "" || c.ProxyPasswordSecretRef != "" {
		password, err := x.secret(ctx, c.ProxyPasswordSecretRef)
		if err != nil {
			return nil, err
		}
		u.User = url.UserPassword(c.ProxyUsername, password)
	}
	return u, nil
}

// dialDirect changes only the socket endpoint. Request Host and TLS SNI stay
// anchored to the configured target. Custom DNS resolves direct targets and a
// proxy's address; HTTP CONNECT/SOCKS resolve target names on the proxy side.
func dialDirect(ctx context.Context, c domain.ConnectionConfig, network, address string) (net.Conn, error) {
	if c.FixedIP != "" {
		_, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		address = net.JoinHostPort(c.FixedIP, port)
	}
	return makeDialer(c).DialContext(ctx, network, address)
}

func bindCancellation(ctx context.Context, conn net.Conn) func() {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	return func() { stop() }
}

func (x *execution) dial(ctx context.Context, c domain.ConnectionConfig, address string) (net.Conn, error) {
	p, err := x.proxyURL(ctx, c)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return dialDirect(ctx, c, "tcp", address)
	}
	port := p.Port()
	if port == "" {
		switch p.Scheme {
		case "https":
			port = "443"
		case "socks5", "socks5h":
			port = "1080"
		default:
			port = "80"
		}
	}
	conn, err := makeDialer(c).DialContext(ctx, "tcp", net.JoinHostPort(p.Hostname(), port))
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = conn.Close()
		}
	}()
	stop := bindCancellation(ctx, conn)
	defer stop()
	if p.Scheme == "https" {
		secured := tls.Client(conn, &tls.Config{ServerName: p.Hostname(), MinVersion: tls.VersionTLS12})
		if err = secured.HandshakeContext(ctx); err != nil {
			return nil, errors.New("proxy TLS handshake failed")
		}
		conn = secured
	}
	if p.Scheme == "socks5" || p.Scheme == "socks5h" {
		err = socksConnect(conn, p, address)
	} else {
		req := &http.Request{Method: http.MethodConnect, URL: &url.URL{Opaque: address}, Host: address, Header: make(http.Header)}
		if p.User != nil {
			password, _ := p.User.Password()
			req.SetBasicAuth(p.User.Username(), password)
			req.Header.Set("Proxy-Authorization", req.Header.Get("Authorization"))
			req.Header.Del("Authorization")
		}
		if err = req.Write(conn); err == nil {
			var response *http.Response
			var reader = bufio.NewReader(conn)
			response, err = http.ReadResponse(reader, req)
			if err == nil {
				if response.StatusCode != 200 {
					err = fmt.Errorf("proxy CONNECT rejected (%d)", response.StatusCode)
				} else {
					conn = &bufferedConn{Conn: conn, reader: reader}
				}
			}
		}
	}
	if err != nil {
		return nil, err
	}
	success = true
	return conn, nil
}

type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) { return c.reader.Read(p) }

func socksConnect(conn net.Conn, p *url.URL, address string) error {
	methods := []byte{5, 1, 0}
	if p.User != nil {
		methods = []byte{5, 2, 0, 2}
	}
	if _, err := conn.Write(methods); err != nil {
		return err
	}
	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return err
	}
	if reply[0] != 5 {
		return errors.New("invalid SOCKS proxy response")
	}
	if reply[1] == 2 {
		if p.User == nil {
			return errors.New("SOCKS proxy requires authentication")
		}
		password, _ := p.User.Password()
		username := p.User.Username()
		if len(username) > 255 || len(password) > 255 {
			return errors.New("SOCKS credentials exceed protocol limit")
		}
		packet := []byte{1, byte(len(username))}
		packet = append(packet, username...)
		packet = append(packet, byte(len(password)))
		packet = append(packet, password...)
		if _, err := conn.Write(packet); err != nil {
			return err
		}
		if _, err := io.ReadFull(conn, reply); err != nil {
			return err
		}
		if reply[0] != 1 || reply[1] != 0 {
			return errors.New("SOCKS authentication failed")
		}
	} else if reply[1] != 0 {
		return errors.New("SOCKS proxy has no supported authentication method")
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	packet := []byte{5, 1, 0}
	ip := net.ParseIP(host)
	if v4 := ip.To4(); v4 != nil {
		packet = append(packet, 1)
		packet = append(packet, v4...)
	} else if ip != nil {
		packet = append(packet, 4)
		packet = append(packet, ip.To16()...)
	} else {
		if len(host) > 255 {
			return errors.New("SOCKS target name too long")
		}
		packet = append(packet, 3, byte(len(host)))
		packet = append(packet, host...)
	}
	packet = binary.BigEndian.AppendUint16(packet, uint16(portNumber))
	if _, err = conn.Write(packet); err != nil {
		return err
	}
	header := make([]byte, 4)
	if _, err = io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[0] != 5 || header[1] != 0 {
		return errors.New("SOCKS proxy connect failed")
	}
	length := 0
	switch header[3] {
	case 1:
		length = 4
	case 4:
		length = 16
	case 3:
		var size [1]byte
		if _, err = io.ReadFull(conn, size[:]); err != nil {
			return err
		}
		length = int(size[0])
	default:
		return errors.New("invalid SOCKS address type")
	}
	_, err = io.CopyN(io.Discard, conn, int64(length+2))
	return err
}

func tlsVersion(v uint16) string {
	switch v {
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return "unknown"
	}
}
func redactURL(u *url.URL) string {
	clean := *u
	clean.User = nil
	clean.RawQuery = ""
	clean.Fragment = ""
	return clean.String()
}
func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}
