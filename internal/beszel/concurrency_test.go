package beszel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

func concurrentHub(t *testing.T, records http.HandlerFunc) (*Client, *atomic.Int64) {
	t.Helper()
	authCalls := &atomic.Int64{}
	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/users/auth-with-password":
			authCalls.Add(1)
			json.NewEncoder(w).Encode(map[string]any{"token": jwt(time.Now().Add(time.Hour).Unix()), "record": map[string]string{"collectionName": "users"}})
		case "/api/beszel/info":
			json.NewEncoder(w).Encode(map[string]string{"v": "0.20.0"})
		default:
			records(w, r)
		}
	}))
	t.Cleanup(hub.Close)
	return adapter(t, hub.URL, "reader@example.invalid", "password"), authCalls
}

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for concurrent operation")
		var zero T
		return zero
	}
}

func TestConcurrentResourcesCacheAndCallerCancellation(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	var slowCalls atomic.Int64
	c, authCalls := concurrentHub(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("filter"), `system="slow"`) {
			slowCalls.Add(1)
			select {
			case started <- struct{}{}:
			default:
			}
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		if strings.Contains(r.URL.Path, "system_stats") {
			json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"created": time.Now().UnixMilli(), "stats": map[string]int{"cpu": 12}}}, "totalPages": 1})
		} else {
			json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "totalPages": 1})
		}
	})
	ctx := context.Background()
	cached, err := c.History(ctx, "warm", "24h")
	if err != nil || cached.Stale {
		t.Fatal(cached, err)
	}

	results := make(chan error, 16)
	for i := 0; i < cap(results); i++ {
		go func() {
			result, err := c.History(ctx, "slow", "24h")
			if err == nil && (result.Stale || len(result.Items) != 1) {
				err = fmt.Errorf("unexpected history: %+v", result)
			}
			results <- err
		}()
	}
	receive(t, started)

	// A departing follower must not cancel the leader or the other followers.
	cancelCtx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()
	if _, err := c.History(cancelCtx, "slow", "24h"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("caller did not cancel", err)
	}
	quickCtx, stop := context.WithTimeout(ctx, time.Second)
	defer stop()
	if result, err := c.Containers(quickCtx, "other"); err != nil || result.Stale {
		t.Fatal("unrelated request blocked", result, err)
	}
	cached.Items[0].CPU = 100
	if result, err := c.History(quickCtx, "warm", "24h"); err != nil || result.Items[0].CPU != 12 {
		t.Fatal("cache blocked or shared mutable data", result, err)
	}

	once.Do(func() { close(release) })
	for i := 0; i < cap(results); i++ {
		if err := receive(t, results); err != nil {
			t.Fatal(err)
		}
	}
	if slowCalls.Load() != 1 || authCalls.Load() != 1 {
		t.Fatal("duplicate upstream work", slowCalls.Load(), authCalls.Load())
	}
}

func TestRemoteConcurrencyIsBoundedAndInvalidationCancelsWaiters(t *testing.T) {
	started := make(chan struct{}, 16)
	var active, peak atomic.Int64
	c, _ := concurrentHub(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "system_stats") && !strings.Contains(r.URL.Query().Get("filter"), `system="warm"`) {
			count := active.Add(1)
			defer active.Add(-1)
			for current := peak.Load(); count > current && !peak.CompareAndSwap(current, count); current = peak.Load() {
			}
			started <- struct{}{}
			<-r.Context().Done()
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "totalPages": 1})
	})
	defer c.Invalidate()
	if _, err := c.History(context.Background(), "warm", "24h"); err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 12)
	for i := 0; i < cap(results); i++ {
		go func(i int) {
			_, err := c.History(context.Background(), fmt.Sprintf("system%d", i), "24h")
			results <- err
		}(i)
	}
	for i := 0; i < 4; i++ {
		receive(t, started)
	}
	quickCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if result, err := c.History(quickCtx, "warm", "24h"); err != nil || result.Stale {
		t.Fatal("cache waited for admission", result, err)
	}
	if _, err := c.Containers(quickCtx, "other"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("admission wait ignored deadline", err)
	}
	if peak.Load() != 4 {
		t.Fatal("unexpected concurrency", peak.Load())
	}
	c.Invalidate()
	for i := 0; i < cap(results); i++ {
		if err := receive(t, results); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
	if result, err := c.Containers(context.Background(), "other"); err != nil || result.Stale {
		t.Fatal("new generation did not recover", result, err)
	}
}

