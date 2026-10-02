package probe

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/octoplorer/octopulse/internal/domain"
)

func tcpMonitor(address string) domain.Monitor {
	host, port, _ := net.SplitHostPort(address)
	p, _ := strconv.Atoi(port)
	m := domain.Monitor{Name: "tcp", Type: domain.MonitorTCP, IntervalSeconds: 30, TimeoutSeconds: 2, TCP: &domain.TCPConfig{Host: host, Port: p}}
	m.Defaults()
	return m
}
func TestTCPBinarySendAndChunkedReceive(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, e := listener.Accept()
		if e != nil {
			return
		}
		defer conn.Close()
		data := make([]byte, 3)
		if _, e = io.ReadFull(conn, data); e != nil || string(data) != "\x00\x01\x02" {
			t.Error("binary send failed")
		}
		_, _ = conn.Write([]byte("ser"))
		time.Sleep(5 * time.Millisecond)
		_, _ = conn.Write([]byte("vice ready\n"))
	}()
	m := tcpMonitor(listener.Addr().String())
	m.TCP.SendBase64 = base64.StdEncoding.EncodeToString([]byte{0, 1, 2})
	m.TCP.ReceiveContains = "ready"
	m.TCP.ReceiveRegex = `service ready\s`
	runMonitor(t, NewRunner(nil), m)
	<-done
}
func TestTCPReceiveLimitAndCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, e := listener.Accept()
			if e != nil {
				return
			}
			go func() {
				defer conn.Close()
				_, _ = conn.Write([]byte(strings.Repeat("x", 100)))
				_, _ = io.Copy(io.Discard, conn)
			}()
		}
	}()
	m := tcpMonitor(listener.Addr().String())
	m.TCP.ReceiveContains = "never"
	m.TCP.MaxReceiveBytes = 10
	result := NewRunner(nil).Run(context.Background(), m)
	if result.Success || !strings.Contains(result.Error, "size limit") {
		t.Fatal(result)
	}
	m.TCP.MaxReceiveBytes = 200
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	result = NewRunner(nil).Run(ctx, m)
	// The socket's absolute deadline and the context timer use the same
	// deadline. Under load the socket can return before ctx.Err is published;
	// both timeout classifications prove bounded cancellation.
	if result.Success || (result.Error != "probe timeout" && result.Error != "network timeout") || time.Since(started) > time.Second {
		t.Fatal(result)
	}
}

func TestDNSRecordTypesProtocolsAndAssertions(t *testing.T) {
	records := map[string]string{"A": "test.example. 60 IN A 127.0.0.9", "AAAA": "test.example. 60 IN AAAA ::1", "CNAME": "test.example. 60 IN CNAME target.example.", "MX": "test.example. 60 IN MX 10 mail.example.", "TXT": "test.example. 60 IN TXT \"check \" \"ready\"", "NS": "test.example. 60 IN NS ns.example.", "SRV": "test.example. 60 IN SRV 1 2 443 target.example.", "PTR": "test.example. 60 IN PTR target.example.", "SOA": "test.example. 60 IN SOA ns.example. hostmaster.example. 42 10 20 30 40", "CAA": "test.example. 60 IN CAA 0 issue \"letsencrypt.org\""}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	packet, err := net.ListenPacket("udp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	handler := dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		response := new(dns.Msg)
		response.SetReply(r)
		q := r.Question[0]
		if q.Name == "missing.example." {
			response.Rcode = dns.RcodeNameError
		} else {
			record, e := dns.NewRR(records[dns.TypeToString[q.Qtype]])
			if e != nil {
				t.Error(e)
			} else {
				response.Answer = []dns.RR{record}
			}
		}
		_ = w.WriteMsg(response)
	})
	udp := &dns.Server{PacketConn: packet, Handler: handler}
	tcp := &dns.Server{Listener: listener, Handler: handler}
	go func() { _ = udp.ActivateAndServe() }()
	go func() { _ = tcp.ActivateAndServe() }()
	defer udp.Shutdown()
	defer tcp.Shutdown()
	for recordType, source := range records {
		for _, protocol := range []string{"udp", "tcp"} {
			t.Run(recordType+"/"+protocol, func(t *testing.T) {
				record, _ := dns.NewRR(source)
				m := domain.Monitor{Name: "dns", Type: domain.MonitorDNS, IntervalSeconds: 30, TimeoutSeconds: 2, DNS: &domain.DNSConfig{Name: "test.example", RecordType: recordType, Server: listener.Addr().String(), Protocol: protocol, ExpectedValues: []string{dnsValue(record)}, MatchMode: "exact"}}
				m.Defaults()
				runMonitor(t, NewRunner(nil), m)
				m.DNS.ExpectedValues = []string{"unexpected"}
				result := NewRunner(nil).Run(context.Background(), m)
				if result.Success {
					t.Fatal("incorrect DNS record accepted")
				}
			})
		}
	}
	m := domain.Monitor{Name: "dns", Type: domain.MonitorDNS, IntervalSeconds: 30, TimeoutSeconds: 2, DNS: &domain.DNSConfig{Name: "missing.example", RecordType: "A", Server: listener.Addr().String(), Protocol: "udp", ExpectedRCode: "NXDOMAIN"}}
	m.Defaults()
	runMonitor(t, NewRunner(nil), m)
	m.DNS.ExpectedRCode = "NOERROR"
	if NewRunner(nil).Run(context.Background(), m).Success {
		t.Fatal("incorrect DNS response code accepted")
	}
}

func TestDNSUDPTruncationFallsBackToTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	packet, err := net.ListenPacket("udp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	var tcpQueries atomic.Int32
	udp := &dns.Server{PacketConn: packet, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		reply := new(dns.Msg)
		reply.SetReply(r)
		reply.Truncated = true
		_ = w.WriteMsg(reply)
	})}
	tcp := &dns.Server{Listener: listener, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		tcpQueries.Add(1)
		reply := new(dns.Msg)
		reply.SetReply(r)
		record, _ := dns.NewRR("test.example. 60 IN A 127.0.0.9")
		reply.Answer = []dns.RR{record}
		_ = w.WriteMsg(reply)
	})}
	go func() { _ = udp.ActivateAndServe() }()
	go func() { _ = tcp.ActivateAndServe() }()
	defer udp.Shutdown()
	defer tcp.Shutdown()
	m := domain.Monitor{Name: "dns", Type: domain.MonitorDNS, IntervalSeconds: 30, TimeoutSeconds: 2, DNS: &domain.DNSConfig{Name: "test.example", Server: listener.Addr().String(), ExpectedValues: []string{"127.0.0.9"}}}
	m.Defaults()
	runMonitor(t, NewRunner(nil), m)
	if tcpQueries.Load() != 1 {
		t.Error("TCP fallback did not run")
	}
}

type certFixture struct {
	pair          tls.Certificate
	ca, cert, key string
}

func issueFixture(t *testing.T, notAfter time.Time, client bool) certFixture {
	t.Helper()
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	caTemplate := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "probe test CA"}, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, NotBefore: time.Now().Add(-365 * 24 * time.Hour), NotAfter: time.Now().Add(365 * 24 * time.Hour)}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	ca, _ := x509.ParseCertificate(caDER)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	template := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "probe.test"}, DNSNames: []string{"probe.test"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, NotBefore: time.Now().Add(-180 * 24 * time.Hour), NotAfter: notAfter, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	if client {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, _ := x509.MarshalPKCS8PrivateKey(key)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return certFixture{pair: pair, ca: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})), cert: string(certPEM), key: string(keyPEM)}
}
func fixtureServer(t *testing.T, fixture certFixture) *httptest.Server {
	t.Helper()
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("tls ready")) }))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{fixture.pair}, MinVersion: tls.VersionTLS12}
	server.StartTLS()
	t.Cleanup(server.Close)
	return server
}

