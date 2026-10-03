package store

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/octoplorer/octopulse/db"
	"github.com/pressly/goose/v3"
)

func TestReadSnapshotRetainsPolicyAcrossConcurrentWrite(t *testing.T) {
	backends := []string{"sqlite"}
	if os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN") != "" {
		backends = append(backends, "postgres")
	}
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			s, err := Open(ctx, testConfig(t, backend))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if err := s.Put(ctx, "maintenance", "policy", map[string]int{"version": 1}); err != nil {
				t.Fatal(err)
			}
			err = s.withReadSnapshot(ctx, func(q dbtx) error {
				var before, after map[string]int
				if err := s.get(ctx, q, "maintenance", "policy", &before); err != nil {
					return err
				}
				if err := s.Put(ctx, "maintenance", "policy", map[string]int{"version": 2}); err != nil {
					return err
				}
				if err := s.get(ctx, q, "maintenance", "policy", &after); err != nil {
					return err
				}
				if before["version"] != 1 || after["version"] != 1 {
					return fmt.Errorf("read snapshot changed: before=%v after=%v", before, after)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			var current map[string]int
			if err := s.Get(ctx, "maintenance", "policy", &current); err != nil || current["version"] != 2 {
				t.Fatal(current, err)
			}
		})
	}
}

func TestFinishedRoundIndexUpgradeAcrossDatabases(t *testing.T) {
	backends := []string{"sqlite"}
	if os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN") != "" {
		backends = append(backends, "postgres")
	}
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			cfg := testConfig(t, backend)
			s, err := Open(ctx, cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { s.Close() })
			seedMonitor(t, s, "migration")
			if err := s.WithTx(ctx, func(tx *Tx) error {
				return tx.PutRound(ctx, Round{ID: "preserved", MonitorID: "migration", ConfigVersion: 1, Generation: 1, StartedAt: 10, FinishedAt: 20, Success: true, LatencyMS: 10})
			}); err != nil {
				t.Fatal(err)
			}
			dialect := goose.DialectSQLite3
			if backend == "postgres" {
				dialect = goose.DialectPostgres
			}
			migrations, err := fs.Sub(db.Migrations, backend+"/migrations")
			if err != nil {
				t.Fatal(err)
			}
			provider, err := goose.NewProvider(dialect, s.write, migrations)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := provider.DownTo(ctx, 2); err != nil {
				t.Fatal(err)
			}
			var count int
			indexQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='rounds_monitor_finished'`
			if backend == "postgres" {
				indexQuery = `SELECT COUNT(*) FROM pg_indexes WHERE schemaname=current_schema() AND indexname='rounds_monitor_finished'`
			}
			if err := s.read.QueryRowContext(ctx, indexQuery).Scan(&count); err != nil || count != 0 {
				t.Fatal("down migration", count, err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = Open(ctx, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.read.QueryRowContext(ctx, indexQuery).Scan(&count); err != nil || count != 1 {
				t.Fatal("up migration", count, err)
			}
			rows, err := s.RoundBuckets(ctx, "migration", 0, 100, 100)
			if err != nil || len(rows) != 1 || rows[0].Count != 1 || rows[0].Successes != 1 {
				t.Fatal("preserved data", rows, err)
			}
			if backend == "sqlite" {
				planRows, err := s.read.QueryContext(ctx, `EXPLAIN QUERY PLAN SELECT finished_at,latency_ms FROM rounds WHERE monitor_id=? AND finished_at>=? AND finished_at<?`, "migration", 0, 100)
				if err != nil {
					t.Fatal(err)
				}
				defer planRows.Close()
				found := false
				for planRows.Next() {
					var id, parent, unused int
					var detail string
					if err := planRows.Scan(&id, &parent, &unused, &detail); err != nil {
						t.Fatal(err)
					}
					found = found || strings.Contains(detail, "rounds_monitor_finished")
				}
				if err := planRows.Err(); err != nil {
					t.Fatal(err)
				}
				if !found {
					t.Fatal("finished-time range query did not select the new monitor index")
				}
			}
		})
	}
}
