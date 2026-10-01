// Package beszel provides an independent, read-only projection of one Beszel
// Hub. It never changes monitor states, availability, or notification events.
package beszel

type SystemInfo struct {
	Hostname      string `json:"hostname"`
	Kernel        string `json:"kernel"`
	CPUModel      string `json:"cpuModel"`
	Cores         int    `json:"cores"`
	Threads       int    `json:"threads"`
	UptimeSeconds int64  `json:"uptimeSeconds"`
	AgentVersion  string `json:"agentVersion"`
}
type System struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Host       string     `json:"host"`
	Status     string     `json:"status"`
	CPU        float64    `json:"cpu"`
	Memory     float64    `json:"memory"`
	Disk       float64    `json:"disk"`
	NetworkIn  *float64   `json:"networkIn,omitempty"`
	NetworkOut *float64   `json:"networkOut,omitempty"`
	UpdatedAt  int64      `json:"updatedAt"`
	Stale      bool       `json:"stale"`
	Info       SystemInfo `json:"info"`
}
type HistoryPoint struct {
	At         int64   `json:"at"`
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
	Disk       float64 `json:"disk"`
	NetworkIn  float64 `json:"networkIn"`
	NetworkOut float64 `json:"networkOut"`
}
type Container struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Image     string  `json:"image"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	UpdatedAt int64   `json:"updatedAt"`
}
type SystemsResponse struct {
	Items    []System `json:"items"`
	Source   string   `json:"source"`
	Version  string   `json:"version"`
	SyncedAt int64    `json:"syncedAt"`
	Stale    bool     `json:"stale"`
	Error    string   `json:"error,omitempty"`
}
type HistoryResponse struct {
	Items      []HistoryPoint `json:"items"`
	Source     string         `json:"source"`
	SyncedAt   int64          `json:"syncedAt"`
	Stale      bool           `json:"stale"`
	Error      string         `json:"error,omitempty"`
	Range      string         `json:"range"`
	IntervalMS int64          `json:"intervalMs"`
}
type ContainersResponse struct {
	Items    []Container `json:"items"`
	Source   string      `json:"source"`
	SyncedAt int64       `json:"syncedAt"`
	Stale    bool        `json:"stale"`
	Error    string      `json:"error,omitempty"`
}
