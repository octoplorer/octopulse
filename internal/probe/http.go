package probe

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/octoplorer/octopulse/internal/domain"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

func encodeText(text, charset string) ([]byte, error) {
	if charset == "" || strings.EqualFold(charset, "utf-8") || strings.EqualFold(charset, "utf8") {
		if !utf8.ValidString(text) {
			return nil, errors.New("invalid UTF-8 request text")
		}
		return []byte(text), nil
	}
	encoding, err := htmlindex.Get(charset)
	if err != nil {
		return nil, errors.New("unsupported request charset")
	}
	result, _, err := transform.Bytes(encoding.NewEncoder(), []byte(text))
	if err != nil {
		return nil, errors.New("request contains characters not representable in charset")
	}
	return result, nil
}
func decodeText(data []byte, charset string) (string, error) {
	if charset == "" || strings.EqualFold(charset, "utf-8") || strings.EqualFold(charset, "utf8") {
		return string(data), nil
	}
	encoding, err := htmlindex.Get(charset)
	if err != nil {
		return "", errors.New("unsupported response charset")
	}
	result, _, err := transform.Bytes(encoding.NewDecoder(), data)
	if err != nil {
		return "", errors.New("response charset decoding failed")
	}
	return string(result), nil
}

func (x *execution) body(ctx context.Context, c domain.HTTPBody) ([]byte, string, error) {
	text := c.Text
	if c.SecretRef != "" {
		var err error
		text, err = x.secret(ctx, c.SecretRef)
		if err != nil {
			return nil, "", err
		}
	}
	charset := c.Charset
	if charset == "" {
		charset = "utf-8"
	}
	switch c.Format {
	case "none", "":
		return nil, "", nil
	case "json":
		if !json.Valid([]byte(text)) {
			return nil, "", errors.New("invalid JSON request body")
		}
		return []byte(text), "application/json", nil
	case "text":
		data, err := encodeText(text, charset)
		return data, mime.FormatMediaType("text/plain", map[string]string{"charset": charset}), err
	case "raw":
		if c.SecretRef != "" {
			return []byte(text), "application/octet-stream", nil
		}
		data, err := base64.StdEncoding.DecodeString(c.Base64)
		return data, "application/octet-stream", err
	case "form":
		values := url.Values{}
		for _, field := range c.Fields {
			value, err := x.value(ctx, field)
			if err != nil {
				return nil, "", err
			}
			encodedName, err := encodeText(field.Name, charset)
			if err != nil {
				return nil, "", err
			}
			encodedValue, err := encodeText(value, charset)
			if err != nil {
				return nil, "", err
			}
			values.Add(string(encodedName), string(encodedValue))
		}
		return []byte(values.Encode()), mime.FormatMediaType("application/x-www-form-urlencoded", map[string]string{"charset": charset}), nil
	case "multipart":
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		for _, field := range c.Fields {
			value, err := x.value(ctx, field)
			if err != nil {
				return nil, "", err
			}
			data, err := encodeText(value, charset)
			if err != nil {
				return nil, "", err
			}
			header := textproto.MIMEHeader{}
			header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": field.Name}))
			header.Set("Content-Type", mime.FormatMediaType("text/plain", map[string]string{"charset": charset}))
			part, err := writer.CreatePart(header)
			if err != nil {
				return nil, "", err
			}
			if _, err = part.Write(data); err != nil {
				return nil, "", err
			}
		}
		for _, file := range c.Files {
			var data []byte
			var err error
			if file.SecretRef != "" {
				value, e := x.secret(ctx, file.SecretRef)
				data = []byte(value)
				err = e
			} else {
				data, err = base64.StdEncoding.DecodeString(file.Base64)
			}
			if err != nil {
				return nil, "", err
			}
			header := textproto.MIMEHeader{}
			header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": file.Field, "filename": file.Filename}))
			contentType := file.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			if strings.ContainsAny(contentType, "\r\n") {
				return nil, "", errors.New("invalid multipart content type")
			}
			header.Set("Content-Type", contentType)
			part, err := writer.CreatePart(header)
			if err != nil {
				return nil, "", err
			}
			if _, err = part.Write(data); err != nil {
				return nil, "", err
			}
		}
		if err := writer.Close(); err != nil {
			return nil, "", err
		}
		return buffer.Bytes(), writer.FormDataContentType(), nil
	}
	return nil, "", errors.New("unsupported request body")
}

