package statistics

import (
	"context"
	"errors"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
)

func dailyAvailability(t *testing.T, h *harness, ids []string, from, to int64) map[string][]domain.Availability {
	t.Helper()
	result, err := h.service.DailyAvailabilityBatch(
		context.Background(),
		ids,
		h.base+from,
		h.base+to,
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertDailyConservation(t *testing.T, h *harness, id string, from, to int64, buckets []domain.Availability) {
	t.Helper()
	whole, err := h.service.Availability(
		context.Background(),
		id,
		h.base+from,
		h.base+to,
	)
	if err != nil {
		t.Fatal(err)
	}
	var sum domain.Availability
	cursor := whole.From
	for _, bucket := range buckets {
		startsAtCursor := bucket.From == cursor
		validEnd := bucket.To > bucket.From && bucket.To <= whole.To
		if !startsAtCursor || !validEnd {
			t.Fatalf("noncontiguous daily bucket: cursor=%d bucket=%+v", cursor, bucket)
		}
		if bucket.To < whole.To && bucket.To%dayMS != 0 {
			t.Fatalf("daily boundary is not UTC midnight: %+v", bucket)
		}
		if bucket.UpMs+bucket.DownMs+bucket.UnknownMs+bucket.ExcludedMs != bucket.To-bucket.From {
			t.Fatalf("daily duration conservation failed: %+v", bucket)
		}
		sum.UpMs += bucket.UpMs
		sum.DownMs += bucket.DownMs
		sum.UnknownMs += bucket.UnknownMs
		sum.ExcludedMs += bucket.ExcludedMs
		sum.EffectiveMs += bucket.EffectiveMs
		cursor = bucket.To
	}
	stateTotalsMatch := sum.UpMs == whole.UpMs && sum.DownMs == whole.DownMs
	otherTotalsMatch := sum.UnknownMs == whole.UnknownMs && sum.ExcludedMs == whole.ExcludedMs
	effectiveTotalMatches := sum.EffectiveMs == whole.EffectiveMs
	durationsMatch := stateTotalsMatch && otherTotalsMatch
	completeWindow := cursor == whole.To && effectiveTotalMatches
	if !durationsMatch || !completeWindow {
		t.Fatalf(
			"daily totals=%+v cursor=%d whole=%+v",
			sum,
			cursor,
			whole,
		)
	}
	if whole.Uptime != nil {
		closeTo(t, whole.Uptime, float64(sum.UpMs)/float64(sum.EffectiveMs)*100)
	}
	if whole.Coverage != nil {
		closeTo(t, whole.Coverage, float64(sum.EffectiveMs)/float64(whole.To-whole.From-sum.ExcludedMs)*100)
	}
}

func TestDailyAvailabilitySplitsCrossMidnightDowntimeAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.now = h.base + dayMS + 2*hourMS
			h.watermark(t, h.now)
			h.interval(
				t,
				domain.StateUp,
				0,
				dayMS-hourMS/2,
			)
			h.interval(
				t,
				domain.StateDown,
				dayMS-hourMS/2,
				dayMS+hourMS,
			)
			h.interval(
				t,
				domain.StateUp,
				dayMS+hourMS,
				-1,
			)
			from, to := dayMS-hourMS, dayMS+2*hourMS
			buckets := dailyAvailability(
				t,
				h,
				[]string{h.m.ID},
				from,
				to,
			)[h.m.ID]
			if len(buckets) != 2 {
				t.Fatalf("partial UTC buckets=%+v", buckets)
			}
			firstDayBoundsMatch := buckets[0].From == h.base+from && buckets[0].To == h.base+dayMS
			if !firstDayBoundsMatch || buckets[1].To != h.now {
				t.Fatalf("partial UTC buckets=%+v", buckets)
			}
			firstDayDurationsMatch := buckets[0].UpMs == hourMS/2 && buckets[0].DownMs == hourMS/2
			secondDayDurationsMatch := buckets[1].UpMs == hourMS && buckets[1].DownMs == hourMS
			if !firstDayDurationsMatch || !secondDayDurationsMatch {
				t.Fatalf("cross-midnight downtime=%+v", buckets)
			}
			for _, bucket := range buckets {
				closeTo(t, bucket.Uptime, 50)
				closeTo(t, bucket.Coverage, 100)
			}
			assertDailyConservation(
				t,
				h,
				h.m.ID,
				from,
				to,
				buckets,
			)
		})
	}
}

