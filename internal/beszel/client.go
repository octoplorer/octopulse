package beszel

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type SecretResolver interface {
	ResolveSecret(context.Context, string) (string, error)
}
type Client struct {
	Store      *store.Store
	Secrets    SecretResolver
	HTTP       *http.Client
	Now        func() time.Time
	op         sync.Mutex
	mu         sync.RWMutex
	config     domain.BeszelConfig
	token      string
	tokenUntil int64
	version    string
	systems    SystemsResponse
	history    map[string]HistoryResponse
	containers map[string]ContainersResponse
	wake       chan struct{}
}

func New(s *store.Store, secrets SecretResolver) *Client {
	c := &Client{Store: s, Secrets: secrets, HTTP: &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, Now: time.Now, history: map[string]HistoryResponse{}, containers: map[string]ContainersResponse{}, wake: make(chan struct{}, 1)}
	if s != nil {
		_ = s.Get(context.Background(), "beszelSnapshots", "systems", &c.systems)
		if c.systems.SyncedAt > 0 {
			c.systems.Stale = true
			c.systems.Error = "Awaiting initial Beszel synchronization"
		}
	}
	return c
}
func (c *Client) Invalidate() {
	c.op.Lock()
	c.config = domain.BeszelConfig{}
	c.token = ""
	c.version = ""
	c.mu.Lock()
	c.systems = SystemsResponse{}
	c.history = map[string]HistoryResponse{}
	c.containers = map[string]ContainersResponse{}
	c.mu.Unlock()
	c.op.Unlock()
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

func (c *Client) loadConfig(ctx context.Context) (domain.BeszelConfig, error) {
	var cfg domain.BeszelConfig
	if err := c.Store.Get(ctx, "beszel", "config", &cfg); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return cfg, ErrDisabled
		}
		return cfg, err
	}
	if err := ValidateConfig(&cfg); err != nil {
		return cfg, err
	}
	if !cfg.Enabled {
		return cfg, ErrDisabled
	}
	return cfg, nil
}
func (c *Client) prepare(ctx context.Context, cfg domain.BeszelConfig) error {
	if c.config != cfg {
		c.token = ""
		c.version = ""
		c.config = cfg
	}
	if err := c.authenticate(ctx, cfg); err != nil {
		return err
	}
	{
		var info struct {
			Version string `json:"v"`
		}
		if err := c.request(ctx, cfg, "GET", "/api/beszel/info", nil, nil, &info, true); err != nil {
			return err
		}
		version := info.Version
		if len(version) > 0 && version[0] == 'v' {
			version = version[1:]
		}
		if len(version) < 5 || version[:5] != "0.20." {
			return ErrVersion
		}
		c.version = version
	}
	return c.authenticate(ctx, cfg)
}

func (c *Client) Start(ctx context.Context) {
	for {
		cfg, err := c.loadConfig(ctx)
		interval := 30 * time.Second
		if err == nil {
			interval = time.Duration(cfg.PollSeconds) * time.Second
			_ = c.Poll(ctx)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-c.wake:
			timer.Stop()
		case <-timer.C:
		}
	}
}

type rawSystem struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Host    string          `json:"host"`
	Status  string          `json:"status"`
	Updated json.RawMessage `json:"updated"`
	Info    struct {
		Hostname string  `json:"h"`
		Kernel   string  `json:"k"`
		CPU      float64 `json:"cpu"`
		Memory   float64 `json:"mp"`
		Disk     float64 `json:"dp"`
		CPUModel string  `json:"m"`
		Cores    int     `json:"c"`
		Threads  int     `json:"t"`
		Uptime   int64   `json:"u"`
		Version  string  `json:"v"`
	} `json:"info"`
}