func (x *execution) http(ctx context.Context, c domain.HTTPConfig) (Result, error) {
	start := time.Now()
	result := Result{}
	u, err := url.Parse(c.URL)
	if err != nil {
		return result, errors.New("invalid target URL")
	}
	query := u.Query()
	for _, pair := range c.Query {
		value, e := x.value(ctx, pair)
		if e != nil {
			return result, e
		}
		query.Add(pair.Name, value)
	}
	u.RawQuery = query.Encode()
	body, contentType, err := x.body(ctx, c.Body)
	if err != nil {
		return result, err
	}
	if c.Body.ContentType != "" && c.Body.Format != "multipart" {
		contentType = c.Body.ContentType
	}
	if c.RequestGzip {
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		if _, err = writer.Write(body); err != nil {
			return result, err
		}
		if err = writer.Close(); err != nil {
			return result, err
		}
		body = buffer.Bytes()
	}
	req, err := http.NewRequestWithContext(ctx, c.Method, u.String(), bytes.NewReader(body))
	if err != nil {
		return result, errors.New("invalid HTTP request")
	}
	for _, pair := range c.Headers {
		value, e := x.value(ctx, pair)
		if e != nil {
			return result, e
		}
		if strings.ContainsAny(value, "\r\n") {
			return result, errors.New("request header contains an invalid value")
		}
		if strings.EqualFold(pair.Name, "Host") {
			req.Host = value
		} else {
			req.Header.Add(pair.Name, value)
		}
	}
	if c.Host != "" {
		req.Host = c.Host
	}
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.RequestGzip {
		req.Header.Set("Content-Encoding", "gzip")
	}
	if c.AcceptEncoding != "" {
		req.Header.Set("Accept-Encoding", c.AcceptEncoding)
	} else if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	if c.Auth.Type != "" && c.Auth.Type != "none" {
		value, e := x.secret(ctx, c.Auth.SecretRef)
		if e != nil {
			return result, e
		}
		switch c.Auth.Type {
		case "basic":
			username := c.Auth.Username
			if c.Auth.UsernameSecretRef != "" {
				username, e = x.secret(ctx, c.Auth.UsernameSecretRef)
				if e != nil {
					return result, e
				}
			}
			req.SetBasicAuth(username, value)
		case "bearer":
			req.Header.Set("Authorization", "Bearer "+value)
		case "header":
			req.Header.Set(c.Auth.Header, c.Auth.Prefix+value)
		}
	}
	tlsConfig, err := x.tlsConfig(ctx, c.TLS, "")
	if err != nil {
		return result, err
	}
	proxyURL, err := x.proxyURL(ctx, c.Connection)
	if err != nil {
		return result, err
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig, DisableCompression: true, ForceAttemptHTTP2: true, MaxIdleConns: 2, IdleConnTimeout: 5 * time.Second, ResponseHeaderTimeout: 0, DialContext: func(dctx context.Context, network, address string) (net.Conn, error) {
		return dialDirect(dctx, c.Connection, network, address)
	}}
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	defer transport.CloseIdleConnections()
	initial := *u
	credentialHeaders := map[string]bool{"Authorization": true, "Cookie": true, "Proxy-Authorization": true}
	if c.Auth.Type == "header" {
		credentialHeaders[http.CanonicalHeaderKey(c.Auth.Header)] = true
	}
	for _, h := range c.Headers {
		if h.SecretRef != "" {
			credentialHeaders[http.CanonicalHeaderKey(h.Name)] = true
		}
	}
	client := &http.Client{Transport: transport, CheckRedirect: func(next *http.Request, via []*http.Request) error {
		if !c.Redirects.Enabled {
			return http.ErrUseLastResponse
		}
		if len(via) > c.Redirects.MaxHops {
			return errors.New("redirect hop limit exceeded")
		}
		switch c.Redirects.Scope {
		case "same-origin":
			if !sameOrigin(&initial, next.URL) {
				return errors.New("redirect outside permitted origin")
			}
		case "same-host":
			if !strings.EqualFold(initial.Hostname(), next.URL.Hostname()) {
				return errors.New("redirect outside permitted host")
			}
		}
		// net/http copies initial headers anew on each redirect. Compare every
		// destination to the initial origin, otherwise a second redirect within
		// the foreign origin could accidentally restore a custom secret header.
		if !sameOrigin(&initial, next.URL) {
			for name := range credentialHeaders {
				next.Header.Del(name)
			}
			next.Host = ""
		} else if c.Host != "" {
			next.Host = c.Host
		}
		if next.URL.User != nil {
			return errors.New("redirect URL contains credentials")
		}
		result.Diagnostics.Redirects = len(via)
		return nil
	}}
	var traceMu sync.Mutex
	var dnsStart, connectStart, tlsStart time.Time
	trace := &httptrace.ClientTrace{DNSStart: func(httptrace.DNSStartInfo) { traceMu.Lock(); dnsStart = time.Now(); traceMu.Unlock() }, DNSDone: func(httptrace.DNSDoneInfo) {
		traceMu.Lock()
		if !dnsStart.IsZero() {
			result.Diagnostics.DNSMs += time.Since(dnsStart).Milliseconds()
		}
		traceMu.Unlock()
	}, ConnectStart: func(string, string) { traceMu.Lock(); connectStart = time.Now(); traceMu.Unlock() }, ConnectDone: func(string, string, error) {
		traceMu.Lock()
		if !connectStart.IsZero() {
			result.Diagnostics.ConnectMs += time.Since(connectStart).Milliseconds()
		}
		traceMu.Unlock()
	}, TLSHandshakeStart: func() { traceMu.Lock(); tlsStart = time.Now(); traceMu.Unlock() }, TLSHandshakeDone: func(tls.ConnectionState, error) {
		traceMu.Lock()
		if !tlsStart.IsZero() {
			result.Diagnostics.TLSMs += time.Since(tlsStart).Milliseconds()
		}
		traceMu.Unlock()
	}, GotFirstResponseByte: func() {
		traceMu.Lock()
		if result.Diagnostics.FirstByteMs == 0 {
			result.Diagnostics.FirstByteMs = time.Since(start).Milliseconds()
		}
		traceMu.Unlock()
	}}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	response, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	result.Diagnostics.StatusCode = response.StatusCode
	result.Diagnostics.FinalURL = redactURL(response.Request.URL)
	if response.TLS != nil {
		result.Diagnostics.TLSVersion = tlsVersion(response.TLS.Version)
	}
	var reader io.Reader = response.Body
	switch strings.ToLower(strings.TrimSpace(response.Header.Get("Content-Encoding"))) {
	case "gzip":
		decompressor, e := gzip.NewReader(response.Body)
		if e != nil {
			return result, errors.New("invalid gzip response")
		}
		defer decompressor.Close()
		reader = decompressor
	case "", "identity":
	default:
		return result, errors.New("unsupported response Content-Encoding")
	}
	data, err := io.ReadAll(io.LimitReader(reader, c.MaxResponseBytes+1))
	if err != nil {
		return result, errors.New("response body read failed")
	}
	if int64(len(data)) > c.MaxResponseBytes {
		return result, errors.New("response exceeds decompressed size limit")
	}
	result.Diagnostics.ResponseBytes = int64(len(data))
	charset := c.ResponseCharset
	if charset == "" {
		_, params, e := mime.ParseMediaType(response.Header.Get("Content-Type"))
		if e == nil {
			charset = params["charset"]
		}
	}
	text, err := decodeText(data, charset)
	if err != nil {
		return result, err
	}
	if int64(len(text)) > c.MaxResponseBytes {
		return result, errors.New("response exceeds decoded size limit")
	}
	if err = httpAssertions(response, text, time.Since(start).Milliseconds(), c.Assertions); err != nil {
		return result, err
	}
	result.Success = true
	return result, nil
}

