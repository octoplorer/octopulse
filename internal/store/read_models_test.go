package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestBatchReadModelsAcrossDatabases(t *testing.T) {
	backends := []string{"sqlite"}
	if os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN") != "" {
		backends = append(backends, "postgres")
	}
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			s, err := Open(ctx, testConfig(t, backend))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			a := seedMonitor(t, s, "a")
			b := seedMonitor(t, s, "b")
			empty := Monitor{ID: "without-runtime", ConfigVersion: 1, Generation: 1, Kind: "http", IntervalMS: 30000}
			if err := s.UpsertMonitor(ctx, empty); err != nil {
				t.Fatal(err)
			}
			if err := s.WithTx(ctx, func(tx *Tx) error {
				if err := tx.Put(ctx, "engineMonitor", a.ID, map[string]string{"state": "up"}); err != nil {
					return err
				}
				if err := tx.Put(ctx, "maintenance", "shared", map[string]any{"monitorIds": []string{a.ID, b.ID}, "startsAt": 100, "endsAt": 150}); err != nil {
					return err
				}
				if err := tx.PutWatermark(ctx, "collection", 250); err != nil {
					return err
				}
				for i, id := range []string{a.ID, b.ID} {
					if err := tx.ReplaceInterval(ctx, Interval{ID: id + "-up", MonitorID: id, State: "up", StartedAt: 100}); err != nil {
						return err
					}
					for j, at := range []int64{99, 100, 220, 300} {
						if err := tx.PutRound(ctx, Round{ID: fmt.Sprintf("%s-%d", id, j), MonitorID: id, ConfigVersion: 1, Generation: 1, StartedAt: at, FinishedAt: at, Success: j%2 == 0, LatencyMS: int64(i+1) * 10}); err != nil {
							return err
						}
					}
					if err := tx.PutAggregate(ctx, Aggregate{MonitorID: id, BucketAt: 100, WidthMS: 100, LatencyTotalMS: int64(i+1) * 50, RoundCount: 5, SuccessfulRoundCount: 4}); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			all, err := s.ListMonitorSnapshots(ctx)
			if err != nil || len(all) != 3 || all[0].Monitor.ID != "a" || all[1].Monitor.ID != "b" {
				t.Fatal(all, err)
			}
			if all[0].Runtime == nil || all[0].Runtime.State != "unknown" || !json.Valid(all[0].EngineJSON) {
				t.Fatalf("joined optional rows: %+v", all[0])
			}
			if all[1].Runtime == nil || all[1].EngineJSON != nil || all[2].Runtime != nil || all[2].EngineJSON != nil {
				t.Fatalf("missing optional rows: %+v", all)
			}
			ids := []string{a.ID}
			for i := 0; i < 900; i++ {
				ids = append(ids, fmt.Sprintf("missing-%d", i))
			}
			ids = append(ids, b.ID, a.ID, empty.ID)
			selected, err := s.ReadMonitorSnapshots(ctx, ids)
			if err != nil || len(selected) != 3 || selected[b.ID].Runtime == nil {
				t.Fatal(selected, err)
			}
			snapshots, err := s.ReadStatisticsBatch(ctx, ids, 100, 300, 100)
			if err != nil || len(snapshots) != 3 {
				t.Fatal(snapshots, err)
			}
			for i, id := range []string{a.ID, b.ID} {
				got := snapshots[id]
				if len(got.Maintenance) != 1 || !got.HasCollectionWatermark || got.CollectionThrough != 250 || len(got.Intervals) != 1 || got.Intervals[0].StartedAt != 100 {
					t.Fatalf("%s snapshot=%+v", id, got)
				}
				if len(got.Rounds) != 2 || got.Rounds[0].At != 100 || got.Rounds[1].At != 200 || got.Rounds[0].LatencyTotalMS != int64(i+1)*10 || got.Rounds[0].Successes != 0 || got.Rounds[1].Successes != 1 {
					t.Fatalf("%s buckets=%+v", id, got.Rounds)
				}
			}
			latency, err := s.ReadLatencyBatch(ctx, ids, 150, 300, 200, 100)
			if err != nil || len(latency) != 3 {
				t.Fatal(latency, err)
			}
			for i, id := range []string{a.ID, b.ID} {
				got := latency[id]
				if len(got.Aggregates) != 1 || got.Aggregates[0].LatencyTotalMS != int64(i+1)*50 || len(got.Rounds) != 1 || got.Rounds[0].At != 200 {
					t.Fatalf("%s latency=%+v", id, got)
				}
			}
			if got, err := s.ReadMonitorSnapshots(ctx, nil); err != nil || len(got) != 0 {
				t.Fatal(got, err)
			}
			if got, err := s.ReadStatisticsBatch(ctx, nil, 100, 300, 0); err != nil || len(got) != 0 {
				t.Fatal(got, err)
			}
			if got, err := s.ReadLatencyBatch(ctx, nil, 100, 300, 200, 100); err != nil || len(got) != 0 {
				t.Fatal(got, err)
			}
			if _, err := s.ReadStatistics(ctx, "absent", 100, 300, 0); !errors.Is(err, ErrNotFound) {
				t.Fatal(err)
			}
			if _, err := s.ReadLatency(ctx, "absent", 100, 300, 200, 100); !errors.Is(err, ErrNotFound) {
				t.Fatal(err)
			}
			if _, err := s.ReadStatisticsBatch(ctx, ids, 300, 100, 0); err == nil {
				t.Fatal("accepted inverted window")
			}
			if _, err := s.ReadLatencyBatch(ctx, ids, 100, 300, 200, 0); err == nil {
				t.Fatal("accepted zero bucket width")
			}
		})
	}
}

func TestDeliveryBacklogAcrossDatabases(t *testing.T) {
	backends := []string{"sqlite"}
	if os.Getenv("OCTOPULSE_TEST_POSTGRES_DSN") != "" {
		backends = append(backends, "postgres")
	}
	for _, backend := range backends {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			s, err := Open(ctx, testConfig(t, backend))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if count, oldest, err := s.DeliveryBacklog(ctx); err != nil || count != 0 || oldest != 0 {
				t.Fatal(count, oldest, err)
			}
			m := seedMonitor(t, s, "backlog")
			if err := s.WithTx(ctx, func(tx *Tx) error {
				if err := tx.PutEvent(ctx, Event{ID: "event", MonitorID: m.ID, Generation: 1, Kind: "down"}); err != nil {
					return err
				}
				for i, state := range []string{"pending", "sent", "pending", "failed", "cancelled", "sending"} {
					if err := tx.PutDelivery(ctx, Delivery{ID: fmt.Sprintf("d-%d", i), EventID: "event", ChannelID: fmt.Sprintf("c-%d", i), Generation: 1, State: state, DueAt: int64(i) * 100}); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if count, oldest, err := s.DeliveryBacklog(ctx); err != nil || count != 2 || oldest != 0 {
				t.Fatal(count, oldest, err)
			}
		})
	}
}
