package store

import (
	"context"
	"encoding/json"
	"fmt"
)

// Document keeps the original serialized payload so maintenance can delete a
// scanned record only if another operation has not updated it in the meantime.
type Document struct {
	ID      string
	Payload json.RawMessage
}

// DocumentPage reads one bounded page using the existing (kind,id) primary key.
// An empty afterID starts a scan; callers resume from the final returned ID.
func (s *Store) DocumentPage(ctx context.Context, kind, afterID string, limit int) ([]Document, error) {
	if limit <= 0 || limit > 1000 {
		return nil, fmt.Errorf("document page limit must be between 1 and 1000")
	}
	rows, err := s.read.QueryContext(
		ctx,
		s.sql(`SELECT id,payload FROM documents WHERE kind=? AND id>? ORDER BY id LIMIT ?`),
		kind,
		afterID,
		limit,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	documents := []Document{}
	for rows.Next() {
		var document Document
		var payload string
		if err := rows.Scan(&document.ID, &payload); err != nil {
			return nil, mapError(err)
		}
		document.Payload = json.RawMessage(payload)
		documents = append(documents, document)
	}
	return documents, mapError(rows.Err())
}

// DeleteDocumentsIfUnchanged leaves records that changed after the scan intact.
// Each call is bounded to the maximum supported maintenance page size.
func (s *Store) DeleteDocumentsIfUnchanged(ctx context.Context, kind string, documents []Document) error {
	if len(documents) > 1000 {
		return fmt.Errorf("document cleanup batch exceeds 1000 records")
	}
	if len(documents) == 0 {
		return nil
	}
	return s.WithTx(ctx, func(tx *Tx) error {
		for _, document := range documents {
			if _, err := tx.tx.ExecContext(
				ctx,
				s.sql(`DELETE FROM documents WHERE kind=? AND id=? AND payload=?`),
				kind,
				document.ID,
				string(document.Payload),
			); err != nil {
				return mapError(err)
			}
		}
		return nil
	})
}

// PruneOperationHistory removes a bounded batch of terminal deliveries whose
// event and due time predate before. An event with active jobs keeps its entire
// delivery history. This policy measures event age, not time since completion.
// Recovery deliveryMarkers are separate documents and are never touched here.
func (s *Store) PruneOperationHistory(ctx context.Context, before int64, batch int) error {
	if before <= 0 || batch <= 0 || batch > 1000 {
		return fmt.Errorf("history cleanup requires a positive cutoff and batch between 1 and 1000")
	}
	if err := s.WithTx(ctx, func(tx *Tx) error {
		_, err := tx.tx.ExecContext(
			ctx,
			s.sql(`DELETE FROM deliveries WHERE id IN (
			SELECT d.id FROM deliveries d JOIN events e ON e.id=d.event_id
			WHERE d.state IN ('sent','failed','cancelled') AND d.due_at<? AND e.created_at<?
			AND NOT EXISTS (SELECT 1 FROM deliveries active WHERE active.event_id=e.id AND active.state IN ('pending','sending'))
			ORDER BY d.due_at,d.id LIMIT ?
		) AND state IN ('sent','failed','cancelled')`),
			before,
			before,
			batch,
		)
		return mapError(err)
	}); err != nil {
		return err
	}
	return s.WithTx(ctx, func(tx *Tx) error {
		query := `SELECT e.id FROM events e WHERE e.created_at<? AND NOT EXISTS (SELECT 1 FROM deliveries d ` +
			`WHERE d.event_id=e.id) ORDER BY e.created_at,e.id LIMIT ?`
		if s.driver == "postgres" {
			query += ` FOR UPDATE OF e SKIP LOCKED`
		}
		rows, err := tx.tx.QueryContext(
			ctx,
			s.sql(query),
			before,
			batch,
		)
		if err != nil {
			return mapError(err)
		}
		ids := make([]string, 0, batch)
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return mapError(err)
			}
			ids = append(ids, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return mapError(err)
		}
		for _, id := range ids {
			// A fresh statement after locking rechecks references committed while
			// the candidate query ran. The lock blocks subsequent FK inserts, so
			// ON DELETE CASCADE cannot remove a newly queued delivery.
			if _, err := tx.tx.ExecContext(
				ctx,
				s.sql(`DELETE FROM events WHERE id=? AND created_at<? AND NOT EXISTS (SELECT 1 FROM deliveries WHERE event_id=?)`),
				id,
				before,
				id,
			); err != nil {
				return mapError(err)
			}
		}
		return nil
	})
}
