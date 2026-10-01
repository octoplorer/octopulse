package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreContract(t *testing.T) {
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
			if _, err = Open(ctx, cfg); !errors.Is(err, ErrLocked) {
				t.Fatalf("second instance: %v", err)
			}
			t.Run("atomic_round_state_event_outbox", func(t *testing.T) {
				m := seedMonitor(t, s, "atomic")
				r := Round{ID: "round-atomic", MonitorID: m.ID, ConfigVersion: 1, Generation: 1, StartedAt: 100, FinishedAt: 130, Success: false, LatencyMS: 30, Attempts: []Attempt{{Number: 1, StartedAt: 100, FinishedAt: 130, LatencyMS: 30, Error: "connection refused"}}}
				abort := errors.New("abort")
				err = s.WithTx(ctx, func(tx *Tx) error {
					if err := tx.PutRound(ctx, r); err != nil {
						return err
					}
					if err := tx.PutRuntime(ctx, Runtime{MonitorID: m.ID, ConfigVersion: 1, Generation: 1, State: "down", Failures: 1, LastRoundID: r.ID, LastCollectedAt: 130}); err != nil {
						return err
					}
					if err := tx.ReplaceInterval(ctx, Interval{ID: "interval-atomic", MonitorID: m.ID, State: "down", StartedAt: 130}); err != nil {
						return err
					}
					if err := tx.PutEvent(ctx, Event{ID: "event-atomic", MonitorID: m.ID, Generation: 1, Kind: "down", CreatedAt: 130}); err != nil {
						return err
					}
					if err := tx.PutDelivery(ctx, Delivery{ID: "delivery-atomic", EventID: "event-atomic", ChannelID: "mail", Generation: 1, DueAt: 130}); err != nil {
						return err
					}
					return abort
				})
				if !errors.Is(err, abort) {
					t.Fatal(err)
				}
				if exists, _ := s.RoundExists(ctx, r.ID); exists {
					t.Fatal("round survived rollback")
				}
				if _, err = s.GetEvent(ctx, "event-atomic"); !errors.Is(err, ErrNotFound) {
					t.Fatal("event survived rollback", err)
				}
				state, err := s.GetRuntime(ctx, m.ID)
				if err != nil || state.State != "unknown" {
					t.Fatal(state, err)
				}
				if err = s.WithTx(ctx, func(tx *Tx) error { return tx.PutRound(ctx, r) }); err != nil {
					t.Fatal(err)
				}
				if err = s.WithTx(ctx, func(tx *Tx) error { return tx.PutRound(ctx, r) }); err != nil {
					t.Fatal("idempotent round", err)
				}
				rounds, err := s.ListRounds(ctx, m.ID, 0, 100)
				if err != nil || len(rounds) != 1 || len(rounds[0].Attempts) != 1 {
					t.Fatal(rounds, err)
				}
			})
			t.Run("constraints_and_stale_results", func(t *testing.T) {
				m := seedMonitor(t, s, "constraints")
				old, err := s.GetRuntime(ctx, m.ID)
				if err != nil {
					t.Fatal(err)
				}
				if err = s.WithTx(ctx, func(tx *Tx) error {
					return tx.PutMonitor(ctx, Monitor{ID: m.ID, ConfigVersion: 2, Generation: 2, Kind: "http", Enabled: true, IntervalMS: 30000})
				}); err != nil {
					t.Fatal(err)
				}
				next := old
				next.State = "down"
				next.LastRoundID = "obsolete"
				if err = s.WithTx(ctx, func(tx *Tx) error { return tx.CompareRuntime(ctx, old, next) }); !errors.Is(err, ErrConflict) {
					t.Fatalf("stale runtime applied: %v", err)
				}
				if err = s.WithTx(ctx, func(tx *Tx) error {
					return tx.PutInterval(ctx, Interval{ID: "second-open", MonitorID: m.ID, State: "up", StartedAt: 100})
				}); !errors.Is(err, ErrConflict) {
					t.Fatalf("two open intervals: %v", err)
				}
				if err = s.WithTx(ctx, func(tx *Tx) error {
					return tx.PutRound(ctx, Round{ID: "orphan", MonitorID: "missing", ConfigVersion: 1, Generation: 1, StartedAt: 1, FinishedAt: 2})
				}); !errors.Is(err, ErrConflict) {
					t.Fatalf("orphan round: %v", err)
				}
			})
			t.Run("unique_page_bindings_rollback", func(t *testing.T) {
				if err = s.BindPage(ctx, PageBinding{PageID: "p1", Slug: "Status1", Domains: []string{"Status1.Example.com."}}); err != nil {
					t.Fatal(err)
				}
				if err = s.BindPage(ctx, PageBinding{PageID: "p2", Slug: "status2", Domains: []string{"status2.example.com"}}); err != nil {
					t.Fatal(err)
				}
				if err = s.BindPage(ctx, PageBinding{PageID: "p2", Slug: "renamed", Domains: []string{"STATUS1.EXAMPLE.COM"}}); !errors.Is(err, ErrConflict) {
					t.Fatalf("domain clash: %v", err)
				}
				b, err := s.PageBinding(ctx, "p2")
				if err != nil || b.Slug != "status2" || b.Domains[0] != "status2.example.com" {
					t.Fatal("binding edit partially committed", b, err)
				}
				if _, err = s.PageIDBySlug(ctx, "renamed"); !errors.Is(err, ErrNotFound) {
					t.Fatal(err)
				}
			})
			t.Run("lease_recovery_and_completion_marker", func(t *testing.T) {
				m := seedMonitor(t, s, "leases")
				err = s.WithTx(ctx, func(tx *Tx) error {
					if err := tx.PutEvent(ctx, Event{ID: "lease-event", MonitorID: m.ID, Generation: 1, Kind: "down", CreatedAt: 100}); err != nil {
						return err
					}
					return tx.PutDelivery(ctx, Delivery{ID: "lease-job", EventID: "lease-event", ChannelID: "mail", Generation: 1, DueAt: 100})
				})
				if err != nil {
					t.Fatal(err)
				}
				first, err := s.ClaimDeliveries(ctx, 100, 50, 10, "worker-a")
				if err != nil || len(first) != 1 {
					t.Fatal(first, err)
				}
				before, err := s.ClaimDeliveries(ctx, 120, 50, 10, "worker-b")
				if err != nil || len(before) != 0 {
					t.Fatal(before, err)
				}
				after, err := s.ClaimDeliveries(ctx, 151, 50, 10, "worker-b")
				if err != nil || len(after) != 1 || after[0].Attempts != 2 {
					t.Fatal(after, err)
				}
				if err = s.CompleteDelivery(ctx, "lease-job", "worker-a", 152, "sent", 0, ""); !errors.Is(err, ErrLeaseLost) {
					t.Fatal("old token completed", err)
				}
				abort := errors.New("marker failed")
				err = s.WithTx(ctx, func(tx *Tx) error {
					if err := tx.CompleteDelivery(ctx, "lease-job", "worker-b", 152, "sent", 0, ""); err != nil {
						return err
					}
					return abort
				})
				if !errors.Is(err, abort) {
					t.Fatal(err)
				}
				d, err := s.GetDelivery(ctx, "lease-job")
				if err != nil || d.State != "sending" {
					t.Fatal(d, err)
				}
				err = s.WithTx(ctx, func(tx *Tx) error {
					if err := tx.CompleteDelivery(ctx, "lease-job", "worker-b", 153, "sent", 0, ""); err != nil {
						return err
					}
					return tx.Put(ctx, "delivery_marker", m.ID+"/mail", map[string]any{"event": "lease-event"})
				})
				if err != nil {
					t.Fatal(err)
				}
				d, err = s.GetDelivery(ctx, "lease-job")
				if err != nil || d.State != "sent" {
					t.Fatal(d, err)
				}
			})
			t.Run("aggregate_watermark_and_retention", func(t *testing.T) {
				m := seedMonitor(t, s, "aggregate")
				err = s.WithTx(ctx, func(tx *Tx) error {
					for i := int64(1); i <= 2; i++ {
						if err := tx.PutRound(ctx, Round{ID: fmt.Sprintf("agg-round-%d", i), MonitorID: m.ID, ConfigVersion: 1, Generation: 1, StartedAt: i * 100, FinishedAt: i*100 + 10, LatencyMS: 10, Attempts: []Attempt{{Number: 1, StartedAt: i * 100, FinishedAt: i*100 + 10, LatencyMS: 10}}}); err != nil {
							return err
						}
					}
					if err := tx.PutAggregate(ctx, Aggregate{MonitorID: m.ID, BucketAt: 0, WidthMS: 300000, UpMS: 150, DownMS: 50, RoundCount: 1}); err != nil {
						return err
					}
					return tx.PutWatermark(ctx, "aggregate", 150)
				})
				if err != nil {
					t.Fatal(err)
				}
				abort := errors.New("aggregation interrupted")
				err = s.WithTx(ctx, func(tx *Tx) error {
					if err := tx.PutAggregate(ctx, Aggregate{MonitorID: m.ID, BucketAt: 0, WidthMS: 300000, UpMS: 250, DownMS: 50, RoundCount: 2}); err != nil {
						return err
					}
					if err := tx.PutWatermark(ctx, "aggregate", 300); err != nil {
						return err
					}
					return abort
				})
				if !errors.Is(err, abort) {
					t.Fatal(err)
				}
				if watermark, err := s.Watermark(ctx, "aggregate"); err != nil || watermark != 150 {
					t.Fatal("watermark partially committed", watermark, err)
				}
				if err = s.WithTx(ctx, func(tx *Tx) error { return tx.PruneHistory(ctx, 300, 300, 150, 10) }); err != nil {
					t.Fatal(err)
				}
				rounds, err := s.ListRounds(ctx, m.ID, 0, 10)
				if err != nil || len(rounds) != 1 || rounds[0].StartedAt != 200 {
					t.Fatal("pruned beyond watermark", rounds, err)
				}
				aggs, err := s.Aggregates(ctx, m.ID, 0, 300000, 300000)
				if err != nil || len(aggs) != 1 || aggs[0].UpMS != 150 {
					t.Fatal(aggs, err)
				}
			})
			if backend == "sqlite" {
				t.Run("connection_pragmas_and_backup_restore", func(t *testing.T) {
					var connections []*sql.Conn
					for i := 0; i < 4; i++ {
						conn, err := s.read.Conn(ctx)
						if err != nil {
							t.Fatal(err)
						}
						connections = append(connections, conn)
						var fk, syncMode, busy, queryOnly int
						if err = conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fk); err != nil {
							t.Fatal(err)
						}
						conn.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&syncMode)
						conn.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busy)
						conn.QueryRowContext(ctx, "PRAGMA query_only").Scan(&queryOnly)
						if fk != 1 || syncMode != 2 || busy != 5000 || queryOnly != 1 {
							t.Fatal(fk, syncMode, busy, queryOnly)
						}
					}
					for _, conn := range connections {
						conn.Close()
					}
					if err = s.Put(ctx, "setting", "backup", map[string]string{"value": "durable"}); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(t.TempDir(), "restored.db")
					if err = s.BackupSQLite(ctx, path); err != nil {
						t.Fatal(err)
					}
					restored, err := Open(ctx, Config{Driver: "sqlite", DSN: path})
					if err != nil {
						t.Fatal(err)
					}
					defer restored.Close()
					var got map[string]string
					if err = restored.Get(ctx, "setting", "backup", &got); err != nil || got["value"] != "durable" {
						t.Fatal(got, err)
					}
				})
			} else {
				t.Run("postgres_backup_restore", func(t *testing.T) {
					dump, err := postgresCommand("pg_dump")
					if err != nil {
						t.Skip("pg_dump is not installed")
					}
					restore, err := postgresCommand("pg_restore")
					if err != nil {
						t.Skip("pg_restore is not installed")
					}
					u, err := url.Parse(cfg.DSN)
					if err != nil {
						t.Fatal(err)
					}
					schema := u.Query().Get("search_path")
					if schema == "" {
						t.Fatal("restore test requires isolated schema")
					}
					if err = s.Put(ctx, "settings", "restore", map[string]string{"value": "committed"}); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(t.TempDir(), "postgres.dump")
					dumpCmd := exec.Command(dump, "--format=custom", "--no-owner", "--schema="+schema, "--file="+path)
					dumpCmd.Env = postgresEnvironment(u)
					if output, err := dumpCmd.CombinedOutput(); err != nil {
						t.Fatalf("pg_dump: %s: %v", output, err)
					}
					if err = os.Chmod(path, 0600); err != nil {
						t.Fatal(err)
					}
					if err = s.Close(); err != nil {
						t.Fatal(err)
					}
					admin, err := sql.Open("pgx", cfg.DSN)
					if err != nil {
						t.Fatal(err)
					}
					if _, err = admin.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
						admin.Close()
						t.Fatal(err)
					}
					admin.Close()
					restoreCmd := exec.Command(restore, "--no-owner", "--exit-on-error", "--dbname="+strings.TrimPrefix(u.Path, "/"), path)
					restoreCmd.Env = postgresEnvironment(u)
					if output, err := restoreCmd.CombinedOutput(); err != nil {
						t.Fatalf("pg_restore: %s: %v", output, err)
					}
					s, err = Open(ctx, cfg)
					if err != nil {
						t.Fatal(err)
					}
					var value map[string]string
					if err = s.Get(ctx, "settings", "restore", &value); err != nil || value["value"] != "committed" {
						t.Fatal(value, err)
					}
					if err = s.Put(ctx, "settings", "restore-after", map[string]bool{"writable": true}); err != nil {
						t.Fatal(err)
					}
				})
				t.Run("postgres_runtime_lock_loss", func(t *testing.T) {
					var pid int
					if err := s.lock.conn.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&pid); err != nil {
						t.Fatal(err)
					}
					admin, err := sql.Open("pgx", cfg.DSN)
					if err != nil {
						t.Fatal(err)
					}
					var terminated bool
					err = admin.QueryRowContext(ctx, "SELECT pg_terminate_backend($1)", pid).Scan(&terminated)
					admin.Close()
					if err != nil || !terminated {
						t.Fatal("terminate lock session", terminated, err)
					}
					select {
					case <-s.LockLost():
					case <-time.After(12 * time.Second):
						t.Fatal("lost runtime lock was not detected")
					}
					_ = s.Close()
					s, err = Open(ctx, cfg)
					if err != nil {
						t.Fatal("reacquire released lock", err)
					}
				})
			}
			// A newer program may have upgraded the DB. Older binaries fail closed.
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			driver, dsn := "pgx", cfg.DSN
			if backend == "sqlite" {
				driver = "sqlite"
				path, _ := sqlitePath(cfg.DSN)
				dsn = sqliteDSN(path, false)
			}
			pool, err := sql.Open(driver, dsn)
			if err != nil {
				t.Fatal(err)
			}
			_, err = pool.ExecContext(ctx, `INSERT INTO goose_db_version(version_id,is_applied) VALUES(2,TRUE)`)
			pool.Close()
			if err != nil {
				t.Fatal(err)
			}
			if rejected, err := Open(ctx, cfg); err == nil {
				rejected.Close()
				t.Fatal("accepted newer database schema")
			}
		})
	}
}

