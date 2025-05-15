package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// PeerDiscovery manages peer discovery and connection
type PeerDiscovery struct {
	host           host.Host
	dht            *dht.IpfsDHT
	peers          map[peer.ID]*Peer
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	bootstrapPeers []string
}

// Peer represents a connected peer
type Peer struct {
	ID        peer.ID
	Addrs     []multiaddr.Multiaddr
	LastSeen  time.Time
	Latency   time.Duration
	IsActive  bool
	UserAgent string
	Version   string
}

// NewPeerDiscovery creates a new peer discovery system
func NewPeerDiscovery(bootstrapPeers []string) (*PeerDiscovery, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.NATPortMap(),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %v", err)
	}

	// Create DHT
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create DHT: %v", err)
	}

	return &PeerDiscovery{
		host:           h,
		dht:            kademliaDHT,
		peers:          make(map[peer.ID]*Peer),
		ctx:            ctx,
		cancel:         cancel,
		bootstrapPeers: bootstrapPeers,
	}, nil
}

// Start starts the peer discovery system
func (pd *PeerDiscovery) Start() error {
	// Bootstrap DHT
	if err := pd.dht.Bootstrap(pd.ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %v", err)
	}

	// Connect to bootstrap peers
	for _, addr := range pd.bootstrapPeers {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			continue
		}

		if err := pd.host.Connect(pd.ctx, *peerInfo); err != nil {
			continue
		}
	}

	// Start peer discovery loop
	go pd.discoveryLoop()

	return nil
}

// Stop stops the peer discovery system
func (pd *PeerDiscovery) Stop() {
	pd.cancel()
	pd.dht.Close()
	pd.host.Close()
}

// discoveryLoop continuously discovers new peers
func (pd *PeerDiscovery) discoveryLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-pd.ctx.Done():
			return
		case <-ticker.C:
			pd.discoverPeers()
		}
	}
}

// discoverPeers discovers new peers using DHT
func (pd *PeerDiscovery) discoverPeers() {
	// Generate a random peer ID to search for
	randomID := peer.ID(fmt.Sprintf("random-%d", time.Now().UnixNano()))

	// Find closest peers
	peers, err := pd.dht.FindClosestPeers(pd.ctx, randomID)
	if err != nil {
		return
	}

	// Connect to discovered peers
	for _, p := range peers {
		if err := pd.connectToPeer(p); err != nil {
			continue
		}
	}
}

// connectToPeer connects to a peer and adds it to the peer list
func (pd *PeerDiscovery) connectToPeer(p peer.ID) error {
	// Check if already connected
	pd.mu.RLock()
	if _, exists := pd.peers[p]; exists {
		pd.mu.RUnlock()
		return nil
	}
	pd.mu.RUnlock()

	// Connect to peer
	if err := pd.host.Connect(pd.ctx, peer.AddrInfo{ID: p}); err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	// Get peer addresses
	addrs := pd.host.Peerstore().Addrs(p)
	if len(addrs) == 0 {
		return fmt.Errorf("no addresses found for peer")
	}

	// Measure latency
	latency, err := pd.measureLatency(p)
	if err != nil {
		return fmt.Errorf("failed to measure latency: %v", err)
	}

	// Add peer to list
	pd.mu.Lock()
	pd.peers[p] = &Peer{
		ID:       p,
		Addrs:    addrs,
		LastSeen: time.Now(),
		Latency:  latency,
		IsActive: true,
	}
	pd.mu.Unlock()

	return nil
}

// measureLatency measures the latency to a peer
func (pd *PeerDiscovery) measureLatency(p peer.ID) (time.Duration, error) {
	start := time.Now()

	// Create a temporary stream to measure latency
	stream, err := pd.host.NewStream(pd.ctx, p, "/ping/1.0.0")
	if err != nil {
		return 0, err
	}
	defer stream.Close()

	// Send ping
	if _, err := stream.Write([]byte("ping")); err != nil {
		return 0, err
	}

	// Read pong
	buf := make([]byte, 4)
	if _, err := stream.Read(buf); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// GetPeers returns all connected peers
func (pd *PeerDiscovery) GetPeers() []*Peer {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	peers := make([]*Peer, 0, len(pd.peers))
	for _, p := range pd.peers {
		peers = append(peers, p)
	}
	return peers
}

// GetActivePeers returns all active peers
func (pd *PeerDiscovery) GetActivePeers() []*Peer {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	var activePeers []*Peer
	for _, p := range pd.peers {
		if p.IsActive {
			activePeers = append(activePeers, p)
		}
	}
	return activePeers
}

// RemovePeer removes a peer from the peer list
func (pd *PeerDiscovery) RemovePeer(p peer.ID) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	delete(pd.peers, p)
}

// UpdatePeer updates peer information
func (pd *PeerDiscovery) UpdatePeer(p peer.ID, userAgent, version string) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if peer, exists := pd.peers[p]; exists {
		peer.LastSeen = time.Now()
		peer.UserAgent = userAgent
		peer.Version = version
	}
}
