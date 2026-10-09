package statistics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type progress struct {
	FiveMinuteThrough    int64  `json:"fiveMinuteThrough"`
	HourThrough          int64  `json:"hourThrough"`
	FiveMinuteSourceHash string `json:"fiveMinuteSourceHash"`
	HourSourceHash       string `json:"hourSourceHash"`
}
type maintenanceSnapshot struct {
	Items []domain.Maintenance `json:"items"`
}

func (s *Service) RunMaintenance(ctx context.Context) error {
	s.workMu.Lock()
	defer s.workMu.Unlock()
	retention, err := s.retention(ctx)
	if err != nil {
		return err
	}
	monitors, err := s.Store.ListMonitors(ctx)
	if err != nil {
		return err
	}
	raw, err := s.Store.List(ctx, "maintenance")
	if err != nil {
		return err
	}
	current := maintenanceSnapshot{Items: []domain.Maintenance{}}
	for _, data := range raw {
		var item domain.Maintenance
		if err = json.Unmarshal(data, &item); err != nil {
			return err
		}
		current.Items = append(current.Items, item)
	}
	if err = s.invalidateMaintenance(ctx, current, monitors); err != nil {
		return err
	}
	now := s.now()
	watermark, err := s.Store.Watermark(ctx, "collection")
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if watermark > now.UnixMilli() {
		watermark = now.UnixMilli()
	}
	globalThrough := watermark
	limit := s.BatchBuckets
	if limit < 1 || limit > 2880 {
		limit = 288
	}
	for _, record := range monitors {
		m, err := decodeMonitor(record)
		if err != nil {
			return err
		}
		var p progress
		err = s.Store.Get(
			ctx,
			"statisticsMonitor",
			m.ID,
			&p,
		)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
		earliest := m.CreatedAt
		if earliest == 0 {
			earliest, err = s.Store.EarliestStatisticsAt(ctx, m.ID)
			if errors.Is(err, store.ErrNotFound) {
				earliest = watermark
			} else if err != nil {
				return err
			}
		}
		for _, width := range []int64{fiveMinuteMS, hourMS} {
			cutoff := monthCutoff(now, retention.HistoryMonths).UnixMilli()
			if width == fiveMinuteMS {
				cutoff = now.UnixMilli() - int64(retention.FiveMinuteDays)*dayMS
			}
			start := p.HourThrough
			if width == fiveMinuteMS {
				start = p.FiveMinuteThrough
			}
			initial := earliest
			if initial < cutoff {
				initial = cutoff
			}
			initial = floor(initial, width)
			sourceHash := p.HourSourceHash
			if width == fiveMinuteMS {
				sourceHash = p.FiveMinuteSourceHash
			}
			currentHash := maintenanceHash(raw, m.ID)
			if sourceHash != "" && sourceHash != currentHash {
				start = initial
			}
			if start < initial {
				start = initial
			}
			final := floor(watermark, width)
			if start > final {
				start = final
			}
			end := start + int64(limit)*width
			if end > final {
				end = final
			}
			if start < end {
				var committedHash string
				committedHash, err = s.aggregate(
					ctx,
					record,
					start,
					end,
					width,
					retention,
					p,
				)
				if err != nil {
					return err
				}
				if width == fiveMinuteMS {
					p.FiveMinuteSourceHash = committedHash
				} else {
					p.HourSourceHash = committedHash
				}
			}
			if width == fiveMinuteMS {
				p.FiveMinuteThrough = end
			} else {
				p.HourThrough = end
			}
			if end < globalThrough {
				globalThrough = end
			}
		}
		// Even no-source/new monitors advance to their first closed UTC bucket.
		if err = s.Store.WithTx(ctx, func(tx *store.Tx) error {
			if _, err := tx.GetMonitor(ctx, m.ID); err != nil {
				return err
			}
			return tx.Put(
				ctx,
				"statisticsMonitor",
				m.ID,
				p,
			)
		}); err != nil {
			return err
		}
	}
	// All required resolutions and monitors have committed through this point.
	// Pruning is bounded and cannot cross the shared progress watermark.
	return s.Store.WithTx(ctx, func(tx *store.Tx) error {
		if err := tx.PutWatermark(ctx, "statistics", globalThrough); err != nil {
			return err
		}
		if err := tx.PruneHistory(
			ctx,
			now.UnixMilli()-int64(retention.RoundDays)*dayMS,
			now.UnixMilli()-int64(retention.AttemptDays)*dayMS,
			globalThrough,
			1000,
		); err != nil {
			return err
		}
		return tx.PruneSummaries(
			ctx,
			now.UnixMilli()-int64(retention.FiveMinuteDays)*dayMS,
			monthCutoff(now, retention.HistoryMonths).UnixMilli(),
			monthCutoff(now, retention.HistoryMonths).UnixMilli(),
			globalThrough,
			1000,
		)
	})
}

