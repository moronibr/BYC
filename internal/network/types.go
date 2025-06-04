package network

import (
	"net"
	"sync"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
)

// MessageType represents the type of network message
type MessageType string

const (
	MessageTypePing      MessageType = "PING"
	MessageTypePong      MessageType = "PONG"
	MessageTypeBlock     MessageType = "BLOCK"
	MessageTypeTx        MessageType = "TX"
	MessageTypeGetBlocks MessageType = "GET_BLOCKS"
	MessageTypeBlocks    MessageType = "BLOCKS"
	MessageTypeGetData   MessageType = "GET_DATA"
	MessageTypeInv       MessageType = "INV"
	MessageTypeAddr      MessageType = "ADDR"
	MessageTypeGetAddr   MessageType = "GET_ADDR"
	MessageTypeVerAck    MessageType = "VERACK"
	MessageTypeVersion   MessageType = "VERSION"
	MessageTypeGetHeight MessageType = "GET_HEIGHT"
)

// Message represents a network message
type Message struct {
	Type    MessageType
	Payload []byte
}

// NetworkMessage represents a message sent over the network
type NetworkMessage struct {
	Type      MessageType
	From      string
	To        string
	Payload   []byte
	Timestamp time.Time
}

// NetworkConfig holds configuration for the network
type NetworkConfig struct {
	NodeID         string
	ListenAddr     string
	ListenPort     int
	MaxPeers       int
	PingTimeout    time.Duration
	PingInterval   time.Duration
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	BootstrapPeers []string
}

// Node represents a network node
type Node struct {
	Config     *Config
	Blockchain *blockchain.Blockchain
	Peers      map[string]*Peer
	server     net.Listener
	mu         sync.RWMutex
	isMining   bool
}

// Peer represents a network peer
type Peer struct {
	ID          string
	Address     string
	LastSeen    time.Time
	Latency     time.Duration
	Version     string
	IsActive    bool
	IsBootstrap bool
	conn        net.Conn
	Node        *Node
	handlers    map[MessageType]MessageHandler
	Height      int64
	mu          sync.RWMutex
}

// Config represents the node configuration
type Config struct {
	Address        string
	BlockType      blockchain.BlockType
	BootstrapPeers []string
}

// MessageHandler is a function that handles a message
type MessageHandler func(*Peer, []byte) error

// NewNetworkMessage creates a new network message
func NewNetworkMessage(msgType MessageType, from, to string, payload []byte) *NetworkMessage {
	return &NetworkMessage{
		Type:      msgType,
		From:      from,
		To:        to,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// NewNetworkConfig creates a new network configuration
func NewNetworkConfig(nodeID, listenAddr string, listenPort, maxPeers int) *NetworkConfig {
	return &NetworkConfig{
		NodeID:         nodeID,
		ListenAddr:     listenAddr,
		ListenPort:     listenPort,
		MaxPeers:       maxPeers,
		PingTimeout:    30 * time.Second,
		PingInterval:   60 * time.Second,
		DialTimeout:    10 * time.Second,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		BootstrapPeers: make([]string, 0),
	}
}

// NewPeer creates a new peer instance
func NewPeer(id, address string, port int) *Peer {
	return &Peer{
		ID:       id,
		Address:  address,
		LastSeen: time.Now(),
		IsActive: true,
		handlers: make(map[MessageType]MessageHandler),
		Height:   0,
		mu:       sync.RWMutex{},
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
	p.conn = conn
}

// GetConnection returns the peer's connection
func (p *Peer) GetConnection() net.Conn {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.conn
}
