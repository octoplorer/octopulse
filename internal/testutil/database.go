// Package testutil supplies an isolated database for shared business acceptance
// tests. Set OCTOPULSE_TEST_DB_DRIVER=postgres to run the same tests on PostgreSQL.
package testutil

import (
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

func Database(t testing.TB) store.Config {
	t.Helper()
	if os.Getenv("OCTOPULSE_TEST_DB_DRIVER") != "postgres" {
		return store.Config{Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "test.db")}
	}
	dsn := os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Fatal("PostgreSQL business tests require OCTOPULSE_TEST_POSTGRES_DSN")
	}
	pool, e := sql.Open("pgx", dsn)
	if e != nil {
		t.Fatal(e)
	}
	schema := "octopulse_business_" + domain.ID()
	if _, e = pool.Exec("CREATE SCHEMA " + schema); e != nil {
		pool.Close()
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if _, e := pool.Exec("DROP SCHEMA " + schema + " CASCADE"); e != nil {
			t.Errorf("drop test schema: %v", e)
		}
		pool.Close()
	})
	u, e := url.Parse(dsn)
	if e != nil {
		t.Fatal(e)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return store.Config{Driver: "postgres", DSN: u.String()}
}