func TestDailyAvailabilityKeepsCreationUnknownAndWatermarkAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.now = h.base + 2*dayMS + 3*hourMS
			h.m.CreatedAt = h.base + dayMS + hourMS
			h.monitor(t, h.m)
			h.interval(
				t,
				"unknown",
				dayMS+hourMS,
				dayMS+2*hourMS,
			)
			h.interval(
				t,
				domain.StateUp,
				dayMS+2*hourMS,
				-1,
			)
			h.watermark(t, h.base+2*dayMS+hourMS)
			buckets := dailyAvailability(
				t,
				h,
				[]string{h.m.ID},
				0,
				2*dayMS+3*hourMS,
			)[h.m.ID]
			if len(buckets) != 3 {
				t.Fatalf("buckets=%+v", buckets)
			}
			isEntirelyUnknown := buckets[0].UnknownMs == dayMS && buckets[0].ExcludedMs == 0
			if !isEntirelyUnknown || buckets[0].Uptime != nil {
				t.Fatalf("pre-creation day=%+v", buckets[0])
			}
			closeTo(t, buckets[0].Coverage, 0)
			creationDayDurationsMatch := buckets[1].UpMs == 22*hourMS && buckets[1].UnknownMs == 2*hourMS
			watermarkDayDurationsMatch := buckets[2].UpMs == hourMS && buckets[2].UnknownMs == 2*hourMS
			if !creationDayDurationsMatch || !watermarkDayDurationsMatch {
				t.Fatalf("creation, unknown state or collection watermark lost: %+v", buckets)
			}
			closeTo(t, buckets[1].Uptime, 100)
			closeTo(t, buckets[1].Coverage, 100.0*22/24)
			closeTo(t, buckets[2].Uptime, 100)
			closeTo(t, buckets[2].Coverage, 100.0/3)
			assertDailyConservation(
				t,
				h,
				h.m.ID,
				0,
				2*dayMS+3*hourMS,
				buckets,
			)
		})
	}
}

func TestDailyAvailabilityUnionsMaintenanceAndPausedAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.now = h.base + 3*dayMS
			h.watermark(t, h.now)
			h.interval(
				t,
				domain.StateDown,
				0,
				12*hourMS,
			)
			h.interval(
				t,
				"paused",
				12*hourMS,
				dayMS+hourMS,
			)
			h.interval(
				t,
				domain.StateUp,
				dayMS+hourMS,
				2*dayMS,
			)
			h.interval(
				t,
				"paused",
				2*dayMS,
				-1,
			)
			h.maintenance(
				t,
				"overlap-first",
				6*hourMS,
				18*hourMS,
			)
			h.maintenance(
				t,
				"overlap-next",
				dayMS+hourMS/2,
				dayMS+2*hourMS,
			)
			buckets := dailyAvailability(
				t,
				h,
				[]string{h.m.ID},
				0,
				3*dayMS,
			)[h.m.ID]
			if len(buckets) != 3 {
				t.Fatalf("maintenance and paused union=%+v", buckets)
			}
			firstDayDurationsMatch := buckets[0].DownMs == 6*hourMS && buckets[0].ExcludedMs == 18*hourMS
			secondDayDurationsMatch := buckets[1].UpMs == 22*hourMS && buckets[1].ExcludedMs == 2*hourMS
			if !firstDayDurationsMatch || !secondDayDurationsMatch {
				t.Fatalf("maintenance and paused union=%+v", buckets)
			}
			closeTo(t, buckets[0].Uptime, 0)
			closeTo(t, buckets[1].Uptime, 100)
			for _, bucket := range buckets[:2] {
				closeTo(t, bucket.Coverage, 100)
			}
			isEntirelyPaused := buckets[2].ExcludedMs == dayMS && buckets[2].UnknownMs == 0
			hasNullRatios := buckets[2].Uptime == nil && buckets[2].Coverage == nil
			if !isEntirelyPaused || !hasNullRatios {
				t.Fatalf("entirely paused day must have null ratios: %+v", buckets[2])
			}
			assertDailyConservation(
				t,
				h,
				h.m.ID,
				0,
				3*dayMS,
				buckets,
			)
		})
	}
}

