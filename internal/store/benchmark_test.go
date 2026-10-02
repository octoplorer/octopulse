package store

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

// BenchmarkRoundBatch100 measures the specified monitor count with concurrent
// history reads and aggregation/cleanup. It measures persistence, not the total
// HTTP probe scheduler or notification network throughput.
func BenchmarkRoundBatch100(b *testing.B) {
	ctx := context.Background()
	s, err := Open(ctx, Config{Driver: "sqlite", DSN: filepath.Join(b.TempDir(), "load.db")})
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	for index := 0; index < 100; index++ {
		id := fmt.Sprintf("monitor-%03d", index)
		if err = s.WithTx(ctx, func(tx *Tx) error {
			if err := tx.PutMonitor(ctx, Monitor{ID: id, ConfigVersion: 1, Generation: 1, Kind: "http", Enabled: true, IntervalMS: 30000}); err != nil {
				return err
			}
			if err := tx.PutRuntime(ctx, Runtime{MonitorID: id, ConfigVersion: 1, Generation: 1, State: "unknown"}); err != nil {
				return err
			}
			return tx.PutInterval(ctx, Interval{ID: id + "/initial", MonitorID: id, State: "unknown", StartedAt: 0})
		}); err != nil {
			b.Fatal(err)
		}
	}
	done := make(chan struct{})
	readStopped := make(chan struct{})
	readError := make(chan error, 1)
	go func() {
		defer close(readStopped)
		for {
			select {
			case <-done:
				return
			default:
			}
			if _, err := s.ListRounds(ctx, "monitor-050", 0, 20); err != nil {
				select {
				case readError <- err:
				default:
				}
				return
			}
			if _, err := s.Intervals(ctx, "monitor-050", 0, 1<<62); err != nil {
				select {
				case readError <- err:
				default:
				}
				return
			}
		}
	}()
	defer func() { close(done); <-readStopped }()
	b.ResetTimer()
	for batch := 0; batch < b.N; batch++ {
		for index := 0; index < 100; index++ {
			id := fmt.Sprintf("monitor-%03d", index)
			roundID := fmt.Sprintf("%s/round/%d", id, batch)
			at := int64(batch+1) * 30000
			r := Round{ID: roundID, MonitorID: id, ConfigVersion: 1, Generation: 1, StartedAt: at, FinishedAt: at + 100, Success: batch%2 == 0, LatencyMS: 100, Attempts: []Attempt{{Number: 1, StartedAt: at, FinishedAt: at + 30, LatencyMS: 30}, {Number: 2, StartedAt: at + 30, FinishedAt: at + 60, LatencyMS: 30}, {Number: 3, StartedAt: at + 60, FinishedAt: at + 100, LatencyMS: 40, Success: batch%2 == 0}}}
			state := "down"
			if r.Success {
				state = "up"
			}
			err = s.WithTx(ctx, func(tx *Tx) error {
				if err := tx.PutRound(ctx, r); err != nil {
					return err
				}
				if err := tx.PutRuntime(ctx, Runtime{MonitorID: id, ConfigVersion: 1, Generation: 1, State: state, LastRoundID: r.ID, LastCollectedAt: r.FinishedAt}); err != nil {
					return err
				}
				if err := tx.ReplaceInterval(ctx, Interval{ID: roundID + "/interval", MonitorID: id, State: state, StartedAt: r.FinishedAt}); err != nil {
					return err
				}
				if err := tx.PutEvent(ctx, Event{ID: roundID + "/event", MonitorID: id, Generation: 1, Kind: state, CreatedAt: r.FinishedAt, Payload: json.RawMessage(`{"message":"probe state"}`)}); err != nil {
					return err
				}
				for channel := 0; channel < 3; channel++ {
					if err := tx.PutDelivery(ctx, Delivery{ID: fmt.Sprintf("%s/delivery/%d", roundID, channel), EventID: roundID + "/event", ChannelID: fmt.Sprint(channel), Generation: 1, DueAt: r.FinishedAt}); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
		if err = s.WithTx(ctx, func(tx *Tx) error {
			if err := tx.PutAggregate(ctx, Aggregate{MonitorID: "monitor-050", BucketAt: 0, WidthMS: 300000, UpMS: int64(batch+1) * 30000, RoundCount: int64(batch + 1)}); err != nil {
				return err
			}
			if err := tx.PutWatermark(ctx, "load", int64(batch+1)*30000); err != nil {
				return err
			}
			return tx.PruneHistory(ctx, 0, 0, 0, 1000)
		}); err != nil {
			b.Fatal(err)
		}
		select {
		case err := <-readError:
			b.Fatal(err)
		default:
		}
	}
	b.ReportMetric(float64(b.N*100)/b.Elapsed().Seconds(), "rounds/s")
}
