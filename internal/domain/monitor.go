package domain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

const (
	MonitorHTTP            = "http"
	MonitorTCP             = "tcp"
	MonitorDNS             = "dns"
	MonitorHeartbeat       = "heartbeat"
	MonitorCertificate     = "certificate"
	StateUnknown           = "unknown"
	StateUp                = "up"
	StateDown              = "down"
	CertificateHealthy     = "healthy"
	CertificateExpiring    = "expiring"
	CertificateExpired     = "expired"
	CertificateCheckFailed = "check_failed"
)

// Monitor includes configuration and confirmed state. All persisted timestamps
// use UTC Unix milliseconds. Config fields contain secret references, never values.
type Monitor struct {
	ID                     string             `json:"id" readOnly:"true"`
	Name                   string             `json:"name"`
	Description            string             `json:"description" required:"false"`
	Type                   string             `json:"type"`
	Tags                   []string           `json:"tags" required:"false"`
	Group                  string             `json:"group" required:"false"`
	Enabled                bool               `json:"enabled" required:"false"`
	IntervalSeconds        int                `json:"intervalSeconds" required:"false"`
	TimeoutSeconds         int                `json:"timeoutSeconds" required:"false"`
	Retries                int                `json:"retries" required:"false"`
	RetryDelaySeconds      int                `json:"retryDelaySeconds" required:"false"`
	FailureThreshold       int                `json:"failureThreshold" required:"false"`
	RecoveryThreshold      int                `json:"recoveryThreshold" required:"false"`
	NotificationChannelIDs []string           `json:"notificationChannelIds" required:"false"`
	NotifyRecovery         bool               `json:"notifyRecovery" required:"false"`
	ReminderSeconds        int                `json:"reminderSeconds" required:"false"`
	ConfigVersion          int64              `json:"configVersion" readOnly:"true"`
	State                  string             `json:"state" readOnly:"true"`
	FailureCount           int                `json:"failureCount" readOnly:"true"`
	SuccessCount           int                `json:"successCount" readOnly:"true"`
	LastCheckedAt          int64              `json:"lastCheckedAt" readOnly:"true"`
	NextCheckAt            int64              `json:"nextCheckAt" readOnly:"true"`
	CreatedAt              int64              `json:"createdAt" readOnly:"true"`
	UpdatedAt              int64              `json:"updatedAt" readOnly:"true"`
	HTTP                   *HTTPConfig        `json:"http,omitempty"`
	TCP                    *TCPConfig         `json:"tcp,omitempty"`
	DNS                    *DNSConfig         `json:"dns,omitempty"`
	Heartbeat              *HeartbeatConfig   `json:"heartbeat,omitempty"`
	Certificate            *CertificateConfig `json:"certificate,omitempty"`
}

// NameValue preserves repeated query/form/header values. SecretRef substitutes
// its entire value during probing and is not returned by diagnostics.
type NameValue struct {
	Name      string `json:"name"`
	Value     string `json:"value" required:"false"`
	SecretRef string `json:"secretRef" required:"false"`
}

type HTTPConfig struct {
	URL              string           `json:"url"`
	Method           string           `json:"method" required:"false"`
	Query            []NameValue      `json:"query" required:"false"`
	Headers          []NameValue      `json:"headers" required:"false"`
	Host             string           `json:"host" required:"false"`
	Body             HTTPBody         `json:"body" required:"false"`
	Auth             HTTPAuth         `json:"auth" required:"false"`
	TLS              TLSConfig        `json:"tls" required:"false"`
	Connection       ConnectionConfig `json:"connection" required:"false"`
	Redirects        RedirectConfig   `json:"redirects" required:"false"`
	AcceptEncoding   string           `json:"acceptEncoding" required:"false"`
	RequestGzip      bool             `json:"requestGzip" required:"false"`
	ResponseCharset  string           `json:"responseCharset" required:"false"`
	MaxResponseBytes int64            `json:"maxResponseBytes" required:"false"`
	Assertions       HTTPAssertions   `json:"assertions" required:"false"`
}

type HTTPBody struct {
	Format      string          `json:"format"` // none, json, form, multipart, text, raw
	Text        string          `json:"text" required:"false"`
	SecretRef   string          `json:"secretRef" required:"false"`
	Base64      string          `json:"base64" required:"false"`
	Charset     string          `json:"charset" required:"false"`
	ContentType string          `json:"contentType" required:"false"`
	Fields      []NameValue     `json:"fields" required:"false"`
	Files       []MultipartFile `json:"files" required:"false"`
}

