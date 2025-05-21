package network

import (
	"context"
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

// Peer represents a network peer
type Peer struct {
	ID         string
	Address    string
	LastSeen   time.Time
	Connection net.Conn
	Node       *Node
	handlers   map[MessageType]MessageHandler
}

// NetworkManager manages the P2P network
type NetworkManager struct {
	config         *NetworkConfig
	peers          map[string]*Peer
	bootstrapPeers []*Peer
	connections    map[string]net.Conn
	messageChan    chan *NetworkMessage
	stopChan       chan struct{}
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex
}

// Node represents a P2P node
type Node struct {
	Config     *Config
	Blockchain *blockchain.Blockchain
	Peers      map[string]*Peer
	server     net.Listener
	mu         sync.RWMutex
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
