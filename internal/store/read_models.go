package store

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/octoplorer/octopulse/internal/store/postgresquery"
	"github.com/octoplorer/octopulse/internal/store/sqlitequery"
)

// MonitorSnapshot reads configuration and its associated runtime and engine
// document together. Missing optional rows are represented by nil values.
type MonitorSnapshot struct {
	Monitor    Monitor
	Runtime    *Runtime
	EngineJSON json.RawMessage
}

func monitorSnapshot(r sqlitequery.ListMonitorSnapshotsRow) MonitorSnapshot {
	result := MonitorSnapshot{
		Monitor: Monitor{
			ID:            r.ID,
			ConfigVersion: r.ConfigVersion,
			Generation:    r.Generation,
			Kind:          r.Kind,
			Enabled:       r.Enabled != 0,
			IntervalMS:    r.IntervalMs,
			ConfigJSON:    json.RawMessage(r.ConfigJson),
		},
	}
	if r.RuntimeMonitorID.Valid {
		result.Runtime = &Runtime{
			MonitorID:        r.RuntimeMonitorID.String,
			ConfigVersion:    r.RuntimeConfigVersion.Int64,
			Generation:       r.RuntimeGeneration.Int64,
			State:            r.State.String,
			Failures:         r.Failures.Int64,
			Successes:        r.Successes.Int64,
			LastRoundID:      r.LastRoundID.String,
			LastCollectedAt:  r.LastCollectedAt.Int64,
			HeartbeatVersion: r.HeartbeatVersion.Int64,
			HeartbeatAt:      r.HeartbeatAt.Int64,
		}
	}
	if r.EngineJson.Valid {
		result.EngineJSON = json.RawMessage(r.EngineJson.String)
	}
	return result
}

// ListMonitorSnapshots uses one statement, so even Read Committed observes
// configuration, runtime and engine state from the same database snapshot.
func (s *Store) ListMonitorSnapshots(ctx context.Context) ([]MonitorSnapshot, error) {
	result := []MonitorSnapshot{}
	if s.driver == "sqlite" {
		rows, err := sqlitequery.New(s.read).ListMonitorSnapshots(ctx)
		if err != nil {
			return nil, mapError(err)
		}
		for _, row := range rows {
			result = append(result, monitorSnapshot(row))
		}
	} else {
		rows, err := postgresquery.New(s.read).ListMonitorSnapshots(ctx)
		if err != nil {
			return nil, mapError(err)
		}
		for _, row := range rows {
			result = append(result, monitorSnapshot(sqlitequery.ListMonitorSnapshotsRow(row)))
		}
	}
	return result, nil
}

// ReadMonitorSnapshots omits unknown IDs. An empty input reads no records.
func (s *Store) ReadMonitorSnapshots(ctx context.Context, ids []string) (map[string]MonitorSnapshot, error) {
	result := map[string]MonitorSnapshot{}
	if len(ids) == 0 {
		return result, nil
	}
	err := s.withReadSnapshot(ctx, func(q dbtx) error {
		return forMonitorBatches(ids, func(batch []string) error {
			if s.driver == "sqlite" {
				rows, err := sqlitequery.New(q).ReadMonitorSnapshots(ctx, batch)
				if err != nil {
					return mapError(err)
				}
				for _, row := range rows {
					result[row.ID] = monitorSnapshot(sqlitequery.ListMonitorSnapshotsRow(row))
				}
			} else {
				rows, err := postgresquery.New(q).ReadMonitorSnapshots(ctx, monitorIDsJSON(batch))
				if err != nil {
					return mapError(err)
				}
				for _, row := range rows {
					result[row.ID] = monitorSnapshot(sqlitequery.ListMonitorSnapshotsRow(row))
				}
			}
			return nil
		})
	})
	return result, err
}

// Multiple reads, including SQLite parameter-limit batches, share one snapshot.
// PostgreSQL Read Committed alone would allow policy edits between statements.
func (s *Store) withReadSnapshot(ctx context.Context, read func(dbtx) error) error {
	options := &sql.TxOptions{ReadOnly: true}
	if s.driver == "postgres" {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := s.read.BeginTx(ctx, options)
	if err != nil {
		return mapError(err)
	}
	defer tx.Rollback()
	if err = read(tx); err != nil {
		return err
	}
	return mapError(tx.Commit())
}

func forMonitorBatches(ids []string, fn func([]string) error) error {
	const batchSize = 900 // Leave room for fixed parameters on SQLite builds with a 999-variable limit.
	unique := make([]string, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	for len(unique) > 0 {
		n := min(len(unique), batchSize)
		if err := fn(unique[:n]); err != nil {
			return err
		}
		unique = unique[n:]
	}
	return nil
}

// PostgreSQL accepts one JSON array parameter without depending on a second
// PostgreSQL driver solely to encode sqlc's database/sql text[] arguments.
func monitorIDsJSON(ids []string) json.RawMessage {
	data, _ := json.Marshal(ids)
	return data
}

func (s *Store) readMonitors(ctx context.Context, q dbtx, ids []string) ([]Monitor, error) {
	result := []Monitor{}
	if s.driver == "sqlite" {
		rows, err := sqlitequery.New(q).ReadMonitors(ctx, ids)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			result = append(
				result,
				Monitor{
					ID:            r.ID,
					ConfigVersion: r.ConfigVersion,
					Generation:    r.Generation,
					Kind:          r.Kind,
					Enabled:       r.Enabled != 0,
					IntervalMS:    r.IntervalMs,
					ConfigJSON:    json.RawMessage(r.ConfigJson),
				},
			)
		}
	} else {
		rows, err := postgresquery.New(q).ReadMonitors(ctx, monitorIDsJSON(ids))
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			result = append(
				result,
				Monitor{
					ID:            r.ID,
					ConfigVersion: r.ConfigVersion,
					Generation:    r.Generation,
					Kind:          r.Kind,
					Enabled:       r.Enabled != 0,
					IntervalMS:    r.IntervalMs,
					ConfigJSON:    json.RawMessage(r.ConfigJson),
				},
			)
		}
	}
	return result, nil
}

