package statistics

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

const fiveMinuteMS int64 = 300000
const hourMS int64 = 3600000
const dayMS int64 = 86400000

var ErrInvalidWindow = errors.New("statistics window must be nonempty, end at or before now, and span at most 410 days")

type Service struct {
	Store        *store.Store
	Now          func() time.Time
	Interval     time.Duration
	BatchBuckets int
	OnError      func(error)
	mu           sync.Mutex
	workMu       sync.Mutex
	wake         chan struct{}
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func New(st *store.Store) *Service {
	return &Service{Store: st, Now: time.Now, Interval: time.Minute, BatchBuckets: 288, wake: make(chan struct{}, 1)}
}
func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}
func (s *Service) window(from, to int64) (int64, int64, error) {
	if from < 0 || to <= from || to-from > 410*dayMS {
		return from, to, ErrInvalidWindow
	}
	now := s.now().UnixMilli()
	if to > now {
		to = now
	}
	if to <= from {
		return from, to, ErrInvalidWindow
	}
	return from, to, nil
}

func (s *Service) Availability(ctx context.Context, id string, from, to int64) (domain.Availability, error) {
	from, to, err := s.window(from, to)
	if err != nil {
		return domain.Availability{}, err
	}
	snapshot, err := s.Store.ReadStatistics(ctx, id, from, to, 0)
	if err != nil {
		return domain.Availability{}, err
	}
	if snapshot.Monitor.Kind == domain.MonitorCertificate {
		return domain.Availability{From: from, To: to, UnknownMs: to - from}, nil
	}
	return calculate(snapshot, from, to), nil
}

func (s *Service) retention(ctx context.Context) (domain.Retention, error) {
	settings := domain.DefaultSettings()
	err := s.Store.Get(ctx, "settings", "organization", &settings)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return domain.Retention{}, err
	}
	r := settings.Retention
	if r.RoundDays < 1 {
		r.RoundDays = 14
	}
	if r.AttemptDays < 1 {
		r.AttemptDays = 3
	}
	if r.FiveMinuteDays < 1 {
		r.FiveMinuteDays = 90
	}
	if r.HistoryMonths < 1 {
		r.HistoryMonths = 13
	}
	return r, nil
}

func (s *Service) Latency(ctx context.Context, id string, from, to int64) ([]domain.LatencyPoint, error) {
	from, to, err := s.window(from, to)
	if err != nil {
		return nil, err
	}
	retention, err := s.retention(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now().UnixMilli()
	fiveMinuteBefore := now - int64(retention.FiveMinuteDays)*dayMS
	rawBefore := now - int64(retention.RoundDays)*dayMS
	width := fiveMinuteMS
	if from < fiveMinuteBefore {
		width = hourMS
	}
	points := map[int64]domain.LatencyPoint{}
	// Retained latency is a full UTC bucket average. Raw data replaces every
	// recent bucket, including exact partial-window edges, without a row limit.
	rawFrom := from
	if rawFrom < rawBefore {
		rawFrom = rawBefore
	}
	snapshot, err := s.Store.ReadLatency(ctx, id, from, to, rawFrom, width)
	if err != nil {
		return nil, err
	}
	for _, bucket := range snapshot.Aggregates {
		if bucket.RoundCount == 0 {
			continue
		}
		points[bucket.BucketAt] = domain.LatencyPoint{At: bucket.BucketAt, LatencyMs: float64(bucket.LatencyTotalMS) / float64(bucket.RoundCount), Success: bucket.SuccessfulRoundCount == bucket.RoundCount}
	}
	if rawFrom < to {
		for _, bucket := range snapshot.Rounds {
			if bucket.Count == 0 {
				continue
			}
			if bucket.At < rawFrom && rawFrom > from {
				if _, retained := points[bucket.At]; retained {
					continue
				}
			}
			points[bucket.At] = domain.LatencyPoint{At: bucket.At, LatencyMs: float64(bucket.LatencyTotalMS) / float64(bucket.Count), Success: bucket.Successes == bucket.Count}
		}
	}
	result := make([]domain.LatencyPoint, 0, len(points))
	for _, point := range points {
		result = append(result, point)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].At < result[j].At })
	return result, nil
}

func floor(at, width int64) int64 { return at / width * width }
func decodeMonitor(record store.Monitor) (domain.Monitor, error) {
	var m domain.Monitor
	err := json.Unmarshal(record.ConfigJSON, &m)
	m.ID = record.ID
	m.Type = record.Kind
	return m, err
}

// Start launches bounded aggregation/retention work asynchronously. The API can
// calculate exact interval statistics before historical backfill finishes.
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		return errors.New("statistics service already started")
	}
	child, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		interval := s.Interval
		if interval <= 0 {
			interval = time.Minute
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := s.RunMaintenance(child); err != nil && child.Err() == nil && s.OnError != nil {
				s.OnError(err)
			}
			select {
			case <-child.Done():
				return
			case <-ticker.C:
			case <-s.wake:
			}
		}
	}()
	return nil
}
func (s *Service) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
}
func (s *Service) Wake() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}
