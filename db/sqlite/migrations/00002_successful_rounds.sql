-- +goose Up
ALTER TABLE aggregates ADD COLUMN successful_round_count INTEGER NOT NULL DEFAULT 0 CHECK(successful_round_count>=0 AND successful_round_count<=round_count);
-- Existing raw rounds permit a trustworthy backfill. If their retained history
-- is unavailable the bucket must be recomputed by the application.
UPDATE aggregates SET successful_round_count=(SELECT CASE WHEN COUNT(*)>aggregates.round_count THEN aggregates.round_count ELSE COUNT(*) END FROM rounds WHERE rounds.monitor_id=aggregates.monitor_id AND rounds.finished_at>=aggregates.bucket_at AND rounds.finished_at<aggregates.bucket_at+aggregates.width_ms AND rounds.success=1);

-- +goose Down
ALTER TABLE aggregates DROP COLUMN successful_round_count;
