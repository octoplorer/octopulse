package store

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/octoplorer/octopulse/db"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

const SchemaVersion int64 = 1

// Open takes a runtime lock before applying migrations. It never logs the DSN.
func Open(ctx context.Context, cfg Config) (_ *Store, err error) {
	if cfg.Driver == "" {
		cfg.Driver = "sqlite"
	}
	if cfg.DSN == "" && cfg.Driver == "sqlite" {
		cfg.DSN = "file:data/octopulse.db"
	}
	s := &Store{driver: cfg.Driver}
	defer func() {
		if err != nil {
			_ = s.Close()
		}
	}()
	var migrationDialect goose.Dialect
	var migrationDir string
	switch cfg.Driver {
	case "sqlite":
		path, pathErr := sqlitePath(cfg.DSN)
		if pathErr != nil {
			return nil, pathErr
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, fmt.Errorf("prepare data directory: %w", err)
		}
		s.lock, err = acquireFileLock(path + ".lock")
		if err != nil {
			return nil, err
		}
		writerDSN := sqliteDSN(path, false)
		s.write, err = sql.Open("sqlite", writerDSN)
		if err != nil {
			return nil, fmt.Errorf("open SQLite: %w", err)
		}
		s.write.SetMaxOpenConns(1)
		s.write.SetMaxIdleConns(1)
		if err = s.write.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("connect SQLite: %w", err)
		}
		if err = os.Chmod(path, 0600); err != nil {
			return nil, fmt.Errorf("secure SQLite database: %w", err)
		}
		var journal, version string
		if err = s.write.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journal); err != nil {
			return nil, err
		}
		if strings.ToLower(journal) != "wal" {
			return nil, fmt.Errorf("SQLite requires WAL journal mode")
		}
		if err = s.write.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version); err != nil {
			return nil, err
		}
		if !sqliteVersionSafe(version) {
			return nil, fmt.Errorf("SQLite %s lacks required WAL reset fix (minimum 3.51.3)", version)
		}
		migrationDialect, migrationDir = goose.DialectSQLite3, "sqlite/migrations"
	case "postgres", "postgresql":
		s.driver = "postgres"
		s.write, err = sql.Open("pgx", cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("open PostgreSQL")
		}
		maxConns := cfg.MaxConnections
		if maxConns == 0 {
			maxConns = 10
		}
		if maxConns < 2 {
			return nil, fmt.Errorf("PostgreSQL requires at least two pool connections including its runtime lock")
		}
		s.write.SetMaxOpenConns(maxConns)
		s.write.SetMaxIdleConns(min(maxConns, 4))
		s.write.SetConnMaxLifetime(30 * time.Minute)
		if err = s.write.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("connect PostgreSQL: %w", err)
		}
		s.lock, err = acquirePostgresLock(ctx, s.write)
		if err != nil {
			return nil, err
		}
		migrationDialect, migrationDir = goose.DialectPostgres, "postgres/migrations"
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
	migrations, err := fs.Sub(db.Migrations, migrationDir)
	if err != nil {
		return nil, err
	}
	provider, err := goose.NewProvider(migrationDialect, s.write, migrations)
	if err != nil {
		return nil, fmt.Errorf("initialize migrations: %w", err)
	}
	version, err := provider.GetDBVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("read database version: %w", err)
	}
	if version > SchemaVersion {
		return nil, fmt.Errorf("database schema version %d is newer than supported %d", version, SchemaVersion)
	}
	if _, err = provider.Up(ctx); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if s.driver == "sqlite" {
		path, _ := sqlitePath(cfg.DSN)
		s.read, err = sql.Open("sqlite", sqliteDSN(path, true))
		if err != nil {
			return nil, err
		}
		s.read.SetMaxOpenConns(4)
		s.read.SetMaxIdleConns(4)
		if err = s.read.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("open SQLite read pool: %w", err)
		}
	} else {
		s.read = s.write
	}
	return s, nil
}

func sqlitePath(dsn string) (string, error) {
	if strings.Contains(dsn, "mode=memory") || strings.Contains(dsn, ":memory:") {
		return "", fmt.Errorf("SQLite requires a persistent local file")
	}
	raw := strings.SplitN(dsn, "?", 2)[0]
	raw = strings.TrimPrefix(raw, "file:")
	decoded, err := url.PathUnescape(raw)
	if err != nil || decoded == "" {
		return "", fmt.Errorf("invalid SQLite database path")
	}
	path, err := filepath.Abs(decoded)
	if err != nil {
		return "", err
	}
	// Existing paths, including symlinks, must resolve to a common process lock.
	if resolved, e := filepath.EvalSymlinks(path); e == nil {
		path = resolved
	} else if parent, e := filepath.EvalSymlinks(filepath.Dir(path)); e == nil {
		path = filepath.Join(parent, filepath.Base(path))
	}
	return path, nil
}

func sqliteDSN(path string, readOnly bool) string {
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "synchronous(FULL)")
	if readOnly {
		q.Set("mode", "ro")
		q.Add("_pragma", "query_only(1)")
	} else {
		q.Add("_pragma", "journal_mode(WAL)")
		q.Set("_txlock", "immediate")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func sqliteVersionSafe(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	n := make([]int, 3)
	for i, p := range parts {
		value, err := strconv.Atoi(p)
		if err != nil {
			return false
		}
		n[i] = value
	}
	return n[0] > 3 || (n[0] == 3 && (n[1] > 51 || (n[1] == 51 && n[2] >= 3)))
}

// LockLost closes if the PostgreSQL lock connection disappears. A caller must
// stop scheduling and terminate when it receives this signal.
func (s *Store) LockLost() <-chan struct{} {
	if s.lock == nil {
		return nil
	}
	return s.lock.lost
}

func (s *Store) Close() error {
	if s == nil {
		return nil
	}
	var first error
	if s.read != nil && s.read != s.write {
		first = s.read.Close()
	}
	if s.lock != nil {
		if err := s.lock.close(); first == nil {
			first = err
		}
		s.lock = nil
	}
	if s.write != nil {
		if err := s.write.Close(); first == nil {
			first = err
		}
		s.write = nil
	}
	return first
}
