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

func (s *Store) ReadLatency(ctx context.Context, id string, from, to, rawFrom, widthMS int64) (LatencySnapshot, error) {
	snapshots, err := s.ReadLatencyBatch(
		ctx,
		[]string{id},
		from,
		to,
		rawFrom,
		widthMS,
	)
	if err != nil {
		return LatencySnapshot{}, err
	}
	snapshot, ok := snapshots[id]
	if !ok {
		return LatencySnapshot{}, ErrNotFound
	}
	return snapshot, nil
}

// ReadLatencyBatch returns retained and raw latency in a consistent snapshot.
// Unknown monitor IDs are omitted and duplicate IDs are read only once.
func (s *Store) ReadLatencyBatch(
	ctx context.Context,
	ids []string,
	from, to, rawFrom, widthMS int64,
) (map[string]LatencySnapshot, error) {
	if widthMS <= 0 || from >= to {
		return nil, fmt.Errorf("invalid latency window")
	}
	result := map[string]LatencySnapshot{}
	if len(ids) == 0 {
		return result, nil
	}
	err := s.withReadSnapshot(ctx, func(q dbtx) error {
		return forMonitorBatches(ids, func(batch []string) error {
			monitors, err := s.readMonitors(ctx, q, batch)
			if err != nil {
				return err
			}
			for _, m := range monitors {
				result[m.ID] = LatencySnapshot{Aggregates: []Aggregate{}, Rounds: []RoundBucket{}}
			}
			aggregates, err := s.readLatencyAggregates(
				ctx,
				q,
				batch,
				from/widthMS*widthMS,
				to,
				widthMS,
			)
			if err != nil {
				return err
			}
			for _, row := range aggregates {
				snapshot := result[row.MonitorID]
				snapshot.Aggregates = append(snapshot.Aggregates, row)
				result[row.MonitorID] = snapshot
			}
			if rawFrom < to {
				buckets, err := s.readRoundBuckets(
					ctx,
					q,
					batch,
					rawFrom,
					to,
					widthMS,
				)
				if err != nil {
					return err
				}
				for id, rows := range buckets {
					snapshot := result[id]
					snapshot.Rounds = rows
					result[id] = snapshot
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) ReadStatistics(ctx context.Context, id string, from, to, widthMS int64) (StatisticsSnapshot, error) {
	snapshots, err := s.ReadStatisticsBatch(
		ctx,
		[]string{id},
		from,
		to,
		widthMS,
	)
	if err != nil {
		return StatisticsSnapshot{}, err
	}
	snapshot, ok := snapshots[id]
	if !ok {
		return StatisticsSnapshot{}, ErrNotFound
	}
	return snapshot, nil
}

// ReadStatisticsBatch loads maintenance and the collection watermark once for
// the whole request. The shared maintenance slice is immutable to callers.
func (s *Store) ReadStatisticsBatch(
	ctx context.Context,
	ids []string,
	from, to, widthMS int64,
) (map[string]StatisticsSnapshot, error) {
	if from >= to || widthMS < 0 {
		return nil, fmt.Errorf("invalid statistics window")
	}
	result := map[string]StatisticsSnapshot{}
	if len(ids) == 0 {
		return result, nil
	}
	err := s.withReadSnapshot(ctx, func(q dbtx) error {
		maintenance, err := s.list(ctx, q, "maintenance")
		if err != nil {
			return err
		}
		var collectionThrough int64
		err = q.QueryRowContext(ctx, `SELECT at_ms FROM watermarks WHERE name='collection'`).Scan(&collectionThrough)
		hasWatermark := err == nil
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return mapError(err)
		}
		return forMonitorBatches(ids, func(batch []string) error {
			monitors, err := s.readMonitors(ctx, q, batch)
			if err != nil {
				return err
			}
			for _, m := range monitors {
				result[m.ID] = StatisticsSnapshot{
					Monitor:                m,
					Intervals:              []Interval{},
					Rounds:                 []RoundBucket{},
					Maintenance:            maintenance,
					CollectionThrough:      collectionThrough,
					HasCollectionWatermark: hasWatermark,
				}
			}
			intervals, err := s.readStatisticsIntervals(
				ctx,
				q,
				batch,
				from,
				to,
			)
			if err != nil {
				return err
			}
			for _, row := range intervals {
				snapshot := result[row.MonitorID]
				snapshot.Intervals = append(snapshot.Intervals, row)
				result[row.MonitorID] = snapshot
			}
			if widthMS > 0 {
				buckets, err := s.readRoundBuckets(
					ctx,
					q,
					batch,
					from,
					to,
					widthMS,
				)
				if err != nil {
					return err
				}
				for id, rows := range buckets {
					snapshot := result[id]
					snapshot.Rounds = rows
					result[id] = snapshot
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) RoundBuckets(ctx context.Context, id string, from, to, widthMS int64) ([]RoundBucket, error) {
	if widthMS <= 0 || from >= to {
		return nil, fmt.Errorf("invalid round bucket window")
	}
	buckets, err := s.readRoundBuckets(
		ctx,
		s.read,
		[]string{id},
		from,
		to,
		widthMS,
	)
	if err != nil {
		return nil, err
	}
	rows := buckets[id]
	if rows == nil {
		rows = []RoundBucket{}
	}
	return rows, nil
}

func (s *Store) EarliestStatisticsAt(ctx context.Context, id string) (int64, error) {
	var at sql.NullInt64
	err := s.read.QueryRowContext(
		ctx,
		s.sql(
			`SELECT MIN(at) FROM (SELECT MIN(started_at) AS at FROM state_intervals WHERE monitor_id=? `+
				`UNION ALL SELECT MIN(finished_at) AS at FROM rounds WHERE monitor_id=?) AS boundaries`,
		),
		id,
		id,
	).Scan(&at)
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
	_, err := t.tx.ExecContext(
		ctx,
		t.s.sql(`DELETE FROM aggregates WHERE monitor_id=? AND width_ms=? AND bucket_at<? AND bucket_at+width_ms>?`),
		id,
		widthMS,
		to,
		from,
	)
	return mapError(err)
}
