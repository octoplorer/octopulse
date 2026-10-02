package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/engine"
	"github.com/octoplorer/octopulse/internal/notify"
	"github.com/octoplorer/octopulse/internal/probe"
	"github.com/octoplorer/octopulse/internal/statistics"
	"github.com/octoplorer/octopulse/internal/store"
)

// This is deliberately an opt-in wall-clock service acceptance test, not a
// benchmark of isolated persistence or a scheduler using a simulated clock.
func TestServiceCapacity100Monitors(t *testing.T) {
	if os.Getenv("OCTOPULSE_RUN_CAPACITY") != "1" {
		t.Skip("set OCTOPULSE_RUN_CAPACITY=1 to run the 65-second service capacity test")
	}
	s, httpServer, admin := testServer(t)
	admin.Timeout = 10 * time.Second
	csrf := bootstrap(t, admin, httpServer.URL)
	var requests, active [100]atomic.Int64
	var overlaps, webhooks atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/monitor/"))
		if err != nil || i < 0 || i >= len(requests) {
			http.NotFound(w, r)
			return
		}
		if active[i].Add(1) != 1 {
			overlaps.Add(1)
		}
		defer active[i].Add(-1)
		n := requests[i].Add(1)
		time.Sleep(20 * time.Millisecond)
		// 70 stable services; 20 require both retries each round; 10 fail
		// two complete rounds and recover during the third round.
		if (i >= 70 && i < 90 && n%3 != 0) || (i >= 90 && n <= 6) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_, _ = io.WriteString(w, "capacity fixture")
	}))
	defer target.Close()
	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		webhooks.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer webhook.Close()
	var secret domain.Secret
	capacityCreate(t, admin, httpServer.URL, csrf, "secrets", map[string]any{
		"name": "capacity webhook", "value": "generic://" + strings.TrimPrefix(webhook.URL, "http://") + "?disabletls=yes",
	}, &secret)
	var channel domain.Channel
	capacityCreate(t, admin, httpServer.URL, csrf, "channels", domain.Channel{
		Name: "capacity webhook", ServiceURLSecretID: secret.ID, Enabled: true,
	}, &channel)
	monitors := make([]domain.Monitor, 100)
	refs := make([]domain.PageMonitor, 100)
	for i := range monitors {
		capacityCreate(t, admin, httpServer.URL, csrf, "monitors", map[string]any{
			"name": fmt.Sprintf("capacity-%03d", i), "type": "http", "intervalSeconds": 30,
			"timeoutSeconds": 5, "retries": 2, "retryDelaySeconds": 1,
			"failureThreshold": 1, "recoveryThreshold": 1, "notificationChannelIds": []string{channel.ID},
			"http": map[string]any{"url": fmt.Sprintf("%s/monitor/%d", target.URL, i), "method": "GET"},
		}, &monitors[i])
		refs[i] = domain.PageMonitor{MonitorID: monitors[i].ID, ShowUptime: true, ShowLatency: true}
	}
	var page domain.Page
	capacityCreate(t, admin, httpServer.URL, csrf, "pages", domain.Page{
		Name: "Capacity status", Slug: "capacity-status", Draft: domain.PageConfig{
			Title: "Capacity status", BrandColor: "#2563eb", ColorScheme: "system", Links: []domain.Link{},
			Groups: []domain.PageGroup{{ID: "services", Name: "Services", Monitors: refs}},
		},
	}, &page)
	if status, body := request(t, admin, "POST", httpServer.URL+"/api/v1/pages/"+page.ID+"/publish", csrf, nil); status != 200 {
		t.Fatalf("publish: %d %s", status, body)
	}
	// Seed retained, already-processed history plus three kinds of source rows:
	// a recent round to aggregate, a 4-day round whose attempts expire, and a
	// 15-day round which expires entirely. This exercises actual housekeeping
	// while the service is handling live probes and public requests.
	seedAt := time.Now().UnixMilli()
	oldRound, attemptsRound := capacitySeedHistory(t, s.Store, monitors, seedAt)
	ctx, cancel := context.WithCancel(context.Background())
	scheduler := engine.New(s.Store, probe.NewRunner(s))
	stats := statistics.New(s.Store)
	worker := notify.New(s.Store, s)
	measure := &capacityMeasurements{}
	scheduler.OnError = measure.backgroundError
	stats.OnError = measure.backgroundError
	s.Check, s.Changed, s.Heartbeat = scheduler.Check, scheduler.NotifyConfigurationChanged, scheduler.Heartbeat
	s.NextCheck = scheduler.NextCheckAt
	s.Stats, s.Latency, s.TestChannel = stats.Availability, stats.Latency, worker.TestChannel
	s.Wake = func() { scheduler.Wake(); stats.Wake() }
	start := time.Now()
	if err := scheduler.Start(ctx); err != nil {
		cancel()
		t.Fatal(err)
	}
	if err := stats.Start(ctx); err != nil {
		cancel()
		scheduler.Stop()
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	workers.Add(1)
	go func() { defer workers.Done(); worker.Start(ctx) }()
	stop := func() {
		cancel()
		scheduler.Stop()
		stats.Stop()
		workers.Wait()
	}
	defer stop()
	public := &http.Client{Timeout: 10 * time.Second}
	for i := 0; i < 6; i++ {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			client, path := public, "/api/public/pages/capacity-status"
			if index >= 4 {
				client, path = admin, "/api/v1/monitors"
			}
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					measure.query(ctx, client, httpServer.URL+path, index < 4)
				}
			}
		}(i)
	}
	measure.sample(s.Store)
	ticker := time.NewTicker(500 * time.Millisecond)
	deadline := time.NewTimer(65 * time.Second)
