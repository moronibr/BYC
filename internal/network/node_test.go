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

	node1.Start()
	node2.Start()

	time.Sleep(100 * time.Millisecond)

	err := node1.Connect("localhost:3002")
	if err != nil {
		t.Errorf("Node.Connect failed: %v", err)
	}

	// Wait for connection to establish
	time.Sleep(100 * time.Millisecond)

	peers := getPeerAddresses(node1)
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(peers))
	}
	if peers[0] != "localhost:3002" {
		t.Errorf("Expected peer localhost:3002, got %s", peers[0])
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
