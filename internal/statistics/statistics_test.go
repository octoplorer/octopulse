package statistics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type harness struct {
	s       *store.Store
	service *Service
	now     int64
	base    int64
	m       domain.Monitor
}

func backends() []string {
	result := []string{"sqlite"}
	if os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN") != "" {
		result = append(result, "postgres")
	}
	return result
}
func testConfig(t *testing.T, backend string) store.Config {
	t.Helper()
	if backend == "sqlite" {
		return store.Config{Driver: backend, DSN: filepath.Join(t.TempDir(), "statistics.db")}
	}
	dsn := os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN")
	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("statistics_test_%d", time.Now().UnixNano())
	if _, err = pool.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec("DROP SCHEMA " + schema + " CASCADE")
		_ = pool.Close()
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	return store.Config{Driver: backend, DSN: u.String()}
}
func newHarness(t *testing.T, backend string) *harness {
	t.Helper()
	s, err := store.Open(context.Background(), testConfig(t, backend))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	base := time.Date(
		2026,
		10,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	).UnixMilli()
	h := &harness{
		s:    s,
		base: base,
		now:  base + 2*hourMS,
		m: domain.Monitor{
			ID:            "monitor",
			Name:          "test",
			Type:          domain.MonitorHTTP,
			CreatedAt:     base,
			ConfigVersion: 1,
			Enabled:       true,
		},
	}
	h.service = New(s)
	h.service.Now = func() time.Time { return time.UnixMilli(h.now) }
	h.monitor(t, h.m)
	h.watermark(t, h.now)
	return h
}
func (h *harness) monitor(t *testing.T, m domain.Monitor) {
	t.Helper()
	data, _ := json.Marshal(m)
	if err := h.s.UpsertMonitor(
		context.Background(),
		store.Monitor{
			ID:            m.ID,
			ConfigVersion: 1,
			Generation:    1,
			Kind:          m.Type,
			Enabled:       m.Enabled,
			IntervalMS:    30000,
			ConfigJSON:    data,
		},
	); err != nil {
		t.Fatal(err)
	}
	h.m = m
}
func (h *harness) watermark(t *testing.T, at int64) {
	t.Helper()
	if err := h.s.WithTx(
		context.Background(),
		func(tx *store.Tx) error {
			return tx.PutWatermark(context.Background(), "collection", at)
		},
	); err != nil {
		t.Fatal(err)
	}
}
func (h *harness) interval(t *testing.T, state string, from, to int64) {
	t.Helper()
	var end *int64
	if to >= 0 {
		value := h.base + to
		end = &value
	}
	if err := h.s.WithTx(context.Background(), func(tx *store.Tx) error {
		return tx.PutInterval(
			context.Background(),
			store.Interval{
				ID:        domain.ID(),
				MonitorID: h.m.ID,
				State:     state,
				StartedAt: h.base + from,
				EndedAt:   end,
			},
		)
	}); err != nil {
		t.Fatal(err)
	}
}
func (h *harness) maintenance(t *testing.T, id string, from, to int64) {
	t.Helper()
	if err := h.s.Put(
		context.Background(),
		"maintenance",
		id,
		domain.Maintenance{
			ID:         id,
			MonitorIDs: []string{h.m.ID},
			StartsAt:   h.base + from,
			EndsAt:     h.base + to,
		},
	); err != nil {
		t.Fatal(err)
	}
}
func (h *harness) availability(t *testing.T, from, to int64) domain.Availability {
	t.Helper()
	result, err := h.service.Availability(
		context.Background(),
		h.m.ID,
		h.base+from,
		h.base+to,
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func closeTo(t *testing.T, actual *float64, expected float64) {
	t.Helper()
	if actual == nil || math.Abs(*actual-expected) > 1e-7 {
		t.Fatalf("value=%v want=%v", actual, expected)
	}
}

func TestDurationStatisticsAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.m.CreatedAt = h.base + 1000
			h.monitor(t, h.m)
			h.interval(
				t,
				"up",
				1000,
				4000,
			)
			h.interval(
				t,
				"down",
				4000,
				7000,
			)
			h.interval(
				t,
				"unknown",
				7000,
				8000,
			)
			h.interval(
				t,
				"paused",
				8000,
				10000,
			)
			h.maintenance(
				t,
				"first",
				2000,
				5000,
			)
			h.maintenance(
				t,
				"overlap",
				3000,
				6000,
			)
			h.maintenance(
				t,
				"pause-overlap",
				9000,
				12000,
			)
			result := h.availability(t, 0, 12000)
			stateDurationsMatch := result.UpMs == 1000 && result.DownMs == 1000
			otherDurationsMatch := result.UnknownMs == 2000 && result.ExcludedMs == 8000
			durationsMatch := stateDurationsMatch && otherDurationsMatch
			if !durationsMatch || result.EffectiveMs != 2000 {
				t.Fatalf("durations=%+v", result)
			}
			closeTo(t, result.Uptime, 50)
			closeTo(t, result.Coverage, 50)
			partial := h.availability(t, 1500, 6500)
			partialStatesMatch := partial.UpMs == 500 && partial.DownMs == 500
			partialOtherDurationsMatch := partial.ExcludedMs == 4000 && partial.UnknownMs == 0
			if !partialStatesMatch || !partialOtherDurationsMatch {
				t.Fatal(partial)
			}
			closeTo(t, partial.Coverage, 100)
			unknown := h.availability(t, 0, 1000)
			if unknown.Uptime != nil || unknown.UnknownMs != 1000 {
				t.Fatal(unknown)
			}
			closeTo(t, unknown.Coverage, 0)
			paused := h.availability(t, 9000, 10000)
			hasNullRatios := paused.Uptime == nil && paused.Coverage == nil
			isEntirelyPaused := paused.EffectiveMs == 0 && paused.ExcludedMs == 1000
			if !hasNullRatios || !isEntirelyPaused {
				t.Fatal(paused)
			}
			for _, window := range [][2]int64{{0, 12000}, {1000, 2000}, {1900, 6100}, {7500, 10500}} {
				a := h.availability(t, window[0], window[1])
				if a.UpMs+a.DownMs+a.UnknownMs+a.ExcludedMs != window[1]-window[0] {
					t.Fatal("statistics duration conservation failed", a)
				}
			}
		})
	}
}

