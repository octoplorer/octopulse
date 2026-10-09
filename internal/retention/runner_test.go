package retention

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
	"github.com/octoplorer/octopulse/internal/testutil"
)

func runner(t *testing.T) (*Runner, int64) {
	t.Helper()
	s, err := store.Open(context.Background(), testutil.Database(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	r := New(s)
	now := time.Date(
		2026,
		10,
		4,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	r.Now = func() time.Time { return now }
	return r, now.UnixMilli()
}

func put(t *testing.T, r *Runner, kind, id string, value any) {
	t.Helper()
	if err := r.Store.Put(
		context.Background(),
		kind,
		id,
		value,
	); err != nil {
		t.Fatal(err)
	}
}

func exists(t *testing.T, r *Runner, kind, id string, want bool) {
	t.Helper()
	var raw json.RawMessage
	err := r.Store.Get(
		context.Background(),
		kind,
		id,
		&raw,
	)
	if want && err != nil || !want && !errors.Is(err, store.ErrNotFound) {
		t.Fatalf(
			"%s/%s existence = %v, want %v",
			kind,
			id,
			err,
			want,
		)
	}
}

func TestExpiredSessionsAlwaysCleanedAndHistoryOptIn(t *testing.T) {
	r, now := runner(t)
	old := now - (90 * 24 * time.Hour).Milliseconds()
	put(
		t,
		r,
		"sessions",
		"expired",
		domain.Session{ExpiresAt: now - 1},
	)
	put(
		t,
		r,
		"sessions",
		"boundary",
		domain.Session{ExpiresAt: now},
	)
	put(
		t,
		r,
		"sessions",
		"active",
		domain.Session{ExpiresAt: now + 1},
	)
	put(
		t,
		r,
		"sessions",
		"unknown",
		map[string]string{"unexpected": "schema"},
	)
	put(
		t,
		r,
		"audit",
		"old",
		domain.Audit{CreatedAt: old},
	)
	put(
		t,
		r,
		"audit",
		"recent",
		domain.Audit{CreatedAt: now},
	)
	for _, state := range []string{"sent", "failed", "sending"} {
		put(
			t,
			r,
			"channelTests",
			state,
			map[string]any{"status": state, "createdAt": old},
		)
	}
	put(
		t,
		r,
		"channelTests",
		"recent",
		map[string]any{"status": "sent", "createdAt": now},
	)
	put(
		t,
		r,
		"deliveryMarkers",
		"recovery",
		domain.DeliveryMarker{DownSentAt: old},
	)
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	exists(
		t,
		r,
		"sessions",
		"expired",
		false,
	)
	exists(
		t,
		r,
		"sessions",
		"boundary",
		false,
	)
	exists(
		t,
		r,
		"sessions",
		"active",
		true,
	)
	exists(
		t,
		r,
		"sessions",
		"unknown",
		true,
	)
	exists(
		t,
		r,
		"audit",
		"old",
		true,
	)
	exists(
		t,
		r,
		"channelTests",
		"sent",
		true,
	)
	r.OperationHistoryDays = 30
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	exists(
		t,
		r,
		"audit",
		"old",
		false,
	)
	exists(
		t,
		r,
		"audit",
		"recent",
		true,
	)
	exists(
		t,
		r,
		"channelTests",
		"sent",
		false,
	)
	exists(
		t,
		r,
		"channelTests",
		"failed",
		false,
	)
	exists(
		t,
		r,
		"channelTests",
		"sending",
		true,
	)
	exists(
		t,
		r,
		"channelTests",
		"recent",
		true,
	)
	exists(
		t,
		r,
		"deliveryMarkers",
		"recovery",
		true,
	)
}

func TestDocumentCursorEventuallyPassesLiveRowsAndWraps(t *testing.T) {
	r, now := runner(t)
	r.BatchSize = 2
	for i := 0; i < 5; i++ {
		put(
			t,
			r,
			"sessions",
			fmt.Sprintf("a%d", i),
			domain.Session{ExpiresAt: now + 1000},
		)
	}
	for i := 0; i < 5; i++ {
		put(
			t,
			r,
			"sessions",
			fmt.Sprintf("z%d", i),
			domain.Session{ExpiresAt: now - 1},
		)
	}
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	exists(
		t,
		r,
		"sessions",
		"z0",
		true,
	)
	// A newly inserted record behind the cursor must be visited on the next pass.
	put(
		t,
		r,
		"sessions",
		"a00",
		domain.Session{ExpiresAt: now - 1},
	)
	for i := 0; i < 11; i++ {
		if err := r.RunOnce(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 5; i++ {
		exists(
			t,
			r,
			"sessions",
			fmt.Sprintf("a%d", i),
			true,
		)
		exists(
			t,
			r,
			"sessions",
			fmt.Sprintf("z%d", i),
			false,
		)
	}
	exists(
		t,
		r,
		"sessions",
		"a00",
		false,
	)
}

func TestDocumentDeletionPreservesConcurrentUpdates(t *testing.T) {
	r, now := runner(t)
	put(
		t,
		r,
		"sessions",
		"session",
		domain.Session{ExpiresAt: now - 1},
	)
	page, err := r.Store.DocumentPage(
		context.Background(),
		"sessions",
		"",
		1,
	)
	if err != nil || len(page) != 1 {
		t.Fatal(page, err)
	}
	put(
		t,
		r,
		"sessions",
		"session",
		domain.Session{ExpiresAt: now + 1000},
	)
	if err := r.Store.DeleteDocumentsIfUnchanged(context.Background(), "sessions", page); err != nil {
		t.Fatal(err)
	}
	exists(
		t,
		r,
		"sessions",
		"session",
		true,
	)
	page, err = r.Store.DocumentPage(
		context.Background(),
		"sessions",
		"",
		1,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Store.DeleteDocumentsIfUnchanged(context.Background(), "sessions", page); err != nil {
		t.Fatal(err)
	}
	exists(
		t,
		r,
		"sessions",
		"session",
		false,
	)
}

func TestHistoryPrunesOnlyOldTerminalEventsAndDeliveries(t *testing.T) {
	r, now := runner(t)
	r.BatchSize = 2
	old := now - (90 * 24 * time.Hour).Milliseconds()
	ctx := context.Background()
	type fixture struct {
		id, state      string
		eventAt, dueAt int64
	}
	fixtures := []fixture{
		{
			id:      "sent",
			state:   "sent",
			eventAt: old,
			dueAt:   0,
		}, {
			id:      "failed",
			state:   "failed",
			eventAt: old,
			dueAt:   old,
		}, {
			id:      "cancelled",
			state:   "cancelled",
			eventAt: old,
			dueAt:   old,
		},
		{
			id:      "pending",
			state:   "pending",
			eventAt: old,
			dueAt:   old,
		}, {
			id:      "sending",
			state:   "sending",
			eventAt: old,
			dueAt:   old,
		},
		{
			id:      "recentEvent",
			state:   "sent",
			eventAt: now,
			dueAt:   old,
		}, {
			id:      "recentDue",
			state:   "sent",
			eventAt: old,
			dueAt:   now,
		}, {
			id:      "orphan",
			state:   "",
			eventAt: old,
			dueAt:   0,
		},
		{id: "mixed", state: "sent", eventAt: old, dueAt: old},
	}
	if err := r.Store.WithTx(ctx, func(tx *store.Tx) error {
		if err := tx.PutMonitor(
			ctx,
			store.Monitor{
				ID:            "monitor",
				ConfigVersion: 1,
				Generation:    1,
				Kind:          "http",
				Enabled:       true,
				IntervalMS:    30000,
			},
		); err != nil {
			return err
		}
		for _, f := range fixtures {
			if err := tx.PutEvent(
				ctx,
				store.Event{
					ID:         f.id,
					MonitorID:  "monitor",
					Generation: 1,
					Kind:       "down",
					CreatedAt:  f.eventAt,
				},
			); err != nil {
				return err
			}
			if f.state != "" {
				if err := tx.PutDelivery(
					ctx,
					store.Delivery{
						ID:         f.id,
						EventID:    f.id,
						ChannelID:  "channel",
						Generation: 1,
						State:      f.state,
						DueAt:      f.dueAt,
						LeaseUntil: old,
						LeaseToken: "lease",
					},
				); err != nil {
					return err
				}
			}
		}
		return tx.PutDelivery(
			ctx,
			store.Delivery{
				ID:         "mixedPending",
				EventID:    "mixed",
				ChannelID:  "other",
				Generation: 1,
				State:      "pending",
				DueAt:      old,
			},
		)
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Store.GetDelivery(ctx, "sent"); err != nil {
		t.Fatal("default policy pruned history", err)
	}
	if _, err := r.Store.GetEvent(ctx, "orphan"); err != nil {
		t.Fatal("default policy pruned events", err)
	}
	r.OperationHistoryDays = 30
	if err := r.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	remaining := 0
	for _, id := range []string{"sent", "failed", "cancelled"} {
		if _, err := r.Store.GetDelivery(ctx, id); err == nil {
			remaining++
		}
	}
	if remaining != 1 {
		t.Fatal("cleanup exceeded its bounded batch", remaining)
	}
	for i := 0; i < 4; i++ {
		if err := r.RunOnce(ctx); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"sent", "failed", "cancelled", "orphan"} {
		if _, err := r.Store.GetEvent(ctx, id); !errors.Is(err, store.ErrNotFound) {
			t.Fatal("old unreferenced event remains", id, err)
		}
	}
	for _, id := range []string{"pending", "sending", "recentEvent", "recentDue", "mixed", "mixedPending"} {
		if _, err := r.Store.GetDelivery(ctx, id); err != nil {
			t.Fatal("active or recent delivery removed", id, err)
		}
	}
}

func TestMalformedDocumentDoesNotBlockCursor(t *testing.T) {
	r, now := runner(t)
	r.BatchSize = 1
	put(
		t,
		r,
		"sessions",
		"a",
		"invalid-session-schema",
	)
	put(
		t,
		r,
		"sessions",
		"b",
		domain.Session{ExpiresAt: now - 1},
	)
	if err := r.RunOnce(context.Background()); err == nil {
		t.Fatal("expected schema error")
	}
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	exists(
		t,
		r,
		"sessions",
		"a",
		true,
	)
	exists(
		t,
		r,
		"sessions",
		"b",
		false,
	)
}
