package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"URnetworkPC/pkg/urnetwork"
)

type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "Disconnected"
	case StatusConnecting:
		return "Connecting..."
	case StatusConnected:
		return "Connected"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

type URnetworkConfig struct {
	DeviceID  string
	ServerURL string
	RelayID   string
	Timeout   time.Duration
}

func DefaultConfig() *URnetworkConfig {
	return &URnetworkConfig{
		DeviceID:  generateDeviceID(),
		ServerURL: "wss://connect.urnetwork.io",
		Timeout:   30 * time.Second,
	}
}

type URnetworkClient struct {
	config      *URnetworkConfig
	ctx         context.Context
	cancel      context.CancelFunc
	client      *urnetwork.Client
	status      ConnectionStatus
	statusMutex sync.RWMutex
	logCallback func(string)
	connected   bool
	connMutex   sync.RWMutex
	bytesIn     uint64
	bytesOut    uint64
	statsMutex  sync.RWMutex
}

func NewURnetworkClient(config *URnetworkConfig) *URnetworkClient {
	if config == nil {
		config = DefaultConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &URnetworkClient{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		status: StatusDisconnected,
	}
}

func (c *URnetworkClient) SetLogCallback(callback func(string)) {
	c.logCallback = callback
}

func (c *URnetworkClient) log(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMsg := fmt.Sprintf("[%s] %s", timestamp, message)
	log.Println(logMsg)
	if c.logCallback != nil {
		c.logCallback(logMsg)
	}
}

func (c *URnetworkClient) GetStatus() ConnectionStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.status
}

func (c *URnetworkClient) setStatus(status ConnectionStatus) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = status
}

func (c *URnetworkClient) IsConnected() bool {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.connected
}

func (c *URnetworkClient) setConnected(connected bool) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	c.connected = connected
}

func (c *URnetworkClient) GetStats() (bytesIn, bytesOut uint64) {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()
	return c.bytesIn, c.bytesOut
}

func (c *URnetworkClient) Connect() error {
	c.setStatus(StatusConnecting)
	c.log("Initializing URnetwork client...")
	c.log(fmt.Sprintf("Device ID: %s", c.config.DeviceID))
	c.log(fmt.Sprintf("Server: %s", c.config.ServerURL))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log(fmt.Sprintf("Recovered from panic: %v", r))
				c.setStatus(StatusError)
				c.setConnected(false)
			}
		}()

		if err := c.connectInternal(); err != nil {
			c.log(fmt.Sprintf("Connection failed: %v", err))
			c.setStatus(StatusError)
			c.setConnected(false)
			return
		}

		c.setStatus(StatusConnected)
		c.setConnected(true)
		c.log("Successfully connected to URnetwork!")
		c.log("Ready to route traffic through decentralized network")

		c.monitorConnection()
	}()

	return nil
}

func (c *URnetworkClient) connectInternal() error {
	c.log("Establishing connection to URnetwork...")
	c.log(fmt.Sprintf("Device ID: %s", c.config.DeviceID))

	cfg := &urnetwork.Config{
		DeviceID:  c.config.DeviceID,
		ServerURL: c.config.ServerURL,
		Timeout:   c.config.Timeout,
	}

	if c.config.RelayID != "" {
		cfg.RelayID = c.config.RelayID
		c.log(fmt.Sprintf("Using specific relay: %s", c.config.RelayID))
	} else {
		relays, err := urnetwork.GetRelays(c.ctx)
		if err != nil {
			return fmt.Errorf("failed to get available relays: %w", err)
		}

		if len(relays) == 0 {
			return fmt.Errorf("no relays available")
		}

		c.log(fmt.Sprintf("Found %d available relays", len(relays)))
		for i, relay := range relays {
			c.log(fmt.Sprintf("  [%d] %s (%s) - Ping: %dms",
				i+1, relay.Name, relay.Country, relay.PingMs))
		}

		bestRelay := relays[0]
		cfg.RelayID = bestRelay.ID
		c.log(fmt.Sprintf("Auto-selected best relay: %s (%s)",
			bestRelay.Name, bestRelay.Country))
	}

	c.log("Initializing URnetwork SDK client...")
	client, err := urnetwork.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create URnetwork client: %w", err)
	}

	c.client = client

	c.log("Performing handshake with relay...")
	if err := c.client.Connect(c.ctx); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.log("Authenticating device...")
	if err := c.client.Authenticate(c.ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.log("Registering with network coordinator...")
	if err := c.client.Register(c.ctx); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	peers, err := c.client.GetPeers(c.ctx)
	if err != nil {
		c.log(fmt.Sprintf("Warning: could not retrieve peer list: %v", err))
	} else {
		c.log(fmt.Sprintf("Connected to %d peer nodes", len(peers)))
		for i, peer := range peers {
			if i < 5 {
				c.log(fmt.Sprintf("  - Peer %d: %s (latency: %dms)",
					i+1, peer.ID, peer.LatencyMs))
			}
		}
		if len(peers) > 5 {
			c.log(fmt.Sprintf("  ... and %d more peers", len(peers)-5))
		}
	}

	c.client.OnStats(func(stats urnetwork.Stats) {
		c.statsMutex.Lock()
		c.bytesIn += stats.BytesIn
		c.bytesOut += stats.BytesOut
		c.statsMutex.Unlock()
	})

	c.log("Connection established successfully")
	return nil
}

