package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/config"
)

// Info represents peer information
type Info struct {
	Address  string
	Height   uint64
	Version  string
	LastSeen time.Time
}

// Peer represents a network peer
type Peer struct {
	conn   net.Conn
	info   Info
	mu     sync.RWMutex
	stopCh chan struct{}
}

// NewPeer creates a new peer
func NewPeer(conn net.Conn, info Info) *Peer {
	return &Peer{
		conn:   conn,
		info:   info,
		stopCh: make(chan struct{}),
	}
}

// GetHeight returns the peer's height
func (p *Peer) GetHeight() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Height
}

// SetHeight sets the peer's height
func (p *Peer) SetHeight(height uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.Height = height
	p.info.LastSeen = time.Now()
}

// Send sends a message to the peer
func (p *Peer) Send(message []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Set write deadline
	if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("failed to set write deadline: %v", err)
	}

	// Write message length
	length := uint32(len(message))
	if err := binary.Write(p.conn, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Write message
	if _, err := p.conn.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

// Close closes the peer connection
func (p *Peer) Close() error {
	close(p.stopCh)
	return p.conn.Close()
}

// Manager manages network peers
type Manager struct {
	config *config.Config
	peers  map[string]*Peer
	mu     sync.RWMutex
}

// NewManager creates a new peer manager
func NewManager(config *config.Config) *Manager {
	return &Manager{
		config: config,
		peers:  make(map[string]*Peer),
	}
}

// AddPeer adds a peer
func (m *Manager) AddPeer(conn net.Conn, info Info) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[info.Address] = NewPeer(conn, info)
}

// RemovePeer removes a peer
func (m *Manager) RemovePeer(address string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if peer, ok := m.peers[address]; ok {
		peer.Close()
		delete(m.peers, address)
	}
}

// GetPeer gets a peer
func (m *Manager) GetPeer(address string) (*Peer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	peer, ok := m.peers[address]
	if !ok {
		return nil, fmt.Errorf("peer not found: %s", address)
	}
	return peer, nil
}

// GetPeers gets all peers
func (m *Manager) GetPeers() []*Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	peers := make([]*Peer, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}
	return peers
}

// Broadcast broadcasts a message to all peers
func (m *Manager) Broadcast(message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, peer := range m.peers {
		go func(p *Peer) {
			if err := p.Send(message); err != nil {
				m.RemovePeer(p.info.Address)
			}
		}(peer)
	}
}

// Start starts the peer manager
func (m *Manager) Start() error {
	// Start peer discovery
	go m.discoverPeers()
	return nil
}

// Stop stops the peer manager
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, peer := range m.peers {
		peer.Close()
	}
	m.peers = make(map[string]*Peer)
}

// discoverPeers discovers new peers
func (m *Manager) discoverPeers() {
	// TODO: Implement peer discovery using DNS seeds and hardcoded nodes
}