type MultipartFile struct {
	Field       string `json:"field"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType" required:"false"`
	Base64      string `json:"base64" required:"false"`
	SecretRef   string `json:"secretRef" required:"false"`
}

type HTTPAuth struct {
	Type              string `json:"type"` // none, basic, bearer, header
	Username          string `json:"username" required:"false"`
	UsernameSecretRef string `json:"usernameSecretRef" required:"false"`
	SecretRef         string `json:"secretRef" required:"false"`
	Header            string `json:"header" required:"false"`
	Prefix            string `json:"prefix" required:"false"`
}

type TLSConfig struct {
	Enabled                    bool   `json:"enabled" required:"false"`
	InsecureSkipVerify         bool   `json:"insecureSkipVerify" required:"false"`
	CASecretRef                string `json:"caSecretRef" required:"false"`
	ClientCertificateSecretRef string `json:"clientCertificateSecretRef" required:"false"`
	ClientKeySecretRef         string `json:"clientKeySecretRef" required:"false"`
	ServerName                 string `json:"serverName" required:"false"`
	MinVersion                 string `json:"minVersion"` // 1.2, 1.3
	MaxVersion                 string `json:"maxVersion" required:"false"`
}

type ConnectionConfig struct {
	ProxyURL               string `json:"proxyUrl"` // http, https, socks5, socks5h
	ProxyUsername          string `json:"proxyUsername" required:"false"`
	ProxyPasswordSecretRef string `json:"proxyPasswordSecretRef" required:"false"`
	DNSServer              string `json:"dnsServer"` // host:port; local target resolution except proxy protocol
	FixedIP                string `json:"fixedIp"`   // only direct target dialing; rejected with a proxy
}

type RedirectConfig struct {
	Enabled bool   `json:"enabled" required:"false"`
	MaxHops int    `json:"maxHops" required:"false"`
	Scope   string `json:"scope"` // same-origin (default), same-host, any
}

type StatusRange struct {
	Min int `json:"min" required:"false"`
	Max int `json:"max" required:"false"`
}

type ValueAssertion struct {
	Name     string `json:"name"`
	Operator string `json:"operator"` // exists, equals, contains, not_contains, regex
	Value    string `json:"value" required:"false"`
}

type JSONAssertion struct {
	Pointer  string          `json:"pointer"`  // RFC 6901 JSON Pointer; empty means the document
	Operator string          `json:"operator"` // exists, equals, not_equals, contains, regex
	Value    json.RawMessage `json:"value" required:"false"`
}

type HTTPAssertions struct {
	StatusCodes     []int            `json:"statusCodes" required:"false"`
	StatusRanges    []StatusRange    `json:"statusRanges" required:"false"`
	Headers         []ValueAssertion `json:"headers" required:"false"`
	TextContains    []string         `json:"textContains" required:"false"`
	TextNotContains []string         `json:"textNotContains" required:"false"`
	Regex           []string         `json:"regex" required:"false"`
	JSON            []JSONAssertion  `json:"json" required:"false"`
	MaxLatencyMs    int64            `json:"maxLatencyMs" required:"false"`
}

type TCPConfig struct {
	Host            string           `json:"host"`
	Port            int              `json:"port"`
	TLS             TLSConfig        `json:"tls" required:"false"`
	Connection      ConnectionConfig `json:"connection" required:"false"`
	SendText        string           `json:"sendText" required:"false"`
	SendBase64      string           `json:"sendBase64" required:"false"`
	SendSecretRef   string           `json:"sendSecretRef" required:"false"`
	Charset         string           `json:"charset" required:"false"`
	ReceiveContains string           `json:"receiveContains" required:"false"`
	ReceiveRegex    string           `json:"receiveRegex" required:"false"`
	MaxReceiveBytes int64            `json:"maxReceiveBytes" required:"false"`
}

type DNSConfig struct {
	Name           string   `json:"name"`
	RecordType     string   `json:"recordType"`
	Server         string   `json:"server" required:"false"`
	Protocol       string   `json:"protocol" required:"false"`
	ExpectedRCode  string   `json:"expectedRCode" required:"false"`
	ExpectedValues []string `json:"expectedValues" required:"false"`
	MatchMode      string   `json:"matchMode"` // contains (default), exact
}