func (s *Service) aggregate(
	ctx context.Context,
	record store.Monitor,
	from, to, width int64,
	retention domain.Retention,
	p progress,
) (string, error) {
	snapshot, err := s.Store.ReadStatistics(
		ctx,
		record.ID,
		from,
		to,
		width,
	)
	if err != nil {
		return "", err
	}
	if snapshot.HasCollectionWatermark && to > snapshot.CollectionThrough {
		return "", errors.New("aggregation crosses reliable collection watermark")
	}
	raw := map[int64]store.RoundBucket{}
	for _, bucket := range snapshot.Rounds {
		raw[bucket.At] = bucket
	}
	existing, err := s.Store.Aggregates(
		ctx,
		record.ID,
		from,
		to,
		width,
	)
	if err != nil {
		return "", err
	}
	retained := map[int64]store.Aggregate{}
	for _, bucket := range existing {
		retained[bucket.BucketAt] = bucket
	}
	rawBefore := s.now().UnixMilli() - int64(retention.RoundDays)*dayMS
	prepared := prepareAvailability(snapshot, maintenanceSpans(snapshot.Maintenance)[record.ID])
	rows := make([]store.Aggregate, 0, (to-from)/width)
	for at := from; at < to; at += width {
		availability := prepared.calculate(at, at+width)
		bucket := raw[at]
		row := store.Aggregate{
			MonitorID:            record.ID,
			BucketAt:             at,
			WidthMS:              width,
			UpMS:                 availability.UpMs,
			DownMS:               availability.DownMs,
			UnknownMS:            availability.UnknownMs,
			ExcludedMS:           availability.ExcludedMs,
			LatencyTotalMS:       bucket.LatencyTotalMS,
			RoundCount:           bucket.Count,
			SuccessfulRoundCount: bucket.Successes,
		}
		if at < rawBefore {
			if old, ok := retained[at]; ok {
				row.LatencyTotalMS = old.LatencyTotalMS
				row.RoundCount = old.RoundCount
				row.SuccessfulRoundCount = old.SuccessfulRoundCount
			}
		}
		rows = append(rows, row)
	}
	if width == fiveMinuteMS {
		p.FiveMinuteThrough = to
		p.FiveMinuteSourceHash = maintenanceHash(snapshot.Maintenance, record.ID)
	} else {
		p.HourThrough = to
		p.HourSourceHash = maintenanceHash(snapshot.Maintenance, record.ID)
	}
	err = s.Store.WithTx(ctx, func(tx *store.Tx) error {
		if _, err := tx.GetMonitor(ctx, record.ID); err != nil {
			return err
		}
		for _, row := range rows {
			if err := tx.PutAggregate(ctx, row); err != nil {
				return err
			}
		}
		return tx.Put(
			ctx,
			"statisticsMonitor",
			record.ID,
			p,
		)
	})
	hash := p.HourSourceHash
	if width == fiveMinuteMS {
		hash = p.FiveMinuteSourceHash
	}
	return hash, err
}