func TestDailyAvailabilityBatchIDsAndCertificatesAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.now = h.base + dayMS + hourMS
			availabilityID := h.m.ID
			h.monitor(t, domain.Monitor{ID: "certificate", Type: domain.MonitorCertificate, CreatedAt: h.base})
			h.interval(
				t,
				domain.StateUp,
				0,
				-1,
			)
			h.maintenance(
				t,
				"certificate-policy",
				dayMS-hourMS,
				dayMS+hourMS,
			)
			from, to := dayMS-hourMS, dayMS+hourMS
			results := dailyAvailability(
				t,
				h,
				[]string{availabilityID, "certificate", "missing", availabilityID},
				from,
				to,
			)
			hasExpectedMonitors := len(results) == 2
			hasDailyBuckets := len(results[availabilityID]) == 2 && len(results["certificate"]) == 2
			if !hasExpectedMonitors || !hasDailyBuckets {
				t.Fatalf("deduplicated result=%+v", results)
			}
			for _, bucket := range results[availabilityID] {
				if bucket.UnknownMs != hourMS || bucket.Uptime != nil {
					t.Fatalf("unobserved availability=%+v", bucket)
				}
				closeTo(t, bucket.Coverage, 0)
			}
			for _, bucket := range results["certificate"] {
				isEntirelyUnknown := bucket.UnknownMs == hourMS && bucket.UpMs == 0
				hasNullRatios := bucket.Uptime == nil && bucket.Coverage == nil
				hasNoAvailability := isEntirelyUnknown && bucket.ExcludedMs == 0
				if !hasNoAvailability || !hasNullRatios {
					t.Fatalf("certificate day=%+v", bucket)
				}
			}
			assertDailyConservation(
				t,
				h,
				availabilityID,
				from,
				to,
				results[availabilityID],
			)
			assertDailyConservation(
				t,
				h,
				"certificate",
				from,
				to,
				results["certificate"],
			)
			if result := dailyAvailability(
				t,
				h,
				nil,
				from,
				to,
			); result == nil || len(result) != 0 {
				t.Fatal("empty IDs must return an empty map", result)
			}
			if result := dailyAvailability(
				t,
				h,
				[]string{"missing"},
				from,
				to,
			); len(result) != 0 {
				t.Fatal("missing IDs must be omitted", result)
			}
		})
	}
}

func TestDailyAvailabilityIncludesNinetyDaysAndClipsFutureAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			h := newHarness(t, backend)
			h.now = h.base + 89*dayMS + 2*hourMS
			h.watermark(t, h.now)
			h.interval(
				t,
				domain.StateUp,
				0,
				-1,
			)
			to := 90 * dayMS
			buckets := dailyAvailability(
				t,
				h,
				[]string{h.m.ID},
				0,
				to,
			)[h.m.ID]
			if len(buckets) != 90 {
				t.Fatalf("90 days through clipped today: count=%d", len(buckets))
			}
			if buckets[89].From != h.base+89*dayMS || buckets[89].To != h.now {
				t.Fatalf("90 days through clipped today: count=%d last=%+v", len(buckets), buckets[len(buckets)-1])
			}
			for i, bucket := range buckets {
				want := dayMS
				if i == 89 {
					want = 2 * hourMS
				}
				if bucket.UpMs != want || bucket.UnknownMs != 0 {
					t.Fatalf("day %d=%+v", i, bucket)
				}
				closeTo(t, bucket.Uptime, 100)
				closeTo(t, bucket.Coverage, 100)
			}
			assertDailyConservation(
				t,
				h,
				h.m.ID,
				0,
				to,
				buckets,
			)
			for _, window := range [][2]int64{
				{-1, h.now},
				{h.now, h.now},
				{h.now, h.now + hourMS},
				{h.base, h.base + 411*dayMS},
			} {
				if _, err := h.service.DailyAvailabilityBatch(
					context.Background(),
					[]string{h.m.ID},
					window[0],
					window[1],
				); !errors.Is(err, ErrInvalidWindow) {
					t.Fatalf("invalid window=%v err=%v", window, err)
				}
			}
		})
	}
}