type HeartbeatConfig struct {
	SecretHash     string `json:"-"`
	PeriodSeconds  int    `json:"periodSeconds" required:"false"`
	GraceSeconds   int    `json:"graceSeconds" required:"false"`
	LastReceivedAt int64  `json:"lastReceivedAt" readOnly:"true"`
	LastSuccess    bool   `json:"lastSuccess" readOnly:"true"`
	Description    string `json:"description" readOnly:"true"`
}

type CertificateConfig struct {
	Host          string           `json:"host"`
	Port          int              `json:"port"`
	TLS           TLSConfig        `json:"tls" required:"false"`
	Connection    ConnectionConfig `json:"connection" required:"false"`
	WarningDays   []int            `json:"warningDays" required:"false"`
	NotifyRenewal bool             `json:"notifyRenewal" required:"false"`
	State         string           `json:"state" readOnly:"true"`
	ExpiresAt     int64            `json:"expiresAt" readOnly:"true"`
	Fingerprint   string           `json:"fingerprint" readOnly:"true"`
	DaysRemaining float64          `json:"daysRemaining" readOnly:"true"`
}

// Defaults fills omitted defaults; zero retries is deliberately preserved so
// callers can choose no additional attempts. Creation DTOs supply the default 2.
func (m *Monitor) Defaults() {
	if m.IntervalSeconds == 0 {
		if m.Type == MonitorCertificate {
			m.IntervalSeconds = 86400
		} else {
			m.IntervalSeconds = 60
		}
	}
	if m.TimeoutSeconds == 0 {
		m.TimeoutSeconds = 10
	}
	if m.FailureThreshold == 0 {
		m.FailureThreshold = 1
	}
	if m.RecoveryThreshold == 0 {
		m.RecoveryThreshold = 1
	}
	if m.State == "" {
		m.State = StateUnknown
	}
	if m.HTTP != nil {
		if m.HTTP.Method == "" {
			m.HTTP.Method = "GET"
		}
		if m.HTTP.Body.Format == "" {
			m.HTTP.Body.Format = "none"
		}
		if m.HTTP.MaxResponseBytes == 0 {
			m.HTTP.MaxResponseBytes = 2 << 20
		}
		if m.HTTP.Redirects.MaxHops == 0 {
			m.HTTP.Redirects.MaxHops = 5
		}
		if m.HTTP.Redirects.Scope == "" {
			m.HTTP.Redirects.Scope = "same-origin"
		}
	}
	if m.TCP != nil && m.TCP.MaxReceiveBytes == 0 {
		m.TCP.MaxReceiveBytes = 64 << 10
	}
	if m.DNS != nil {
		if m.DNS.RecordType == "" {
			m.DNS.RecordType = "A"
		}
		if m.DNS.Protocol == "" {
			m.DNS.Protocol = "udp"
		}
		if m.DNS.ExpectedRCode == "" {
			m.DNS.ExpectedRCode = "NOERROR"
		}
		if m.DNS.MatchMode == "" {
			m.DNS.MatchMode = "contains"
		}
	}
	if m.Heartbeat != nil && m.Heartbeat.PeriodSeconds == 0 {
		m.Heartbeat.PeriodSeconds = 60
	}
	if m.Certificate != nil {
		if m.Certificate.Port == 0 {
			m.Certificate.Port = 443
		}
		if len(m.Certificate.WarningDays) == 0 {
			m.Certificate.WarningDays = []int{30, 14, 7, 1}
		}
	}
}

func (m Monitor) IsAvailability() bool { return m.Type != MonitorCertificate }
func (m Monitor) IsActive() bool {
	return m.Type == MonitorHTTP || m.Type == MonitorTCP || m.Type == MonitorDNS || m.Type == MonitorCertificate
}

// SecretReferences lets authorization and referential-integrity checks inspect
// every credential reference without serializing or resolving secret values.
func (m Monitor) SecretReferences() []string {
	refs := map[string]bool{}
	add := func(ref string) {
		if ref != "" {
			refs[ref] = true
		}
	}
	addTLS := func(c TLSConfig) {
		add(c.CASecretRef)
		add(c.ClientCertificateSecretRef)
		add(c.ClientKeySecretRef)
	}
	addPairs := func(pairs []NameValue) {
		for _, pair := range pairs {
			add(pair.SecretRef)
		}
	}
	if c := m.HTTP; c != nil {
		addPairs(c.Query)
		addPairs(c.Headers)
		addPairs(c.Body.Fields)
		add(c.Body.SecretRef)
		for _, file := range c.Body.Files {
			add(file.SecretRef)
		}
		add(c.Auth.SecretRef)
		add(c.Auth.UsernameSecretRef)
		addTLS(c.TLS)
		add(c.Connection.ProxyPasswordSecretRef)
	}
	if c := m.TCP; c != nil {
		add(c.SendSecretRef)
		addTLS(c.TLS)
		add(c.Connection.ProxyPasswordSecretRef)
	}
	if c := m.Certificate; c != nil {
		addTLS(c.TLS)
		add(c.Connection.ProxyPasswordSecretRef)
	}
	result := make([]string, 0, len(refs))
	for ref := range refs {
		result = append(result, ref)
	}
	sort.Strings(result)
	return result
}

