package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func ID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func Now() int64 { return time.Now().UTC().UnixMilli() }

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Role      Role   `json:"role" enum:"admin,operator,viewer"`
	Locale    string `json:"locale" enum:"zh-CN,en"`
	Timezone  string `json:"timezone"`
	Enabled   bool   `json:"enabled"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}
type UserRecord struct {
	User
	PasswordHash string `json:"passwordHash"`
}
type Session struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	CSRFToken string `json:"csrfToken"`
	ExpiresAt int64  `json:"expiresAt"`
	CreatedAt int64  `json:"createdAt"`
}
type Secret struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}
type SecretRecord struct {
	Secret
	Ciphertext string `json:"ciphertext"`
}
type Channel struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ServiceURLSecretID string `json:"serviceUrlSecretId"`
	Enabled            bool   `json:"enabled"`
	CreatedAt          int64  `json:"createdAt"`
	UpdatedAt          int64  `json:"updatedAt"`
}
type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}
type PageMonitor struct {
	MonitorID   string `json:"monitorId"`
	Alias       string `json:"alias"`
	ShowUptime  bool   `json:"showUptime"`
	ShowLatency bool   `json:"showLatency"`
}
type PageGroup struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Monitors []PageMonitor `json:"monitors"`
}
type PageConfig struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	LogoURL     string      `json:"logoUrl"`
	BrandColor  string      `json:"brandColor"`
	ColorScheme string      `json:"colorScheme" enum:"system,light,dark"`
	Links       []Link      `json:"links"`
	Groups      []PageGroup `json:"groups"`
}
type Page struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Domain      string      `json:"domain"`
	Draft       PageConfig  `json:"draft"`
	Published   *PageConfig `json:"published,omitempty"`
	PublishedAt int64       `json:"publishedAt"`
	Version     int64       `json:"version"`
	CreatedAt   int64       `json:"createdAt"`
	UpdatedAt   int64       `json:"updatedAt"`
}
type Maintenance struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	MonitorIDs  []string `json:"monitorIds"`
	PageIDs     []string `json:"pageIds"`
	StartsAt    int64    `json:"startsAt"`
	EndsAt      int64    `json:"endsAt"`
	Timezone    string   `json:"timezone"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
}
type IncidentUpdate struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	Status    string `json:"status" enum:"investigating,identified,monitoring,resolved"`
	CreatedAt int64  `json:"createdAt"`
}
type Incident struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Body       string           `json:"body"`
	Status     string           `json:"status" enum:"investigating,identified,monitoring,resolved"`
	Impact     string           `json:"impact" enum:"none,partial,outage"`
	PageIDs    []string         `json:"pageIds"`
	MonitorIDs []string         `json:"monitorIds"`
	Updates    []IncidentUpdate `json:"updates"`
	CreatedAt  int64            `json:"createdAt"`
	UpdatedAt  int64            `json:"updatedAt"`
	ResolvedAt int64            `json:"resolvedAt"`
}
type Retention struct {
	RoundDays      int `json:"roundDays" minimum:"1"`
	AttemptDays    int `json:"attemptDays" minimum:"1"`
	FiveMinuteDays int `json:"fiveMinuteDays" minimum:"1"`
	HistoryMonths  int `json:"historyMonths" minimum:"1"`
}
type Settings struct {
	OrganizationName string    `json:"organizationName"`
	Timezone         string    `json:"timezone"`
	Locale           string    `json:"locale" enum:"zh-CN,en"`
	Retention        Retention `json:"retention"`
	AllowedDomains   []string  `json:"allowedDomains"`
}

func DefaultSettings() Settings {
	return Settings{OrganizationName: "Octopulse", Timezone: "UTC", Locale: "zh-CN", Retention: Retention{RoundDays: 14, AttemptDays: 3, FiveMinuteDays: 90, HistoryMonths: 13}, AllowedDomains: []string{}}
}

type Audit struct {
	ID           string `json:"id"`
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	Action       string `json:"action"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	CreatedAt    int64  `json:"createdAt"`
}
type BeszelConfig struct {
	URL              string `json:"url"`
	Email            string `json:"email"`
	PasswordSecretID string `json:"passwordSecretId"`
	Enabled          bool   `json:"enabled"`
	PollSeconds      int    `json:"pollSeconds" minimum:"30"`
}
type Availability struct {
	From        int64    `json:"from"`
	To          int64    `json:"to"`
	UpMs        int64    `json:"upMs"`
	DownMs      int64    `json:"downMs"`
	UnknownMs   int64    `json:"unknownMs"`
	ExcludedMs  int64    `json:"excludedMs"`
	EffectiveMs int64    `json:"effectiveMs"`
	Uptime      *float64 `json:"uptime"`
	Coverage    *float64 `json:"coverage"`
}
type LatencyPoint struct {
	At        int64   `json:"at"`
	LatencyMs float64 `json:"latencyMs"`
	Success   bool    `json:"success"`
}
type PublicMonitor struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	State        string             `json:"state"`
	Paused       bool               `json:"paused"`
	Maintenance  bool               `json:"maintenance"`
	Availability Availability       `json:"availability"`
	Latency      []LatencyPoint     `json:"latency"`
	Certificate  *PublicCertificate `json:"certificate,omitempty"`
}
type PublicCertificate struct {
	State         string  `json:"state"`
	ExpiresAt     int64   `json:"expiresAt"`
	DaysRemaining float64 `json:"daysRemaining"`
}
type PublicGroup struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Monitors []PublicMonitor `json:"monitors"`
}
type PublicIncident struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Body       string           `json:"body"`
	Status     string           `json:"status"`
	Impact     string           `json:"impact"`
	Updates    []IncidentUpdate `json:"updates"`
	CreatedAt  int64            `json:"createdAt"`
	ResolvedAt int64            `json:"resolvedAt"`
}
type PublicMaintenance struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartsAt    int64  `json:"startsAt"`
	EndsAt      int64  `json:"endsAt"`
}
type PublicPage struct {
	ID          string              `json:"id"`
	Slug        string              `json:"slug"`
	Config      PageConfig          `json:"config"`
	State       string              `json:"state"`
	Groups      []PublicGroup       `json:"groups"`
	Incidents   []PublicIncident    `json:"incidents"`
	Maintenance []PublicMaintenance `json:"maintenance"`
	UpdatedAt   int64               `json:"updatedAt"`
}
