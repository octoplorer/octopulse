// Package engine schedules local probes and atomically commits observations,
// confirmed states, state intervals and durable notification jobs.
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/store"
)

var (
	ErrBusy       = errors.New("monitor already has an active check")
	ErrPaused     = errors.New("monitor is paused")
	ErrSuperseded = errors.New("check superseded by current configuration")
)

const StatePaused = "paused"
const collectionWatermark = "collection"

type inFlight struct {
	cancel              context.CancelFunc
	version, generation int64
}
type schedule struct {
	version           int64
	next              int64
	enabled           bool
	maintenance       bool
	pendingEvaluation bool
}
type Engine struct {
	Store  *store.Store
	Runner *probe.Runner
	// Now and Attempt permit deterministic state-machine tests without replacing
	// production scheduling. Attempt defaults to Runner.Run.
	Now          func() time.Time
	Attempt      func(context.Context, domain.Monitor) probe.Result
	PollInterval time.Duration
	OnError      func(error)
	mu           sync.Mutex
	running      map[string]inFlight
	schedules    map[string]schedule
	semaphore    chan struct{}
	wake         chan struct{}
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	started      bool
}

type Metadata struct {
	ConfigVersion          int64                    `json:"configVersion"`
	InitializedAt          int64                    `json:"initializedAt"`
	CycleID                string                   `json:"cycleId"`
	FaultQueued            bool                     `json:"faultQueued"`
	LastReminderAt         int64                    `json:"lastReminderAt"`
	MaintenanceActive      bool                     `json:"maintenanceActive"`
	EvaluationAfter        int64                    `json:"evaluationAfter"`
	PendingRecovery        bool                     `json:"pendingRecovery"`
	HasHeartbeatReport     bool                     `json:"hasHeartbeatReport"`
	HeartbeatSuccess       bool                     `json:"heartbeatSuccess"`
	CertificateFingerprint string                   `json:"certificateFingerprint"`
	CertificateWarnings    []int                    `json:"certificateWarnings"`
	Certificate            *probe.CertificateResult `json:"certificate,omitempty"`
}

// NotificationPayload is copied into the event and each channel's outbox job.
// The worker checks generation/state/maintenance and records sent down messages
// by cycleId + channelId before accepting recovery jobs.
type NotificationPayload = domain.NotificationPayload

func New(st *store.Store, runner *probe.Runner) *Engine {
	e := &Engine{Store: st, Runner: runner, Now: time.Now, PollInterval: time.Second, running: map[string]inFlight{}, schedules: map[string]schedule{}, semaphore: make(chan struct{}, 100), wake: make(chan struct{}, 1)}
	if runner != nil {
		e.Attempt = runner.Run
	}
	return e
}
func (e *Engine) now() int64 {
	if e.Now == nil {
		return time.Now().UTC().UnixMilli()
	}
	return e.Now().UTC().UnixMilli()
}
func (e *Engine) report(err error) {
	if err != nil && !errors.Is(err, ErrBusy) && !errors.Is(err, ErrPaused) && !errors.Is(err, ErrSuperseded) && !errors.Is(err, context.Canceled) && e.OnError != nil {
		e.OnError(err)
	}
}

func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.started {
		e.mu.Unlock()
		return errors.New("engine already started")
	}
	e.ctx, e.cancel = context.WithCancel(ctx)
	e.started = true
	e.mu.Unlock()
	if err := e.initialize(e.ctx); err != nil {
		e.cancel()
		e.mu.Lock()
		e.started = false
		e.mu.Unlock()
		return err
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		interval := e.PollInterval
		if interval <= 0 {
			interval = time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		e.report(e.poll(e.ctx))
		for {
			select {
			case <-e.ctx.Done():
				e.mu.Lock()
				for _, active := range e.running {
					active.cancel()
				}
				e.mu.Unlock()
				e.closeCollection()
				return
			case <-ticker.C:
				e.report(e.poll(e.ctx))
			case <-e.wake:
				e.report(e.poll(e.ctx))
			}
		}
	}()
	return nil
}
func (e *Engine) Stop() {
	e.mu.Lock()
	cancel := e.cancel
	e.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	e.wg.Wait()
}
func (e *Engine) Wake() {
	select {
	case e.wake <- struct{}{}:
	default:
	}
}