func (m Monitor) Validate() error {
	if strings.TrimSpace(m.Name) == "" || len(m.Name) > 200 {
		return errors.New("monitor name must be between 1 and 200 characters")
	}
	if m.IntervalSeconds < 30 || m.IntervalSeconds > 2592000 {
		return errors.New("interval must be between 30 seconds and 30 days")
	}
	if m.TimeoutSeconds < 1 || m.TimeoutSeconds > m.IntervalSeconds {
		return errors.New("timeout must fit the check interval")
	}
	invalidRetries := m.Retries < 0 || m.Retries > 10
	invalidRetryDelay := m.RetryDelaySeconds < 0 || m.RetryDelaySeconds > m.IntervalSeconds
	if invalidRetries || invalidRetryDelay {
		return errors.New("invalid retry budget")
	}
	invalidFailureThreshold := m.FailureThreshold < 1 || m.FailureThreshold > 100
	invalidRecoveryThreshold := m.RecoveryThreshold < 1 || m.RecoveryThreshold > 100
	if invalidFailureThreshold || invalidRecoveryThreshold {
		return errors.New("confirmation thresholds must be between 1 and 100")
	}
	if m.ReminderSeconds != 0 && m.ReminderSeconds < 30 {
		return errors.New("reminder interval must be zero or at least 30 seconds")
	}
	var count int
	for _, present := range []bool{m.HTTP != nil, m.TCP != nil, m.DNS != nil, m.Heartbeat != nil, m.Certificate != nil} {
		if present {
			count++
		}
	}
	if count != 1 {
		return errors.New("exactly one monitor configuration is required")
	}
	switch m.Type {
	case MonitorHTTP:
		if m.HTTP == nil {
			return errors.New("http configuration required")
		}
		return m.HTTP.Validate()
	case MonitorTCP:
		if m.TCP == nil {
			return errors.New("tcp configuration required")
		}
		c := m.TCP
		if err := validateTarget(c.Host, c.Port); err != nil {
			return err
		}
		if c.SendText != "" && c.SendBase64 != "" {
			return errors.New("choose text or binary TCP payload")
		}
		if c.SendBase64 != "" {
			if _, err := base64.StdEncoding.DecodeString(c.SendBase64); err != nil {
				return errors.New("invalid TCP payload base64")
			}
		}
		if c.ReceiveRegex != "" {
			if _, err := regexp.Compile(c.ReceiveRegex); err != nil {
				return errors.New("invalid receive regex")
			}
		}
		if c.MaxReceiveBytes < 1 || c.MaxReceiveBytes > 16<<20 {
			return errors.New("TCP receive limit must be 1 byte to 16 MiB")
		}
		if err := validateCharset(c.Charset); err != nil {
			return err
		}
		if err := c.TLS.Validate(); err != nil {
			return err
		}
		return c.Connection.Validate()
	case MonitorDNS:
		if m.DNS == nil {
			return errors.New("dns configuration required")
		}
		c := m.DNS
		invalidNameLength := c.Name == "" || len(c.Name) > 253
		if invalidNameLength || strings.ContainsAny(c.Name, " \t\r\n") {
			return errors.New("invalid DNS query name")
		}
		if !contains(
			[]string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SRV", "PTR", "SOA", "CAA"},
			strings.ToUpper(c.RecordType),
		) {
			return errors.New("unsupported DNS record type")
		}
		if c.Protocol != "udp" && c.Protocol != "tcp" {
			return errors.New("DNS protocol must be udp or tcp")
		}
		if c.Server != "" {
			if _, _, err := net.SplitHostPort(c.Server); err != nil {
				return errors.New("DNS server must be host:port")
			}
		}
		if !contains(
			[]string{"NOERROR", "FORMERR", "SERVFAIL", "NXDOMAIN", "NOTIMP", "REFUSED"},
			strings.ToUpper(c.ExpectedRCode),
		) {
			return errors.New("unsupported DNS response code")
		}
		if c.MatchMode != "contains" && c.MatchMode != "exact" {
			return errors.New("DNS match mode must be contains or exact")
		}
	case MonitorHeartbeat:
		if m.Heartbeat == nil {
			return errors.New("heartbeat configuration required")
		}
		if m.Heartbeat.PeriodSeconds < 30 || m.Heartbeat.GraceSeconds < 0 {
			return errors.New("invalid heartbeat period or grace")
		}
	case MonitorCertificate:
		if m.Certificate == nil {
			return errors.New("certificate configuration required")
		}
		c := m.Certificate
		if err := validateTarget(c.Host, c.Port); err != nil {
			return err
		}
		for _, d := range c.WarningDays {
			if d < 1 || d > 3650 {
				return errors.New("invalid certificate warning day")
			}
		}
		if err := c.TLS.Validate(); err != nil {
			return err
		}
		return c.Connection.Validate()
	default:
		return errors.New("unsupported monitor type")
	}
	return nil
}

