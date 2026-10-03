-- name: ListMonitorSnapshots :many
SELECT m.id,m.config_version,m.generation,m.kind,m.enabled,m.interval_ms,m.config_json,
       r.monitor_id AS runtime_monitor_id,r.config_version AS runtime_config_version,r.generation AS runtime_generation,
       r.state,r.failures,r.successes,r.last_round_id,r.last_collected_at,r.heartbeat_version,r.heartbeat_at,
       d.payload AS engine_json
FROM monitors m
LEFT JOIN monitor_runtime r ON r.monitor_id=m.id
LEFT JOIN documents d ON d.kind='engineMonitor' AND d.id=m.id
ORDER BY m.id;

-- name: ReadMonitorSnapshots :many
SELECT m.id,m.config_version,m.generation,m.kind,m.enabled,m.interval_ms,m.config_json,
       r.monitor_id AS runtime_monitor_id,r.config_version AS runtime_config_version,r.generation AS runtime_generation,
       r.state,r.failures,r.successes,r.last_round_id,r.last_collected_at,r.heartbeat_version,r.heartbeat_at,
       d.payload AS engine_json
FROM monitors m
LEFT JOIN monitor_runtime r ON r.monitor_id=m.id
LEFT JOIN documents d ON d.kind='engineMonitor' AND d.id=m.id
WHERE m.id IN (SELECT value FROM jsonb_array_elements_text(CAST(sqlc.arg(ids_json) AS jsonb)) AS ids(value))
ORDER BY m.id;

-- name: ReadMonitors :many
SELECT * FROM monitors WHERE id IN (SELECT value FROM jsonb_array_elements_text(CAST(sqlc.arg(ids_json) AS jsonb)) AS ids(value)) ORDER BY id;

-- name: ReadStatisticsIntervals :many
SELECT id,monitor_id,state,started_at,ended_at FROM state_intervals
WHERE started_at<sqlc.arg(to_ms) AND (ended_at IS NULL OR ended_at>sqlc.arg(from_ms)) AND monitor_id IN (SELECT value FROM jsonb_array_elements_text(CAST(sqlc.arg(ids_json) AS jsonb)) AS ids(value))
ORDER BY monitor_id,started_at,id;

-- name: ReadLatencyAggregates :many
SELECT * FROM aggregates
WHERE bucket_at>=sqlc.arg(from_ms) AND bucket_at<sqlc.arg(to_ms) AND width_ms=sqlc.arg(width_ms) AND monitor_id IN (SELECT value FROM jsonb_array_elements_text(CAST(sqlc.arg(ids_json) AS jsonb)) AS ids(value))
ORDER BY monitor_id,bucket_at;

-- name: ReadRoundBuckets :many
SELECT monitor_id,
       CAST((finished_at / CAST(sqlc.arg(width_ms) AS BIGINT)) * CAST(sqlc.arg(width_ms) AS BIGINT) AS BIGINT) AS bucket_at,
       CAST(SUM(latency_ms) AS BIGINT) AS latency_total_ms,
       COUNT(*) AS round_count,
       CAST(SUM(CASE WHEN success=1 THEN 1 ELSE 0 END) AS BIGINT) AS successful_round_count
FROM rounds
WHERE finished_at>=sqlc.arg(from_ms) AND finished_at<sqlc.arg(to_ms) AND monitor_id IN (SELECT value FROM jsonb_array_elements_text(CAST(sqlc.arg(ids_json) AS jsonb)) AS ids(value))
GROUP BY monitor_id,bucket_at
ORDER BY monitor_id,bucket_at;

-- name: DeliveryBacklog :one
SELECT COUNT(*) AS pending,CAST(COALESCE(MIN(due_at),0) AS BIGINT) AS oldest_due_at FROM deliveries WHERE state='pending';
