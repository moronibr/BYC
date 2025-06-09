package network

import (
	"testing"
	"time"

	"byc/internal/blockchain"
)

func getPeerAddresses(n *Node) []string {
	peers := make([]string, 0, len(n.Peers))
	for addr := range n.Peers {
		peers = append(peers, addr)
	}
	return peers
}

func TestNewNode(t *testing.T) {
	config := &Config{
		Address:        "localhost:8000",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	node, err := NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	if node.Config != config {
		t.Error("Node config not set correctly")
	}

	if node.Blockchain == nil {
		t.Error("Blockchain not initialized")
	}

	if node.Peers == nil {
		t.Error("Peers map not initialized")
	}
}

func TestNodeStartStop(t *testing.T) {
	config := &Config{
		Address:        "localhost:8001",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	node, err := NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	// Node starts automatically when created
	if node.server == nil {
		t.Error("Server not started")
	}

	// Stop node
	node.Stop()
}

func TestNodeConnect(t *testing.T) {
	config1 := &Config{
		Address:        "localhost:8002",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	config2 := &Config{
		Address:        "localhost:8003",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	node1, err := NewNode(config1)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	defer node1.Stop()

	node2, err := NewNode(config2)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	defer node2.Stop()

	// Connect node2 to node1
	err = node2.ConnectToPeer("localhost:8002")
	if err != nil {
		t.Fatalf("Failed to connect nodes: %v", err)
	}

	// Wait for connection to be established
	time.Sleep(100 * time.Millisecond)

	if len(node2.Peers) != 1 {
		t.Error("Node2 should have one peer")
	}

	if len(node1.Peers) != 1 {
		t.Error("Node1 should have one peer")
	}
}

func TestNodeBroadcast(t *testing.T) {
	config1 := &Config{
		Address:        "localhost:8004",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	config2 := &Config{
		Address:        "localhost:8005",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	node1, err := NewNode(config1)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	defer node1.Stop()

	node2, err := NewNode(config2)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	defer node2.Stop()

	// Connect node2 to node1
	err = node2.ConnectToPeer("localhost:8004")
	if err != nil {
		t.Fatalf("Failed to connect nodes: %v", err)
	}

	// Wait for connection to be established
	time.Sleep(100 * time.Millisecond)

	// Broadcast a message
	msg := &Message{
		Type:    PingMsg,
		Payload: []byte("test"),
	}

	err = node1.BroadcastMessage(msg)
	if err != nil {
		t.Fatalf("Failed to broadcast message: %v", err)
	}

	// Wait for message to be received
	time.Sleep(100 * time.Millisecond)
}

func TestPeerDisconnect(t *testing.T) {
	config1 := &Config{
		Address:        "localhost:8006",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	config2 := &Config{
		Address:        "localhost:8007",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	node1, err := NewNode(config1)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	defer node1.Stop()

	node2, err := NewNode(config2)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	defer node2.Stop()

	// Connect node2 to node1
	err = node2.ConnectToPeer("localhost:8006")
	if err != nil {
		t.Fatalf("Failed to connect nodes: %v", err)
	}

	// Wait for connection to be established
	time.Sleep(100 * time.Millisecond)

	// Disconnect node2
	node2.Stop()

	// Wait for disconnect to be detected
	time.Sleep(100 * time.Millisecond)

	if len(node1.Peers) != 0 {
		t.Error("Node1 should have no peers after disconnect")
	}
}

func TestConnectToInvalidPeer(t *testing.T) {
	config := &Config{
		Address:        "localhost:8008",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	node, err := NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer node.Stop()

	// Try to connect to non-existent peer
	err = node.ConnectToPeer("localhost:9999")
	if err == nil {
		t.Error("Should fail to connect to non-existent peer")
	}
}

func TestMultiplePeerConnections(t *testing.T) {
	config := &Config{
		Address:        "localhost:3030",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}
	node, err := NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer node.Stop()

	peerAddrs := []string{"localhost:3031", "localhost:3032", "localhost:3033"}
	peers := []*Node{}
	for _, addr := range peerAddrs {
		peerConfig := &Config{
			Address:        addr,
			BlockType:      blockchain.GoldenBlock,
			BootstrapPeers: []string{},
		}
		peer, err := NewNode(peerConfig)
		if err != nil {
			t.Fatalf("Failed to create peer %s: %v", addr, err)
		}
		peers = append(peers, peer)
	}
	time.Sleep(100 * time.Millisecond)

	for _, addr := range peerAddrs {
		err := node.ConnectToPeer(addr)
		if err != nil {
			t.Errorf("Failed to connect to peer %s: %v", addr, err)
		}
	}
	time.Sleep(100 * time.Millisecond)

	if len(node.Peers) != len(peerAddrs) {
		t.Errorf("Expected %d peers, got %d", len(peerAddrs), len(node.Peers))
	}

	// Clean up
	for _, peer := range node.Peers {
		peer.Conn.Close()
	}
	for _, p := range peers {
		for _, peer := range p.Peers {
			peer.Conn.Close()
		}
	}
}

// Note: TestBroadcastBlockToPeers and TestHandleInvalidMessage would require more advanced mocking or integration testing
// because the current implementation does not expose hooks for message receipt. These are left as TODOs for future work.
