package store

import (
	"context"

	"github.com/octoplorer/octopulse/internal/store/postgresquery"
	"github.com/octoplorer/octopulse/internal/store/sqlitequery"
)

// DeliveryBacklog counts pending work. oldestDueAt is zero when the queue is empty.
func (s *Store) DeliveryBacklog(ctx context.Context) (pending int64, oldestDueAt int64, err error) {
	if s.driver == "sqlite" {
		row, err := sqlitequery.New(s.read).DeliveryBacklog(ctx)
		return row.Pending, row.OldestDueAt, mapError(err)
	}
	row, err := postgresquery.New(s.read).DeliveryBacklog(ctx)
	return row.Pending, row.OldestDueAt, mapError(err)
}
