package telemetry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
	"modernc.org/sqlite"
)

type requestKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestKey{}, id)
}

// LogError deliberately excludes driver messages, SQL parameters, paths and
// network URLs: they may contain secret values. Error classes and database
// codes retain actionable causes without exposing those values.
func LogError(ctx context.Context, operation string, err error) {
	if err == nil {
		return
	}
	attrs := []slog.Attr{slog.String("operation", operation)}
	if id, ok := ctx.Value(requestKey{}).(string); ok {
		attrs = append(attrs, slog.String("request_id", id))
	}
	leaf := err
	for errors.Unwrap(leaf) != nil {
		leaf = errors.Unwrap(leaf)
	}
	attrs = append(attrs, slog.String("error_type", fmt.Sprintf("%T", leaf)))
	var pg *pgconn.PgError
	var sq *sqlite.Error
	var network net.Error
	var path *os.PathError
	level, message := slog.LevelError, "operation failed"
	switch {
	case errors.Is(err, context.Canceled):
		// Client disconnects and normal shutdown are not platform failures.
		level, message = slog.LevelDebug, "operation cancelled"
		attrs = append(attrs, slog.String("error_class", "cancelled"))
	case errors.Is(err, context.DeadlineExceeded):
		attrs = append(attrs, slog.String("error_class", "deadline_exceeded"))
	case errors.Is(err, sql.ErrConnDone):
		attrs = append(attrs, slog.String("error_class", "connection_closed"))
	case errors.As(err, &pg):
		attrs = append(attrs, slog.String("error_class", "postgres"), slog.String("error_code", pg.Code))
	case errors.As(err, &sq):
		attrs = append(attrs, slog.String("error_class", "sqlite"), slog.Int("error_code", sq.Code()))
	case errors.As(err, &network) && network.Timeout():
		attrs = append(attrs, slog.String("error_class", "network_timeout"))
	case errors.As(err, &path):
		attrs = append(attrs, slog.String("error_class", "filesystem"), slog.String("filesystem_operation", path.Op))
	default:
		attrs = append(attrs, slog.String("error_class", "operation_failed"))
	}
	slog.LogAttrs(
		ctx,
		level,
		message,
		attrs...,
	)
}