func httpAssertions(response *http.Response, text string, latency int64, a domain.HTTPAssertions) error {
	statusOK := false
	if len(a.StatusCodes) == 0 && len(a.StatusRanges) == 0 {
		statusOK = response.StatusCode >= 200 && response.StatusCode < 400
	} else {
		for _, code := range a.StatusCodes {
			if response.StatusCode == code {
				statusOK = true
			}
		}
		for _, r := range a.StatusRanges {
			if response.StatusCode >= r.Min && response.StatusCode <= r.Max {
				statusOK = true
			}
		}
	}
	if !statusOK {
		return errors.New("HTTP status assertion failed")
	}
	for i, assertion := range a.Headers {
		values := response.Header.Values(assertion.Name)
		if assertion.Operator == "exists" {
			if len(values) == 0 {
				return assertionError("header", i)
			}
			continue
		}
		matched := false
		for _, v := range values {
			if valueMatches(v, assertion.Operator, assertion.Value) {
				matched = true
			}
		}
		if assertion.Operator == "not_contains" {
			matched = true
			for _, v := range values {
				if strings.Contains(v, assertion.Value) {
					matched = false
				}
			}
		}
		if !matched {
			return assertionError("header", i)
		}
	}
	for i, s := range a.TextContains {
		if !strings.Contains(text, s) {
			return assertionError("text contains", i)
		}
	}
	for i, s := range a.TextNotContains {
		if strings.Contains(text, s) {
			return assertionError("text excludes", i)
		}
	}
	for i, s := range a.Regex {
		r, err := regexp.Compile(s)
		if err != nil || !r.MatchString(text) {
			return assertionError("regex", i)
		}
	}
	if len(a.JSON) > 0 {
		var document any
		decoder := json.NewDecoder(strings.NewReader(text))
		decoder.UseNumber()
		if err := decoder.Decode(&document); err != nil {
			return errors.New("response is not valid JSON")
		}
		if decoder.Decode(new(any)) != io.EOF {
			return errors.New("response contains trailing JSON data")
		}
		for i, assertion := range a.JSON {
			actual, ok := jsonPointer(document, assertion.Pointer)
			if assertion.Operator == "exists" {
				if !ok {
					return assertionError("JSON", i)
				}
				continue
			}
			if !ok {
				return assertionError("JSON", i)
			}
			var expected any
			decoder := json.NewDecoder(bytes.NewReader(assertion.Value))
			decoder.UseNumber()
			if err := decoder.Decode(&expected); err != nil {
				return assertionError("JSON", i)
			}
			matched := false
			switch assertion.Operator {
			case "equals":
				matched = equalJSON(actual, expected)
			case "not_equals":
				matched = !equalJSON(actual, expected)
			case "contains", "regex":
				left, leftOK := actual.(string)
				right, rightOK := expected.(string)
				matched = leftOK && rightOK && valueMatches(left, assertion.Operator, right)
			}
			if !matched {
				return assertionError("JSON", i)
			}
		}
	}
	if a.MaxLatencyMs > 0 && latency > a.MaxLatencyMs {
		return errors.New("response latency assertion failed")
	}
	return nil
}
func valueMatches(value, operator, expected string) bool {
	switch operator {
	case "equals":
		return value == expected
	case "contains":
		return strings.Contains(value, expected)
	case "not_contains":
		return !strings.Contains(value, expected)
	case "regex":
		r, err := regexp.Compile(expected)
		return err == nil && r.MatchString(value)
	}
	return false
}
func jsonPointer(document any, pointer string) (any, bool) {
	if pointer == "" {
		return document, true
	}
	current := document
	for _, part := range strings.Split(pointer[1:], "/") {
		for i := 0; i < len(part); i++ {
			if part[i] == '~' && (i+1 >= len(part) || (part[i+1] != '0' && part[i+1] != '1')) {
				return nil, false
			}
			if part[i] == '~' {
				i++
			}
		}
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, false
			}
		case []any:
			if part == "" || (len(part) > 1 && part[0] == '0') {
				return nil, false
			}
			for _, digit := range part {
				if digit < '0' || digit > '9' {
					return nil, false
				}
			}
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(v) {
				return nil, false
			}
			current = v[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func equalJSON(a, b any) bool {
	switch left := a.(type) {
	case json.Number:
		right, ok := b.(json.Number)
		if !ok {
			return false
		}
		if left == right {
			return true
		}
		lc, le, lok := canonicalNumber(left)
		rc, re, rok := canonicalNumber(right)
		return lok && rok && lc == rc && le == re
	case map[string]any:
		right, ok := b.(map[string]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for key, value := range left {
			other, present := right[key]
			if !present || !equalJSON(value, other) {
				return false
			}
		}
		return true
	case []any:
		right, ok := b.([]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for i, value := range left {
			if !equalJSON(value, right[i]) {
				return false
			}
		}
		return true
	case string:
		right, ok := b.(string)
		return ok && left == right
	case bool:
		right, ok := b.(bool)
		return ok && left == right
	case nil:
		return b == nil
	}
	return false
}

// Compare numeric JSON values without expanding an attacker-controlled exponent
// into a gigantic integer (for example 1e999999999). Significant digits and a
// decimal exponent are sufficient for exact equality.
func canonicalNumber(number json.Number) (string, string, bool) {
	value := string(number)
	sign := ""
	if strings.HasPrefix(value, "-") {
		sign = "-"
		value = value[1:]
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == 'e' || r == 'E' })
	exponent := new(big.Int)
	if len(parts) == 2 {
		if len(parts[1]) > 128 {
			return "", "", false
		}
		if _, ok := exponent.SetString(parts[1], 10); !ok {
			return "", "", false
		}
	}
	coefficient := parts[0]
	decimal := strings.IndexByte(coefficient, '.')
	fractional := 0
	if decimal >= 0 {
		fractional = len(coefficient) - decimal - 1
		coefficient = coefficient[:decimal] + coefficient[decimal+1:]
	}
	coefficient = strings.TrimLeft(coefficient, "0")
	if coefficient == "" {
		return "0", "0", true
	}
	trimmed := strings.TrimRight(coefficient, "0")
	exponent.Add(exponent, big.NewInt(int64(len(coefficient)-len(trimmed)-fractional)))
	return sign + trimmed, exponent.String(), true
}
