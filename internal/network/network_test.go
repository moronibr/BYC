package network

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/moroni/BYC/internal/network/common"
)

func TestNetworkManager(t *testing.T) {
	config := &common.NetworkConfig{
		NodeID:         "test-node",
		ListenPort:     8080,
		MaxPeers:       10,
		BootstrapPeers: []string{"localhost:8081"},
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	nm := NewNetworkManager(config)
	if nm == nil {
		t.Fatal("failed to create network manager")
	}

	// Test peer management
	peer := common.NewPeer("test-peer", "localhost", 8081)
	peer.IsBootstrap = true

	nm.mu.Lock()
	nm.peers[peer.Address] = peer
	nm.mu.Unlock()

	peers := nm.GetPeers()
	if len(peers) != 1 {
		t.Errorf("expected 1 peer, got %d", len(peers))
	}

	// Test message handling
	msg := &common.NetworkMessage{
		Type:      common.MessageTypePing,
		From:      "test-peer",
		To:        "test-node",
		Payload:   []byte("ping"),
		Timestamp: time.Now(),
	}

	if err := nm.handleMessage(msg); err != nil {
		t.Errorf("failed to handle message: %v", err)
	}
}

func TestSecureNetworkManager(t *testing.T) {
	netConfig := &common.NetworkConfig{
		NodeID:         "test-node",
		ListenPort:     8080,
		MaxPeers:       10,
		BootstrapPeers: []string{"localhost:8081"},
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	secureConfig := NewSecureConfig()
	snm, err := NewSecureNetworkManager(netConfig, secureConfig)
	if err != nil {
		t.Fatalf("failed to create secure network manager: %v", err)
	}

	// Start the secure network manager
	err = snm.Start()
	if err != nil {
		t.Fatalf("failed to start secure network manager: %v", err)
	}
	defer snm.Stop()

	// Test secure peer creation
	peer := common.NewPeer("test-peer", "localhost", 8081)
	peer.IsBootstrap = true

	securePeer, err := NewSecurePeer(peer)
	if err != nil {
		t.Fatalf("failed to create secure peer: %v", err)
	}

	// Test message signing and verification
	msg := &common.NetworkMessage{
		Type:      common.MessageTypePing,
		From:      "test-peer",
		To:        "test-node",
		Payload:   []byte("ping"),
		Timestamp: time.Now(),
	}

	signedMsg, err := securePeer.SignMessage(msg)
	if err != nil {
		t.Fatalf("failed to sign message: %v", err)
	}

	if !securePeer.VerifyMessage(signedMsg) {
		t.Error("message verification failed")
	}

	// Test sending secure message
	err = snm.SendSecureMessage(msg)
	if err != nil {
		t.Fatalf("failed to send secure message: %v", err)
	}
}

func TestMultiplexedConn(t *testing.T) {
	// Create a test connection
	conn1, conn2 := net.Pipe()
	defer conn1.Close()
	defer conn2.Close()

	// Create multiplexed connections
	mc1 := NewMultiplexedConn(conn1, true)
	mc2 := NewMultiplexedConn(conn2, true)

	// Start the multiplexers
	mc1.Start()
	mc2.Start()

	// Create a stream
	stream := mc1.CreateStream(1)
	if stream == nil {
		t.Fatal("failed to create stream")
	}

	// Test data transfer
	testData := []byte("test data")
	if err := mc1.Write(stream.ID, testData); err != nil {
		t.Fatalf("failed to write data: %v", err)
	}

	receivedData, err := mc2.Read(stream.ID)
	if err != nil {
		t.Fatalf("failed to read data: %v", err)
	}

	if string(receivedData) != string(testData) {
		t.Errorf("expected %s, got %s", testData, receivedData)
	}

	// Test stream closure
	mc1.CloseStream(stream.ID)
	if _, err := mc2.Read(stream.ID); err != io.EOF {
		t.Error("expected EOF after stream closure")
	}
}

func TestPartitionManager(t *testing.T) {
	config := &common.NetworkConfig{
		NodeID:         "test-node",
		ListenPort:     8080,
		MaxPeers:       10,
		BootstrapPeers: []string{"localhost:8081"},
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	nm := NewNetworkManager(config)
	pm := NewPartitionManager(nm)

	// Test partition detection
	pm.Start()
	defer pm.Stop()

	// Add a peer
	peer := common.NewPeer("test-peer", "localhost", 8081)
	peer.IsBootstrap = true

	nm.mu.Lock()
	nm.peers[peer.Address] = peer
	nm.mu.Unlock()

	// Wait for partition check
	time.Sleep(pm.checkInterval + time.Second)

	// Check partition state
	state := pm.GetPartitionState()
	if state == nil {
		t.Fatal("failed to get partition state")
	}

	// Test partition recovery
	if err := pm.RecoverPartition(); err != nil {
		t.Errorf("failed to recover partition: %v", err)
	}
}

func TestPeerDiscovery(t *testing.T) {
	// Create two network managers
	config1 := &common.NetworkConfig{
		NodeID:         "node1",
		ListenPort:     8000,
		BootstrapPeers: []string{},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	config2 := &common.NetworkConfig{
		NodeID:         "node2",
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
	config1 := &common.NetworkConfig{
		NodeID:         "node1",
		ListenPort:     8000,
		BootstrapPeers: []string{},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	config2 := &common.NetworkConfig{
		NodeID:         "node2",
		ListenPort:     8001,
		BootstrapPeers: []string{"localhost:8000"},
		MaxPeers:       10,
		PingInterval:   time.Second,
		DialTimeout:    time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	}

	config3 := &common.NetworkConfig{
		NodeID:         "node3",
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
	config := &common.NetworkConfig{
		NodeID:         "test-node",
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
	peer := common.NewPeer("test-peer", "localhost", 8001)

	// Add peer
	err = nm.AddPeer(peer)
	if err != nil {
		t.Fatalf("Failed to add peer: %v", err)
	}

	// Test different message types
	messageTypes := []common.MessageType{
		common.MessageTypePing,
		common.MessageTypePeerDiscovery,
		common.MessageTypePeerList,
		common.MessageTypePong,
	}

	for _, msgType := range messageTypes {
		msg := &common.NetworkMessage{
			Type:      msgType,
			From:      "self",
			To:        peer.ID,
			Timestamp: time.Now(),
		}

		err = nm.SendMessage(msg)
		if err != nil {
			t.Fatalf("Failed to send message type %d: %v", msgType, err)
		}
	}

	// Stop network manager
	nm.Stop()
}

func TestConnectionManagement(t *testing.T) {
	// Create test configuration
	config := &common.NetworkConfig{
		NodeID:         "test-node",
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
		peer := common.NewPeer(fmt.Sprintf("test-peer-%d", i), "localhost", 8001+i)

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