func (e *Engine) initialize(ctx context.Context) error {
	monitors, err := e.Store.ListMonitors(ctx)
	if err != nil {
		return err
	}
	now := e.now()
	return e.Store.WithTx(ctx, func(tx *store.Tx) error {
		watermark, err := tx.Watermark(ctx, collectionWatermark)
		if errors.Is(err, store.ErrNotFound) {
			watermark = now
		} else if err != nil {
			return err
		}
		if watermark > now {
			watermark = now
		}
		if err = tx.CutOpenIntervals(ctx, watermark); err != nil {
			return err
		}
		for _, record := range monitors {
			current, err := tx.GetMonitor(ctx, record.ID)
			if err != nil {
				return err
			}
			m, err := decodeMonitor(current)
			if err != nil {
				return err
			}
			previous, err := tx.GetRuntime(ctx, m.ID)
			exists := err == nil
			if err != nil && !errors.Is(err, store.ErrNotFound) {
				return err
			}
			current.Generation++
			if err = tx.PutMonitor(ctx, current); err != nil {
				return err
			}
			next := store.Runtime{MonitorID: m.ID, ConfigVersion: current.ConfigVersion, Generation: current.Generation, State: domain.StateUnknown, HeartbeatVersion: previous.HeartbeatVersion + 1, HeartbeatAt: previous.HeartbeatAt}
			if !current.Enabled {
				next.State = StatePaused
			}
			if err = tx.PutRuntime(ctx, next); err != nil {
				return err
			}
			meta, err := getMetadata(ctx, tx, m, current, now)
			if err != nil {
				return err
			}
			if err = tx.Put(ctx, "engineMonitor", m.ID, meta); err != nil {
				return err
			}
			if m.IsAvailability() {
				start := now
				if exists {
					start = watermark
				}
				if m.CreatedAt > start {
					start = m.CreatedAt
				}
				if err = tx.PutInterval(ctx, store.Interval{ID: domain.ID(), MonitorID: m.ID, State: next.State, StartedAt: start}); err != nil {
					return err
				}
			}
		}
		return tx.PutWatermark(ctx, collectionWatermark, now)
	})
}

func (e *Engine) closeCollection() {
	// Last persisted scheduler tick is the safe crash boundary. A cancelled
	// in-flight observation is never turned into target Down during shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	e.report(e.Store.WithTx(ctx, func(tx *store.Tx) error {
		at, err := tx.Watermark(ctx, collectionWatermark)
		if err != nil {
			return err
		}
		return tx.CutOpenIntervals(ctx, at)
	}))
}

func decodeMonitor(record store.Monitor) (domain.Monitor, error) {
	var m domain.Monitor
	if err := json.Unmarshal(record.ConfigJSON, &m); err != nil {
		return m, fmt.Errorf("invalid stored monitor configuration: %w", err)
	}
	m.ID = record.ID
	m.Type = record.Kind
	m.Enabled = record.Enabled
	m.ConfigVersion = record.ConfigVersion
	m.Defaults()
	if err := m.Validate(); err != nil {
		return m, fmt.Errorf("invalid monitor %s: %w", record.ID, err)
	}
	return m, nil
}

func getMetadata(ctx context.Context, tx *store.Tx, m domain.Monitor, record store.Monitor, now int64) (Metadata, error) {
	var meta Metadata
	err := tx.Get(ctx, "engineMonitor", m.ID, &meta)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return meta, err
	}
	if meta.ConfigVersion != record.ConfigVersion {
		meta = Metadata{ConfigVersion: record.ConfigVersion, InitializedAt: now}
	}
	if meta.InitializedAt == 0 {
		meta.InitializedAt = now
	}
	return meta, nil
}

