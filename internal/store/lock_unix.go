//go:build !windows

package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

type instanceLock struct {
	file *os.File
	conn *sql.Conn
	stop chan struct{}
	done chan struct{}
	lost chan struct{}
	once sync.Once
}

func acquireFileLock(path string) (*instanceLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, ErrLocked
	}
	return &instanceLock{file: f, lost: make(chan struct{})}, nil
}

func acquirePostgresLock(ctx context.Context, pool *sql.DB) (*instanceLock, error) {
	conn, err := pool.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var acquired bool
	if err = conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(850087,1)").Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("acquire PostgreSQL runtime lock: %w", err)
	}
	if !acquired {
		_ = conn.Close()
		return nil, ErrLocked
	}
	l := &instanceLock{conn: conn, stop: make(chan struct{}), done: make(chan struct{}), lost: make(chan struct{})}
	go func() {
		defer close(l.done)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-l.stop:
				return
			case <-ticker.C:
				checkCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				var held bool
				err := conn.QueryRowContext(
					checkCtx,
					"SELECT EXISTS(SELECT 1 FROM pg_locks WHERE locktype='advisory' AND pid=pg_backend_pid() "+
						"AND classid=850087 AND objid=1 AND granted)",
				).Scan(&held)
				cancel()
				if err != nil || !held {
					close(l.lost)
					return
				}
			}
		}
	}()
	return l, nil
}

func (l *instanceLock) close() error {
	var err error
	l.once.Do(func() {
		if l.conn != nil {
			close(l.stop)
			<-l.done
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_, _ = l.conn.ExecContext(ctx, "SELECT pg_advisory_unlock(850087,1)")
			err = l.conn.Close()
		}
		if l.file != nil {
			_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
			err = l.file.Close()
		}
	})
	return err
}
