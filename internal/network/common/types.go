package common

import (
	"net"
	"sync"
	"time"
)

// MessageType represents the type of network message
type MessageType uint8

const (
	MessageTypePing MessageType = iota
	MessageTypePong
	MessageTypeBlock
	MessageTypeTransaction
	MessageTypePeerDiscovery
	MessageTypePeerList
)

// NetworkMessage represents a message sent over the network
type NetworkMessage struct {
	Type      MessageType
	From      string
	To        string
	Payload   []byte
	Timestamp time.Time
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	NodeID         string
	ListenPort     int
	MaxPeers       int
	BootstrapPeers []string
	PingInterval   time.Duration
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// Peer represents a network peer
type Peer struct {
	ID          string
	Address     string
	Port        int
	LastSeen    time.Time
	Latency     time.Duration
	Version     string
	IsActive    bool
	IsBootstrap bool
	Conn        net.Conn
	mu          sync.RWMutex
}

// NewPeer creates a new peer instance
func NewPeer(id, address string, port int) *Peer {
	return &Peer{
		ID:       id,
		Address:  address,
		Port:     port,
		LastSeen: time.Now(),
		IsActive: true,
	}
}

// UpdateLastSeen updates the last seen timestamp
func (p *Peer) UpdateLastSeen() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.LastSeen = time.Now()
}

// SetConnection sets the peer's connection
func (p *Peer) SetConnection(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Conn = conn
}

// GetConnection returns the peer's connection
func (p *Peer) GetConnection() net.Conn {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Conn
}