func TestTLSCustomCAFixedIPSNIAndVersions(t *testing.T) {
	fixture := issueFixture(t, time.Now().Add(90*24*time.Hour), false)
	server := fixtureServer(t, fixture)
	u := strings.Replace(server.URL, "127.0.0.1", "probe.test", 1)
	m := httpMonitor(u)
	m.HTTP.Connection.FixedIP = "127.0.0.1"
	m.HTTP.TLS = domain.TLSConfig{CASecretRef: "ca", MinVersion: "1.2", MaxVersion: "1.2"}
	m.HTTP.Assertions.TextContains = []string{"tls ready"}
	runner := NewRunner(SecretResolverFunc(func(context.Context, string) (string, error) { return fixture.ca, nil }))
	result := runMonitor(t, runner, m)
	if result.Diagnostics.TLSVersion != "1.2" {
		t.Fatal(result.Diagnostics)
	}
	m.HTTP.TLS.ServerName = "wrong.example"
	if runner.Run(context.Background(), m).Success {
		t.Fatal("wrong SNI accepted")
	}
	m.HTTP.TLS.InsecureSkipVerify = true
	runMonitor(t, runner, m)
	tcp := tcpMonitor(server.Listener.Addr().String())
	tcp.TCP.TLS = domain.TLSConfig{Enabled: true, CASecretRef: "ca", ServerName: "probe.test"}
	tcp.TCP.SendText = "GET / HTTP/1.0\r\nHost: probe.test\r\n\r\n"
	tcp.TCP.ReceiveContains = "tls ready"
	runMonitor(t, runner, tcp)
}

func TestCertificateFourStates(t *testing.T) {
	for _, test := range []struct {
		name    string
		expiry  time.Duration
		trusted bool
		want    string
	}{{"healthy", 90 * 24 * time.Hour, true, domain.CertificateHealthy}, {"expiring", 7 * 24 * time.Hour, true, domain.CertificateExpiring}, {"expired", -24 * time.Hour, true, domain.CertificateExpired}, {"check-failed", 90 * 24 * time.Hour, false, domain.CertificateCheckFailed}} {
		t.Run(test.name, func(t *testing.T) {
			fixture := issueFixture(t, time.Now().Add(test.expiry), false)
			server := fixtureServer(t, fixture)
			host, port, _ := net.SplitHostPort(server.Listener.Addr().String())
			p, _ := strconv.Atoi(port)
			m := domain.Monitor{Name: "certificate", Type: domain.MonitorCertificate, IntervalSeconds: 86400, TimeoutSeconds: 2, Certificate: &domain.CertificateConfig{Host: host, Port: p}}
			m.Defaults()
			if test.trusted {
				m.Certificate.TLS.CASecretRef = "ca"
			}
			runner := NewRunner(SecretResolverFunc(func(context.Context, string) (string, error) { return fixture.ca, nil }))
			result := runner.Run(context.Background(), m)
			if result.Certificate == nil || result.Certificate.State != test.want {
				t.Fatalf("certificate result=%+v error=%s", result.Certificate, result.Error)
			}
			if result.Success != (test.want != domain.CertificateCheckFailed) {
				t.Fatal(result)
			}
			if test.trusted && (result.Certificate.ExpiresAt == 0 || result.Certificate.Fingerprint == "") {
				t.Fatal("missing certificate details")
			}
		})
	}
}

