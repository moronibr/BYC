package wallet

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient represents a WebSocket client
type WSClient struct {
	conn     *websocket.Conn
	send     chan []byte
	deviceID string
	ip       string
	lastPing time.Time
	mu       sync.Mutex
}

// WSEvent represents a WebSocket event
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	IPWhitelist    []string
	RequestTimeout time.Duration
	MaxConnections int
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteWait      time.Duration
	MaxMessageSize int64
	AllowedOrigins []string
	APIKeyRequired bool
	APIKeys        map[string]string // API key -> device ID
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RequestTimeout: 30 * time.Second,
		MaxConnections: 100,
		PingInterval:   30 * time.Second,
		PongWait:       60 * time.Second,
		WriteWait:      10 * time.Second,
		MaxMessageSize: 512 * 1024, // 512KB
		AllowedOrigins: []string{"*"},
		APIKeyRequired: true,
		APIKeys:        make(map[string]string),
	}
}

// WSServer represents the WebSocket server
type WSServer struct {
	upgrader   websocket.Upgrader
	clients    map[*WSClient]bool
	broadcast  chan *WSEvent
	register   chan *WSClient
	unregister chan *WSClient
	config     *SecurityConfig
	mu         sync.RWMutex
}

// NewWSServer creates a new WebSocket server
func NewWSServer(config *SecurityConfig) *WSServer {
	return &WSServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, allowed := range config.AllowedOrigins {
					if allowed == "*" || allowed == origin {
						return true
					}
				}
				return false
			},
		},
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan *WSEvent),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		config:     config,
	}
}

// ServeWS handles WebSocket connections
func (s *WSServer) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Check IP whitelist
	if !s.isIPAllowed(r.RemoteAddr) {
		http.Error(w, "IP not allowed", http.StatusForbidden)
		return
	}

	// Check API key if required
	if s.config.APIKeyRequired {
		apiKey := r.Header.Get("X-API-Key")
		if !s.validateAPIKey(apiKey) {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &WSClient{
		conn:     conn,
		send:     make(chan []byte, 256),
		ip:       r.RemoteAddr,
		lastPing: time.Now(),
	}

	s.register <- client

	// Start goroutines for reading and writing
	go s.readPump(client)
	go s.writePump(client)
}

// readPump pumps messages from the WebSocket connection
func (s *WSServer) readPump(c *WSClient) {
	defer func() {
		s.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(s.config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v\n", err)
			}
			break
		}

		// Handle message
		var event WSEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		// Process event
		s.handleEvent(c, &event)
	}
}

// writePump pumps messages to the WebSocket connection
func (s *WSServer) writePump(c *WSClient) {
	ticker := time.NewTicker(s.config.PingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(s.config.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(s.config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleEvent processes WebSocket events
func (s *WSServer) handleEvent(c *WSClient, event *WSEvent) {
	switch event.Type {
	case "subscribe":
		if deviceID, ok := event.Payload.(string); ok {
			c.deviceID = deviceID
		}
	case "unsubscribe":
		c.deviceID = ""
	case "ping":
		c.lastPing = time.Now()
		c.send <- []byte(`{"type":"pong"}`)
	}
}

// BroadcastEvent broadcasts an event to all clients
func (s *WSServer) BroadcastEvent(event *WSEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for client := range s.clients {
		if client.deviceID == event.Type {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(s.clients, client)
			}
		}
	}
}

// isIPAllowed checks if an IP is allowed
func (s *WSServer) isIPAllowed(addr string) bool {
	if len(s.config.IPWhitelist) == 0 {
		return true
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}

	for _, allowed := range s.config.IPWhitelist {
		if allowed == host {
			return true
		}
	}

	return false
}

// validateAPIKey validates an API key
func (s *WSServer) validateAPIKey(key string) bool {
	if !s.config.APIKeyRequired {
		return true
	}

	_, exists := s.config.APIKeys[key]
	return exists
}

// SignRequest signs a request with HMAC-SHA256
func SignRequest(apiKey string, method string, params []byte) string {
	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(method))
	h.Write(params)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyRequest verifies a request signature
func (s *WSServer) VerifyRequest(apiKey, method string, params []byte, signature string) bool {
	expected := SignRequest(apiKey, method, params)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// AddAPIKey adds an API key
func (s *WSServer) AddAPIKey(key, deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.APIKeys[key] = deviceID
}

// RemoveAPIKey removes an API key
func (s *WSServer) RemoveAPIKey(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.config.APIKeys, key)
}

// AddIPToWhitelist adds an IP to the whitelist
func (s *WSServer) AddIPToWhitelist(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.IPWhitelist = append(s.config.IPWhitelist, ip)
}

// RemoveIPFromWhitelist removes an IP from the whitelist
func (s *WSServer) RemoveIPFromWhitelist(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, allowed := range s.config.IPWhitelist {
		if allowed == ip {
			s.config.IPWhitelist = append(s.config.IPWhitelist[:i], s.config.IPWhitelist[i+1:]...)
			break
		}
	}
}
