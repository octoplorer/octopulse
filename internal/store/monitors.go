package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/octoplorer/octopulse/internal/store/postgresquery"
	"github.com/octoplorer/octopulse/internal/store/sqlitequery"
)

func (t *Tx) PutMonitor(ctx context.Context, m Monitor) error {
	payload, err := jsonText(m.ConfigJSON)
	if err != nil {
		return err
	}
	if m.ID == "" {
		return fmt.Errorf("monitor ID is required")
	}
	if t.s.driver == "sqlite" {
		err = sqlitequery.New(t.tx).PutMonitor(ctx, sqlitequery.PutMonitorParams{ID: m.ID, ConfigVersion: m.ConfigVersion, Generation: m.Generation, Kind: m.Kind, Enabled: boolInt(m.Enabled), IntervalMs: m.IntervalMS, ConfigJson: payload})
	} else {
		err = postgresquery.New(t.tx).PutMonitor(ctx, postgresquery.PutMonitorParams{ID: m.ID, ConfigVersion: m.ConfigVersion, Generation: m.Generation, Kind: m.Kind, Enabled: boolInt(m.Enabled), IntervalMs: m.IntervalMS, ConfigJson: payload})
	}
	return mapError(err)
}

func (s *Store) getMonitor(ctx context.Context, q dbtx, id string) (Monitor, error) {
	if s.driver == "sqlite" {
		r, err := sqlitequery.New(q).GetMonitor(ctx, id)
		return Monitor{r.ID, r.ConfigVersion, r.Generation, r.Kind, r.Enabled != 0, r.IntervalMs, json.RawMessage(r.ConfigJson)}, mapError(err)
	}
	r, err := postgresquery.New(q).GetMonitor(ctx, id)
	return Monitor{r.ID, r.ConfigVersion, r.Generation, r.Kind, r.Enabled != 0, r.IntervalMs, json.RawMessage(r.ConfigJson)}, mapError(err)
}

func (s *Store) ListMonitors(ctx context.Context) ([]Monitor, error) {
	result := []Monitor{}
	if s.driver == "sqlite" {
		records, err := sqlitequery.New(s.read).ListMonitors(ctx)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range records {
			result = append(result, Monitor{r.ID, r.ConfigVersion, r.Generation, r.Kind, r.Enabled != 0, r.IntervalMs, json.RawMessage(r.ConfigJson)})
		}
	} else {
		records, err := postgresquery.New(s.read).ListMonitors(ctx)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range records {
			result = append(result, Monitor{r.ID, r.ConfigVersion, r.Generation, r.Kind, r.Enabled != 0, r.IntervalMs, json.RawMessage(r.ConfigJson)})
		}
	}
	return result, nil
}

func (t *Tx) DeleteMonitor(ctx context.Context, id string) error {
	var n int64
	var err error
	if t.s.driver == "sqlite" {
		n, err = sqlitequery.New(t.tx).DeleteMonitor(ctx, id)
	} else {
		n, err = postgresquery.New(t.tx).DeleteMonitor(ctx, id)
	}
	if err != nil {
		return mapError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteMonitor(ctx context.Context, id string) error {
	return s.WithTx(ctx, func(t *Tx) error { return t.DeleteMonitor(ctx, id) })
}

func (s *Store) getRuntime(ctx context.Context, q dbtx, id string) (Runtime, error) {
	if s.driver == "sqlite" {
		r, err := sqlitequery.New(q).GetRuntime(ctx, id)
		return Runtime{r.MonitorID, r.ConfigVersion, r.Generation, r.State, r.Failures, r.Successes, r.LastRoundID, r.LastCollectedAt, r.HeartbeatVersion, r.HeartbeatAt}, mapError(err)
	}
	r, err := postgresquery.New(q).GetRuntime(ctx, id)
	return Runtime{r.MonitorID, r.ConfigVersion, r.Generation, r.State, r.Failures, r.Successes, r.LastRoundID, r.LastCollectedAt, r.HeartbeatVersion, r.HeartbeatAt}, mapError(err)
}

func (t *Tx) PutRuntime(ctx context.Context, r Runtime) error {
	_, err := t.tx.ExecContext(ctx, t.s.sql(`INSERT INTO monitor_runtime(monitor_id,config_version,generation,state,failures,successes,last_round_id,last_collected_at,heartbeat_version,heartbeat_at) VALUES(?,?,?,?,?,?,?,?,?,?) ON CONFLICT(monitor_id) DO UPDATE SET config_version=excluded.config_version,generation=excluded.generation,state=excluded.state,failures=excluded.failures,successes=excluded.successes,last_round_id=excluded.last_round_id,last_collected_at=excluded.last_collected_at,heartbeat_version=excluded.heartbeat_version,heartbeat_at=excluded.heartbeat_at`), r.MonitorID, r.ConfigVersion, r.Generation, r.State, r.Failures, r.Successes, r.LastRoundID, r.LastCollectedAt, r.HeartbeatVersion, r.HeartbeatAt)
	return mapError(err)
}

// CompareRuntime rejects an in-flight result from an obsolete configuration,
// generation, heartbeat version, or concurrently replaced last round.
func (t *Tx) CompareRuntime(ctx context.Context, previous Runtime, next Runtime) error {
	result, err := t.tx.ExecContext(ctx, t.s.sql(`UPDATE monitor_runtime SET config_version=?,generation=?,state=?,failures=?,successes=?,last_round_id=?,last_collected_at=?,heartbeat_version=?,heartbeat_at=? WHERE monitor_id=? AND config_version=? AND generation=? AND heartbeat_version=? AND last_round_id=? AND EXISTS(SELECT 1 FROM monitors WHERE id=? AND config_version=? AND generation=?)`), next.ConfigVersion, next.Generation, next.State, next.Failures, next.Successes, next.LastRoundID, next.LastCollectedAt, next.HeartbeatVersion, next.HeartbeatAt, previous.MonitorID, previous.ConfigVersion, previous.Generation, previous.HeartbeatVersion, previous.LastRoundID, previous.MonitorID, next.ConfigVersion, next.Generation)
	if err != nil {
		return mapError(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrConflict
	}
	return nil
}
