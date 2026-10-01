package store

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Store struct {
	write  *sql.DB
	read   *sql.DB
	driver string
	lock   *instanceLock
}

type Tx struct {
	s  *Store
	tx *sql.Tx
}

// WithTx commits all mutations together and rolls back on any error or panic.
func (s *Store) WithTx(ctx context.Context, fn func(*Tx) error) error {
	tx, err := s.write.BeginTx(ctx, nil)
	if err != nil {
		return mapError(err)
	}
	defer tx.Rollback()
	if err = fn(&Tx{s: s, tx: tx}); err != nil {
		return err
	}
	return mapError(tx.Commit())
}

func (s *Store) Driver() string { return s.driver }

func (s *Store) Stats() PoolStats { return PoolStats{Write: s.write.Stats(), Read: s.read.Stats()} }

func (s *Store) Get(ctx context.Context, kind, id string, out any) error {
	return s.get(ctx, s.read, kind, id, out)
}
func (t *Tx) Get(ctx context.Context, kind, id string, out any) error {
	return t.s.get(ctx, t.tx, kind, id, out)
}
func (s *Store) List(ctx context.Context, kind string) ([]json.RawMessage, error) {
	return s.list(ctx, s.read, kind)
}
func (t *Tx) List(ctx context.Context, kind string) ([]json.RawMessage, error) {
	return t.s.list(ctx, t.tx, kind)
}
func (s *Store) Put(ctx context.Context, kind, id string, value any) error {
	return s.WithTx(ctx, func(t *Tx) error { return t.Put(ctx, kind, id, value) })
}
func (t *Tx) Put(ctx context.Context, kind, id string, value any) error {
	return t.s.put(ctx, t.tx, kind, id, value)
}
func (s *Store) Delete(ctx context.Context, kind, id string) error {
	return s.WithTx(ctx, func(t *Tx) error { return t.Delete(ctx, kind, id) })
}
func (t *Tx) Delete(ctx context.Context, kind, id string) error {
	return t.s.delete(ctx, t.tx, kind, id)
}

func (s *Store) UpsertMonitor(ctx context.Context, m Monitor) error {
	return s.WithTx(ctx, func(t *Tx) error { return t.PutMonitor(ctx, m) })
}

func (s *Store) GetMonitor(ctx context.Context, id string) (Monitor, error) {
	return s.getMonitor(ctx, s.read, id)
}
func (t *Tx) GetMonitor(ctx context.Context, id string) (Monitor, error) {
	if t.s.driver == "postgres" {
		var m Monitor
		var enabled int64
		var payload string
		err := t.tx.QueryRowContext(ctx, `SELECT id,config_version,generation,kind,enabled,interval_ms,config_json FROM monitors WHERE id=$1 FOR UPDATE`, id).Scan(&m.ID, &m.ConfigVersion, &m.Generation, &m.Kind, &enabled, &m.IntervalMS, &payload)
		m.Enabled = enabled != 0
		m.ConfigJSON = json.RawMessage(payload)
		return m, mapError(err)
	}
	return t.s.getMonitor(ctx, t.tx, id)
}
func (s *Store) GetRuntime(ctx context.Context, id string) (Runtime, error) {
	return s.getRuntime(ctx, s.read, id)
}
func (t *Tx) GetRuntime(ctx context.Context, id string) (Runtime, error) {
	return t.s.getRuntime(ctx, t.tx, id)
}

type dbtx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
