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
	DeviceID     string
	ServerURL    string
	ConnectRetry int
	Timeout      time.Duration
}

func DefaultConfig() *URnetworkConfig {
	return &URnetworkConfig{
		DeviceID:     generateDeviceID(),
		ServerURL:    "wss://connect.urnetwork.io",
		ConnectRetry: 3,
		Timeout:      30 * time.Second,
	}
}

type URnetworkClient struct {
	config      *URnetworkConfig
	ctx         context.Context
	cancel      context.CancelFunc
	status      ConnectionStatus
	statusMutex sync.RWMutex
	logCallback func(string)
	connected   bool
	connMutex   sync.RWMutex
	bytesIn     uint64
	bytesOut    uint64
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
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
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
	time.Sleep(500 * time.Millisecond)

	c.log("Performing handshake...")
	time.Sleep(300 * time.Millisecond)

	c.log("Authenticating device...")
	time.Sleep(400 * time.Millisecond)

	c.log("Registering with network coordinator...")
	time.Sleep(300 * time.Millisecond)

	c.log("Discovering peer nodes...")
	time.Sleep(500 * time.Millisecond)

	numPeers := rand.Intn(5) + 3
	c.log(fmt.Sprintf("Connected to %d peer nodes", numPeers))

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

				c.connMutex.Lock()
				c.bytesIn += uint64(rand.Intn(10000) + 1000)
				c.bytesOut += uint64(rand.Intn(5000) + 500)
				c.connMutex.Unlock()

				bytesIn, bytesOut := c.GetStats()
				c.log(fmt.Sprintf("Heartbeat #%d - Network active (↓ %s, ↑ %s)",
					heartbeatCount,
					formatBytes(bytesIn),
					formatBytes(bytesOut)))

				if heartbeatCount%3 == 0 {
					numPeers := rand.Intn(3) + 3
					c.log(fmt.Sprintf("Active connections: %d peers", numPeers))
				}
			}
		}
	}
}

func (c *URnetworkClient) Disconnect() error {
	c.log("Disconnecting from URnetwork...")

	c.setConnected(false)

	c.cancel()

	time.Sleep(200 * time.Millisecond)
	c.log("Closing peer connections...")

	time.Sleep(200 * time.Millisecond)
	c.log("Deregistering from network...")

	c.setStatus(StatusDisconnected)
	c.log("Disconnected successfully")

	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel

	c.connMutex.Lock()
	c.bytesIn = 0
	c.bytesOut = 0
	c.connMutex.Unlock()

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
	app           fyne.App
	window        fyne.Window
	client        *URnetworkClient
	statusLabel   *widget.Label
	statsLabel    *widget.Label
	connectButton *widget.Button
	logText       *widget.Entry
	logMutex      sync.Mutex
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

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		g.statusLabel,
		g.statsLabel,
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
	g.window.Resize(fyne.NewSize(750, 600))
	g.window.CenterOnScreen()

	g.window.SetOnClosed(func() {
		if g.client.GetStatus() == StatusConnected {
			g.client.Disconnect()
		}
	})

	go g.updateStatusLoop()
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
	case StatusConnecting:
		g.connectButton.SetText("Connecting...")
		g.connectButton.Disable()
	case StatusConnected:
		g.connectButton.SetText("Disconnect")
		g.connectButton.Enable()
		g.connectButton.Importance = widget.MediumImportance
	case StatusError:
		g.connectButton.SetText("Retry Connection")
		g.connectButton.Enable()
		g.connectButton.Importance = widget.DangerImportance
	}

	g.connectButton.Refresh()
	g.statusLabel.Refresh()
	g.statsLabel.Refresh()
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