run:
	for {
		select {
		case <-ticker.C:
			measure.sample(s.Store)
		case <-deadline.C:
			break run
		}
	}
	ticker.Stop()
	stop()
	elapsed := time.Since(start)
	var roundCount, attemptCount int
	var maxLateness, maxRoundMS int64
	for i, m := range monitors {
		rounds, err := s.Store.ListRounds(context.Background(), m.ID, start.UnixMilli(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(rounds) != 3 {
			starts, generations, attempts := make([]int64, len(rounds)), make([]int64, len(rounds)), make([]int, len(rounds))
			for n, round := range rounds {
				starts[n], generations[n], attempts[n] = round.StartedAt-start.UnixMilli(), round.Generation, len(round.Attempts)
			}
			t.Fatalf("monitor %d collected %d rounds, want 3; startOffsetsMs=%v generations=%v attempts=%v actualRequests=%d", i, len(rounds), starts, generations, attempts, requests[i].Load())
		}
		sort.Slice(rounds, func(a, b int) bool { return rounds[a].StartedAt < rounds[b].StartedAt })
		for n, round := range rounds {
			wantAttempts := 1
			if (i >= 70 && i < 90) || (i >= 90 && n < 2) {
				wantAttempts = 3
			}
			if len(round.Attempts) != wantAttempts || round.Success != !(i >= 90 && n < 2) {
				t.Fatalf("monitor %d round %d: attempts=%d success=%v", i, n, len(round.Attempts), round.Success)
			}
			if n > 0 && round.StartedAt < rounds[n-1].FinishedAt {
				t.Fatalf("monitor %d has overlapping rounds", i)
			}
			maxLateness = max(maxLateness, round.StartedAt-rounds[0].StartedAt-int64(n)*30000)
			maxRoundMS = max(maxRoundMS, round.FinishedAt-round.StartedAt)
			attemptCount += len(round.Attempts)
		}
		roundCount += len(rounds)
		if int(requests[i].Load()) != func() int {
			if i < 70 {
				return 3
			}
			if i < 90 {
				return 9
			}
			return 7
		}() {
			t.Fatalf("monitor %d HTTP request count mismatch: %d", i, requests[i].Load())
		}
		intervals, err := s.Store.Intervals(context.Background(), m.ID, start.UnixMilli(), time.Now().UnixMilli())
		if err != nil {
			t.Fatal(err)
		}
		for n := 1; n < len(intervals); n++ {
			if intervals[n-1].EndedAt == nil || *intervals[n-1].EndedAt > intervals[n].StartedAt {
				t.Fatalf("monitor %d has overlapping state intervals", i)
			}
		}
	}
	if overlaps.Load() != 0 || roundCount != 300 || attemptCount != 460 {
		t.Fatalf("overlap=%d rounds=%d attempts=%d", overlaps.Load(), roundCount, attemptCount)
	}
	deliveries, err := s.Store.ListDeliveries(context.Background(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 20 || webhooks.Load() != 20 {
		t.Fatalf("deliveries=%d actual webhook sends=%d, want 20 down/recovery notifications", len(deliveries), webhooks.Load())
	}
	deliveredKinds := map[string]int{}
	for _, delivery := range deliveries {
		if delivery.State != "sent" || delivery.Attempts != 1 {
			t.Fatalf("delivery %s state=%s attempts=%d", delivery.ID, delivery.State, delivery.Attempts)
		}
		var payload domain.NotificationPayload
		if err := json.Unmarshal(delivery.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		deliveredKinds[payload.Kind]++
		if payload.Kind == "up" {
			var marker domain.DeliveryMarker
			if err := s.Store.Get(context.Background(), "deliveryMarkers", domain.DeliveryMarkerID(payload.MonitorID, channel.ID, payload.CycleID), &marker); err != nil {
				t.Fatal(err)
			}
			if marker.DownSentAt <= 0 || marker.RecoverySentAt < marker.DownSentAt {
				t.Fatalf("recovery lacks a completed down delivery from the same cycle: %+v", marker)
			}
		}
	}
	if deliveredKinds["down"] != 10 || deliveredKinds["up"] != 10 {
		t.Fatalf("unexpected delivered notification kinds: %v", deliveredKinds)
	}
	if exists, err := s.Store.RoundExists(context.Background(), oldRound); err != nil || exists {
		t.Fatalf("15-day history retention not applied: exists=%v err=%v", exists, err)
	}
	history, err := s.Store.ListRounds(context.Background(), monitors[0].ID, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	prunedAttempts := false
	for _, round := range history {
		if round.ID == attemptsRound {
			prunedAttempts = len(round.Attempts) == 0
		}
	}
	if !prunedAttempts {
		t.Fatal("4-day attempts were not pruned while retaining their round")
	}
	aggregatedRounds := int64(0)
	for _, m := range monitors {
		rows, err := s.Store.Aggregates(context.Background(), m.ID, (seedAt-10*60000)/300000*300000, time.Now().UnixMilli(), 300000)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			aggregatedRounds += row.RoundCount
		}
	}
	if aggregatedRounds < 100 {
		t.Fatalf("background aggregation lost seeded rounds: %d", aggregatedRounds)
	}
	measure.mu.Lock()
	defer measure.mu.Unlock()
	if len(measure.errors) != 0 {
		t.Fatalf("service errors: %v", measure.errors)
	}
	p95 := capacityPercentile(measure.publicLatency, .95)
	// Race instrumentation radically changes modernc's query cost. Record the
	// latency distribution rather than treating an instrumented run as a
	// production performance SLA; still require enough successful samples.
	if len(measure.publicLatency) < 20 || len(measure.adminLatency) < 20 {
		t.Fatalf("insufficient request samples: public=%d admin=%d public p95=%s", len(measure.publicLatency), len(measure.adminLatency), p95)
	}
	driver := os.Getenv("OCTOPULSE_TEST_DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}
	t.Logf("capacity driver=%s duration=%s monitors=100 interval=30s rounds=%d attempts=%d overlapping=0 max_schedule_lateness_ms=%d max_round_ms=%d public_requests=%d public_p95_ms=%.2f admin_requests=%d admin_p95_ms=%.2f max_goroutines=%d peak_heap_mib=%.2f peak_go_sys_mib=%.2f peak_pending_or_sending=%d notifications_sent=%d aggregated_rounds=%d retention=rounds_and_attempts_verified",
		driver, elapsed.Round(time.Millisecond), roundCount, attemptCount, maxLateness, maxRoundMS,
		len(measure.publicLatency), float64(p95.Microseconds())/1000, len(measure.adminLatency), float64(capacityPercentile(measure.adminLatency, .95).Microseconds())/1000,
		measure.maxGoroutines, float64(measure.peakHeap)/1048576, float64(measure.peakSys)/1048576, measure.peakQueue, webhooks.Load(), aggregatedRounds)
}

func capacityCreate(t *testing.T, client *http.Client, base, csrf, kind string, body, out any) {
	t.Helper()
	status, response := request(t, client, http.MethodPost, base+"/api/v1/"+kind, csrf, body)
	if status != 200 {
		t.Fatalf("create %s: %d %s", kind, status, response)
	}
	if err := json.Unmarshal(response, out); err != nil {
		t.Fatal(err)
	}
}

func capacitySeedHistory(t *testing.T, st *store.Store, monitors []domain.Monitor, now int64) (string, string) {
	t.Helper()
	old, attempts := domain.ID(), domain.ID()
	err := st.WithTx(context.Background(), func(tx *store.Tx) error {
		for i, m := range monitors {
			record, err := tx.GetMonitor(context.Background(), m.ID)
			if err != nil {
				return err
			}
			m.CreatedAt = now - 16*86400000
			record.ConfigJSON, err = json.Marshal(m)
			if err != nil {
				return err
			}
			if err = tx.PutMonitor(context.Background(), record); err != nil {
				return err
			}
			if err = tx.Put(context.Background(), "monitors", m.ID, m); err != nil {
				return err
			}
			if err = tx.Put(context.Background(), "statisticsMonitor", m.ID, map[string]any{
				"fiveMinuteThrough": (now - 10*60000) / 300000 * 300000, "hourThrough": (now - 10*60000) / 3600000 * 3600000,
			}); err != nil {
				return err
			}
			end := now - 5*60000
			if err = tx.PutInterval(context.Background(), store.Interval{ID: domain.ID(), MonitorID: m.ID, State: domain.StateUp, StartedAt: now - 10*60000, EndedAt: &end}); err != nil {
				return err
			}
			at := now - 8*60000
			if err = tx.PutRound(context.Background(), store.Round{ID: domain.ID(), MonitorID: m.ID, ConfigVersion: 1, Generation: 1, StartedAt: at, FinishedAt: at + 20, Success: true, LatencyMS: 20}); err != nil {
				return err
			}
			if i != 0 {
				continue
			}
			for _, item := range []struct {
				id   string
				days int64
			}{{old, 15}, {attempts, 4}} {
				at = now - item.days*86400000
				if err = tx.PutRound(context.Background(), store.Round{ID: item.id, MonitorID: m.ID, ConfigVersion: 1, Generation: 1, StartedAt: at, FinishedAt: at + 20, Success: true, LatencyMS: 20, Attempts: []store.Attempt{{Number: 1, StartedAt: at, FinishedAt: at + 20, Success: true, LatencyMS: 20}}}); err != nil {
					return err
				}
			}
		}
		return tx.PutWatermark(context.Background(), "collection", now)
	})
	if err != nil {
		t.Fatal(err)
	}
	return old, attempts
}

type capacityMeasurements struct {
	mu                          sync.Mutex
	publicLatency, adminLatency []time.Duration
	errors                      []string
	maxGoroutines               int
	peakHeap, peakSys           uint64
	peakQueue                   int
}

func (m *capacityMeasurements) backgroundError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.errors) < 20 {
		m.errors = append(m.errors, err.Error())
	}
}

func (m *capacityMeasurements) query(ctx context.Context, client *http.Client, url string, public bool) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		m.backgroundError(err)
		return
	}
	response, err := client.Do(req)
	if err != nil {
		if ctx.Err() == nil {
			m.backgroundError(err)
		}
		return
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if ctx.Err() != nil {
		return
	}
	if err != nil {
		m.backgroundError(err)
		return
	}
	if response.StatusCode != 200 {
		m.backgroundError(fmt.Errorf("query status=%d", response.StatusCode))
		return
	}
	count := 0
	if public {
		var page domain.PublicPage
		if err = json.Unmarshal(data, &page); err == nil {
			for _, group := range page.Groups {
				count += len(group.Monitors)
			}
		}
	} else {
		var list Items[domain.Monitor]
		if err = json.Unmarshal(data, &list); err == nil {
			count = len(list.Items)
		}
	}
	if err != nil || count != 100 {
		m.backgroundError(fmt.Errorf("invalid query projection: monitors=%d error=%v", count, err))
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if public {
		m.publicLatency = append(m.publicLatency, time.Since(start))
	} else {
		m.adminLatency = append(m.adminLatency, time.Since(start))
	}
}

func (m *capacityMeasurements) sample(st *store.Store) {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	rows, err := st.ListDeliveries(context.Background(), 1000)
	if err != nil {
		m.backgroundError(err)
		return
	}
	queue := 0
	for _, row := range rows {
		if row.State == "pending" || row.State == "sending" {
			queue++
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxGoroutines = max(m.maxGoroutines, runtime.NumGoroutine())
	m.peakHeap, m.peakSys = max(m.peakHeap, memory.HeapAlloc), max(m.peakSys, memory.Sys)
	m.peakQueue = max(m.peakQueue, queue)
}

func capacityPercentile(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	copy := append([]time.Duration(nil), values...)
	sort.Slice(copy, func(i, j int) bool { return copy[i] < copy[j] })
	return copy[int(math.Ceil(float64(len(copy))*percentile))-1]
}
