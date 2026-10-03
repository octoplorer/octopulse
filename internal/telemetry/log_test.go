package telemetry

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestErrorDiagnosticsRetainCodeWithoutCredentials(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	err := fmt.Errorf("postgres://admin:top-secret@db: %w", &pgconn.PgError{Code: "40001", Message: "top-secret serialized payload"})
	LogError(WithRequestID(context.Background(), "request-123"), "collection.commit", err)
	got := output.String()
	for _, required := range []string{"40001", "postgres", "request-123", "collection.commit"} {
		if !strings.Contains(got, required) {
			t.Errorf("missing %q in %s", required, got)
		}
	}
	if strings.Contains(got, "top-secret") || strings.Contains(got, "admin:") {
		t.Fatal("credentials leaked into operational log")
	}
}
