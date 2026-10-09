package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func (t *Tx) PutAggregate(ctx context.Context, a Aggregate) error {
	_, err := t.tx.ExecContext(
		ctx,
		t.s.sql(
			`INSERT INTO aggregates(monitor_id,bucket_at,width_ms,up_ms,down_ms,unknown_ms,excluded_ms,`+
				`latency_total_ms,round_count,successful_round_count) VALUES(?,?,?,?,?,?,?,?,?,?) `+
				`ON CONFLICT(monitor_id,bucket_at,width_ms) `+
				`DO UPDATE SET up_ms=excluded.up_ms,down_ms=excluded.down_ms,unknown_ms=excluded.unknown_ms,`+
				`excluded_ms=excluded.excluded_ms,latency_total_ms=excluded.latency_total_ms,`+
				`round_count=excluded.round_count,successful_round_count=excluded.successful_round_count`,
		),
		a.MonitorID,
		a.BucketAt,
		a.WidthMS,
		a.UpMS,
		a.DownMS,
		a.UnknownMS,
		a.ExcludedMS,
		a.LatencyTotalMS,
		a.RoundCount,
		a.SuccessfulRoundCount,
	)
	return mapError(err)
}

func (s *Store) Aggregates(ctx context.Context, monitorID string, start, end, widthMS int64) ([]Aggregate, error) {
	rows, err := s.read.QueryContext(
		ctx,
		s.sql(
			`SELECT monitor_id,bucket_at,width_ms,up_ms,down_ms,unknown_ms,excluded_ms,latency_total_ms,`+
				`round_count,successful_round_count FROM aggregates WHERE monitor_id=? AND bucket_at>=? `+
				`AND bucket_at<? AND width_ms=? ORDER BY bucket_at`,
		),
		monitorID,
		start,
		end,
		widthMS,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := []Aggregate{}
	for rows.Next() {
		var a Aggregate
		if err = rows.Scan(
			&a.MonitorID,
			&a.BucketAt,
			&a.WidthMS,
			&a.UpMS,
			&a.DownMS,
			&a.UnknownMS,
			&a.ExcludedMS,
			&a.LatencyTotalMS,
			&a.RoundCount,
			&a.SuccessfulRoundCount,
		); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, mapError(rows.Err())
}

func (t *Tx) DeleteAggregates(ctx context.Context, monitorID string, start, end int64) error {
	_, err := t.tx.ExecContext(
		ctx,
		t.s.sql(`DELETE FROM aggregates WHERE monitor_id=? AND bucket_at<? AND bucket_at+width_ms>?`),
		monitorID,
		end,
		start,
	)
	return mapError(err)
}

func (t *Tx) PutWatermark(ctx context.Context, name string, at int64) error {
	_, err := t.tx.ExecContext(
		ctx,
		t.s.sql(`INSERT INTO watermarks(name,at_ms) VALUES(?,?) ON CONFLICT(name) DO UPDATE SET at_ms=excluded.at_ms`),
		name,
		at,
	)
	return mapError(err)
}

func (s *Store) Watermark(ctx context.Context, name string) (int64, error) {
	var at int64
	err := s.read.QueryRowContext(ctx, s.sql(`SELECT at_ms FROM watermarks WHERE name=?`), name).Scan(&at)
	return at, mapError(err)
}

func (t *Tx) Watermark(ctx context.Context, name string) (int64, error) {
	var at int64
	err := t.tx.QueryRowContext(ctx, t.s.sql(`SELECT at_ms FROM watermarks WHERE name=?`), name).Scan(&at)
	return at, mapError(err)
}

// PruneHistory deletes only a bounded batch behind a confirmed aggregation
// watermark. Attempts and rounds have independent retention cutoffs.
func (t *Tx) PruneHistory(ctx context.Context, roundBefore, attemptBefore, watermark int64, batch int) error {
	if batch <= 0 || batch > 10000 {
		batch = 1000
	}
	_, err := t.tx.ExecContext(
		ctx,
		t.s.sql(
			`DELETE FROM attempts WHERE (round_id,number) IN (SELECT a.round_id,a.number FROM attempts a `+
				`JOIN rounds r ON r.id=a.round_id WHERE a.finished_at<? AND r.finished_at<=? `+
				`ORDER BY a.finished_at,a.round_id,a.number LIMIT ?)`,
		),
		attemptBefore,
		watermark,
		batch,
	)
	if err != nil {
		return mapError(err)
	}
	_, err = t.tx.ExecContext(
		ctx,
		t.s.sql(
			`DELETE FROM rounds WHERE id IN (SELECT id FROM rounds WHERE finished_at<? `+
				`AND finished_at<=? ORDER BY finished_at,id LIMIT ?)`,
		),
		roundBefore,
		watermark,
		batch,
	)
	return mapError(err)
}

// PruneSummaries applies the two aggregate retention windows and closed-state
// interval retention in bounded batches, behind the reliable aggregate boundary.
func (t *Tx) PruneSummaries(
	ctx context.Context,
	fiveMinuteBefore, hourlyBefore, intervalBefore, watermark int64,
	batch int,
) error {
	if batch <= 0 || batch > 10000 {
		batch = 1000
	}
	for _, window := range []struct{ width, before int64 }{
		{
			width:  300000,
			before: fiveMinuteBefore,
		},
		{
			width:  3600000,
			before: hourlyBefore,
		},
	} {
		_, err := t.tx.ExecContext(
			ctx,
			t.s.sql(
				`DELETE FROM aggregates WHERE (monitor_id,bucket_at,width_ms) IN (`+
					`SELECT monitor_id,bucket_at,width_ms FROM aggregates `+
					`WHERE width_ms=? AND bucket_at+width_ms<? AND bucket_at+width_ms<=? `+
					`ORDER BY bucket_at,monitor_id,width_ms LIMIT ?)`,
			),
			window.width,
			window.before,
			watermark,
			batch,
		)
		if err != nil {
			return mapError(err)
		}
	}
	_, err := t.tx.ExecContext(
		ctx,
		t.s.sql(
			`DELETE FROM state_intervals WHERE id IN (SELECT id FROM state_intervals `+
				`WHERE ended_at IS NOT NULL AND ended_at<? AND ended_at<=? ORDER BY ended_at,id LIMIT ?)`,
		),
		intervalBefore,
		watermark,
		batch,
	)
	return mapError(err)
}

// BackupSQLite produces a consistent standalone copy including committed WAL
// content. The destination must be new; a running .db file is never copied.
func (s *Store) BackupSQLite(ctx context.Context, destination string) error {
	if s.driver != "sqlite" {
		return fmt.Errorf("SQLite backup is unavailable for this database")
	}
	path, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err == nil {
		return fmt.Errorf("backup destination already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	_, err = s.write.ExecContext(ctx, `VACUUM INTO ?`, path)
	if err != nil {
		return mapError(err)
	}
	return os.Chmod(path, 0600)
}
