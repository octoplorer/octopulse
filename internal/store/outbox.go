package store

import (
	"context"
	"encoding/json"
	"fmt"
)

func (t *Tx) PutEvent(ctx context.Context, e Event) error {
	payload, err := jsonText(e.Payload)
	if err != nil {
		return err
	}
	_, err = t.tx.ExecContext(
		ctx,
		t.s.sql(
			`INSERT INTO events(id,monitor_id,generation,kind,created_at,payload) VALUES(?,?,?,?,?,?) `+
				`ON CONFLICT(id) DO NOTHING`,
		),
		e.ID,
		e.MonitorID,
		e.Generation,
		e.Kind,
		e.CreatedAt,
		payload,
	)
	return mapError(err)
}

func (s *Store) GetEvent(ctx context.Context, id string) (Event, error) {
	var e Event
	var payload string
	err := s.read.QueryRowContext(
		ctx,
		s.sql(`SELECT id,monitor_id,generation,kind,created_at,payload FROM events WHERE id=?`),
		id,
	).Scan(
		&e.ID,
		&e.MonitorID,
		&e.Generation,
		&e.Kind,
		&e.CreatedAt,
		&payload,
	)
	e.Payload = json.RawMessage(payload)
	return e, mapError(err)
}

func (s *Store) ListEvents(ctx context.Context, monitorID string, since int64, limit int) ([]Event, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	rows, err := s.read.QueryContext(
		ctx,
		s.sql(
			`SELECT id,monitor_id,generation,kind,created_at,payload FROM events WHERE monitor_id=? `+
				`AND created_at>=? ORDER BY created_at DESC,id DESC LIMIT ?`,
		),
		monitorID,
		since,
		limit,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := []Event{}
	for rows.Next() {
		var e Event
		var payload string
		if err = rows.Scan(
			&e.ID,
			&e.MonitorID,
			&e.Generation,
			&e.Kind,
			&e.CreatedAt,
			&payload,
		); err != nil {
			return nil, err
		}
		e.Payload = json.RawMessage(payload)
		result = append(result, e)
	}
	return result, mapError(rows.Err())
}

func (t *Tx) PutDelivery(ctx context.Context, d Delivery) error {
	payload, err := jsonText(d.Payload)
	if err != nil {
		return err
	}
	if d.State == "" {
		d.State = "pending"
	}
	_, err = t.tx.ExecContext(
		ctx,
		t.s.sql(
			`INSERT INTO deliveries(id,event_id,channel_id,generation,state,due_at,attempts,lease_token,`+
				`lease_until,last_error,payload) VALUES(?,?,?,?,?,?,?,?,?,?,?) `+
				`ON CONFLICT(event_id,channel_id) DO NOTHING`,
		),
		d.ID,
		d.EventID,
		d.ChannelID,
		d.Generation,
		d.State,
		d.DueAt,
		d.Attempts,
		d.LeaseToken,
		d.LeaseUntil,
		d.LastError,
		payload,
	)
	return mapError(err)
}

const deliveryColumns = `id,event_id,channel_id,generation,state,due_at,attempts,lease_token,lease_until,last_error,` +
	`payload`

func scanDelivery(row interface{ Scan(...any) error }) (Delivery, error) {
	var d Delivery
	var payload string
	err := row.Scan(
		&d.ID,
		&d.EventID,
		&d.ChannelID,
		&d.Generation,
		&d.State,
		&d.DueAt,
		&d.Attempts,
		&d.LeaseToken,
		&d.LeaseUntil,
		&d.LastError,
		&payload,
	)
	d.Payload = json.RawMessage(payload)
	return d, mapError(err)
}

func (s *Store) GetDelivery(ctx context.Context, id string) (Delivery, error) {
	return scanDelivery(s.read.QueryRowContext(ctx, s.sql(`SELECT `+deliveryColumns+` FROM deliveries WHERE id=?`), id))
}

func (s *Store) ListDeliveries(ctx context.Context, limit int) ([]Delivery, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	rows, err := s.read.QueryContext(
		ctx,
		s.sql(`SELECT `+deliveryColumns+` FROM deliveries ORDER BY due_at DESC,id DESC LIMIT ?`),
		limit,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := []Delivery{}
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, mapError(rows.Err())
}

// InFlightDeliveries returns all active sending jobs for one monitor/channel.
// The worker interprets the cycle payload and lease deadline; this query does
// not truncate the list or infer notification policy from JSON in the database.
func (s *Store) InFlightDeliveries(ctx context.Context, monitorID, channelID string) ([]Delivery, error) {
	rows, err := s.read.QueryContext(
		ctx,
		s.sql(
			`SELECT d.id,d.event_id,d.channel_id,d.generation,d.state,d.due_at,d.attempts,d.lease_token,`+
				`d.lease_until,d.last_error,d.payload FROM deliveries d JOIN events e ON e.id=d.event_id `+
				`WHERE e.monitor_id=? AND d.channel_id=? AND d.state='sending' ORDER BY d.due_at,d.id`,
		),
		monitorID,
		channelID,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := []Delivery{}
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, mapError(rows.Err())
}

// ClaimDeliveries atomically leases pending or abandoned jobs. A unique token
// must be supplied per worker batch; completion checks both token and deadline.
func (s *Store) ClaimDeliveries(ctx context.Context, now, leaseMS int64, limit int, token string) ([]Delivery, error) {
	if token == "" || leaseMS <= 0 {
		return nil, fmt.Errorf("a lease token and positive duration are required")
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	result := []Delivery{}
	err := s.WithTx(ctx, func(t *Tx) error {
		query := `SELECT ` + deliveryColumns +
			` FROM deliveries WHERE (state='pending' AND due_at<=?) ` +
			`OR (state='sending' AND lease_until<=?) ORDER BY due_at,id LIMIT ?`
		if s.driver == "postgres" {
			query += ` FOR UPDATE SKIP LOCKED`
		}
		rows, err := t.tx.QueryContext(
			ctx,
			s.sql(query),
			now,
			now,
			limit,
		)
		if err != nil {
			return mapError(err)
		}
		candidates := make([]Delivery, 0, limit)
		for rows.Next() {
			d, e := scanDelivery(rows)
			if e != nil {
				rows.Close()
				return e
			}
			candidates = append(candidates, d)
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		for _, d := range candidates {
			r, err := t.tx.ExecContext(
				ctx,
				s.sql(
					`UPDATE deliveries SET state='sending',lease_token=?,lease_until=?,attempts=attempts+1 `+
						`WHERE id=? AND ((state='pending' AND due_at<=?) OR (state='sending' AND lease_until<=?))`,
				),
				token,
				now+leaseMS,
				d.ID,
				now,
				now,
			)
			if err != nil {
				return mapError(err)
			}
			n, e := r.RowsAffected()
			if e != nil {
				return e
			}
			if n == 0 {
				continue
			}
			d.State = "sending"
			d.LeaseToken = token
			d.LeaseUntil = now + leaseMS
			d.Attempts++
			result = append(result, d)
		}
		return nil
	})
	return result, err
}

// CompleteDelivery can share a transaction with the notification policy's
// durable delivery marker, so a restart cannot lose the recovery prerequisite.
func (t *Tx) CompleteDelivery(
	ctx context.Context,
	id, token string,
	now int64,
	state string,
	nextDue int64,
	lastError string,
) error {
	if state != "sent" && state != "pending" && state != "failed" && state != "cancelled" {
		return fmt.Errorf("invalid delivery completion state")
	}
	result, err := t.tx.ExecContext(
		ctx,
		t.s.sql(
			`UPDATE deliveries SET state=?,due_at=?,last_error=?,lease_token='',lease_until=0 WHERE id=? `+
				`AND state='sending' AND lease_token=? AND lease_until>?`,
		),
		state,
		nextDue,
		lastError,
		id,
		token,
		now,
	)
	if err != nil {
		return mapError(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrLeaseLost
	}
	return nil
}

func (s *Store) CompleteDelivery(
	ctx context.Context,
	id, token string,
	now int64,
	state string,
	nextDue int64,
	lastError string,
) error {
	return s.WithTx(ctx, func(t *Tx) error {
		return t.CompleteDelivery(
			ctx,
			id,
			token,
			now,
			state,
			nextDue,
			lastError,
		)
	})
}

func (t *Tx) CancelDeliveries(ctx context.Context, eventID string) error {
	_, err := t.tx.ExecContext(
		ctx,
		t.s.sql(
			`UPDATE deliveries SET state='cancelled',lease_token='',lease_until=0 WHERE event_id=? `+
				`AND state IN('pending','sending')`,
		),
		eventID,
	)
	return mapError(err)
}