func postgresCommand(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	path := filepath.Join("/opt/homebrew/opt/postgresql@18/bin", name)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("%s unavailable", name)
}

func postgresEnvironment(u *url.URL) []string {
	env := append(os.Environ(), "PGHOST="+u.Hostname(), "PGPORT="+u.Port(), "PGDATABASE="+strings.TrimPrefix(u.Path, "/"))
	if u.User != nil {
		env = append(env, "PGUSER="+u.User.Username())
		if password, ok := u.User.Password(); ok {
			env = append(env, "PGPASSWORD="+password)
		}
	}
	if sslmode := u.Query().Get("sslmode"); sslmode != "" {
		env = append(env, "PGSSLMODE="+sslmode)
	}
	return env
}

func testConfig(t *testing.T, backend string) Config {
	t.Helper()
	if backend == "sqlite" {
		return Config{Driver: backend, DSN: filepath.Join(t.TempDir(), "octopulse.db")}
	}
	dsn := os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN")
	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("octopulse_test_%d", time.Now().UnixNano())
	if _, err = pool.Exec(`CREATE SCHEMA ` + schema); err != nil {
		pool.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); pool.Close() })
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return Config{Driver: backend, DSN: u.String()}
}

func seedMonitor(t *testing.T, s *Store, id string) Monitor {
	t.Helper()
	ctx := context.Background()
	m := Monitor{ID: id, ConfigVersion: 1, Generation: 1, Kind: "http", Enabled: true, IntervalMS: 30000, ConfigJSON: json.RawMessage(`{"name":"test"}`)}
	err := s.WithTx(ctx, func(tx *Tx) error {
		if err := tx.PutMonitor(ctx, m); err != nil {
			return err
		}
		if err := tx.PutRuntime(ctx, Runtime{MonitorID: id, ConfigVersion: 1, Generation: 1, State: "unknown"}); err != nil {
			return err
		}
		return tx.PutInterval(ctx, Interval{ID: id + "-initial", MonitorID: id, State: "unknown", StartedAt: 0})
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}
