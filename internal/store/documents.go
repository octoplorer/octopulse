package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/octoplorer/octopulse/internal/store/postgresquery"
	"github.com/octoplorer/octopulse/internal/store/sqlitequery"
	"modernc.org/sqlite"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && (pg.Code == "23505" || pg.Code == "23503" || pg.Code == "23514") {
		return fmt.Errorf("%w: database constraint", ErrConflict)
	}
	var sq *sqlite.Error
	if errors.As(err, &sq) && sq.Code()&255 == 19 {
		return fmt.Errorf("%w: database constraint", ErrConflict)
	}
	return err
}

func (s *Store) get(ctx context.Context, q dbtx, kind, id string, out any) error {
	var payload string
	var err error
	if s.driver == "sqlite" {
		payload, err = sqlitequery.New(q).GetDocument(ctx, sqlitequery.GetDocumentParams{Kind: kind, ID: id})
	} else {
		payload, err = postgresquery.New(q).GetDocument(ctx, postgresquery.GetDocumentParams{Kind: kind, ID: id})
	}
	if err != nil {
		return mapError(err)
	}
	return json.Unmarshal([]byte(payload), out)
}

func (s *Store) list(ctx context.Context, q dbtx, kind string) ([]json.RawMessage, error) {
	var values []string
	var err error
	if s.driver == "sqlite" {
		values, err = sqlitequery.New(q).ListDocuments(ctx, kind)
	} else {
		values, err = postgresquery.New(q).ListDocuments(ctx, kind)
	}
	if err != nil {
		return nil, mapError(err)
	}
	result := make([]json.RawMessage, 0, len(values))
	for _, v := range values {
		result = append(result, json.RawMessage(v))
	}
	return result, nil
}

func (s *Store) put(ctx context.Context, q dbtx, kind, id string, value any) error {
	if kind == "" || id == "" {
		return fmt.Errorf("record kind and ID are required")
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if s.driver == "sqlite" {
		err = sqlitequery.New(q).PutDocument(ctx, sqlitequery.PutDocumentParams{Kind: kind, ID: id, Payload: string(payload)})
	} else {
		err = postgresquery.New(q).PutDocument(
			ctx,
			postgresquery.PutDocumentParams{
				Kind:    kind,
				ID:      id,
				Payload: string(payload),
			},
		)
	}
	return mapError(err)
}

func (s *Store) delete(ctx context.Context, q dbtx, kind, id string) error {
	var n int64
	var err error
	if s.driver == "sqlite" {
		n, err = sqlitequery.New(q).DeleteDocument(ctx, sqlitequery.DeleteDocumentParams{Kind: kind, ID: id})
	} else {
		n, err = postgresquery.New(q).DeleteDocument(ctx, postgresquery.DeleteDocumentParams{Kind: kind, ID: id})
	}
	if err != nil {
		return mapError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	if kind == "page" || kind == "pages" {
		_, err = q.ExecContext(ctx, s.sql("DELETE FROM page_slugs WHERE page_id=?"), id)
	}
	return mapError(err)
}

// sql converts the project's positional placeholders at the adapter boundary.
// SQL strings are static source text, never user input.
func (s *Store) sql(query string) string {
	if s.driver == "sqlite" {
		return query
	}
	var b strings.Builder
	var n int
	for _, r := range query {
		if r == '?' {
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func boolInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func jsonText(value json.RawMessage) (string, error) {
	if len(value) == 0 {
		return "null", nil
	}
	if !json.Valid(value) {
		return "", fmt.Errorf("invalid JSON payload")
	}
	return string(value), nil
}
