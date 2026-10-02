-- +goose Up
CREATE TABLE documents (
    kind TEXT NOT NULL,
    id TEXT NOT NULL,
    payload TEXT NOT NULL,
    PRIMARY KEY(kind,id)
);
CREATE TABLE monitors (
    id TEXT PRIMARY KEY,
    config_version BIGINT NOT NULL CHECK(config_version>0),
    generation BIGINT NOT NULL CHECK(generation>0),
    kind TEXT NOT NULL,
    enabled BIGINT NOT NULL CHECK(enabled IN (0,1)),
    interval_ms BIGINT NOT NULL CHECK(interval_ms>0),
    config_json TEXT NOT NULL
);
CREATE TABLE monitor_runtime (
    monitor_id TEXT PRIMARY KEY REFERENCES monitors(id) ON DELETE CASCADE,
    config_version BIGINT NOT NULL,
    generation BIGINT NOT NULL,
    state TEXT NOT NULL,
    failures BIGINT NOT NULL CHECK(failures>=0),
    successes BIGINT NOT NULL CHECK(successes>=0),
    last_round_id TEXT NOT NULL,
    last_collected_at BIGINT NOT NULL,
    heartbeat_version BIGINT NOT NULL,
    heartbeat_at BIGINT NOT NULL
);
CREATE TABLE rounds (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    config_version BIGINT NOT NULL,
    generation BIGINT NOT NULL,
    started_at BIGINT NOT NULL,
    finished_at BIGINT NOT NULL CHECK(finished_at>=started_at),
    success BIGINT NOT NULL CHECK(success IN (0,1)),
    latency_ms BIGINT NOT NULL CHECK(latency_ms>=0)
);
CREATE INDEX rounds_monitor_time ON rounds(monitor_id,started_at,id);
CREATE INDEX rounds_cleanup ON rounds(finished_at,id);
CREATE TABLE attempts (
    round_id TEXT NOT NULL REFERENCES rounds(id) ON DELETE CASCADE,
    number BIGINT NOT NULL CHECK(number>=1),
    started_at BIGINT NOT NULL,
    finished_at BIGINT NOT NULL CHECK(finished_at>=started_at),
    success BIGINT NOT NULL CHECK(success IN(0,1)),
    latency_ms BIGINT NOT NULL CHECK(latency_ms>=0),
    error TEXT NOT NULL,
    detail TEXT NOT NULL,
    PRIMARY KEY(round_id,number)
);
CREATE INDEX attempts_cleanup ON attempts(finished_at,round_id,number);
CREATE TABLE state_intervals (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    state TEXT NOT NULL,
    started_at BIGINT NOT NULL,
    ended_at BIGINT CHECK(ended_at IS NULL OR ended_at>=started_at)
);
CREATE INDEX intervals_monitor_time ON state_intervals(monitor_id,started_at,id);
CREATE INDEX intervals_cleanup ON state_intervals(ended_at,id);
CREATE UNIQUE INDEX intervals_one_open ON state_intervals(monitor_id) WHERE ended_at IS NULL;
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    generation BIGINT NOT NULL,
    kind TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    payload TEXT NOT NULL
);
CREATE INDEX events_monitor_time ON events(monitor_id,created_at,id);
CREATE TABLE deliveries (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    channel_id TEXT NOT NULL,
    generation BIGINT NOT NULL,
    state TEXT NOT NULL CHECK(state IN('pending','sending','sent','failed','cancelled')),
    due_at BIGINT NOT NULL,
    attempts BIGINT NOT NULL CHECK(attempts>=0),
    lease_token TEXT NOT NULL,
    lease_until BIGINT NOT NULL,
    last_error TEXT NOT NULL,
    payload TEXT NOT NULL,
    UNIQUE(event_id,channel_id)
);
CREATE INDEX deliveries_due ON deliveries(state,due_at,id);
CREATE INDEX deliveries_expired_lease ON deliveries(state,lease_until,id);
CREATE TABLE page_slugs (
    page_id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE
);
CREATE TABLE page_domains (
    domain TEXT PRIMARY KEY,
    page_id TEXT NOT NULL REFERENCES page_slugs(page_id) ON DELETE CASCADE
);
CREATE INDEX page_domains_page ON page_domains(page_id,domain);
CREATE TABLE aggregates (
    monitor_id TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    bucket_at BIGINT NOT NULL,
    width_ms BIGINT NOT NULL CHECK(width_ms>0),
    up_ms BIGINT NOT NULL CHECK(up_ms>=0),
    down_ms BIGINT NOT NULL CHECK(down_ms>=0),
    unknown_ms BIGINT NOT NULL CHECK(unknown_ms>=0),
    excluded_ms BIGINT NOT NULL CHECK(excluded_ms>=0),
    latency_total_ms BIGINT NOT NULL CHECK(latency_total_ms>=0),
    round_count BIGINT NOT NULL CHECK(round_count>=0),
    PRIMARY KEY(monitor_id,bucket_at,width_ms)
);
CREATE INDEX aggregates_cleanup ON aggregates(width_ms,bucket_at,monitor_id);
CREATE TABLE watermarks (
    name TEXT PRIMARY KEY,
    at_ms BIGINT NOT NULL
);

-- +goose Down
DROP TABLE watermarks;
DROP TABLE aggregates;
DROP TABLE page_domains;
DROP TABLE page_slugs;
DROP TABLE deliveries;
DROP TABLE events;
DROP TABLE state_intervals;
DROP TABLE attempts;
DROP TABLE rounds;
DROP TABLE monitor_runtime;
DROP TABLE monitors;
DROP TABLE documents;
