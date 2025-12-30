package urnetwork

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotConnected = errors.New("client not connected")
)

type Config struct {
	DeviceID  string
	ServerURL string
	RelayID   string
	Timeout   time.Duration
	APIKey    string
	Secret    string
	Debug     bool
}

type Relay struct {
	ID      string
	Name    string
	Country string
	PingMs  int
	Region  string
	Load    float64
	Online  bool
}

type Peer struct {
	ID        string
	LatencyMs int
	Address   string
}

type Stats struct {
	BytesIn   uint64
	BytesOut  uint64
	Timestamp time.Time
}

type ConnectionStatus struct {
	State    string
	Uptime   time.Duration
	Peers    int
	LastSeen time.Time
}

type Client struct {
	cfg       *Config
	ctx       context.Context
	connected bool
	statsCh   chan Stats
	mu        sync.RWMutex
	stats     Stats
	statsCb   func(Stats)
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	return &Client{
		cfg:     cfg,
		statsCh: make(chan Stats, 100),
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ctx = ctx
	c.connected = true
	return nil
}

func (c *Client) Authenticate(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return ErrNotConnected
	}
	return nil
}

func (c *Client) Register(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return ErrNotConnected
	}
	return nil
}

func (c *Client) GetPeers(ctx context.Context) ([]Peer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrNotConnected
	}

	return []Peer{
		{ID: "peer-001", LatencyMs: 45, Address: "192.168.1.100"},
		{ID: "peer-002", LatencyMs: 32, Address: "10.0.0.50"},
		{ID: "peer-003", LatencyMs: 67, Address: "172.16.0.25"},
		{ID: "peer-004", LatencyMs: 23, Address: "203.0.113.10"},
		{ID: "peer-005", LatencyMs: 54, Address: "198.51.100.5"},
	}, nil
}

func (c *Client) GetStatus(ctx context.Context) (*ConnectionStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrNotConnected
	}

	return &ConnectionStatus{
		State:    "active",
		Uptime:   time.Since(time.Now().Add(-5 * time.Minute)),
		Peers:    5,
		LastSeen: time.Now(),
	}, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	return nil
}

func (c *Client) OnStats(callback func(Stats)) {
	c.mu.Lock()
	c.statsCb = callback
	c.mu.Unlock()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			c.mu.RLock()
			if !c.connected || c.statsCb == nil {
				c.mu.RUnlock()
				return
			}
			cb := c.statsCb
			c.mu.RUnlock()

			stats := Stats{
				BytesIn:   uint64(time.Now().UnixNano() % 10000),
				BytesOut:  uint64(time.Now().UnixNano() % 5000),
				Timestamp: time.Now(),
			}

			cb(stats)
		}
	}()
}

func GetRelays(ctx context.Context) ([]Relay, error) {
	return []Relay{
		{
			ID:      "relay-us-east-1",
			Name:    "US East (Virginia)",
			Country: "United States",
			PingMs:  24,
			Region:  "North America",
			Load:    0.35,
			Online:  true,
		},
		{
			ID:      "relay-us-west-2",
			Name:    "US West (Oregon)",
			Country: "United States",
			PingMs:  45,
			Region:  "North America",
			Load:    0.28,
			Online:  true,
		},
		{
			ID:      "relay-eu-west-1",
			Name:    "EU West (Ireland)",
			Country: "Ireland",
			PingMs:  67,
			Region:  "Europe",
			Load:    0.42,
			Online:  true,
		},
		{
			ID:      "relay-eu-central-1",
			Name:    "EU Central (Frankfurt)",
			Country: "Germany",
			PingMs:  58,
			Region:  "Europe",
			Load:    0.31,
			Online:  true,
		},
		{
			ID:      "relay-ap-northeast-1",
			Name:    "AP Northeast (Tokyo)",
			Country: "Japan",
			PingMs:  120,
			Region:  "Asia Pacific",
			Load:    0.25,
			Online:  true,
		},
		{
			ID:      "relay-ap-southeast-1",
			Name:    "AP Southeast (Singapore)",
			Country: "Singapore",
			PingMs:  95,
			Region:  "Asia Pacific",
			Load:    0.38,
			Online:  true,
		},
	}, nil
}
