package statistics

import (
	"context"
	"errors"
	"testing"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

func TestBatchStatisticsAcrossDatabases(t *testing.T) {
	for _, backend := range backends() {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			h := newHarness(t, backend)
			first := h.m.ID
			h.interval(
				t,
				"up",
				0,
				1000,
			)
			h.interval(
				t,
				"down",
				1000,
				4000,
			)
			h.maintenance(
				t,
				"first-policy",
				500,
				1500,
			)
			h.monitor(t, domain.Monitor{ID: "second", Type: domain.MonitorHTTP, CreatedAt: h.base + 500})
			h.interval(
				t,
				"paused",
				500,
				1000,
			)
			h.interval(
				t,
				"down",
				1000,
				-1,
			)
			h.monitor(t, domain.Monitor{ID: "certificate", Type: domain.MonitorCertificate, CreatedAt: h.base})
			h.interval(
				t,
				"up",
				0,
				-1,
			)
			if err := h.s.WithTx(ctx, func(tx *store.Tx) error {
				for i, id := range []string{first, "second", "certificate"} {
					if err := tx.PutRound(
						ctx,
						store.Round{
							ID:            id + "-round",
							MonitorID:     id,
							ConfigVersion: 1,
							Generation:    1,
							StartedAt:     h.base + 100,
							FinishedAt:    h.base + 200,
							LatencyMS:     int64(i+1) * 10,
							Success:       i%2 == 0,
						},
					); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			ids := []string{first, "second", "certificate", "missing", first}
			results, err := h.service.AvailabilityBatch(
				ctx,
				ids,
				h.base,
				h.base+4000,
			)
			if err != nil || len(results) != 3 {
				t.Fatal(results, err)
			}
			a := results[first]
			aStateDurationsMatch := a.UpMs == 500 && a.DownMs == 2500
			aOtherDurationsMatch := a.ExcludedMs == 1000 && a.UnknownMs == 0
			if !aStateDurationsMatch || !aOtherDurationsMatch {
				t.Fatalf("first availability=%+v", a)
			}
			b := results["second"]
			bStateDurationsMatch := b.UpMs == 0 && b.DownMs == 3000
			bOtherDurationsMatch := b.ExcludedMs == 500 && b.UnknownMs == 500
			if !bStateDurationsMatch || !bOtherDurationsMatch {
				t.Fatalf("second availability=%+v", b)
			}
			c := results["certificate"]
			isEntirelyUnknown := c.UnknownMs == 4000 && c.UpMs == 0
			hasNullRatios := c.Uptime == nil && c.Coverage == nil
			if !isEntirelyUnknown || !hasNullRatios {
				t.Fatalf("certificate availability=%+v", c)
			}
			latency, err := h.service.LatencyBatch(
				ctx,
				ids,
				h.base,
				h.base+4000,
			)
			if err != nil || len(latency) != 3 {
				t.Fatal(latency, err)
			}
			for i, id := range []string{first, "second", "certificate"} {
				points := latency[id]
				if len(points) != 1 {
					t.Fatalf("%s points=%+v", id, points)
				}
				latencyMatches := points[0].At == h.base && points[0].LatencyMs == float64((i+1)*10)
				if !latencyMatches || points[0].Success != (i%2 == 0) {
					t.Fatalf("%s points=%+v", id, points)
				}
			}
			if got, err := h.service.AvailabilityBatch(
				ctx,
				nil,
				h.base,
				h.base+4000,
			); err != nil || len(got) != 0 {
				t.Fatal(got, err)
			}
			if got, err := h.service.LatencyBatch(
				ctx,
				nil,
				h.base,
				h.base+4000,
			); err != nil || len(got) != 0 {
				t.Fatal(got, err)
			}
			if _, err := h.service.Availability(
				ctx,
				"missing",
				h.base,
				h.base+4000,
			); !errors.Is(err, store.ErrNotFound) {
				t.Fatal(err)
			}
			if _, err := h.service.Latency(
				ctx,
				"missing",
				h.base,
				h.base+4000,
			); !errors.Is(err, store.ErrNotFound) {
				t.Fatal(err)
			}
			if _, err := h.service.AvailabilityBatch(
				ctx,
				ids,
				h.base,
				h.base,
			); !errors.Is(err, ErrInvalidWindow) {
				t.Fatal(err)
			}
			if _, err := h.service.LatencyBatch(
				ctx,
				ids,
				h.base,
				h.base,
			); !errors.Is(err, ErrInvalidWindow) {
				t.Fatal(err)
			}
		})
	}
}

func TestPreparedAvailabilityFindsOverlappingAndOpenIntervals(t *testing.T) {
	end := int64(1000)
	shortEnd := int64(400)
	snapshot := store.StatisticsSnapshot{
		Monitor: store.Monitor{ID: "monitor"},
		// Deliberately unordered and overlapping: the prefix maximum must keep the
		// earlier long interval when shorter later intervals end before the window.
		Intervals: []store.Interval{
			{State: domain.StateUp, StartedAt: 300, EndedAt: &shortEnd},
			{State: domain.StateUp, StartedAt: 0, EndedAt: &end},
			{State: domain.StateDown, StartedAt: 1000},
		},
		HasCollectionWatermark: true, CollectionThrough: 1300,
	}
	input := prepareAvailability(snapshot, []span{{from: 500, to: 600}, {from: 1100, to: 1200}})
	for _, test := range []struct{ from, to, up, down, excluded, unknown int64 }{
		{from: 450, to: 750, up: 200, down: 0, excluded: 100, unknown: 0},
		{from: 1000, to: 1500, up: 0, down: 200, excluded: 100, unknown: 200},
		{from: 500, to: 600, up: 0, down: 0, excluded: 100, unknown: 0},
		{from: 1500, to: 1600, up: 0, down: 0, excluded: 0, unknown: 100},
	} {
		got := input.calculate(test.from, test.to)
		stateDurationsMatch := got.UpMs == test.up && got.DownMs == test.down
		otherDurationsMatch := got.ExcludedMs == test.excluded && got.UnknownMs == test.unknown
		if !stateDurationsMatch || !otherDurationsMatch {
			t.Fatalf(
				"[%d,%d) availability=%+v",
				test.from,
				test.to,
				got,
			)
		}
	}
	if snapshot.Intervals[0].StartedAt != 300 {
		t.Fatal("preparation mutated caller's interval ordering")
	}
}