func TestOpenIntervalWatermarkAndFutureWindowsAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.interval(
				t,
				"up",
				0,
				-1,
			)
			h.watermark(t, h.base+1000)
			result := h.availability(t, 0, 3000)
			if result.UpMs != 1000 || result.UnknownMs != 2000 {
				t.Fatal("open state crossed reliable collection boundary", result)
			}
			closeTo(t, result.Coverage, 100.0/3)
			h.now = h.base + 4000
			future := h.availability(t, 0, 5000)
			if future.To != h.now {
				t.Fatal("future window was not clipped")
			}
			for _, window := range [][2]int64{{5000, 6000}, {3000, 3000}, {4000, 3000}, {-1, 3000}, {0, 411 * dayMS}} {
				if _, err := h.service.Availability(
					context.Background(),
					h.m.ID,
					h.base+window[0],
					h.base+window[1],
				); !errors.Is(err, ErrInvalidWindow) {
					if window[0] == -1 {
						continue
					}
					t.Fatalf("invalid window=%v err=%v", window, err)
				}
			}
		})
	}
}

func TestAggregateBucketsAndRawResultsRemainEquivalentAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.interval(
				t,
				"up",
				0,
				45*60000,
			)
			h.interval(
				t,
				"down",
				45*60000,
				60*60000,
			)
			h.interval(
				t,
				"paused",
				60*60000,
				90*60000,
			)
			h.interval(
				t,
				"unknown",
				90*60000,
				-1,
			)
			h.maintenance(
				t,
				"one",
				30*60000,
				50*60000,
			)
			h.maintenance(
				t,
				"two",
				40*60000,
				55*60000,
			)
			for i, at := range []int64{0, fiveMinuteMS - 1, fiveMinuteMS, hourMS - 1, hourMS} {
				h.round(
					t,
					fmt.Sprintf("boundary-%d", i),
					h.base+at,
					int64((i+1)*10),
					i != 1,
				)
			}
			if err := h.service.RunMaintenance(context.Background()); err != nil {
				t.Fatal(err)
			}
			want := h.availability(t, 0, 2*hourMS)
			for _, width := range []int64{fiveMinuteMS, hourMS} {
				rows, err := h.s.Aggregates(
					context.Background(),
					h.m.ID,
					h.base,
					h.now,
					width,
				)
				if err != nil {
					t.Fatal(err)
				}
				var up, down, unknown, excluded, total, count, successes int64
				for _, row := range rows {
					up += row.UpMS
					down += row.DownMS
					unknown += row.UnknownMS
					excluded += row.ExcludedMS
					total += row.LatencyTotalMS
					count += row.RoundCount
					successes += row.SuccessfulRoundCount
				}
				stateDurationsMatch := up == want.UpMs && down == want.DownMs
				otherDurationsMatch := unknown == want.UnknownMs && excluded == want.ExcludedMs
				availabilityMatches := stateDurationsMatch && otherDurationsMatch
				roundCountsMatch := count == 5 && successes == 4
				latencyMatches := total == 150 && roundCountsMatch
				if !availabilityMatches || !latencyMatches {
					t.Fatalf(
						"aggregate mismatch width=%d rows=%+v wanted=%+v",
						width,
						rows,
						want,
					)
				}
			}
			points, err := h.service.Latency(
				context.Background(),
				h.m.ID,
				h.base,
				h.now,
			)
			if err != nil {
				t.Fatal(err)
			}
			if len(points) != 4 {
				t.Fatalf("latency buckets=%+v", points)
			}
			firstLatencyMatches := points[0].At == h.base && points[0].LatencyMs == 15
			latencyBoundsMatch := firstLatencyMatches && points[1].At == h.base+fiveMinuteMS
			if !latencyBoundsMatch || points[0].Success {
				t.Fatalf("latency buckets=%+v", points)
			}
			before, _, _ := h.service.Progress(context.Background(), h.m.ID)
			if err := h.service.RunMaintenance(context.Background()); err != nil {
				t.Fatal(err)
			}
			after, _, _ := h.service.Progress(context.Background(), h.m.ID)
			if before != after {
				t.Fatal("idempotent aggregation drifted")
			}
		})
	}
}
func (h *harness) round(t *testing.T, id string, at, latency int64, success bool) {
	t.Helper()
	if err := h.s.WithTx(context.Background(), func(tx *store.Tx) error {
		return tx.PutRound(
			context.Background(),
			store.Round{
				ID:            id,
				MonitorID:     h.m.ID,
				ConfigVersion: 1,
				Generation:    1,
				StartedAt:     at,
				FinishedAt:    at,
				Success:       success,
				LatencyMS:     latency,
				Attempts: []store.Attempt{
					{
						Number:     1,
						StartedAt:  at,
						FinishedAt: at,
						Success:    success,
						LatencyMS:  latency,
					},
				},
			},
		)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLargeHistoryHasNoQueryLimitOrAttemptDetailJoin(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			const count = 12001
			err := h.s.WithTx(context.Background(), func(tx *store.Tx) error {
				for i := 0; i < count; i++ {
					if err := tx.PutRound(
						context.Background(),
						store.Round{
							ID:            fmt.Sprintf("bulk-%d", i),
							MonitorID:     h.m.ID,
							ConfigVersion: 1,
							Generation:    1,
							StartedAt:     h.base + int64(i),
							FinishedAt:    h.base + int64(i),
							Success:       i%2 == 0,
							LatencyMS:     int64(i%100 + 1),
						},
					); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			buckets, err := h.s.RoundBuckets(
				context.Background(),
				h.m.ID,
				h.base,
				h.now,
				fiveMinuteMS,
			)
			if err != nil || len(buckets) != 1 {
				t.Fatalf("large history truncated: %+v err=%v", buckets, err)
			}
			if buckets[0].Count != count || buckets[0].Successes != 6001 {
				t.Fatalf("large history truncated: %+v err=%v", buckets, err)
			}
			points, err := h.service.Latency(
				context.Background(),
				h.m.ID,
				h.base,
				h.now,
			)
			if err != nil || len(points) != 1 {
				t.Fatal(points, err)
			}
			if points[0].Success {
				t.Fatal(points, err)
			}
			expected := float64(buckets[0].LatencyTotalMS) / count
			if points[0].LatencyMs != expected {
				t.Fatal("latency average lost observations")
			}
		})
	}
}

func TestMaintenanceEditDeletionBackfillAndRetainedLatencyAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.base = h.now - 20*dayMS
			h.m.CreatedAt = h.base
			h.monitor(t, h.m)
			h.interval(
				t,
				"up",
				0,
				-1,
			)
			h.round(
				t,
				"old",
				h.base+1000,
				120,
				false,
			)
			h.round(
				t,
				"recent",
				h.now-4*dayMS,
				40,
				true,
			)
			h.service.BatchBuckets = 2880
			for i := 0; i < 3; i++ {
				if err := h.service.RunMaintenance(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			if exists, _ := h.s.RoundExists(context.Background(), "old"); exists {
				t.Fatal("old raw round not pruned after both resolutions completed")
			}
			recent, err := h.s.ListRounds(
				context.Background(),
				h.m.ID,
				h.now-5*dayMS,
				10,
			)
			if err != nil || len(recent) != 1 {
				t.Fatal("attempt retention differs from round retention", recent, err)
			}
			if len(recent[0].Attempts) != 0 {
				t.Fatal("attempt retention differs from round retention", recent, err)
			}
			before, err := h.s.Aggregates(
				context.Background(),
				h.m.ID,
				h.base,
				h.base+fiveMinuteMS,
				fiveMinuteMS,
			)
			if err != nil || len(before) != 1 {
				t.Fatal(before, err)
			}
			if before[0].RoundCount != 1 || before[0].SuccessfulRoundCount != 0 {
				t.Fatal(before, err)
			}
			h.maintenance(
				t,
				"old-maintenance",
				0,
				fiveMinuteMS,
			)
			immediate := h.availability(t, 0, fiveMinuteMS)
			if immediate.ExcludedMs != fiveMinuteMS {
				t.Fatal("API waited for aggregation to apply maintenance edit")
			}
			for i := 0; i < 3; i++ {
				if err := h.service.RunMaintenance(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			after, err := h.s.Aggregates(
				context.Background(),
				h.m.ID,
				h.base,
				h.base+fiveMinuteMS,
				fiveMinuteMS,
			)
			if err != nil || len(after) != 1 {
				t.Fatal("backfill erased retained latency after raw pruning", after, err)
			}
			roundCountsMatch := after[0].RoundCount == 1 && after[0].SuccessfulRoundCount == 0
			retainedLatencyMatches := after[0].LatencyTotalMS == 120 && roundCountsMatch
			if after[0].ExcludedMS != fiveMinuteMS || !retainedLatencyMatches {
				t.Fatal("backfill erased retained latency after raw pruning", after, err)
			}
			if err = h.s.Delete(context.Background(), "maintenance", "old-maintenance"); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				if err = h.service.RunMaintenance(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			restored, _ := h.s.Aggregates(
				context.Background(),
				h.m.ID,
				h.base,
				h.base+fiveMinuteMS,
				fiveMinuteMS,
			)
			if len(restored) != 1 {
				t.Fatal("deleted maintenance did not reaggregate", restored)
			}
			durationsRestored := restored[0].ExcludedMS == 0 && restored[0].UpMS == fiveMinuteMS
			if !durationsRestored || restored[0].LatencyTotalMS != 120 {
				t.Fatal("deleted maintenance did not reaggregate", restored)
			}
			points, err := h.service.Latency(
				context.Background(),
				h.m.ID,
				h.base,
				h.base+fiveMinuteMS,
			)
			if err != nil || len(points) != 1 {
				t.Fatal("retained outcome changed meaning after raw pruning", points, err)
			}
			if points[0].LatencyMs != 120 || points[0].Success {
				t.Fatal("retained outcome changed meaning after raw pruning", points, err)
			}
		})
	}
}

func TestConfiguredRetentionAndBoundedProgressProtectUnaggregatedHistory(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.base = h.now - 5*dayMS
			h.m.CreatedAt = h.base
			h.monitor(t, h.m)
			h.interval(
				t,
				"up",
				0,
				-1,
			)
			h.round(
				t,
				"protected",
				h.base+2*dayMS,
				10,
				true,
			)
			settings := domain.DefaultSettings()
			settings.Retention = domain.Retention{RoundDays: 1, AttemptDays: 1, FiveMinuteDays: 10, HistoryMonths: 1}
			if err := h.s.Put(
				context.Background(),
				"settings",
				"organization",
				settings,
			); err != nil {
				t.Fatal(err)
			}
			h.service.BatchBuckets = 1
			if err := h.service.RunMaintenance(context.Background()); err != nil {
				t.Fatal(err)
			}
			if exists, _ := h.s.RoundExists(context.Background(), "protected"); !exists {
				t.Fatal("raw history deleted before both bucket resolutions committed")
			}
			through, err := h.s.Watermark(context.Background(), "statistics")
			if err != nil || through > h.base+fiveMinuteMS {
				t.Fatal("prune watermark crossed bounded progress", through, err)
			}
			h.service.BatchBuckets = 2880
			for i := 0; i < 2; i++ {
				if err = h.service.RunMaintenance(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			if exists, _ := h.s.RoundExists(context.Background(), "protected"); exists {
				t.Fatal("configured round retention not applied after aggregation")
			}
		})
	}
}

func TestCertificateAvailabilityHasNoUptime(t *testing.T) {
	h := newHarness(t, "sqlite")
	h.m.Type = domain.MonitorCertificate
	h.monitor(t, h.m)
	result := h.availability(t, 0, 1000)
	hasNullRatios := result.Uptime == nil && result.Coverage == nil
	hasNoAvailability := result.UpMs == 0 && result.DownMs == 0
	if !hasNullRatios || !hasNoAvailability {
		t.Fatal("certificate risk contributed to availability", result)
	}
}

func TestLargeStateHistoryIncludesEveryIntervalAndExactEdges(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			const count = 12001
			const width int64 = 30000
			h.now = h.base + count*width
			h.watermark(t, h.now)
			if err := h.s.WithTx(context.Background(), func(tx *store.Tx) error {
				for i := 0; i < count; i++ {
					state := domain.StateUp
					if i%2 == 1 {
						state = domain.StateDown
					}
					end := h.base + int64(i+1)*width
					if err := tx.PutInterval(
						context.Background(),
						store.Interval{
							ID:        fmt.Sprintf("state-%d", i),
							MonitorID: h.m.ID,
							State:     state,
							StartedAt: h.base + int64(i)*width,
							EndedAt:   &end,
						},
					); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			result := h.availability(t, 0, count*width)
			stateDurationsMatch := result.UpMs == 6001*width && result.DownMs == 6000*width
			hasCompleteObservation := result.UnknownMs == 0 && result.EffectiveMs == count*width
			if !stateDurationsMatch || !hasCompleteObservation {
				t.Fatal("long state history was truncated", result)
			}
			edge := h.availability(t, width/2, count*width-width/2)
			if edge.UpMs != 6000*width || edge.DownMs != 6000*width {
				t.Fatal("partial edge intersections lost duration", edge)
			}
		})
	}
}

func TestMonthRetentionClampsCalendarEnds(t *testing.T) {
	for _, test := range []struct {
		input, want string
		months      int
	}{
		{
			input:  "2026-03-31",
			want:   "2026-02-28",
			months: 1,
		},
		{
			input:  "2024-03-31",
			want:   "2024-02-29",
			months: 1,
		},
		{
			input:  "2026-10-31",
			want:   "2025-09-30",
			months: 13,
		},
	} {
		input, err := time.Parse("2006-01-02", test.input)
		if err != nil {
			t.Fatal(err)
		}
		if got := monthCutoff(input, test.months).Format("2006-01-02"); got != test.want {
			t.Fatalf("calendar cutoff=%s want=%s", got, test.want)
		}
	}
}

func TestMaintenanceSourceReversalCannotLeaveAnIntermediatePolicyBucket(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.interval(
				t,
				"up",
				0,
				-1,
			)
			if err := h.service.RunMaintenance(context.Background()); err != nil {
				t.Fatal(err)
			}
			var p progress
			if err := h.s.Get(
				context.Background(),
				"statisticsMonitor",
				h.m.ID,
				&p,
			); err != nil {
				t.Fatal(err)
			}
			p.FiveMinuteSourceHash = "intermediate-maintenance-policy"
			if err := h.s.WithTx(context.Background(), func(tx *store.Tx) error {
				if err := tx.PutAggregate(
					context.Background(),
					store.Aggregate{
						MonitorID:  h.m.ID,
						BucketAt:   h.base,
						WidthMS:    fiveMinuteMS,
						ExcludedMS: fiveMinuteMS,
					},
				); err != nil {
					return err
				}
				return tx.Put(
					context.Background(),
					"statisticsMonitor",
					h.m.ID,
					p,
				)
			}); err != nil {
				t.Fatal(err)
			}
			if err := h.service.RunMaintenance(context.Background()); err != nil {
				t.Fatal(err)
			}
			buckets, err := h.s.Aggregates(
				context.Background(),
				h.m.ID,
				h.base,
				h.base+fiveMinuteMS,
				fiveMinuteMS,
			)
			if err != nil || len(buckets) != 1 {
				t.Fatal("global unchanged policy missed per-bucket intermediate source", buckets, err)
			}
			if buckets[0].UpMS != fiveMinuteMS || buckets[0].ExcludedMS != 0 {
				t.Fatal("global unchanged policy missed per-bucket intermediate source", buckets, err)
			}
		})
	}
}