func TestHTTPProxyAuthenticationAndTCPConnect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Proxy-Authorization") != "" {
			t.Error("proxy credential reached target")
		}
		_, _ = w.Write([]byte("proxied"))
	}))
	defer target.Close()
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Proxy-Authorization")
		if auth != "Basic "+base64.StdEncoding.EncodeToString([]byte("proxy-user:proxy-pass")) {
			t.Error("proxy auth missing")
			w.WriteHeader(407)
			return
		}
		if r.Method == http.MethodConnect {
			conn, err := net.Dial("tcp", r.Host)
			if err != nil {
				w.WriteHeader(502)
				return
			}
			hijacker := w.(http.Hijacker)
			client, buffer, err := hijacker.Hijack()
			if err != nil {
				conn.Close()
				return
			}
			_, _ = buffer.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
			_ = buffer.Flush()
			go func() { defer client.Close(); defer conn.Close(); _, _ = io.Copy(conn, client) }()
			_, _ = io.Copy(client, conn)
			return
		}
		request := r.Clone(r.Context())
		request.RequestURI = ""
		request.Header.Del("Proxy-Authorization")
		response, err := http.DefaultTransport.RoundTrip(request)
		if err != nil {
			w.WriteHeader(502)
			return
		}
		defer response.Body.Close()
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, response.Body)
	}))
	defer proxyServer.Close()
	connection := domain.ConnectionConfig{ProxyURL: proxyServer.URL, ProxyUsername: "proxy-user", ProxyPasswordSecretRef: "proxy"}
	runner := NewRunner(SecretResolverFunc(func(context.Context, string) (string, error) { return "proxy-pass", nil }))
	m := httpMonitor(target.URL)
	m.HTTP.Connection = connection
	m.HTTP.Assertions.TextContains = []string{"proxied"}
	runMonitor(t, runner, m)
	tcp := tcpMonitor(target.Listener.Addr().String())
	tcp.TCP.Connection = connection
	tcp.TCP.SendText = "GET / HTTP/1.0\r\nHost: target\r\n\r\n"
	tcp.TCP.ReceiveContains = "proxied"
	runMonitor(t, runner, tcp)
}

func TestHTTPUsesConfiguredDNSResolver(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Host, "probe.test:") {
			t.Error("URL Host changed during DNS resolution")
		}
		_, _ = w.Write([]byte("custom dns"))
	}))
	defer target.Close()
	packet, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	server := &dns.Server{PacketConn: packet, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		requests.Add(1)
		reply := new(dns.Msg)
		reply.SetReply(r)
		if r.Question[0].Qtype == dns.TypeA {
			record, _ := dns.NewRR("probe.test. 60 IN A 127.0.0.1")
			reply.Answer = []dns.RR{record}
		}
		_ = w.WriteMsg(reply)
	})}
	go func() { _ = server.ActivateAndServe() }()
	defer server.Shutdown()
	m := httpMonitor(strings.Replace(target.URL, "127.0.0.1", "probe.test", 1))
	m.HTTP.Connection.DNSServer = packet.LocalAddr().String()
	m.HTTP.Assertions.TextContains = []string{"custom dns"}
	runMonitor(t, NewRunner(nil), m)
	if requests.Load() == 0 {
		t.Error("custom resolver unused")
	}
}

func TestHTTPMutualTLSSecretReferences(t *testing.T) {
	serverCertificate := issueFixture(t, time.Now().Add(90*24*time.Hour), false)
	clientCertificate := issueFixture(t, time.Now().Add(90*24*time.Hour), true)
	clientRoots := x509.NewCertPool()
	clientRoots.AppendCertsFromPEM([]byte(clientCertificate.ca))
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TLS.PeerCertificates) != 1 {
			t.Error("client certificate missing")
		}
		_, _ = w.Write([]byte("mutual tls"))
	}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{serverCertificate.pair}, MinVersion: tls.VersionTLS12, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: clientRoots}
	server.StartTLS()
	defer server.Close()
	values := map[string]string{"ca": serverCertificate.ca, "certificate": clientCertificate.cert, "key": clientCertificate.key}
	runner := NewRunner(SecretResolverFunc(func(_ context.Context, id string) (string, error) { return values[id], nil }))
	m := httpMonitor(server.URL)
	m.HTTP.TLS = domain.TLSConfig{CASecretRef: "ca", ClientCertificateSecretRef: "certificate", ClientKeySecretRef: "key"}
	runMonitor(t, runner, m)
	m.HTTP.TLS.ClientCertificateSecretRef = ""
	m.HTTP.TLS.ClientKeySecretRef = ""
	if runner.Run(context.Background(), m).Success {
		t.Error("mTLS server accepted a probe without client certificate")
	}
}

