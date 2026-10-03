package beszel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"golang.org/x/sync/singleflight"
)

const requestTimeout = 15 * time.Second

// A session owns credentials for exactly one configuration generation. Old
// requests never read a new account's token, even when the Hub URL is unchanged.
type clientSession struct {
	config     domain.BeszelConfig
	generation uint64
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	token      string
	tokenUntil int64
	auth       singleflight.Group
	prepare    singleflight.Group
}

func (c *Client) resetLocked() {
	if c.session != nil {
		c.session.cancel()
	}
	c.generation++
	c.session = nil
	c.systems = SystemsResponse{}
	c.history = map[string]HistoryResponse{}
	c.containers = map[string]ContainersResponse{}
}

func (c *Client) wakePoller() {
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

// Invalidate discards in-memory state and cancels the previous generation.
// Call UpdateConfig when changing persisted configuration and snapshots together.
func (c *Client) Invalidate() {
	c.commit <- struct{}{}
	c.mu.Lock()
	c.resetLocked()
	c.mu.Unlock()
	<-c.commit
	c.wakePoller()
}

// UpdateConfig serializes configuration transactions with snapshot commits.
// The callback must persist the configuration and delete its previous snapshot
// atomically. Network operations never hold this commit guard.
func (c *Client) UpdateConfig(ctx context.Context, update func() error) error {
	select {
	case c.commit <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-c.commit }()
	if err := update(); err != nil {
		return err
	}
	c.mu.Lock()
	c.resetLocked()
	c.mu.Unlock()
	c.wakePoller()
	return nil
}

func (c *Client) loadSession(ctx context.Context) (*clientSession, error) {
	for {
		c.mu.RLock()
		if c.closed {
			c.mu.RUnlock()
			return nil, context.Canceled
		}
		generation := c.generation
		c.mu.RUnlock()
		cfg, err := c.loadConfig(ctx)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return nil, context.Canceled
		}
		if generation != c.generation {
			c.mu.Unlock()
			continue
		}
		if c.session != nil && c.session.config != cfg {
			c.resetLocked()
		}
		if c.session == nil {
			sessionCtx, cancel := context.WithCancel(context.Background())
			c.session = &clientSession{config: cfg, generation: c.generation, ctx: sessionCtx, cancel: cancel}
		}
		session := c.session
		c.mu.Unlock()
		return session, nil
	}
}

// Close rejects new work, cancels the active generation, and waits for all
// shared operations (including detached authentication) before the store closes.
func (c *Client) Close() {
	c.mu.Lock()
	c.closed = true
	if c.session != nil {
		c.session.cancel()
	}
	c.mu.Unlock()
	c.work.Wait()
}

// Each caller can leave independently. Shared work has its own deadline and is
// cancelled on configuration changes; one departing caller cannot cancel peers.
func (c *Client) fetch(ctx context.Context, session *clientSession, key string, fn func(context.Context) (any, error)) (any, error) {
	return c.sharedWork(ctx, session, &c.flights, fmt.Sprintf("%d:%s", session.generation, key), fn)
}

func (c *Client) sharedWork(ctx context.Context, session *clientSession, group *singleflight.Group, key string, fn func(context.Context) (any, error)) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, context.Canceled
	}
	// Register before launching so Close cannot finish before a queued callback
	// starts. Each waiter is counted, including those whose caller leaves early.
	c.work.Add(1)
	c.mu.Unlock()
	result := make(chan singleflight.Result, 1)
	go func() {
		defer c.work.Done()
		value, err, shared := group.Do(key, func() (any, error) {
			workCtx, cancel := context.WithTimeout(session.ctx, requestTimeout)
			defer cancel()
			if err := workCtx.Err(); err != nil {
				return nil, err
			}
			return fn(workCtx)
		})
		result <- singleflight.Result{Val: value, Err: err, Shared: shared}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-session.ctx.Done():
		return nil, session.ctx.Err()
	case result := <-result:
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := session.ctx.Err(); err != nil {
			return nil, err
		}
		return result.Val, result.Err
	}
}
