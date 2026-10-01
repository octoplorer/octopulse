-- name: GetDocument :one
SELECT payload FROM documents WHERE kind = sqlc.arg(kind) AND id = sqlc.arg(id);

-- name: ListDocuments :many
SELECT payload FROM documents WHERE kind = sqlc.arg(kind) ORDER BY id;

-- name: PutDocument :exec
INSERT INTO documents(kind,id,payload) VALUES(sqlc.arg(kind),sqlc.arg(id),sqlc.arg(payload))
ON CONFLICT(kind,id) DO UPDATE SET payload=excluded.payload;

-- name: DeleteDocument :execrows
DELETE FROM documents WHERE kind = sqlc.arg(kind) AND id = sqlc.arg(id);

-- name: GetMonitor :one
SELECT * FROM monitors WHERE id = sqlc.arg(id);

-- name: ListMonitors :many
SELECT * FROM monitors ORDER BY id;

-- name: PutMonitor :exec
INSERT INTO monitors(id,config_version,generation,kind,enabled,interval_ms,config_json)
VALUES(sqlc.arg(id),sqlc.arg(config_version),sqlc.arg(generation),sqlc.arg(kind),sqlc.arg(enabled),sqlc.arg(interval_ms),sqlc.arg(config_json))
ON CONFLICT(id) DO UPDATE SET config_version=excluded.config_version,generation=excluded.generation,kind=excluded.kind,enabled=excluded.enabled,interval_ms=excluded.interval_ms,config_json=excluded.config_json;

-- name: DeleteMonitor :execrows
DELETE FROM monitors WHERE id = sqlc.arg(id);

-- name: GetRuntime :one
SELECT * FROM monitor_runtime WHERE monitor_id = sqlc.arg(monitor_id);
