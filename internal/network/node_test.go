package network

import (
	"testing"
	"time"

	"github.com/byc/internal/blockchain"
)

func getPeerAddresses(n *Node) []string {
	peers := make([]string, 0, len(n.Peers))
	for addr := range n.Peers {
		peers = append(peers, addr)
	}
	return peers
}

func TestNewNode(t *testing.T) {
	bc := blockchain.NewBlockchain()
	// Test creating node with valid address
	node := NewNode("localhost:3000", bc)
	if node.Address != "localhost:3000" {
		t.Errorf("Expected address localhost:3000, got %s", node.Address)
	}
}

func TestNodeStartStop(t *testing.T) {
	bc := blockchain.NewBlockchain()
	node := NewNode("localhost:3000", bc)

	// Start node
	err := node.Start()
	if err != nil {
		t.Errorf("Node.Start failed: %v", err)
	}

	// Wait for a short time to ensure node has started
	time.Sleep(100 * time.Millisecond)

	// No Stop method in Node, so nothing to call here
}

func TestNodeConnect(t *testing.T) {
	bc1 := blockchain.NewBlockchain()
	bc2 := blockchain.NewBlockchain()
	node1 := NewNode("localhost:3001", bc1)
	node2 := NewNode("localhost:3002", bc2)

	// Start nodes
	err := node1.Start()
	if err != nil {
		t.Fatalf("Failed to start node1: %v", err)
	}
	err = node2.Start()
	if err != nil {
		t.Fatalf("Failed to start node2: %v", err)
	}

	// Wait for nodes to start
	time.Sleep(100 * time.Millisecond)

	// Connect node1 to node2
	err = node1.Connect("localhost:3002")
	if err != nil {
		t.Fatalf("Node.Connect failed: %v", err)
	}

	// Wait for connection to establish
	time.Sleep(100 * time.Millisecond)

	// Check peer count
	peers := getPeerAddresses(node1)
	found := false
	for _, p := range peers {
		if p == "localhost:3002" {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected peer localhost:3002, got peers: %v", peers)
	}

	// Clean up by closing connections
	for _, peer := range node1.Peers {
		peer.Conn.Close()
	}
	for _, peer := range node2.Peers {
		peer.Conn.Close()
	}
}

func TestNodeBroadcast(t *testing.T) {
	bc := blockchain.NewBlockchain()
	node := NewNode("localhost:3000", bc)

	block := blockchain.Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
		PrevHash:     []byte{},
		Hash:         []byte{},
		Nonce:        0,
		BlockType:    blockchain.GoldenBlock,
		Difficulty:   4,
	}

	node.broadcastBlock(block)
}

func TestPeerDisconnect(t *testing.T) {
	bc1 := blockchain.NewBlockchain()
	bc2 := blockchain.NewBlockchain()
	node1 := NewNode("localhost:3011", bc1)
	node2 := NewNode("localhost:3012", bc2)

	err := node1.Start()
	if err != nil {
		t.Fatalf("Failed to start node1: %v", err)
	}
	err = node2.Start()
	if err != nil {
		t.Fatalf("Failed to start node2: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	err = node1.Connect("localhost:3012")
	if err != nil {
		t.Fatalf("Node.Connect failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Disconnect peer from both sides
	for _, peer := range node1.Peers {
		peer.Conn.Close()
	}
	for _, peer := range node2.Peers {
		peer.Conn.Close()
	}

	// Wait for both peer maps to become empty (up to 1 second)
	deadline := time.Now().Add(1 * time.Second)
	for (len(node1.Peers) != 0 || len(node2.Peers) != 0) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if len(node1.Peers) != 0 {
		t.Errorf("Expected 0 peers after disconnect, got %d", len(node1.Peers))
	}
	if len(node2.Peers) != 0 {
		t.Errorf("Expected 0 peers after disconnect, got %d", len(node2.Peers))
	}
}

func TestConnectToInvalidPeer(t *testing.T) {
	bc := blockchain.NewBlockchain()
	node := NewNode("localhost:3020", bc)
	err := node.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	err = node.Connect("localhost:9999") // assuming nothing is listening here
	if err == nil {
		t.Errorf("Expected error when connecting to invalid peer, got nil")
	}
}

func TestMultiplePeerConnections(t *testing.T) {
	bc := blockchain.NewBlockchain()
	node := NewNode("localhost:3030", bc)
	err := node.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}

	peerAddrs := []string{"localhost:3031", "localhost:3032", "localhost:3033"}
	peers := []*Node{}
	for _, addr := range peerAddrs {
		peer := NewNode(addr, blockchain.NewBlockchain())
		err := peer.Start()
		if err != nil {
			t.Fatalf("Failed to start peer %s: %v", addr, err)
		}
		peers = append(peers, peer)
	}
	time.Sleep(100 * time.Millisecond)

	for _, addr := range peerAddrs {
		err := node.Connect(addr)
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
