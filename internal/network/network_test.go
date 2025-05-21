package network

import (
	"fmt"
	"testing"
	"time"
)

func TestNetworkManager(t *testing.T) {
	// Create test configuration
	config := &NetworkConfig{
		ListenPort:     8000,
		BootstrapPeers: []string{"localhost:8001"},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	// Create network manager
	nm := NewNetworkManager(config)

	// Test starting the network manager
	err := nm.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager: %v", err)
	}

	// Test adding a peer
	peer := &Peer{
		ID:       "test-peer",
		Address:  "localhost",
		Port:     8001,
		LastSeen: time.Now(),
		IsActive: true,
		Version:  "1.0.0",
	}

	err = nm.AddPeer(peer)
	if err != nil {
		t.Fatalf("Failed to add peer: %v", err)
	}

	// Test getting peers
	peers := nm.GetPeers()
	if len(peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(peers))
	}

	// Test sending a message
	msg := &NetworkMessage{
		Type:      "ping",
		From:      "self",
		To:        peer.ID,
		Timestamp: time.Now(),
	}

	err = nm.SendMessage(msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Test removing a peer
	nm.RemovePeer(peer.ID)
	peers = nm.GetPeers()
	if len(peers) != 0 {
		t.Fatalf("Expected 0 peers, got %d", len(peers))
	}

	// Test stopping the network manager
	nm.Stop()
}

func TestPeerDiscovery(t *testing.T) {
	// Create two network managers
	config1 := &NetworkConfig{
		ListenPort:     8000,
		BootstrapPeers: []string{},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	config2 := &NetworkConfig{
		ListenPort:     8001,
		BootstrapPeers: []string{"localhost:8000"},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	nm1 := NewNetworkManager(config1)
	nm2 := NewNetworkManager(config2)

	// Start both network managers
	err := nm1.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager 1: %v", err)
	}

	err = nm2.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager 2: %v", err)
	}

	// Wait for peer discovery
	time.Sleep(2 * time.Second)

	// Check if peers are discovered
	peers1 := nm1.GetPeers()
	peers2 := nm2.GetPeers()

	if len(peers1) == 0 || len(peers2) == 0 {
		t.Fatalf("Peer discovery failed: nm1 has %d peers, nm2 has %d peers", len(peers1), len(peers2))
	}

	// Stop both network managers
	nm1.Stop()
	nm2.Stop()
}

func TestNetworkPartition(t *testing.T) {
	// Create three network managers
	config1 := &NetworkConfig{
		ListenPort:     8000,
		BootstrapPeers: []string{},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	config2 := &NetworkConfig{
		ListenPort:     8001,
		BootstrapPeers: []string{"localhost:8000"},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	config3 := &NetworkConfig{
		ListenPort:     8002,
		BootstrapPeers: []string{"localhost:8000"},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	nm1 := NewNetworkManager(config1)
	nm2 := NewNetworkManager(config2)
	nm3 := NewNetworkManager(config3)

	// Start all network managers
	err := nm1.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager 1: %v", err)
	}

	err = nm2.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager 2: %v", err)
	}

	err = nm3.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager 3: %v", err)
	}

	// Wait for network to stabilize
	time.Sleep(2 * time.Second)

	// Simulate network partition by stopping nm1
	nm1.Stop()

	// Wait for partition detection
	time.Sleep(2 * time.Second)

	// Check if nm2 and nm3 can still communicate
	peers2 := nm2.GetPeers()
	peers3 := nm3.GetPeers()

	if len(peers2) == 0 || len(peers3) == 0 {
		t.Fatalf("Network partition recovery failed: nm2 has %d peers, nm3 has %d peers", len(peers2), len(peers3))
	}

	// Stop remaining network managers
	nm2.Stop()
	nm3.Stop()
}

func TestMessageHandling(t *testing.T) {
	// Create test configuration
	config := &NetworkConfig{
		ListenPort:     8000,
		BootstrapPeers: []string{},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	// Create network manager
	nm := NewNetworkManager(config)

	// Start network manager
	err := nm.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager: %v", err)
	}

	// Create test peer
	peer := &Peer{
		ID:       "test-peer",
		Address:  "localhost",
		Port:     8001,
		LastSeen: time.Now(),
		IsActive: true,
		Version:  "1.0.0",
	}

	// Add peer
	err = nm.AddPeer(peer)
	if err != nil {
		t.Fatalf("Failed to add peer: %v", err)
	}

	// Test different message types
	messageTypes := []string{"ping", "get_peers", "peer_list", "pong"}

	for _, msgType := range messageTypes {
		msg := &NetworkMessage{
			Type:      msgType,
			From:      "self",
			To:        peer.ID,
			Timestamp: time.Now(),
		}

		err = nm.SendMessage(msg)
		if err != nil {
			t.Fatalf("Failed to send %s message: %v", msgType, err)
		}
	}

	// Stop network manager
	nm.Stop()
}

func TestConnectionManagement(t *testing.T) {
	// Create test configuration
	config := &NetworkConfig{
		ListenPort:     8000,
		BootstrapPeers: []string{},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	// Create network manager
	nm := NewNetworkManager(config)

	// Start network manager
	err := nm.Start()
	if err != nil {
		t.Fatalf("Failed to start network manager: %v", err)
	}

	// Create test peers
	for i := 0; i < 5; i++ {
		peer := &Peer{
			ID:       fmt.Sprintf("test-peer-%d", i),
			Address:  "localhost",
			Port:     8001 + i,
			LastSeen: time.Now(),
			IsActive: true,
			Version:  "1.0.0",
		}

		err = nm.AddPeer(peer)
		if err != nil {
			t.Fatalf("Failed to add peer %d: %v", i, err)
		}
	}

	// Wait for connection management
	time.Sleep(2 * time.Second)

	// Check number of peers
	peers := nm.GetPeers()
	if len(peers) != 5 {
		t.Fatalf("Expected 5 peers, got %d", len(peers))
	}

	// Simulate peer disconnection
	nm.RemovePeer("test-peer-0")

	// Wait for connection management
	time.Sleep(2 * time.Second)

	// Check number of peers
	peers = nm.GetPeers()
	if len(peers) != 4 {
		t.Fatalf("Expected 4 peers, got %d", len(peers))
	}

	// Stop network manager
	nm.Stop()
}
