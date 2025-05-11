package network

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/youngchain/internal/core/types"
)

// Peer represents a connected peer in the network
type Peer struct {
	ID          uint64
	conn        net.Conn
	Addr        net.Addr
	Version     uint32
	Services    uint64
	UserAgent   string
	lastSeen    time.Time
	messageChan chan *types.Message
	stopChan    chan struct{}
	writeMutex  sync.Mutex
}

// PeerManager manages network peers
type PeerManager struct {
	peers     map[uint64]*Peer
	mu        sync.RWMutex
	maxPeers  int
	bootstrap []string
}

// NewPeerManager creates a new peer manager
func NewPeerManager(maxPeers int, bootstrap []string) *PeerManager {
	return &PeerManager{
		peers:     make(map[uint64]*Peer),
		maxPeers:  maxPeers,
		bootstrap: bootstrap,
	}
}

// AddPeer adds a peer to the manager
func (pm *PeerManager) AddPeer(peer *Peer) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.peers) >= pm.maxPeers {
		return fmt.Errorf("maximum number of peers reached")
	}

	pm.peers[peer.ID] = peer
	return nil
}

// RemovePeer removes a peer from the manager
func (pm *PeerManager) RemovePeer(id uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.peers, id)
}

// GetPeer gets a peer by ID
func (pm *PeerManager) GetPeer(id uint64) (*Peer, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peer, exists := pm.peers[id]
	return peer, exists
}

// GetRandomPeers gets a random subset of peers
func (pm *PeerManager) GetRandomPeers(count int) []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if count > len(pm.peers) {
		count = len(pm.peers)
	}

	peers := make([]*Peer, 0, count)
	for _, peer := range pm.peers {
		peers = append(peers, peer)
		if len(peers) == count {
			break
		}
	}

	return peers
}

// DiscoverPeers discovers new peers
func (pm *PeerManager) DiscoverPeers() error {
	// Try bootstrap nodes first
	for _, addr := range pm.bootstrap {
		if p, err := pm.connectToPeer(addr); err == nil {
			if err := pm.AddPeer(p); err != nil {
				continue
			}
		}
	}

	// Ask existing peers for more peers
	peers := pm.GetRandomPeers(5)
	for _, p := range peers {
		// Send getaddr message
		msg := &types.Message{Type: types.GetAddrMsg, Payload: nil}
		if err := p.SendMessage(msg); err != nil {
			continue
		}
	}

	return nil
}

// connectToPeer connects to a peer
func (pm *PeerManager) connectToPeer(addr string) (*Peer, error) {
	// Generate random peer ID
	id := make([]byte, 8)
	rand.Read(id)
	peerID := binary.BigEndian.Uint64(id)

	// Create peer
	peer := &Peer{
		ID:       peerID,
		lastSeen: time.Now(),
	}

	// Parse address
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	peer.Addr = tcpAddr

	return peer, nil
}

// UpdatePeer updates peer information
func (pm *PeerManager) UpdatePeer(id uint64, version uint32, services uint64, userAgent string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if peer, exists := pm.peers[id]; exists {
		peer.Version = version
		peer.Services = services
		peer.UserAgent = userAgent
		peer.lastSeen = time.Now()
	}
}

// Cleanup removes inactive peers
func (pm *PeerManager) Cleanup(maxAge time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	for id, peer := range pm.peers {
		if now.Sub(peer.lastSeen) > maxAge {
			delete(pm.peers, id)
		}
	}
}

// GetPeerCount returns the number of peers
func (pm *PeerManager) GetPeerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return len(pm.peers)
}

// GetBootstrapPeers returns the bootstrap peers
func (pm *PeerManager) GetBootstrapPeers() []string {
	return pm.bootstrap
}

// SetBootstrapPeers sets the bootstrap peers
func (pm *PeerManager) SetBootstrapPeers(peers []string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.bootstrap = peers
}

// NewPeer creates a new peer from a connection
func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		conn:        conn,
		Addr:        conn.RemoteAddr(),
		lastSeen:    time.Now(),
		messageChan: make(chan *types.Message, 100),
		stopChan:    make(chan struct{}),
	}
}

// Start starts the peer's message handling
func (p *Peer) Start() {
	go p.readLoop()
	go p.writeLoop()
}

// Stop stops the peer's message handling
func (p *Peer) Stop() {
	close(p.stopChan)
	p.conn.Close()
}

// RemoteAddr returns the peer's remote address
func (p *Peer) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

// SendMessage sends a message to the peer
func (p *Peer) SendMessage(msg *types.Message) error {
	p.writeMutex.Lock()
	defer p.writeMutex.Unlock()

	// Encode message
	data, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode message: %v", err)
	}

	// Write message length
	if err := binary.Write(p.conn, binary.BigEndian, uint32(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Write message data
	if _, err := p.conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %v", err)
	}

	return nil
}

// readLoop reads messages from the connection
func (p *Peer) readLoop() {
	for {
		select {
		case <-p.stopChan:
			return
		default:
			// Read message length
			var length uint32
			if err := binary.Read(p.conn, binary.BigEndian, &length); err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading message length: %v\n", err)
				}
				return
			}

			// Read message data
			data := make([]byte, length)
			if _, err := io.ReadFull(p.conn, data); err != nil {
				fmt.Printf("Error reading message data: %v\n", err)
				return
			}

			// Decode message
			msg := &types.Message{}
			if err := msg.Decode(data); err != nil {
				fmt.Printf("Error decoding message: %v\n", err)
				continue
			}

			// Update last seen time
			p.lastSeen = time.Now()

			// Send message to channel
			select {
			case p.messageChan <- msg:
			default:
				fmt.Println("Message channel full, dropping message")
			}
		}
	}
}

// writeLoop writes messages to the connection
func (p *Peer) writeLoop() {
	for {
		select {
		case <-p.stopChan:
			return
		case msg := <-p.messageChan:
			if err := p.SendMessage(msg); err != nil {
				fmt.Printf("Error sending message: %v\n", err)
				return
			}
		}
	}
}

// Add exported LastSeen method
func (p *Peer) LastSeen() time.Time {
	return p.lastSeen
}
