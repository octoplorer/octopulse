// Package store persists shared application records in SQLite or PostgreSQL.
// It does not interpret probe results, permissions, or notification policy.
package store

import (
	"database/sql"
	"encoding/json"
	"errors"
)

var (
	ErrNotFound  = errors.New("record not found")
	ErrConflict  = errors.New("record conflicts with current data")
	ErrLocked    = errors.New("another octopulse instance holds the database")
	ErrLeaseLost = errors.New("delivery lease has expired or changed")
)

type Config struct {
	Driver         string
	DSN            string
	MaxConnections int
}

type PoolStats struct {
	Write sql.DBStats `json:"write"`
	Read  sql.DBStats `json:"read"`
}

type Monitor struct {
	ID            string          `json:"id"`
	ConfigVersion int64           `json:"configVersion"`
	Generation    int64           `json:"generation"`
	Kind          string          `json:"kind"`
	Enabled       bool            `json:"enabled"`
	IntervalMS    int64           `json:"intervalMs"`
	ConfigJSON    json.RawMessage `json:"config"`
}

type Runtime struct {
	MonitorID        string `json:"monitorId"`
	ConfigVersion    int64  `json:"configVersion"`
	Generation       int64  `json:"generation"`
	State            string `json:"state"`
	Failures         int64  `json:"failures"`
	Successes        int64  `json:"successes"`
	LastRoundID      string `json:"lastRoundId"`
	LastCollectedAt  int64  `json:"lastCollectedAt"`
	HeartbeatVersion int64  `json:"heartbeatVersion"`
	HeartbeatAt      int64  `json:"heartbeatAt"`
}

type Attempt struct {
	Number     int64           `json:"number"`
	StartedAt  int64           `json:"startedAt"`
	FinishedAt int64           `json:"finishedAt"`
	Success    bool            `json:"success"`
	LatencyMS  int64           `json:"latencyMs"`
	Error      string          `json:"error,omitempty"`
	Detail     json.RawMessage `json:"detail,omitempty"`
}

type Round struct {
	ID            string    `json:"id"`
	MonitorID     string    `json:"monitorId"`
	ConfigVersion int64     `json:"configVersion"`
	Generation    int64     `json:"generation"`
	StartedAt     int64     `json:"startedAt"`
	FinishedAt    int64     `json:"finishedAt"`
	Success       bool      `json:"success"`
	LatencyMS     int64     `json:"latencyMs"`
	Attempts      []Attempt `json:"attempts"`
}

type Interval struct {
	ID        string `json:"id"`
	MonitorID string `json:"monitorId"`
	State     string `json:"state"`
	StartedAt int64  `json:"startedAt"`
	EndedAt   *int64 `json:"endedAt,omitempty"`
}

type Event struct {
	ID         string          `json:"id"`
	MonitorID  string          `json:"monitorId"`
	Generation int64           `json:"generation"`
	Kind       string          `json:"kind"`
	CreatedAt  int64           `json:"createdAt"`
	Payload    json.RawMessage `json:"payload"`
}

type Delivery struct {
	ID         string          `json:"id"`
	EventID    string          `json:"eventId"`
	ChannelID  string          `json:"channelId"`
	Generation int64           `json:"generation"`
	State      string          `json:"state"`
	DueAt      int64           `json:"dueAt"`
	Attempts   int64           `json:"attempts"`
	LeaseToken string          `json:"leaseToken,omitempty"`
	LeaseUntil int64           `json:"leaseUntil"`
	LastError  string          `json:"lastError,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

type PageBinding struct {
	PageID  string   `json:"pageId"`
	Slug    string   `json:"slug"`
	Domains []string `json:"domains"`
}

type Aggregate struct {
	MonitorID            string `json:"monitorId"`
	BucketAt             int64  `json:"bucketAt"`
	WidthMS              int64  `json:"widthMs"`
	UpMS                 int64  `json:"upMs"`
	DownMS               int64  `json:"downMs"`
	UnknownMS            int64  `json:"unknownMs"`
	ExcludedMS           int64  `json:"excludedMs"`
	LatencyTotalMS       int64  `json:"latencyTotalMs"`
	RoundCount           int64  `json:"roundCount"`
	SuccessfulRoundCount int64  `json:"successfulRoundCount"`
}