func (c *URnetworkClient) monitorConnection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	heartbeatCount := 0

	for {
		select {
		case <-c.ctx.Done():
			c.log("Connection monitoring stopped")
			return
		case <-ticker.C:
			if c.IsConnected() && c.GetStatus() == StatusConnected {
				heartbeatCount++

				bytesIn, bytesOut := c.GetStats()

				if c.client != nil {
					if status, err := c.client.GetStatus(c.ctx); err == nil {
						c.log(fmt.Sprintf("Heartbeat #%d - Network active (↓ %s, ↑ %s) - Status: %s",
							heartbeatCount,
							formatBytes(bytesIn),
							formatBytes(bytesOut),
							status.State))
					}

					if heartbeatCount%3 == 0 {
						if peers, err := c.client.GetPeers(c.ctx); err == nil {
							c.log(fmt.Sprintf("Active connections: %d peers", len(peers)))
						}
					}
				} else {
					c.log(fmt.Sprintf("Heartbeat #%d - Network active (↓ %s, ↑ %s)",
						heartbeatCount,
						formatBytes(bytesIn),
						formatBytes(bytesOut)))
				}
			}
		}
	}
}

func (c *URnetworkClient) GetRelays() ([]urnetwork.Relay, error) {
	return urnetwork.GetRelays(c.ctx)
}

func (c *URnetworkClient) SetRelay(relayID string) {
	c.config.RelayID = relayID
	c.log(fmt.Sprintf("Relay selection changed to: %s", relayID))
}

func (c *URnetworkClient) Disconnect() error {
	c.log("Disconnecting from URnetwork...")

	c.setConnected(false)

	if c.client != nil {
		c.log("Closing peer connections...")

		if err := c.client.Disconnect(c.ctx); err != nil {
			c.log(fmt.Sprintf("Warning: disconnect error: %v", err))
		}

		c.log("Deregistering from network...")
		c.client = nil
	}

	c.cancel()

	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel

	c.statsMutex.Lock()
	c.bytesIn = 0
	c.bytesOut = 0
	c.statsMutex.Unlock()

	c.setStatus(StatusDisconnected)
	c.log("Disconnected successfully")

	return nil
}

func generateDeviceID() string {
	return fmt.Sprintf("urnetwork-win-%d-%04x", time.Now().Unix(), rand.Intn(65536))
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

type URnetworkGUI struct {
	app             fyne.App
	window          fyne.Window
	client          *URnetworkClient
	statusLabel     *widget.Label
	statsLabel      *widget.Label
	connectButton   *widget.Button
	relaySelect     *widget.Select
	logText         *widget.Entry
	logMutex        sync.Mutex
	availableRelays []urnetwork.Relay
}

func NewURnetworkGUI() *URnetworkGUI {
	myApp := app.New()
	myWindow := myApp.NewWindow("URnetwork Client for Windows")

	config := DefaultConfig()
	client := NewURnetworkClient(config)

	gui := &URnetworkGUI{
		app:    myApp,
		window: myWindow,
		client: client,
	}

	gui.setupUI()
	return gui
}

func (g *URnetworkGUI) setupUI() {
	g.statusLabel = widget.NewLabel("Status: Disconnected")
	g.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	g.statsLabel = widget.NewLabel("Network Stats: N/A")

	g.connectButton = widget.NewButton("Connect to URnetwork", g.onConnectClick)
	g.connectButton.Importance = widget.HighImportance

	clearLogsButton := widget.NewButton("Clear Logs", g.onClearLogsClick)

	refreshRelaysButton := widget.NewButton("Refresh Relays", g.onRefreshRelaysClick)

	g.relaySelect = widget.NewSelect([]string{"Auto (Best Relay)"}, g.onRelaySelected)
	g.relaySelect.SetSelected("Auto (Best Relay)")

	g.logText = widget.NewMultiLineEntry()
	g.logText.SetPlaceHolder("Connection logs will appear here...")
	g.logText.Wrapping = fyne.TextWrapWord
	g.logText.Disable()

	g.client.SetLogCallback(func(message string) {
		g.appendLog(message)
	})

	logScroll := container.NewVScroll(g.logText)
	logScroll.SetMinSize(fyne.NewSize(650, 320))

	title := widget.NewLabel("URnetwork Client for Windows")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("Decentralized Network Connection Manager")
	subtitle.Alignment = fyne.TextAlignCenter

	infoLabel := widget.NewLabel("💡 Click 'Connect' to join the URnetwork decentralized network")
	infoLabel.Wrapping = fyne.TextWrapWord

	relayLabel := widget.NewLabel("Relay Server:")
	relayLabel.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		g.statusLabel,
		g.statsLabel,
		container.NewVBox(
			relayLabel,
			container.NewBorder(nil, nil, nil, refreshRelaysButton, g.relaySelect),
		),
		container.NewHBox(
			g.connectButton,
			clearLogsButton,
		),
		widget.NewSeparator(),
		infoLabel,
		widget.NewLabel("Connection Log:"),
		logScroll,
	)

	g.window.SetContent(content)
	g.window.Resize(fyne.NewSize(750, 650))
	g.window.CenterOnScreen()

	g.window.SetOnClosed(func() {
		if g.client.GetStatus() == StatusConnected {
			g.client.Disconnect()
		}
	})

	go g.updateStatusLoop()
	go g.loadAvailableRelays()
}