func TestAuthenticatedSOCKSProxyHTTPAndTCP(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("socks ready")) }))
	defer target.Close()
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyListener.Close()
	var remoteNameSeen atomic.Bool
	go func() {
		for {
			conn, e := proxyListener.Accept()
			if e != nil {
				return
			}
			go func() {
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
				var greeting [2]byte
				if _, e := io.ReadFull(conn, greeting[:]); e != nil {
					return
				}
				methods := make([]byte, int(greeting[1]))
				if _, e := io.ReadFull(conn, methods); e != nil {
					return
				}
				_, _ = conn.Write([]byte{5, 2})
				var auth [2]byte
				if _, e := io.ReadFull(conn, auth[:]); e != nil {
					return
				}
				username := make([]byte, int(auth[1]))
				if _, e := io.ReadFull(conn, username); e != nil {
					return
				}
				var passwordSize [1]byte
				if _, e := io.ReadFull(conn, passwordSize[:]); e != nil {
					return
				}
				password := make([]byte, int(passwordSize[0]))
				if _, e := io.ReadFull(conn, password); e != nil {
					return
				}
				if string(username) != "socks-user" || string(password) != "socks-pass" {
					_, _ = conn.Write([]byte{1, 1})
					return
				}
				_, _ = conn.Write([]byte{1, 0})
				var header [4]byte
				if _, e := io.ReadFull(conn, header[:]); e != nil {
					return
				}
				var host string
				switch header[3] {
				case 1:
					address := make([]byte, 4)
					if _, e := io.ReadFull(conn, address); e != nil {
						return
					}
					host = net.IP(address).String()
				case 4:
					address := make([]byte, 16)
					if _, e := io.ReadFull(conn, address); e != nil {
						return
					}
					host = net.IP(address).String()
				case 3:
					var size [1]byte
					if _, e := io.ReadFull(conn, size[:]); e != nil {
						return
					}
					name := make([]byte, int(size[0]))
					if _, e := io.ReadFull(conn, name); e != nil {
						return
					}
					host = string(name)
					if host == "probe.test" {
						remoteNameSeen.Store(true)
						host = "127.0.0.1"
					}
				default:
					return
				}
				var port [2]byte
				if _, e := io.ReadFull(conn, port[:]); e != nil {
					return
				}
				upstream, e := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(int(binary.BigEndian.Uint16(port[:])))), time.Second)
				if e != nil {
					_, _ = conn.Write([]byte{5, 5, 0, 1, 0, 0, 0, 0, 0, 0})
					return
				}
				defer upstream.Close()
				_, _ = conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
				go func() { _, _ = io.Copy(upstream, conn); _ = upstream.Close() }()
				_, _ = io.Copy(conn, upstream)
			}()
		}
	}()
	connection := domain.ConnectionConfig{ProxyURL: "socks5://" + proxyListener.Addr().String(), ProxyUsername: "socks-user", ProxyPasswordSecretRef: "proxy"}
	runner := NewRunner(SecretResolverFunc(func(context.Context, string) (string, error) { return "socks-pass", nil }))
	m := httpMonitor(strings.Replace(target.URL, "127.0.0.1", "probe.test", 1))
	m.HTTP.Connection = connection
	m.HTTP.Assertions.TextContains = []string{"socks ready"}
	runMonitor(t, runner, m)
	tcp := tcpMonitor(target.Listener.Addr().String())
	tcp.TCP.Host = "probe.test"
	tcp.TCP.Connection = connection
	tcp.TCP.SendText = "GET / HTTP/1.0\r\nHost: probe.test\r\n\r\n"
	tcp.TCP.ReceiveContains = "socks ready"
	runMonitor(t, runner, tcp)
	if !remoteNameSeen.Load() {
		t.Fatal("SOCKS target DNS was resolved locally")
	}
	tcp.TCP.Connection.ProxyPasswordSecretRef = "bad"
	runner.Resolver = SecretResolverFunc(func(context.Context, string) (string, error) { return "bad-pass", nil })
	result := runner.Run(context.Background(), tcp)
	if result.Success || strings.Contains(result.Error, "bad-pass") {
		t.Fatal("SOCKS auth failure missing or leaked secret", result)
	}
}