func TestLatencyRetentionBoundaryKeepsWholeBucketAndRecentEdgesUseRaw(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.now += fiveMinuteMS / 2
			h.base = h.now - 14*dayMS - fiveMinuteMS/2
			h.m.CreatedAt = h.base
			h.monitor(t, h.m)
			h.round(
				t,
				"partial-raw",
				h.base+200000,
				300,
				true,
			)
			if err := h.s.WithTx(context.Background(), func(tx *store.Tx) error {
				return tx.PutAggregate(
					context.Background(),
					store.Aggregate{
						MonitorID:            h.m.ID,
						BucketAt:             h.base,
						WidthMS:              fiveMinuteMS,
						LatencyTotalMS:       300,
						RoundCount:           2,
						SuccessfulRoundCount: 2,
					},
				)
			}); err != nil {
				t.Fatal(err)
			}
			retained, err := h.service.Latency(
				context.Background(),
				h.m.ID,
				h.base,
				h.base+fiveMinuteMS,
			)
			if err != nil || len(retained) != 1 {
				t.Fatal("partial raw retention overwrote full historical bucket", retained, err)
			}
			if retained[0].LatencyMs != 150 {
				t.Fatal("partial raw retention overwrote full historical bucket", retained, err)
			}
			recent, err := h.service.Latency(
				context.Background(),
				h.m.ID,
				h.base+180000,
				h.base+fiveMinuteMS,
			)
			if err != nil || len(recent) != 1 {
				t.Fatal("recent window edge ignored raw intersection", recent, err)
			}
			if recent[0].LatencyMs != 300 {
				t.Fatal("recent window edge ignored raw intersection", recent, err)
			}
		})
	}
}
