package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address       string
	Driver        string
	DSN           string
	DataDir       string
	StaticDir     string
	EncryptionKey string
	AdminHosts    []string
	CookieSecure  bool
}

func Load() (Config, error) {
	c := Config{Address: env("OCTOPULSE_ADDR", "127.0.0.1:8080"), Driver: env("OCTOPULSE_DB_DRIVER", "sqlite"), DSN: env("OCTOPULSE_DB_DSN", "file:.data/octopulse.db"), DataDir: env("OCTOPULSE_DATA_DIR", ".data"), StaticDir: env("OCTOPULSE_STATIC_DIR", "web/dist"), EncryptionKey: os.Getenv("OCTOPULSE_ENCRYPTION_KEY")}
	if c.Driver != "sqlite" && c.Driver != "postgres" {
		return c, fmt.Errorf("OCTOPULSE_DB_DRIVER must be sqlite or postgres")
	}
	if _, _, err := net.SplitHostPort(c.Address); err != nil {
		return c, fmt.Errorf("invalid OCTOPULSE_ADDR: %w", err)
	}
	if v := os.Getenv("OCTOPULSE_COOKIE_SECURE"); v != "" {
		var err error
		c.CookieSecure, err = strconv.ParseBool(v)
		if err != nil {
			return c, fmt.Errorf("invalid OCTOPULSE_COOKIE_SECURE")
		}
	}
	for _, h := range strings.Split(env("OCTOPULSE_ADMIN_HOSTS", "localhost,127.0.0.1"), ",") {
		h = strings.ToLower(strings.TrimSpace(h))
		if h != "" {
			c.AdminHosts = append(c.AdminHosts, h)
		}
	}
	return c, nil
}
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
