package config

import "testing"

func TestOperationalSettings(t *testing.T) {
	t.Setenv("OCTOPULSE_METRICS_ADDR", "127.0.0.1:9090")
	t.Setenv("OCTOPULSE_PROBE_CONCURRENCY", "8")
	t.Setenv("OCTOPULSE_OPERATION_HISTORY_DAYS", "")
	c, err := Load()
	if err != nil || c.MetricsAddress != "127.0.0.1:9090" || c.ProbeConcurrency != 8 || c.OperationHistoryDays != 0 {
		t.Fatalf("unexpected configuration: %+v %v", c, err)
	}
	t.Setenv("OCTOPULSE_OPERATION_HISTORY_DAYS", "30")
	if c, err := Load(); err != nil || c.OperationHistoryDays != 30 {
		t.Fatalf("history retention configuration: %+v %v", c, err)
	}
	t.Setenv("OCTOPULSE_OPERATION_HISTORY_DAYS", "-1")
	if _, err := Load(); err == nil {
		t.Fatal("negative history retention accepted")
	}
	t.Setenv("OCTOPULSE_OPERATION_HISTORY_DAYS", "0")
	t.Setenv("OCTOPULSE_PROBE_CONCURRENCY", "0")
	if _, err := Load(); err == nil {
		t.Fatal("unbounded probe concurrency accepted")
	}
	t.Setenv("OCTOPULSE_PROBE_CONCURRENCY", "8")
	t.Setenv("OCTOPULSE_METRICS_ADDR", "invalid")
	if _, err := Load(); err == nil {
		t.Fatal("invalid metrics address accepted")
	}
}