func (s *Store) readStatisticsIntervals(ctx context.Context, q dbtx, ids []string, from, to int64) ([]Interval, error) {
	result := []Interval{}
	appendInterval := func(id, monitorID, state string, started int64, ended sql.NullInt64) {
		row := Interval{ID: id, MonitorID: monitorID, State: state, StartedAt: started}
		if ended.Valid {
			row.EndedAt = &ended.Int64
		}
		result = append(result, row)
	}
	if s.driver == "sqlite" {
		rows, err := sqlitequery.New(q).ReadStatisticsIntervals(
			ctx,
			sqlitequery.ReadStatisticsIntervalsParams{
				Ids: ids,
				FromMs: sql.NullInt64{
					Int64: from,
					Valid: true,
				},
				ToMs: to,
			},
		)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			appendInterval(
				r.ID,
				r.MonitorID,
				r.State,
				r.StartedAt,
				r.EndedAt,
			)
		}
	} else {
		rows, err := postgresquery.New(q).ReadStatisticsIntervals(
			ctx,
			postgresquery.ReadStatisticsIntervalsParams{
				IdsJson: monitorIDsJSON(ids),
				FromMs: sql.NullInt64{
					Int64: from,
					Valid: true,
				},
				ToMs: to,
			},
		)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			appendInterval(
				r.ID,
				r.MonitorID,
				r.State,
				r.StartedAt,
				r.EndedAt,
			)
		}
	}
	return result, nil
}

func (s *Store) readLatencyAggregates(
	ctx context.Context,
	q dbtx,
	ids []string,
	from, to, width int64,
) ([]Aggregate, error) {
	result := []Aggregate{}
	appendAggregate := func(r sqlitequery.Aggregate) {
		result = append(
			result,
			Aggregate{
				MonitorID:            r.MonitorID,
				BucketAt:             r.BucketAt,
				WidthMS:              r.WidthMs,
				UpMS:                 r.UpMs,
				DownMS:               r.DownMs,
				UnknownMS:            r.UnknownMs,
				ExcludedMS:           r.ExcludedMs,
				LatencyTotalMS:       r.LatencyTotalMs,
				RoundCount:           r.RoundCount,
				SuccessfulRoundCount: r.SuccessfulRoundCount,
			},
		)
	}
	if s.driver == "sqlite" {
		rows, err := sqlitequery.New(q).ReadLatencyAggregates(
			ctx,
			sqlitequery.ReadLatencyAggregatesParams{
				Ids:     ids,
				FromMs:  from,
				ToMs:    to,
				WidthMs: width,
			},
		)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			appendAggregate(r)
		}
	} else {
		rows, err := postgresquery.New(q).ReadLatencyAggregates(
			ctx,
			postgresquery.ReadLatencyAggregatesParams{
				IdsJson: monitorIDsJSON(ids),
				FromMs:  from,
				ToMs:    to,
				WidthMs: width,
			},
		)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			appendAggregate(sqlitequery.Aggregate(r))
		}
	}
	return result, nil
}

func (s *Store) readRoundBuckets(
	ctx context.Context,
	q dbtx,
	ids []string,
	from, to, width int64,
) (map[string][]RoundBucket, error) {
	result := map[string][]RoundBucket{}
	appendBucket := func(r sqlitequery.ReadRoundBucketsRow) {
		result[r.MonitorID] = append(
			result[r.MonitorID],
			RoundBucket{
				At:             r.BucketAt,
				LatencyTotalMS: r.LatencyTotalMs,
				Count:          r.RoundCount,
				Successes:      r.SuccessfulRoundCount,
			},
		)
	}
	if s.driver == "sqlite" {
		rows, err := sqlitequery.New(q).ReadRoundBuckets(
			ctx,
			sqlitequery.ReadRoundBucketsParams{
				Ids:     ids,
				FromMs:  from,
				ToMs:    to,
				WidthMs: width,
			},
		)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			appendBucket(r)
		}
	} else {
		rows, err := postgresquery.New(q).ReadRoundBuckets(
			ctx,
			postgresquery.ReadRoundBucketsParams{
				IdsJson: monitorIDsJSON(ids),
				FromMs:  from,
				ToMs:    to,
				WidthMs: width,
			},
		)
		if err != nil {
			return nil, mapError(err)
		}
		for _, r := range rows {
			appendBucket(sqlitequery.ReadRoundBucketsRow(r))
		}
	}
	return result, nil
}