func (s *Service) invalidateMaintenance(
	ctx context.Context,
	current maintenanceSnapshot,
	monitors []store.Monitor,
) error {
	var previous maintenanceSnapshot
	err := s.Store.Get(
		ctx,
		"statisticsMaintenance",
		"snapshot",
		&previous,
	)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}
	if reflect.DeepEqual(previous, current) {
		return nil
	}
	old := map[string]domain.Maintenance{}
	next := map[string]domain.Maintenance{}
	for _, m := range previous.Items {
		old[m.ID] = m
	}
	for _, m := range current.Items {
		next[m.ID] = m
	}
	affected := map[string]int64{}
	mark := func(m domain.Maintenance) {
		for _, id := range m.MonitorIDs {
			start, exists := affected[id]
			if !exists || m.StartsAt < start {
				affected[id] = m.StartsAt
			}
		}
	}
	for id, m := range old {
		if newer, ok := next[id]; !ok || !reflect.DeepEqual(m, newer) {
			mark(m)
			if ok {
				mark(newer)
			}
		}
	}
	for id, m := range next {
		if _, exists := old[id]; !exists {
			mark(m)
		}
	}
	return s.Store.WithTx(ctx, func(tx *store.Tx) error {
		for _, record := range monitors {
			from, exists := affected[record.ID]
			if !exists {
				continue
			}
			if _, err := tx.GetMonitor(ctx, record.ID); err != nil {
				return err
			}
			var p progress
			err := tx.Get(
				ctx,
				"statisticsMonitor",
				record.ID,
				&p,
			)
			if errors.Is(err, store.ErrNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			fiveFrom := floor(from, fiveMinuteMS)
			hourFrom := floor(from, hourMS)
			if fiveFrom < p.FiveMinuteThrough {
				p.FiveMinuteThrough = fiveFrom
			}
			if hourFrom < p.HourThrough {
				p.HourThrough = hourFrom
			}
			data := make([]json.RawMessage, 0, len(current.Items))
			for _, item := range current.Items {
				encoded, _ := json.Marshal(item)
				data = append(data, encoded)
			}
			hash := maintenanceHash(data, record.ID)
			p.FiveMinuteSourceHash = hash
			p.HourSourceHash = hash
			if err = tx.Put(
				ctx,
				"statisticsMonitor",
				record.ID,
				p,
			); err != nil {
				return err
			}
		}
		return tx.Put(
			ctx,
			"statisticsMaintenance",
			"snapshot",
			current,
		)
	})
}

// Per-resolution source hashes catch a maintenance edit/reversal between the
// global change scan and either read snapshot; matching the global snapshot
// alone could otherwise leave a bucket from an intermediate policy forever.
func maintenanceHash(raw []json.RawMessage, id string) string {
	type value struct {
		ID       string
		From, To int64
	}
	items := []value{}
	for _, data := range raw {
		var m domain.Maintenance
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		for _, monitorID := range m.MonitorIDs {
			if monitorID == id {
				items = append(items, value{ID: m.ID, From: m.StartsAt, To: m.EndsAt})
				break
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	encoded, _ := json.Marshal(items)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func (s *Service) Progress(ctx context.Context, id string) (int64, int64, error) {
	var p progress
	err := s.Store.Get(
		ctx,
		"statisticsMonitor",
		id,
		&p,
	)
	return p.FiveMinuteThrough, p.HourThrough, err
}

func monthCutoff(now time.Time, months int) time.Time {
	now = now.UTC()
	first := time.Date(
		now.Year(),
		now.Month()-time.Month(months),
		1,
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
		time.UTC,
	)
	lastDay := time.Date(
		first.Year(),
		first.Month()+1,
		0,
		0,
		0,
		0,
		0,
		time.UTC,
	).Day()
	day := now.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(
		first.Year(),
		first.Month(),
		day,
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
		time.UTC,
	)
}
