package beszel

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
)

var (
	ErrDisabled    = errors.New("Beszel integration is disabled")
	ErrAuth        = errors.New("Beszel password authentication failed; verify the account or disable MFA for this integration account")
	ErrPermission  = errors.New("Beszel account cannot access the requested records")
	ErrUnavailable = errors.New("Beszel Hub is unavailable")
	ErrSchema      = errors.New("Beszel response does not match the supported schema")
	ErrVersion     = errors.New("Beszel Hub version is unsupported; adapter supports the 0.20 release family")
	ErrNotFound    = errors.New("Beszel system is unavailable to this account")
)
var systemID = regexp.MustCompile(`^[a-zA-Z0-9]{1,100}$`)

func ValidateConfig(cfg *domain.BeszelConfig) error {
	if cfg.PollSeconds == 0 {
		cfg.PollSeconds = 30
	}
	if cfg.PollSeconds < 30 || cfg.PollSeconds > 3600 {
		return errors.New("Beszel poll interval must be between 30 and 3600 seconds")
	}
	if !cfg.Enabled && cfg.URL == "" {
		return nil
	}
	u, err := url.Parse(strings.TrimSpace(cfg.URL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("Beszel URL must be an HTTP(S) Hub URL without embedded credentials or query")
	}
	cfg.URL = strings.TrimRight(u.String(), "/")
	if cfg.Enabled && (cfg.Email == "" || cfg.PasswordSecretID == "") {
		return errors.New("Beszel requires an account email and password secret reference")
	}
	return nil
}

func (c *Client) request(ctx context.Context, cfg domain.BeszelConfig, method, path string, query url.Values, body any, out any, token string) error {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return ErrSchema
		}
	}
	address := cfg.URL + path
	if len(query) > 0 {
		address += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, address, bytes.NewReader(data))
	if err != nil {
		return ErrUnavailable
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	select {
	case c.remote <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-c.remote }()
	response, err := c.HTTP.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode == 401 || response.StatusCode == 400 && strings.Contains(path, "auth-") {
		return ErrAuth
	}
	if response.StatusCode == 403 {
		return ErrPermission
	}
	if response.StatusCode == 404 {
		return ErrNotFound
	}
	if response.StatusCode != 200 {
		return ErrUnavailable
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 4<<20+1))
	if err != nil {
		return ErrUnavailable
	}
	if len(raw) > 4<<20 {
		return ErrSchema
	}
	if err = json.Unmarshal(raw, out); err != nil {
		return ErrSchema
	}
	return nil
}

func (c *Client) authenticate(ctx context.Context, session *clientSession) (string, error) {
	session.mu.Lock()
	token, until := session.token, session.tokenUntil
	session.mu.Unlock()
	if token != "" && until > c.Now().Add(5*time.Minute).Unix() {
		return token, nil
	}
	value, err := c.sharedWork(ctx, session, &session.auth, "", func(ctx context.Context) (any, error) {
		session.mu.Lock()
		token, until := session.token, session.tokenUntil
		session.mu.Unlock()
		if token != "" && until > c.Now().Add(5*time.Minute).Unix() {
			return token, nil
		}
		cfg := session.config
		var result struct {
			Token  string `json:"token"`
			Record struct {
				CollectionName string `json:"collectionName"`
			} `json:"record"`
		}
		if token != "" {
			if err := c.request(ctx, cfg, "POST", "/api/collections/users/auth-refresh", nil, nil, &result, token); err == nil && result.Token != "" {
				c.installToken(session, result.Token)
				return result.Token, nil
			}
			c.clearToken(session, token)
		}
		password, err := c.Secrets.ResolveSecret(ctx, cfg.PasswordSecretID)
		if err != nil {
			return nil, ErrAuth
		}
		if err = c.request(ctx, cfg, "POST", "/api/collections/users/auth-with-password", nil, map[string]string{"identity": cfg.Email, "password": password}, &result, ""); err != nil {
			return nil, err
		}
		if result.Token == "" || result.Record.CollectionName != "users" {
			return nil, ErrAuth
		}
		c.installToken(session, result.Token)
		return result.Token, nil
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func (c *Client) clearToken(session *clientSession, rejected string) {
	session.mu.Lock()
	if session.token == rejected {
		session.token = ""
		session.tokenUntil = 0
	}
	session.mu.Unlock()
}

func (c *Client) installToken(session *clientSession, token string) {
	until := c.Now().Add(15 * time.Minute).Unix()
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		raw, err := base64.RawURLEncoding.DecodeString(parts[1])
		var claims struct {
			Exp int64 `json:"exp"`
		}
		if err == nil && json.Unmarshal(raw, &claims) == nil && claims.Exp > 0 {
			until = claims.Exp
		}
	}
	session.mu.Lock()
	session.token, session.tokenUntil = token, until
	session.mu.Unlock()
}

func listRecords[T any](ctx context.Context, c *Client, session *clientSession, collection string, query url.Values) ([]T, error) {
	token, err := c.authenticate(ctx, session)
	if err != nil {
		return nil, err
	}
	items := []T{}
	for page := 1; page <= 10; page++ {
		query.Set("page", strconv.Itoa(page))
		query.Set("perPage", "500")
		var response struct {
			Items      []T `json:"items"`
			TotalPages int `json:"totalPages"`
		}
		err := c.request(ctx, session.config, "GET", "/api/collections/"+collection+"/records", query, nil, &response, token)
		if errors.Is(err, ErrAuth) {
			c.clearToken(session, token)
			if token, err = c.authenticate(ctx, session); err != nil {
				return nil, err
			}
			err = c.request(ctx, session.config, "GET", "/api/collections/"+collection+"/records", query, nil, &response, token)
		}
		if err != nil {
			return nil, err
		}
		items = append(items, response.Items...)
		if response.TotalPages <= page {
			return items, nil
		}
	}
	return nil, fmt.Errorf("%w: Beszel record page limit exceeded", ErrSchema)
}

func timestamp(raw json.RawMessage) (int64, error) {
	var number int64
	if json.Unmarshal(raw, &number) == nil {
		return number, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return 0, ErrSchema
	}
	for _, format := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999Z", "2006-01-02 15:04:05.999", "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(format, text); err == nil {
			return parsed.UnixMilli(), nil
		}
	}
	return 0, ErrSchema
}