func (c HTTPConfig) Validate() error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return errors.New("HTTP URL must be http(s) with no embedded credentials")
	}
	invalidScheme := u.Scheme != "http" && u.Scheme != "https"
	invalidAuthority := u.Hostname() == "" || u.User != nil
	if invalidScheme || invalidAuthority {
		return errors.New("HTTP URL must be http(s) with no embedded credentials")
	}
	if c.Method == "" || !regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$").MatchString(c.Method) {
		return errors.New("invalid request method")
	}
	if strings.ContainsAny(c.Host, "\r\n") {
		return errors.New("invalid Host override")
	}
	for _, h := range c.Headers {
		if !regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$").MatchString(h.Name) || strings.ContainsAny(h.Value, "\r\n") {
			return errors.New("invalid request header")
		}
		if c.Body.Format == "multipart" && strings.EqualFold(h.Name, "Content-Type") {
			return errors.New("multipart Content-Type is generated automatically")
		}
	}
	if !contains([]string{"none", "json", "form", "multipart", "text", "raw"}, c.Body.Format) {
		return errors.New("unsupported body format")
	}
	hasInlineJSONBody := c.Body.Format == "json" && c.Body.SecretRef == ""
	if hasInlineJSONBody && !json.Valid([]byte(c.Body.Text)) {
		return errors.New("invalid JSON request body")
	}
	hasJSONCharset := c.Body.Format == "json" && c.Body.Charset != ""
	if hasJSONCharset && !strings.EqualFold(c.Body.Charset, "utf-8") {
		return errors.New("JSON charset must be UTF-8")
	}
	if c.Body.Format == "raw" && c.Body.SecretRef == "" {
		if _, err := base64.StdEncoding.DecodeString(c.Body.Base64); err != nil {
			return errors.New("invalid raw request base64")
		}
	}
	for _, f := range c.Body.Files {
		missingFileMetadata := f.Field == "" || f.Filename == ""
		if missingFileMetadata || strings.ContainsAny(f.Filename+f.Field, "\r\n") {
			return errors.New("invalid multipart file")
		}
		if f.SecretRef == "" {
			if _, err := base64.StdEncoding.DecodeString(f.Base64); err != nil {
				return errors.New("invalid multipart base64")
			}
		}
	}
	if err := validateCharset(c.Body.Charset); err != nil {
		return err
	}
	if err := validateCharset(c.ResponseCharset); err != nil {
		return err
	}
	switch c.AcceptEncoding {
	case "", "gzip", "identity":
	default:
		return errors.New("Accept-Encoding must be gzip or identity")
	}
	if c.MaxResponseBytes < 1 || c.MaxResponseBytes > 16<<20 {
		return errors.New("response limit must be 1 byte to 16 MiB")
	}
	if !contains([]string{"", "none", "basic", "bearer", "header"}, c.Auth.Type) {
		return errors.New("unsupported authentication")
	}
	requiresAuthSecret := c.Auth.Type != "" && c.Auth.Type != "none"
	if requiresAuthSecret && c.Auth.SecretRef == "" {
		return errors.New("authentication requires a secret reference")
	}
	if c.Auth.Type == "header" && !regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$").MatchString(c.Auth.Header) {
		return errors.New("invalid authentication header")
	}
	if err := c.TLS.Validate(); err != nil {
		return err
	}
	if err := c.Connection.Validate(); err != nil {
		return err
	}
	if c.Redirects.MaxHops < 1 || c.Redirects.MaxHops > 20 {
		return errors.New("redirect max hops must be 1 to 20")
	}
	if !contains([]string{"same-origin", "same-host", "any"}, c.Redirects.Scope) {
		return errors.New("invalid redirect scope")
	}
	for _, code := range c.Assertions.StatusCodes {
		if code < 100 || code > 599 {
			return errors.New("invalid expected status code")
		}
	}
	for _, r := range c.Assertions.StatusRanges {
		outsideStatusBounds := r.Min < 100 || r.Max > 599
		if outsideStatusBounds || r.Max < r.Min {
			return errors.New("invalid status range")
		}
	}
	for _, r := range c.Assertions.Regex {
		if _, err := regexp.Compile(r); err != nil {
			return errors.New("invalid response regex")
		}
	}
	for _, a := range c.Assertions.Headers {
		if !contains([]string{"exists", "equals", "contains", "not_contains", "regex"}, a.Operator) || a.Name == "" {
			return errors.New("invalid response header assertion")
		}
		if a.Operator == "regex" {
			if _, err := regexp.Compile(a.Value); err != nil {
				return errors.New("invalid header regex")
			}
		}
	}
	for _, a := range c.Assertions.JSON {
		if a.Pointer != "" && !strings.HasPrefix(a.Pointer, "/") {
			return errors.New("JSON pointer must be empty or start with /")
		}
		if !contains([]string{"exists", "equals", "not_equals", "contains", "regex"}, a.Operator) {
			return errors.New("invalid JSON assertion")
		}
		if a.Operator != "exists" && !json.Valid(a.Value) {
			return errors.New("JSON assertion value must be valid JSON")
		}
		if a.Operator == "regex" {
			var s string
			if json.Unmarshal(a.Value, &s) != nil {
				return errors.New("JSON regex value must be a JSON string")
			}
			if _, err := regexp.Compile(s); err != nil {
				return errors.New("invalid JSON regex")
			}
		}
	}
	if c.Assertions.MaxLatencyMs < 0 {
		return errors.New("latency limit must be nonnegative")
	}
	return nil
}

