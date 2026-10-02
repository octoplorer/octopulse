package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type RoundBucket struct {
	At             int64
	LatencyTotalMS int64
	Count          int64
	Successes      int64
}

type StatisticsSnapshot struct {
	Monitor                Monitor
	Intervals              []Interval
	Maintenance            []json.RawMessage
	Rounds                 []RoundBucket
	CollectionThrough      int64
	HasCollectionWatermark bool
}

type LatencySnapshot struct {
	Aggregates []Aggregate
	Rounds     []RoundBucket
}

func (s *Store) ReadLatency(ctx context.Context, id string, from, to, rawFrom, widthMS int64) (snapshot LatencySnapshot, err error) {
	if widthMS <= 0 || from >= to {
		return snapshot, fmt.Errorf("invalid latency window")
	}
	options := &sql.TxOptions{ReadOnly: true}
	if s.driver == "postgres" {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := s.read.BeginTx(ctx, options)
	if err != nil {
		return snapshot, mapError(err)
	}
	defer tx.Rollback()
	if _, err = s.getMonitor(ctx, tx, id); err != nil {
		return snapshot, err
	}
	start := from / widthMS * widthMS
	rows, err := tx.QueryContext(ctx, s.sql(`SELECT monitor_id,bucket_at,width_ms,up_ms,down_ms,unknown_ms,excluded_ms,latency_total_ms,round_count,successful_round_count FROM aggregates WHERE monitor_id=? AND bucket_at>=? AND bucket_at<? AND width_ms=? ORDER BY bucket_at`), id, start, to, widthMS)
	if err != nil {
		return snapshot, mapError(err)
	}
	snapshot.Aggregates = []Aggregate{}
	for rows.Next() {
		var row Aggregate
		if err = rows.Scan(&row.MonitorID, &row.BucketAt, &row.WidthMS, &row.UpMS, &row.DownMS, &row.UnknownMS, &row.ExcludedMS, &row.LatencyTotalMS, &row.RoundCount, &row.SuccessfulRoundCount); err != nil {
			rows.Close()
			return snapshot, err
		}
		snapshot.Aggregates = append(snapshot.Aggregates, row)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return snapshot, err
	}
	rows.Close()
	if rawFrom < to {
		snapshot.Rounds, err = s.roundBuckets(ctx, tx, id, rawFrom, to, widthMS)
		if err != nil {
			return snapshot, err
		}
	}
	if err = tx.Commit(); err != nil {
		return snapshot, mapError(err)
	}
	return snapshot, nil
}

// Statistics reads use a single snapshot. PostgreSQL Read Committed would allow
// an edited maintenance plan between separate reads; Repeatable Read prevents
// that. SQLite's deferred read transaction pins its WAL snapshot on first read.
func (s *Store) ReadStatistics(ctx context.Context, id string, from, to, widthMS int64) (snapshot StatisticsSnapshot, err error) {
	if from >= to || widthMS < 0 {
		return snapshot, fmt.Errorf("invalid statistics window")
	}
	options := &sql.TxOptions{ReadOnly: true}
	if s.driver == "postgres" {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := s.read.BeginTx(ctx, options)
	if err != nil {
		return snapshot, mapError(err)
	}
	defer tx.Rollback()
	if snapshot.Monitor, err = s.getMonitor(ctx, tx, id); err != nil {
		return snapshot, err
	}
	rows, err := tx.QueryContext(ctx, s.sql(`SELECT id,monitor_id,state,started_at,ended_at FROM state_intervals WHERE monitor_id=? AND started_at<? AND (ended_at IS NULL OR ended_at>?) ORDER BY started_at,id`), id, to, from)
	if err != nil {
		return snapshot, mapError(err)
	}
	snapshot.Intervals = []Interval{}
	for rows.Next() {
		var interval Interval
		var ended sql.NullInt64
		if err = rows.Scan(&interval.ID, &interval.MonitorID, &interval.State, &interval.StartedAt, &ended); err != nil {
			rows.Close()
			return snapshot, err
		}
		if ended.Valid {
			interval.EndedAt = &ended.Int64
		}
		snapshot.Intervals = append(snapshot.Intervals, interval)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return snapshot, err
	}
	rows.Close()
	if snapshot.Maintenance, err = s.list(ctx, tx, "maintenance"); err != nil {
		return snapshot, err
	}
	err = tx.QueryRowContext(ctx, `SELECT at_ms FROM watermarks WHERE name='collection'`).Scan(&snapshot.CollectionThrough)
	if err == nil {
		snapshot.HasCollectionWatermark = true
	} else if !errors.Is(err, sql.ErrNoRows) {
		return snapshot, mapError(err)
	}
	if widthMS > 0 {
		snapshot.Rounds, err = s.roundBuckets(ctx, tx, id, from, to, widthMS)
		if err != nil {
			return snapshot, err
		}
	}
	if err = tx.Commit(); err != nil {
		return snapshot, mapError(err)
	}
	return snapshot, nil
}

func (s *Store) RoundBuckets(ctx context.Context, id string, from, to, widthMS int64) ([]RoundBucket, error) {
	return s.roundBuckets(ctx, s.read, id, from, to, widthMS)
}
func (s *Store) roundBuckets(ctx context.Context, q dbtx, id string, from, to, widthMS int64) ([]RoundBucket, error) {
	if widthMS <= 0 || from >= to {
		return nil, fmt.Errorf("invalid round bucket window")
	}
	rows, err := q.QueryContext(ctx, s.sql(`SELECT (finished_at / ?) * ? AS bucket_at,SUM(latency_ms),COUNT(*),SUM(CASE WHEN success=1 THEN 1 ELSE 0 END) FROM rounds WHERE monitor_id=? AND finished_at>=? AND finished_at<? GROUP BY 1 ORDER BY bucket_at`), widthMS, widthMS, id, from, to)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := []RoundBucket{}
	for rows.Next() {
		var bucket RoundBucket
		if err = rows.Scan(&bucket.At, &bucket.LatencyTotalMS, &bucket.Count, &bucket.Successes); err != nil {
			return nil, err
		}
		result = append(result, bucket)
	}
	return result, mapError(rows.Err())
}

func (s *Store) EarliestStatisticsAt(ctx context.Context, id string) (int64, error) {
	var at sql.NullInt64
	err := s.read.QueryRowContext(ctx, s.sql(`SELECT MIN(at) FROM (SELECT MIN(started_at) AS at FROM state_intervals WHERE monitor_id=? UNION ALL SELECT MIN(finished_at) AS at FROM rounds WHERE monitor_id=?) AS boundaries`), id, id).Scan(&at)
	if err != nil {
		return 0, mapError(err)
	}
	if !at.Valid {
		return 0, ErrNotFound
	}
	return at.Int64, nil
}

// DeleteAggregateRange removes only one resolution. Rebuilding 5-minute data
// must not erase retained hourly data outside the rebuild's source window.
func (t *Tx) DeleteAggregateRange(ctx context.Context, id string, from, to, widthMS int64) error {
	_, err := t.tx.ExecContext(ctx, t.s.sql(`DELETE FROM aggregates WHERE monitor_id=? AND width_ms=? AND bucket_at<? AND bucket_at+width_ms>?`), id, widthMS, to, from)
	return mapError(err)
}