func (g *URnetworkGUI) appendLog(message string) {
	g.logMutex.Lock()
	defer g.logMutex.Unlock()

	currentText := g.logText.Text
	if currentText != "" {
		currentText += "\n"
	}
	currentText += message
	g.logText.SetText(currentText)

	g.logText.CursorRow = len(g.logText.Text)
}

func (g *URnetworkGUI) onConnectClick() {
	status := g.client.GetStatus()

	if status == StatusDisconnected || status == StatusError {
		g.connectButton.Disable()
		if err := g.client.Connect(); err != nil {
			g.appendLog(fmt.Sprintf("Error: %v", err))
			g.connectButton.Enable()
		}
	} else if status == StatusConnected {
		g.connectButton.Disable()
		if err := g.client.Disconnect(); err != nil {
			g.appendLog(fmt.Sprintf("Error: %v", err))
		}
		g.connectButton.Enable()
	}
}

func (g *URnetworkGUI) onClearLogsClick() {
	g.logMutex.Lock()
	defer g.logMutex.Unlock()
	g.logText.SetText("")
}

func (g *URnetworkGUI) updateStatusLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		status := g.client.GetStatus()
		g.updateUIStatus(status)
		g.updateStats()
	}
}

func (g *URnetworkGUI) updateStats() {
	if g.client.GetStatus() == StatusConnected {
		bytesIn, bytesOut := g.client.GetStats()
		statsText := fmt.Sprintf("Network Stats: ↓ %s received, ↑ %s sent",
			formatBytes(bytesIn),
			formatBytes(bytesOut))
		g.statsLabel.SetText(statsText)
	} else {
		g.statsLabel.SetText("Network Stats: N/A")
	}
}

func (g *URnetworkGUI) loadAvailableRelays() {
	g.appendLog("Fetching available relays...")
	relays, err := g.client.GetRelays()
	if err != nil {
		g.appendLog(fmt.Sprintf("Failed to load relays: %v", err))
		return
	}

	g.availableRelays = relays
	options := []string{"Auto (Best Relay)"}
	for _, relay := range relays {
		option := fmt.Sprintf("%s (%s) - %dms", relay.Name, relay.Country, relay.PingMs)
		options = append(options, option)
	}

	g.relaySelect.Options = options
	g.relaySelect.Refresh()
	g.appendLog(fmt.Sprintf("Loaded %d available relays", len(relays)))
}

func (g *URnetworkGUI) onRefreshRelaysClick() {
	go g.loadAvailableRelays()
}

func (g *URnetworkGUI) onRelaySelected(option string) {
	if option == "Auto (Best Relay)" {
		g.client.SetRelay("")
		g.appendLog("Relay selection: Auto (Best Relay)")
		return
	}

	for i, relay := range g.availableRelays {
		expectedOption := fmt.Sprintf("%s (%s) - %dms", relay.Name, relay.Country, relay.PingMs)
		if option == expectedOption {
			g.client.SetRelay(relay.ID)
			g.appendLog(fmt.Sprintf("Relay selection: %s (ID: %s)", relay.Name, relay.ID))
			break
		} else if i == len(g.availableRelays)-1 {
			g.appendLog(fmt.Sprintf("Warning: Selected relay not found: %s", option))
		}
	}
}

func (g *URnetworkGUI) updateUIStatus(status ConnectionStatus) {
	statusText := "Status: " + status.String()
	if status == StatusConnected {
		statusText += " ✓"
	}
	g.statusLabel.SetText(statusText)

	switch status {
	case StatusDisconnected:
		g.connectButton.SetText("Connect to URnetwork")
		g.connectButton.Enable()
		g.connectButton.Importance = widget.HighImportance
		g.relaySelect.Enable()
	case StatusConnecting:
		g.connectButton.SetText("Connecting...")
		g.connectButton.Disable()
		g.relaySelect.Disable()
	case StatusConnected:
		g.connectButton.SetText("Disconnect")
		g.connectButton.Enable()
		g.connectButton.Importance = widget.MediumImportance
		g.relaySelect.Disable()
	case StatusError:
		g.connectButton.SetText("Retry Connection")
		g.connectButton.Enable()
		g.connectButton.Importance = widget.DangerImportance
		g.relaySelect.Enable()
	}

	g.connectButton.Refresh()
	g.statusLabel.Refresh()
	g.statsLabel.Refresh()
	g.relaySelect.Refresh()
}

func (g *URnetworkGUI) Run() {
	g.window.ShowAndRun()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	gui := NewURnetworkGUI()
	gui.Run()
}