func (c TLSConfig) Validate() error {
	if (c.ClientCertificateSecretRef == "") != (c.ClientKeySecretRef == "") {
		return errors.New("client certificate and key references are required together")
	}
	for _, v := range []string{c.MinVersion, c.MaxVersion} {
		switch v {
		case "", "1.2", "1.3":
		default:
			return errors.New("TLS version must be 1.2 or 1.3")
		}
	}
	hasVersionRange := c.MinVersion != "" && c.MaxVersion != ""
	if hasVersionRange && c.MinVersion > c.MaxVersion {
		return errors.New("TLS minimum exceeds maximum")
	}
	return nil
}

func (c ConnectionConfig) Validate() error {
	if c.FixedIP != "" && net.ParseIP(c.FixedIP) == nil {
		return errors.New("fixed IP must be a valid IP address")
	}
	if c.DNSServer != "" {
		if _, _, err := net.SplitHostPort(c.DNSServer); err != nil {
			return errors.New("custom DNS server must be host:port")
		}
	}
	if c.ProxyURL != "" {
		u, err := url.Parse(c.ProxyURL)
		if err != nil {
			return errors.New("proxy URL must use http(s)/socks5(h) without embedded credentials")
		}
		invalidAuthority := u.Hostname() == "" || u.User != nil
		if invalidAuthority || !contains([]string{"http", "https", "socks5", "socks5h"}, u.Scheme) {
			return errors.New("proxy URL must use http(s)/socks5(h) without embedded credentials")
		}
		if c.FixedIP != "" {
			return errors.New("fixed IP cannot be combined with a proxy")
		}
	}
	return nil
}

func validateTarget(host string, port int) error {
	invalidHost := host == "" || strings.ContainsAny(host, " \t\r\n/")
	invalidPort := port < 1 || port > 65535
	if invalidHost || invalidPort {
		return errors.New("target requires a host and a port between 1 and 65535")
	}
	return nil
}
func validateCharset(charset string) error {
	if !contains(
		[]string{"", "utf-8", "utf8", "iso-8859-1", "windows-1252", "gbk", "gb18030", "shift_jis", "big5"},
		strings.ToLower(charset),
	) {
		return fmt.Errorf("unsupported charset %q", charset)
	}
	return nil
}
func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
