package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address                   string
	MetricsAddress            string
	ProbeConcurrency          int
	OperationHistoryDays      int
	Driver                    string
	DSN                       string
	DataDir                   string
	StaticDir                 string
	EncryptionKey             string
	AdminHosts                []string
	CookieSecure              bool
	DBMaxConnections          int
	StatisticsIntervalSeconds int
}

func Load() (Config, error) {
	c := Config{Address: env("OCTOPULSE_ADDR", "127.0.0.1:8080"), Driver: env("OCTOPULSE_DB_DRIVER", "sqlite"), DSN: env("OCTOPULSE_DB_DSN", "file:.data/octopulse.db"), DataDir: env("OCTOPULSE_DATA_DIR", ".data"), StaticDir: env("OCTOPULSE_STATIC_DIR", "web/dist"), EncryptionKey: os.Getenv("OCTOPULSE_ENCRYPTION_KEY")}
	if c.Driver != "sqlite" && c.Driver != "postgres" {
		return c, fmt.Errorf("OCTOPULSE_DB_DRIVER must be sqlite or postgres")
	}
	c.MetricsAddress = os.Getenv("OCTOPULSE_METRICS_ADDR")
	if c.MetricsAddress != "" {
		if _, _, err := net.SplitHostPort(c.MetricsAddress); err != nil {
			return c, fmt.Errorf("invalid OCTOPULSE_METRICS_ADDR: %w", err)
		}
	}
	var e error
	c.OperationHistoryDays, e = integer("OCTOPULSE_OPERATION_HISTORY_DAYS", 0, 0, 3650)
	if e != nil {
		return c, e
	}
	c.ProbeConcurrency, e = integer("OCTOPULSE_PROBE_CONCURRENCY", 100, 1, 1000)
	if e != nil {
		return c, e
	}
	c.DBMaxConnections, e = integer("OCTOPULSE_DB_MAX_CONNECTIONS", 10, 2, 100)
	if e != nil {
		return c, e
	}
	c.StatisticsIntervalSeconds, e = integer("OCTOPULSE_STATISTICS_INTERVAL_SECONDS", 60, 5, 3600)
	if e != nil {
		return c, e
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
func integer(key string, defaultValue, min, max int) (int, error) {
	v := defaultValue
	if raw := os.Getenv(key); raw != "" {
		parsed, e := strconv.Atoi(raw)
		if e != nil {
			return 0, fmt.Errorf("invalid %s", key)
		}
		v = parsed
	}
	if v < min || v > max {
		return 0, fmt.Errorf("%s must be between %d and %d", key, min, max)
	}
	return v, nil
}
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