func (c *Client) Poll(ctx context.Context) error {
	c.op.Lock()
	defer c.op.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cfg, err := c.loadConfig(ctx)
	if err != nil {
		c.fail(err)
		return err
	}
	if err = c.prepare(ctx, cfg); err != nil {
		c.fail(err)
		return err
	}
	query := url.Values{"sort": {"name,id"}, "fields": {"id,name,host,status,updated,info"}}
	records, err := listRecords[rawSystem](ctx, c, cfg, "systems", query)
	if err != nil {
		c.fail(err)
		return err
	}
	now := c.Now().UnixMilli()
	result := SystemsResponse{Items: []System{}, Source: cfg.URL, Version: c.version, SyncedAt: now}
	for _, r := range records {
		updated, err := timestamp(r.Updated)
		if err != nil || r.ID == "" {
			c.fail(ErrSchema)
			return ErrSchema
		}
		result.Items = append(result.Items, System{ID: r.ID, Name: r.Name, Host: r.Host, Status: r.Status, CPU: r.Info.CPU, Memory: r.Info.Memory, Disk: r.Info.Disk, UpdatedAt: updated, Stale: r.Status != "up" || now-updated > int64(max(cfg.PollSeconds*2, 120))*1000, Info: SystemInfo{Hostname: r.Info.Hostname, Kernel: r.Info.Kernel, CPUModel: r.Info.CPUModel, Cores: r.Info.Cores, Threads: r.Info.Threads, UptimeSeconds: r.Info.Uptime, AgentVersion: r.Info.Version}})
	}
	if err = c.Store.Put(ctx, "beszelSnapshots", "systems", result); err != nil {
		c.fail(err)
		return err
	}
	c.mu.Lock()
	c.systems = result
	c.mu.Unlock()
	return nil
}
func (c *Client) fail(err error) {
	c.mu.Lock()
	c.systems.Stale = true
	c.systems.Error = err.Error()
	c.mu.Unlock()
}

func (c *Client) Systems(ctx context.Context) (SystemsResponse, error) {
	cfg, err := c.loadConfig(ctx)
	if err != nil {
		return SystemsResponse{Items: []System{}, Stale: true, Error: err.Error()}, nil
	}
	c.mu.RLock()
	result := c.systems
	result.Items = append([]System{}, result.Items...)
	c.mu.RUnlock()
	if result.Source != cfg.URL || result.SyncedAt == 0 {
		_ = c.Poll(ctx)
		c.mu.RLock()
		result = c.systems
		result.Items = append([]System{}, result.Items...)
		c.mu.RUnlock()
	}
	if result.Source != "" && result.Source != cfg.URL {
		result = SystemsResponse{Items: []System{}, Stale: true, Error: result.Error}
	}
	result.Source = cfg.URL
	if result.SyncedAt == 0 || c.Now().UnixMilli()-result.SyncedAt > int64(cfg.PollSeconds*2)*1000 {
		result.Stale = true
	}
	if result.Stale {
		for i := range result.Items {
			result.Items[i].Stale = true
		}
	}
	return result, nil
}

type rawHistory struct {
	Created json.RawMessage `json:"created"`
	Stats   struct {
		CPU             float64   `json:"cpu"`
		Memory          float64   `json:"mp"`
		Disk            float64   `json:"dp"`
		NetworkSent     float64   `json:"ns"`
		NetworkReceived float64   `json:"nr"`
		Bytes           []float64 `json:"b"`
	} `json:"stats"`
}

var ranges = map[string]struct {
	duration time.Duration
	kind     string
	width    int64
}{"1h": {time.Hour, "1m", 60000}, "12h": {12 * time.Hour, "10m", 600000}, "24h": {24 * time.Hour, "20m", 1200000}, "1w": {7 * 24 * time.Hour, "120m", 7200000}, "30d": {30 * 24 * time.Hour, "480m", 28800000}}

