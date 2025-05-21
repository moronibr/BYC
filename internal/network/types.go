package network

import (
	"encoding/json"
	"fmt"
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
	Type    MessageType
	From    string
	To      string
	Payload []byte
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	NodeID         string
	ListenPort     int
	MaxPeers       int
	BootstrapPeers []string
}

// Peer represents a network peer
type Peer struct {
	ID          string
	Address     string
	Port        int
	LastSeen    int64
	Latency     int64
	Version     string
	IsActive    bool
	IsBootstrap bool
	Conn        net.Conn
}

// NetworkManager manages network operations
type NetworkManager struct {
	config *NetworkConfig
	peers  map[string]*Peer
	mu     sync.RWMutex
	server net.Listener
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(config *NetworkConfig) *NetworkManager {
	return &NetworkManager{
		config: config,
		peers:  make(map[string]*Peer),
	}
}

// GetPeers returns the list of peers
func (nm *NetworkManager) GetPeers() []*Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	peers := make([]*Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, peer)
	}
	return peers
}

// Ping sends a ping message to a peer
func (nm *NetworkManager) Ping(peerID string) error {
	msg := &NetworkMessage{
		Type:    MessageTypePing,
		From:    nm.config.NodeID,
		To:      peerID,
		Payload: []byte("ping"),
	}
	return nm.SendMessage(msg)
}

// SendMessage sends a message to a peer
func (nm *NetworkManager) SendMessage(msg *NetworkMessage) error {
	nm.mu.RLock()
	peer, ok := nm.peers[msg.To]
	nm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("peer %s not found", msg.To)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	_, err = peer.Conn.Write(data)
	return err
}

// ConnectToPeer establishes a connection to a peer
func (nm *NetworkManager) ConnectToPeer(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	peer := &Peer{
		Address: address,
		Conn:    conn,
	}

	nm.mu.Lock()
	nm.peers[address] = peer
	nm.mu.Unlock()

	return nil
}

// handleMessage handles a received message
func (nm *NetworkManager) handleMessage(msg *NetworkMessage) error {
	switch msg.Type {
	case MessageTypePing:
		return nm.handlePing(msg)
	case MessageTypePong:
		return nm.handlePong(msg)
	case MessageTypeBlock:
		return nm.handleBlock(msg)
	case MessageTypeTransaction:
		return nm.handleTransaction(msg)
	case MessageTypePeerDiscovery:
		return nm.handlePeerDiscovery(msg)
	case MessageTypePeerList:
		return nm.handlePeerList(msg)
	default:
		return fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

// handlePing handles a ping message
func (nm *NetworkManager) handlePing(msg *NetworkMessage) error {
	pong := &NetworkMessage{
		Type:    MessageTypePong,
		From:    nm.config.NodeID,
		To:      msg.From,
		Payload: []byte("pong"),
	}
	return nm.SendMessage(pong)
}

// handlePong handles a pong message
func (nm *NetworkManager) handlePong(msg *NetworkMessage) error {
	// Update peer latency
	nm.mu.Lock()
	if peer, ok := nm.peers[msg.From]; ok {
		peer.LastSeen = time.Now().UnixNano()
	}
	nm.mu.Unlock()
	return nil
}

// handleBlock handles a block message
func (nm *NetworkManager) handleBlock(msg *NetworkMessage) error {
	// TODO: Implement block handling
	return nil
}

// handleTransaction handles a transaction message
func (nm *NetworkManager) handleTransaction(msg *NetworkMessage) error {
	// TODO: Implement transaction handling
	return nil
}

// handlePeerDiscovery handles a peer discovery message
func (nm *NetworkManager) handlePeerDiscovery(msg *NetworkMessage) error {
	// TODO: Implement peer discovery
	return nil
}

// handlePeerList handles a peer list message
func (nm *NetworkManager) handlePeerList(msg *NetworkMessage) error {
	// TODO: Implement peer list handling
	return nil
}
