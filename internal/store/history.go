package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

func (t *Tx) PutRound(ctx context.Context, r Round) error {
	result, err := t.tx.ExecContext(ctx, t.s.sql(`INSERT INTO rounds(id,monitor_id,config_version,generation,started_at,finished_at,success,latency_ms) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(id) DO NOTHING`), r.ID, r.MonitorID, r.ConfigVersion, r.Generation, r.StartedAt, r.FinishedAt, boolInt(r.Success), r.LatencyMS)
	if err != nil {
		return mapError(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var monitorID string
		var version, generation, started, finished, success, latency int64
		if err = t.tx.QueryRowContext(ctx, t.s.sql(`SELECT monitor_id,config_version,generation,started_at,finished_at,success,latency_ms FROM rounds WHERE id=?`), r.ID).Scan(&monitorID, &version, &generation, &started, &finished, &success, &latency); err != nil {
			return mapError(err)
		}
		if monitorID != r.MonitorID || version != r.ConfigVersion || generation != r.Generation || started != r.StartedAt || finished != r.FinishedAt || success != boolInt(r.Success) || latency != r.LatencyMS {
			return ErrConflict
		}
		return nil
	}
	for _, a := range r.Attempts {
		detail, e := jsonText(a.Detail)
		if e != nil {
			return e
		}
		_, err = t.tx.ExecContext(ctx, t.s.sql(`INSERT INTO attempts(round_id,number,started_at,finished_at,success,latency_ms,error,detail) VALUES(?,?,?,?,?,?,?,?)`), r.ID, a.Number, a.StartedAt, a.FinishedAt, boolInt(a.Success), a.LatencyMS, a.Error, detail)
		if err != nil {
			return mapError(err)
		}
	}
	return nil
}

func (s *Store) RoundExists(ctx context.Context, id string) (bool, error) {
	var n int64
	err := s.read.QueryRowContext(ctx, s.sql(`SELECT COUNT(*) FROM rounds WHERE id=?`), id).Scan(&n)
	return n != 0, mapError(err)
}

func (s *Store) ListRounds(ctx context.Context, monitorID string, since int64, limit int) ([]Round, error) {
	return s.ListRoundsBetween(ctx, monitorID, since, 1<<63-1, limit)
}
func (s *Store) ListRoundsBetween(ctx context.Context, monitorID string, since, until int64, limit int) ([]Round, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	rows, err := s.read.QueryContext(ctx, s.sql(`SELECT id,monitor_id,config_version,generation,started_at,finished_at,success,latency_ms FROM rounds WHERE monitor_id=? AND finished_at>=? AND finished_at<=? ORDER BY started_at DESC,id DESC LIMIT ?`), monitorID, since, until, limit)
	if err != nil {
		return nil, mapError(err)
	}
	result := []Round{}
	byID := map[string]int{}
	for rows.Next() {
		var r Round
		var success int64
		if err = rows.Scan(&r.ID, &r.MonitorID, &r.ConfigVersion, &r.Generation, &r.StartedAt, &r.FinishedAt, &success, &r.LatencyMS); err != nil {
			rows.Close()
			return nil, err
		}
		r.Success = success != 0
		r.Attempts = []Attempt{}
		byID[r.ID] = len(result)
		result = append(result, r)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if len(result) == 0 {
		return result, nil
	}
	// Use exactly the first selection's IDs, so concurrently inserted rounds do
	// not shift the bounded selection between the two queries.
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(result)), ",")
	ids := make([]any, len(result))
	for index, round := range result {
		ids[index] = round.ID
	}
	attemptRows, err := s.read.QueryContext(ctx, s.sql(`SELECT a.round_id,a.number,a.started_at,a.finished_at,a.success,a.latency_ms,a.error,a.detail FROM attempts a WHERE a.round_id IN (`+placeholders+`) ORDER BY a.round_id,a.number`), ids...)
	if err != nil {
		return nil, mapError(err)
	}
	defer attemptRows.Close()
	for attemptRows.Next() {
		var id, detail string
		var a Attempt
		var success int64
		if err = attemptRows.Scan(&id, &a.Number, &a.StartedAt, &a.FinishedAt, &success, &a.LatencyMS, &a.Error, &detail); err != nil {
			return nil, err
		}
		a.Success = success != 0
		a.Detail = json.RawMessage(detail)
		if index, ok := byID[id]; ok {
			result[index].Attempts = append(result[index].Attempts, a)
		}
	}
	return result, mapError(attemptRows.Err())
}

func (t *Tx) ReplaceInterval(ctx context.Context, i Interval) error {
	var latest int64
	err := t.tx.QueryRowContext(ctx, t.s.sql(`SELECT started_at FROM state_intervals WHERE monitor_id=? AND ended_at IS NULL`), i.MonitorID).Scan(&latest)
	if err != nil && err != sql.ErrNoRows {
		return mapError(err)
	}
	if err == nil && i.StartedAt < latest {
		return fmt.Errorf("%w: interval predates current state", ErrConflict)
	}
	if _, err = t.tx.ExecContext(ctx, t.s.sql(`UPDATE state_intervals SET ended_at=? WHERE monitor_id=? AND ended_at IS NULL`), i.StartedAt, i.MonitorID); err != nil {
		return mapError(err)
	}
	return t.PutInterval(ctx, i)
}

func (t *Tx) PutInterval(ctx context.Context, i Interval) error {
	_, err := t.tx.ExecContext(ctx, t.s.sql(`INSERT INTO state_intervals(id,monitor_id,state,started_at,ended_at) VALUES(?,?,?,?,?)`), i.ID, i.MonitorID, i.State, i.StartedAt, i.EndedAt)
	return mapError(err)
}

func (t *Tx) CloseInterval(ctx context.Context, monitorID string, at int64) error {
	_, err := t.tx.ExecContext(ctx, t.s.sql(`UPDATE state_intervals SET ended_at=? WHERE monitor_id=? AND ended_at IS NULL AND started_at<=?`), at, monitorID, at)
	return mapError(err)
}

func (s *Store) Intervals(ctx context.Context, monitorID string, start, end int64) ([]Interval, error) {
	rows, err := s.read.QueryContext(ctx, s.sql(`SELECT id,monitor_id,state,started_at,ended_at FROM state_intervals WHERE monitor_id=? AND started_at<? AND (ended_at IS NULL OR ended_at>?) ORDER BY started_at,id`), monitorID, end, start)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := []Interval{}
	for rows.Next() {
		var i Interval
		var ended sql.NullInt64
		if err = rows.Scan(&i.ID, &i.MonitorID, &i.State, &i.StartedAt, &ended); err != nil {
			return nil, err
		}
		if ended.Valid {
			i.EndedAt = &ended.Int64
		}
		result = append(result, i)
	}
	return result, mapError(rows.Err())
}

// CutOpenIntervals ends known state at the last reliable collection watermark,
// then records an explicit Unknown gap. Runtime values are caller-controlled.
func (t *Tx) CutOpenIntervals(ctx context.Context, at int64) error {
	_, err := t.tx.ExecContext(ctx, t.s.sql(`UPDATE state_intervals SET ended_at=CASE WHEN started_at>? THEN started_at ELSE ? END WHERE ended_at IS NULL`), at, at)
	return mapError(err)
}