func TestConfigUpdateDiscardsInflightResultsAndPersistedSnapshot(t *testing.T) {
	started := make(chan string, 3)
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	var old atomic.Bool
	old.Store(true)
	c, _ := concurrentHub(t, func(w http.ResponseWriter, r *http.Request) {
		if old.Load() {
			started <- r.URL.Path
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "totalPages": 1})
	})
	ctx := context.Background()
	results := make(chan error, 3)
	go func() { results <- c.Poll(ctx) }()
	go func() { _, err := c.History(ctx, "system1", "24h"); results <- err }()
	go func() { _, err := c.Containers(ctx, "system1"); results <- err }()
	for i := 0; i < 3; i++ {
		receive(t, started)
	}
	var cfg domain.BeszelConfig
	if err := c.Store.Get(ctx, "beszel", "config", &cfg); err != nil {
		t.Fatal(err)
	}
	// Same URL, different account: source URL alone must never retain the cache.
	cfg.Email = "new@example.invalid"
	if err := c.UpdateConfig(ctx, func() error {
		return c.Store.WithTx(ctx, func(tx *store.Tx) error {
			if err := tx.Put(ctx, "beszel", "config", cfg); err != nil {
				return err
			}
			if err := tx.Delete(ctx, "beszelSnapshots", "systems"); err != nil && !errors.Is(err, store.ErrNotFound) {
				return err
			}
			return nil
		})
	}); err != nil {
		t.Fatal(err)
	}
	old.Store(false)
	once.Do(func() { close(release) })
	for i := 0; i < 3; i++ {
		if err := receive(t, results); !errors.Is(err, context.Canceled) {
			t.Fatal("old generation completed", err)
		}
	}
	var persisted SystemsResponse
	if err := c.Store.Get(ctx, "beszelSnapshots", "systems", &persisted); !errors.Is(err, store.ErrNotFound) {
		t.Fatal("obsolete snapshot survived", persisted, err)
	}
	c.mu.RLock()
	if len(c.history) != 0 || len(c.containers) != 0 || c.systems.SyncedAt != 0 {
		t.Error("old data repopulated cache")
	}
	c.mu.RUnlock()
	if err := c.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	if err := c.Store.Get(ctx, "beszelSnapshots", "systems", &persisted); err != nil || persisted.SyncedAt == 0 {
		t.Fatal("new generation did not persist", persisted, err)
	}
}

func TestStartWaitsForSharedWorkOnShutdown(t *testing.T) {
	c, _ := concurrentHub(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "totalPages": 1})
	})
	session, err := c.loadSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	started, cancelled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	caller := make(chan error, 1)
	go func() {
		_, err := c.fetch(context.Background(), session, "shutdown", func(ctx context.Context) (any, error) {
			close(started)
			<-ctx.Done()
			close(cancelled)
			<-release // Simulate transport cleanup after cancellation.
			return nil, ctx.Err()
		})
		caller <- err
	}()
	receive(t, started)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { c.Start(ctx); close(done) }()
	cancel()
	receive(t, cancelled)
	select {
	case <-done:
		t.Fatal("Start returned before detached work stopped")
	default:
	}
	once.Do(func() { close(release) })
	receive(t, done)
	if err := receive(t, caller); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := c.Poll(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatal("closed client accepted work", err)
	}
}
