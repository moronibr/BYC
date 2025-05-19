package tests

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/network"
	"github.com/stretchr/testify/assert"
)

func setupTestNetwork(t *testing.T) (*network.Node, string) {
	// Create a random port for testing
	listener, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Create test node
	node, err := network.NewNode(&network.Config{
		Address:        "localhost:" + string(port),
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	})
	assert.NoError(t, err)
	assert.NotNil(t, node)

	return node, listener.Addr().String()
}

func TestNewNode(t *testing.T) {
	node, addr := setupTestNetwork(t)
	defer node.Close()

	assert.NotNil(t, node)
	assert.NotEmpty(t, addr)
	assert.NotNil(t, node.Blockchain)
	assert.NotNil(t, node.Peers)
}

func TestAddPeer(t *testing.T) {
	node, _ := setupTestNetwork(t)
	defer node.Close()

	// Create test peer
	peer := &network.Peer{
		Address:  "localhost:8000",
		LastSeen: time.Now(),
	}

	// Add peer
	node.Peers[peer.Address] = peer
	assert.Equal(t, 1, len(node.Peers))
}

func TestRemovePeer(t *testing.T) {
	node, _ := setupTestNetwork(t)
	defer node.Close()

	// Add test peer
	peer := &network.Peer{
		Address:  "localhost:8000",
		LastSeen: time.Now(),
	}
	node.Peers[peer.Address] = peer

	// Remove peer
	delete(node.Peers, peer.Address)
	assert.Equal(t, 0, len(node.Peers))
}

func TestBroadcastBlock(t *testing.T) {
	node1, addr1 := setupTestNetwork(t)
	defer node1.Close()

	node2, _ := setupTestNetwork(t)
	defer node2.Close()

	// Add node2 as peer to node1
	node1.Peers[addr1] = &network.Peer{
		Address:  addr1,
		LastSeen: time.Now(),
	}

	// Create test block
	block := &blockchain.Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
		PrevHash:     []byte("prev_hash"),
		Hash:         []byte("test_hash"),
		BlockType:    blockchain.GoldenBlock,
		Difficulty:   4,
	}

	// Broadcast block
	blockBytes, err := json.Marshal(block)
	assert.NoError(t, err)
	node1.BroadcastMessage(&network.Message{
		Type:    network.BlockMsg,
		Payload: blockBytes,
	})

	// Wait for block to be received
	time.Sleep(100 * time.Millisecond)

	// Verify block was received by node2
	receivedBlock, err := node2.Blockchain.GetBlock(block.Hash)
	assert.NoError(t, err)
	assert.NotNil(t, receivedBlock)
	assert.Equal(t, block.Hash, receivedBlock.Hash)
}

func TestBroadcastTransaction(t *testing.T) {
	node1, addr1 := setupTestNetwork(t)
	defer node1.Close()

	node2, _ := setupTestNetwork(t)
	defer node2.Close()

	// Add node2 as peer to node1
	node1.Peers[addr1] = &network.Peer{
		Address:  addr1,
		LastSeen: time.Now(),
	}

	// Create test transaction
	tx := &blockchain.Transaction{
		ID:        []byte("test_tx"),
		Inputs:    []blockchain.TxInput{},
		Outputs:   []blockchain.TxOutput{},
		Timestamp: time.Now(),
		BlockType: blockchain.GoldenBlock,
	}

	// Broadcast transaction
	txBytes, err := json.Marshal(tx)
	assert.NoError(t, err)
	node1.BroadcastMessage(&network.Message{
		Type:    network.TxMsg,
		Payload: txBytes,
	})

	// Wait for transaction to be received
	time.Sleep(100 * time.Millisecond)

	// Verify transaction was received by node2
	receivedTx, err := node2.Blockchain.GetTransaction(tx.ID)
	assert.NoError(t, err)
	assert.NotNil(t, receivedTx)
	assert.Equal(t, tx.ID, receivedTx.ID)
}

func TestSyncBlockchain(t *testing.T) {
	node1, addr1 := setupTestNetwork(t)
	defer node1.Close()

	node2, _ := setupTestNetwork(t)
	defer node2.Close()

	// Add node2 as peer to node1
	node1.Peers[addr1] = &network.Peer{
		Address:  addr1,
		LastSeen: time.Now(),
	}

	// Add a block to node1
	block := &blockchain.Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
		PrevHash:     node1.Blockchain.GoldenBlocks[0].Hash,
		Hash:         []byte("test_hash"),
		BlockType:    blockchain.GoldenBlock,
		Difficulty:   4,
	}
	err := node1.Blockchain.AddBlock(*block)
	assert.NoError(t, err)

	// Sync blockchain
	node2.BroadcastMessage(&network.Message{
		Type:    network.GetBlocksMsg,
		Payload: nil,
	})

	// Wait for sync to complete
	time.Sleep(100 * time.Millisecond)

	// Verify blockchain was synced
	assert.Equal(t, len(node1.Blockchain.GoldenBlocks), len(node2.Blockchain.GoldenBlocks))
}

func TestHandleConnection(t *testing.T) {
	node, addr := setupTestNetwork(t)
	defer node.Close()

	// Create test connection
	conn, err := net.Dial("tcp", addr)
	assert.NoError(t, err)
	defer conn.Close()

	// Send test message
	message := &network.Message{
		Type:    network.BlockMsg,
		Payload: []byte("test_data"),
	}
	err = node.BroadcastMessage(message)
	assert.NoError(t, err)

	// Wait for message to be handled
	time.Sleep(100 * time.Millisecond)
}

func TestInvalidMessages(t *testing.T) {
	node, addr := setupTestNetwork(t)
	defer node.Close()

	// Create test connection
	conn, err := net.Dial("tcp", addr)
	assert.NoError(t, err)
	defer conn.Close()

	// Send invalid message type
	message := &network.Message{
		Type:    network.MessageType(999), // Invalid message type
		Payload: []byte("test_data"),
	}
	err = node.BroadcastMessage(message)
	assert.NoError(t, err)

	// Wait for message to be handled
	time.Sleep(100 * time.Millisecond)
}