func (e *Engine) ensureRuntime(ctx context.Context, record store.Monitor, m domain.Monitor) (store.Runtime, error) {
	if existing, err := e.Store.GetRuntime(ctx, m.ID); err == nil && existing.ConfigVersion == record.ConfigVersion && existing.Generation == record.Generation && ((record.Enabled && existing.State != StatePaused) || (!record.Enabled && existing.State == StatePaused)) {
		return existing, nil
	}
	var runtime store.Runtime
	err := e.Store.WithTx(ctx, func(tx *store.Tx) error {
		current, err := tx.GetMonitor(ctx, record.ID)
		if err != nil {
			return err
		}
		if current.ConfigVersion != record.ConfigVersion || current.Generation != record.Generation {
			return ErrSuperseded
		}
		previous, err := tx.GetRuntime(ctx, m.ID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
		state := domain.StateUnknown
		if !current.Enabled {
			state = StatePaused
		}
		if err == nil && previous.ConfigVersion == current.ConfigVersion && previous.Generation == current.Generation && ((current.Enabled && previous.State != StatePaused) || (!current.Enabled && previous.State == StatePaused)) {
			runtime = previous
			return nil
		}
		current.Generation++
		if err = tx.PutMonitor(ctx, current); err != nil {
			return err
		}
		runtime = store.Runtime{MonitorID: m.ID, ConfigVersion: current.ConfigVersion, Generation: current.Generation, State: state, HeartbeatVersion: previous.HeartbeatVersion + 1, HeartbeatAt: previous.HeartbeatAt}
		if err = tx.PutRuntime(ctx, runtime); err != nil {
			return err
		}
		meta, err := getMetadata(ctx, tx, m, record, e.now())
		if err != nil {
			return err
		}
		if err = tx.Put(ctx, "engineMonitor", m.ID, meta); err != nil {
			return err
		}
		if m.IsAvailability() {
			return tx.ReplaceInterval(ctx, store.Interval{ID: domain.ID(), MonitorID: m.ID, State: state, StartedAt: e.now()})
		}
		return nil
	})
	return runtime, err
}

func (e *Engine) NotifyConfigurationChanged(ctx context.Context, id string) error {
	e.mu.Lock()
	if active, ok := e.running[id]; ok {
		active.cancel()
	}
	delete(e.schedules, id)
	e.mu.Unlock()
	err := e.Store.WithTx(ctx, func(tx *store.Tx) error {
		record, err := tx.GetMonitor(ctx, id)
		if err != nil {
			return err
		}
		m, err := decodeMonitor(record)
		if err != nil {
			return err
		}
		old, err := tx.GetRuntime(ctx, id)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
		record.Generation++
		if err = tx.PutMonitor(ctx, record); err != nil {
			return err
		}
		state := domain.StateUnknown
		if !record.Enabled {
			state = StatePaused
		}
		next := store.Runtime{MonitorID: id, ConfigVersion: record.ConfigVersion, Generation: record.Generation, State: state, HeartbeatVersion: old.HeartbeatVersion + 1, HeartbeatAt: old.HeartbeatAt}
		if err = tx.PutRuntime(ctx, next); err != nil {
			return err
		}
		meta, err := getMetadata(ctx, tx, m, record, e.now())
		if err != nil {
			return err
		}
		if err = tx.Put(ctx, "engineMonitor", id, meta); err != nil {
			return err
		}
		if m.IsAvailability() {
			return tx.ReplaceInterval(ctx, store.Interval{ID: domain.ID(), MonitorID: id, State: state, StartedAt: e.now()})
		}
		return nil
	})
	e.Wake()
	return err
}

func (e *Engine) poll(ctx context.Context) error {
	monitors, err := e.Store.ListMonitors(ctx)
	if err != nil {
		return err
	}
	maintenance, err := e.Store.List(ctx, "maintenance")
	if err != nil {
		return err
	}
	now := e.now()
	seen := map[string]bool{}
	for _, record := range monitors {
		seen[record.ID] = true
		m, err := decodeMonitor(record)
		if err != nil {
			e.report(err)
			continue
		}
		runtime, err := e.ensureRuntime(ctx, record, m)
		if err != nil {
			e.report(err)
			continue
		}
		runtime, err = e.expireCollection(ctx, record, m, runtime, now)
		if err != nil {
			e.report(err)
			continue
		}
		record.Generation = runtime.Generation
		activeMaintenance := inMaintenance(maintenance, m.ID, now)
		e.mu.Lock()
		plan, present := e.schedules[m.ID]
		if active, ok := e.running[m.ID]; ok && (active.version != record.ConfigVersion || active.generation != record.Generation || !record.Enabled) {
			active.cancel()
		}
		if !present || plan.version != record.ConfigVersion || plan.enabled != record.Enabled {
			plan = schedule{version: record.ConfigVersion, next: now, enabled: record.Enabled, maintenance: activeMaintenance, pendingEvaluation: true}
		}
		maintenanceEnded := plan.maintenance && !activeMaintenance
		maintenanceEntered := !plan.maintenance && activeMaintenance
		plan.maintenance = activeMaintenance
		if maintenanceEnded {
			plan.pendingEvaluation = true
		}
		if !record.Enabled {
			e.schedules[m.ID] = plan
			e.mu.Unlock()
			continue
		}
		due := now >= plan.next || plan.pendingEvaluation || m.Type == domain.MonitorHeartbeat
		if due {
			interval := record.IntervalMS
			if interval <= 0 {
				interval = int64(m.IntervalSeconds) * 1000
			}
			if plan.next <= now {
				plan.next += ((now-plan.next)/interval + 1) * interval
			}
			e.schedules[m.ID] = plan
		} else {
			e.schedules[m.ID] = plan
		}
		e.mu.Unlock()
		if maintenanceEntered {
			e.report(e.markMaintenance(ctx, m.ID))
		}
		if maintenanceEnded {
			e.report(e.markMaintenanceExit(ctx, m.ID, now))
		}
		if due {
			e.wg.Add(1)
			go func(id string) { defer e.wg.Done(); e.report(e.Check(ctx, id)) }(m.ID)
		}
	}
	e.mu.Lock()
	for id, active := range e.running {
		if !seen[id] {
			active.cancel()
		}
	}
	for id := range e.schedules {
		if !seen[id] {
			delete(e.schedules, id)
		}
	}
	e.mu.Unlock()
	return e.Store.WithTx(ctx, func(tx *store.Tx) error { return putCollectionWatermark(ctx, tx, now) })
}

func inMaintenance(raw []json.RawMessage, id string, now int64) bool {
	for _, data := range raw {
		var m domain.Maintenance
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		if m.StartsAt <= now && now < m.EndsAt {
			for _, monitorID := range m.MonitorIDs {
				if monitorID == id {
					return true
				}
			}
		}
	}
	return false
}

func (e *Engine) Check(ctx context.Context, id string) error {
	record, err := e.Store.GetMonitor(ctx, id)
	if err != nil {
		return err
	}
	m, err := decodeMonitor(record)
	if err != nil {
		return err
	}
	if !record.Enabled {
		return ErrPaused
	}
	roundContext, cancel := context.WithTimeout(ctx, time.Duration(m.IntervalSeconds)*time.Second)
	defer cancel()
	e.mu.Lock()
	if _, ok := e.running[id]; ok {
		e.mu.Unlock()
		return ErrBusy
	}
	e.running[id] = inFlight{cancel: cancel, version: record.ConfigVersion, generation: record.Generation}
	e.mu.Unlock()
	defer func() { e.mu.Lock(); delete(e.running, id); e.mu.Unlock() }()
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-roundContext.Done():
		return roundContext.Err()
	}
	runtime, err := e.ensureRuntime(roundContext, record, m)
	if err != nil {
		return err
	}
	record.Generation = runtime.Generation
	e.mu.Lock()
	active := e.running[id]
	active.generation = record.Generation
	e.running[id] = active
	e.mu.Unlock()
	if m.Type == domain.MonitorHeartbeat {
		return e.evaluateHeartbeat(roundContext, record, m, runtime)
	}
	if e.Attempt == nil {
		return errors.New("probe runner unavailable")
	}
	started := e.now()
	round := store.Round{ID: domain.ID(), MonitorID: id, ConfigVersion: record.ConfigVersion, Generation: record.Generation, StartedAt: started, Attempts: []store.Attempt{}}
	deadline, _ := roundContext.Deadline()
	probeDeadline := deadline.Add(-100 * time.Millisecond)
	attemptContext, stopAttempts := context.WithDeadline(roundContext, probeDeadline)
	defer stopAttempts()
	var result probe.Result
	for number := 0; number <= m.Retries; number++ {
		if attemptContext.Err() != nil {
			break
		}
		attemptStart := e.now()
		result = e.Attempt(attemptContext, m)
		finished := e.now()
		details, _ := json.Marshal(result)
		round.Attempts = append(round.Attempts, store.Attempt{Number: int64(number + 1), StartedAt: attemptStart, FinishedAt: finished, Success: result.Success, LatencyMS: result.LatencyMs, Error: result.Error, Detail: details})
		if result.Success {
			round.Success = true
			break
		}
		if number < m.Retries && m.RetryDelaySeconds > 0 {
			timer := time.NewTimer(time.Duration(m.RetryDelaySeconds) * time.Second)
			select {
			case <-timer.C:
			case <-attemptContext.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if roundContext.Err() != nil {
		return ErrSuperseded
	}
	if len(round.Attempts) == 0 {
		return errors.New("check budget exhausted before first attempt")
	}
	round.FinishedAt = e.now()
	round.LatencyMS = round.FinishedAt - started
	err = e.commit(roundContext, record, m, runtime, round, result, false)
	if err == nil {
		e.clearPendingEvaluation(roundContext, id)
	}
	return err
}

func (e *Engine) clearPendingEvaluation(ctx context.Context, id string) {
	var meta Metadata
	if e.Store.Get(ctx, "engineMonitor", id, &meta) == nil && meta.EvaluationAfter == 0 {
		e.mu.Lock()
		plan := e.schedules[id]
		plan.pendingEvaluation = false
		e.schedules[id] = plan
		e.mu.Unlock()
	}
}