func (c *Client) History(ctx context.Context, id, rangeName string) (HistoryResponse, error) {
	window, ok := ranges[rangeName]
	if !ok || !systemID.MatchString(id) {
		return HistoryResponse{}, ErrSchema
	}
	c.op.Lock()
	defer c.op.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cfg, err := c.loadConfig(ctx)
	if err != nil {
		return HistoryResponse{}, err
	}
	key := id + ":" + rangeName
	c.mu.RLock()
	cached := c.history[key]
	c.mu.RUnlock()
	if cached.Source == cfg.URL && cached.SyncedAt > 0 && c.Now().UnixMilli()-cached.SyncedAt < int64(cfg.PollSeconds)*1000 {
		return cached, nil
	}
	result := HistoryResponse{Items: []HistoryPoint{}, Source: cfg.URL, Range: rangeName, IntervalMS: window.width}
	if err = c.prepare(ctx, cfg); err == nil {
		filter := `system="` + id + `" && created > "` + c.Now().UTC().Add(-window.duration).Format("2006-01-02 15:04:05.000Z") + `" && type="` + window.kind + `"`
		records, fetchErr := listRecords[rawHistory](ctx, c, cfg, "system_stats", url.Values{"filter": {filter}, "sort": {"created,id"}, "fields": {"created,stats"}})
		err = fetchErr
		if err == nil {
			for _, r := range records {
				at, e := timestamp(r.Created)
				if e != nil {
					err = e
					break
				}
				incoming, outgoing := r.Stats.NetworkReceived*1_048_576, r.Stats.NetworkSent*1_048_576
				if len(r.Stats.Bytes) == 2 {
					outgoing, incoming = r.Stats.Bytes[0], r.Stats.Bytes[1]
				}
				result.Items = append(result.Items, HistoryPoint{At: at, CPU: r.Stats.CPU, Memory: r.Stats.Memory, Disk: r.Stats.Disk, NetworkIn: incoming, NetworkOut: outgoing})
			}
		}
	}
	if err != nil {
		if cached.Source == cfg.URL {
			cached.Stale = true
			cached.Error = err.Error()
			return cached, nil
		}
		result.Stale = true
		result.Error = err.Error()
		return result, nil
	}
	result.SyncedAt = c.Now().UnixMilli()
	c.mu.Lock()
	if len(c.history) >= 128 {
		var oldest string
		var at int64
		for key, value := range c.history {
			if oldest == "" || value.SyncedAt < at {
				oldest = key
				at = value.SyncedAt
			}
		}
		delete(c.history, oldest)
	}
	c.history[key] = result
	c.mu.Unlock()
	return result, nil
}

type rawContainer struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Image   string          `json:"image"`
	Status  string          `json:"status"`
	CPU     float64         `json:"cpu"`
	Memory  float64         `json:"memory"`
	Updated json.RawMessage `json:"updated"`
}

func (c *Client) Containers(ctx context.Context, id string) (ContainersResponse, error) {
	if !systemID.MatchString(id) {
		return ContainersResponse{}, ErrSchema
	}
	c.op.Lock()
	defer c.op.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cfg, err := c.loadConfig(ctx)
	if err != nil {
		return ContainersResponse{}, err
	}
	c.mu.RLock()
	cached := c.containers[id]
	c.mu.RUnlock()
	if cached.Source == cfg.URL && cached.SyncedAt > 0 && c.Now().UnixMilli()-cached.SyncedAt < int64(cfg.PollSeconds)*1000 {
		return cached, nil
	}
	result := ContainersResponse{Items: []Container{}, Source: cfg.URL}
	if err = c.prepare(ctx, cfg); err == nil {
		records, fetchErr := listRecords[rawContainer](ctx, c, cfg, "containers", url.Values{"filter": {`system="` + id + `"`}, "sort": {"name,id"}, "fields": {"id,name,image,status,cpu,memory,updated"}})
		err = fetchErr
		if err == nil {
			for _, r := range records {
				updated, e := timestamp(r.Updated)
				if e != nil {
					err = e
					break
				}
				result.Items = append(result.Items, Container{ID: r.ID, Name: r.Name, Image: r.Image, Status: r.Status, CPU: r.CPU, Memory: r.Memory, UpdatedAt: updated})
			}
		}
	}
	if err != nil {
		if cached.Source == cfg.URL {
			cached.Stale = true
			cached.Error = err.Error()
			return cached, nil
		}
		result.Stale = true
		result.Error = err.Error()
		return result, nil
	}
	result.SyncedAt = c.Now().UnixMilli()
	c.mu.Lock()
	if len(c.containers) >= 128 {
		var oldest string
		var at int64
		for key, value := range c.containers {
			if oldest == "" || value.SyncedAt < at {
				oldest = key
				at = value.SyncedAt
			}
		}
		delete(c.containers, oldest)
	}
	c.containers[id] = result
	c.mu.Unlock()
	return result, nil
}
