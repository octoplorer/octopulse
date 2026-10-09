package engine

import (
	"container/list"
	"context"
	"fmt"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type checkTask struct {
	lifetime, ctx context.Context
	cancel        context.CancelFunc
	record        store.Monitor
	monitor       domain.Monitor
	epoch         uint64
	managed       bool
	acceptedAt    time.Time
	queued        *list.Element
	stopQueued    func() bool
	result        chan error
	abandoned     chan struct{}
}

type EngineStats struct {
	Active int
	Queued int
}

// Stats reports round execution and admission, independently of monitor count.
func (e *Engine) Stats() EngineStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	return EngineStats{Active: e.active, Queued: e.queue.Len()}
}

// Check accepts one non-overlapping round and waits for its result. Once the
// engine has started, the round belongs to the engine lifecycle: cancelling the
// caller stops waiting without shortening a target's configured probe budget.
// Without Start, the caller supplies the lifetime for standalone execution.
func (e *Engine) Check(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	record, err := e.Store.GetMonitor(ctx, id)
	if err != nil {
		return err
	}
	m, err := decodeMonitor(record)
	if err != nil {
		return err
	}
	task, err := e.acceptRound(
		ctx,
		record,
		m,
		true,
	)
	if err != nil {
		return err
	}
	if !task.managed {
		return e.runStandalone(task)
	}
	select {
	case err := <-task.result:
		return err
	case <-ctx.Done():
		// Either the caller or the worker drains the detached result, so there is
		// no goroutine waiting on behalf of a caller who has already left.
		close(task.abandoned)
		select {
		case err := <-task.result:
			e.report(err)
		default:
		}
		return fmt.Errorf("%w: %w", ErrCheckAccepted, ctx.Err())
	}
}

func (e *Engine) acceptRound(
	ctx context.Context,
	record store.Monitor,
	m domain.Monitor,
	manual bool,
) (*checkTask, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !record.Enabled {
		return nil, ErrPaused
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	lifetime := e.ctx
	if lifetime == nil {
		lifetime = ctx
	}
	if err := lifetime.Err(); err != nil {
		return nil, err
	}
	if _, ok := e.running[m.ID]; ok {
		return nil, ErrBusy
	}
	roundContext, cancel := context.WithTimeout(lifetime, time.Duration(m.IntervalSeconds)*time.Second)
	task := &checkTask{
		lifetime:   lifetime,
		ctx:        roundContext,
		cancel:     cancel,
		record:     record,
		monitor:    m,
		epoch:      e.schedules[m.ID].evaluationEpoch,
		managed:    e.started,
		acceptedAt: time.Now(),
	}
	if manual && task.managed {
		task.result = make(chan error, 1)
		task.abandoned = make(chan struct{})
	}
	e.running[m.ID] = inFlight{cancel: cancel, version: record.ConfigVersion, generation: record.Generation}
	if task.managed {
		// Admission is limited to one entry per persisted monitor, shared by
		// scheduled and manual checks. Queue size cannot grow with repeated
		// polls or repeated manual requests for the same monitor.
		task.queued = e.queue.PushBack(task)
		// AfterFunc adds no waiting goroutine. Cancelled queued rounds release
		// ownership promptly, even while every worker is occupied.
		e.wg.Add(1)
		task.stopQueued = context.AfterFunc(roundContext, func() {
			defer e.wg.Done()
			e.mu.Lock()
			if task.queued == nil {
				e.mu.Unlock()
				return
			}
			e.queue.Remove(task.queued)
			task.queued = nil
			e.mu.Unlock()
			e.finish(task, roundContext.Err())
		})
		close(e.queueChanged)
		e.queueChanged = make(chan struct{})
	}
	return task, nil
}

func (e *Engine) finish(task *checkTask, err error) {
	cancelled := task.ctx.Err() != nil
	task.cancel()
	e.mu.Lock()
	delete(e.running, task.monitor.ID)
	retryEvaluation := e.schedules[task.monitor.ID].pendingEvaluation && cancelled && task.lifetime.Err() == nil
	e.mu.Unlock()
	if retryEvaluation {
		e.Wake()
	}
	if task.result == nil {
		e.report(err)
		return
	}
	task.result <- err
	select {
	case <-task.abandoned:
		select {
		case err := <-task.result:
			e.report(err)
		default:
		}
	default:
	}
}

func (e *Engine) worker() {
	defer e.wg.Done()
	for {
		e.mu.Lock()
		if e.queue.Len() == 0 {
			changed := e.queueChanged
			e.mu.Unlock()
			select {
			case <-e.ctx.Done():
				return
			case <-changed:
				continue
			}
		}
		task := e.queue.Remove(e.queue.Front()).(*checkTask)
		task.queued = nil
		if task.stopQueued() {
			e.wg.Done()
		}
		e.active++
		e.mu.Unlock()
		if e.OnRoundStart != nil && task.ctx.Err() == nil {
			e.OnRoundStart(time.Since(task.acceptedAt))
		}
		err := e.checkRound(
			task.lifetime,
			task.ctx,
			task.record,
			task.monitor,
			task.epoch,
		)
		e.mu.Lock()
		e.active--
		e.mu.Unlock()
		e.finish(task, err)
	}
}

func (e *Engine) runStandalone(task *checkTask) error {
	defer task.cancel()
	defer func() {
		e.mu.Lock()
		delete(e.running, task.monitor.ID)
		e.mu.Unlock()
	}()
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-task.ctx.Done():
		return task.ctx.Err()
	}
	e.mu.Lock()
	e.active++
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		e.active--
		e.mu.Unlock()
	}()
	return e.checkRound(
		task.lifetime,
		task.ctx,
		task.record,
		task.monitor,
		task.epoch,
	)
}
