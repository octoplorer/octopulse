-- +goose Up
CREATE INDEX rounds_monitor_finished ON rounds(monitor_id,finished_at);

-- +goose Down
DROP INDEX rounds_monitor_finished;
