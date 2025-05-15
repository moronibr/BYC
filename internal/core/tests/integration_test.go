package tests

import (
	"testing"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/consensus"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/core/network"
	"github.com/youngchain/internal/core/transaction"
)

// TestBlockchainIntegration tests the integration of blockchain components
func TestBlockchainIntegration(t *testing.T) {
	// Create blockchain
	blockchain := block.NewBlockchain()
	if blockchain == nil {
		t.Fatal("Failed to create blockchain")
	}

	// Create transaction pool
	txPool := transaction.NewTxPool(1000, 0.001, nil)
	if txPool == nil {
		t.Fatal("Failed to create transaction pool")
	}

	// Create consensus
	consensusConfig := &consensus.ConsensusConfig{
		TargetBits: 24,
		MaxNonce:   1000000,
	}
	consensus := consensus.NewConsensus(consensusConfig)
	if consensus == nil {
		t.Fatal("Failed to create consensus")
	}

	// Create miner
	minerConfig := &mining.MinerConfig{
		MiningAddress: "test_address",
		TargetBits:    24,
		MaxNonce:      1000000,
	}
	miner := mining.NewMiner(minerConfig, blockchain, txPool, nil)
	if miner == nil {
		t.Fatal("Failed to create miner")
	}

	// Create network node
	nodeConfig := &network.NodeConfig{
		ListenPort:       8333,
		MaxPeers:         10,
		HandshakeTimeout: 30 * time.Second,
		PingInterval:     2 * time.Minute,
	}
	node := network.NewNode(nodeConfig)
	if node == nil {
		t.Fatal("Failed to create network node")
	}

	// Test block creation and mining
	t.Run("BlockCreationAndMining", func(t *testing.T) {
		// Create a new block
		newBlock := block.NewBlock(block.GoldenBlock, 1, []byte("prev_hash"), time.Now())
		if newBlock == nil {
			t.Fatal("Failed to create new block")
		}

		// Add transactions to block
		tx := transaction.NewTransaction(1, "Leah")
		if tx == nil {
			t.Fatal("Failed to create transaction")
		}

		if err := newBlock.AddTransaction(tx); err != nil {
			t.Fatalf("Failed to add transaction to block: %v", err)
		}

		// Mine block
		if err := miner.MineBlock(newBlock); err != nil {
			t.Fatalf("Failed to mine block: %v", err)
		}

		// Verify block
		if !newBlock.IsValid() {
			t.Fatal("Block validation failed")
		}

		// Add block to blockchain
		if err := blockchain.AddBlock(newBlock); err != nil {
			t.Fatalf("Failed to add block to blockchain: %v", err)
		}
	})

	// Test transaction handling
	t.Run("TransactionHandling", func(t *testing.T) {
		// Create transaction
		tx := transaction.NewTransaction(1, "Leah")
		if tx == nil {
			t.Fatal("Failed to create transaction")
		}

		// Add transaction to pool
		if err := txPool.AddTransaction(tx); err != nil {
			t.Fatalf("Failed to add transaction to pool: %v", err)
		}

		// Verify transaction in pool
		if txPool.GetSize() != 1 {
			t.Fatal("Transaction not added to pool")
		}

		// Get best transactions
		bestTxs := txPool.GetBest(10)
		if len(bestTxs) != 1 {
			t.Fatal("Failed to get best transactions")
		}
	})

	// Test network communication
	t.Run("NetworkCommunication", func(t *testing.T) {
		// Start node
		if err := node.Start(); err != nil {
			t.Fatalf("Failed to start node: %v", err)
		}
		defer node.Stop()

		// Create test message
		msg := &network.Message{
			Type: network.MsgVersion,
			Payload: &network.VersionPayload{
				Version:   1,
				Services:  0,
				Timestamp: time.Now(),
				AddrRecv:  "127.0.0.1:8333",
				AddrFrom:  "127.0.0.1:8334",
			},
		}

		// Broadcast message
		if err := node.BroadcastMessage(msg); err != nil {
			t.Fatalf("Failed to broadcast message: %v", err)
		}
	})

	// Test consensus
	t.Run("Consensus", func(t *testing.T) {
		// Create test block
		testBlock := block.NewBlock(block.GoldenBlock, 1, []byte("prev_hash"), time.Now())
		if testBlock == nil {
			t.Fatal("Failed to create test block")
		}

		// Validate block
		if err := consensus.ValidateBlock(testBlock); err != nil {
			t.Fatalf("Block validation failed: %v", err)
		}

		// Validate block header
		if err := consensus.ValidateBlockHeader(testBlock.Header); err != nil {
			t.Fatalf("Block header validation failed: %v", err)
		}

		// Validate block size
		if err := consensus.ValidateBlockSize(testBlock); err != nil {
			t.Fatalf("Block size validation failed: %v", err)
		}

		// Validate transactions
		if err := consensus.ValidateTransactions(testBlock.Transactions); err != nil {
			t.Fatalf("Transaction validation failed: %v", err)
		}

		// Validate proof of work
		if err := consensus.ValidateProofOfWork(testBlock); err != nil {
			t.Fatalf("Proof of work validation failed: %v", err)
		}
	})

	// Test blockchain operations
	t.Run("BlockchainOperations", func(t *testing.T) {
		// Get best block
		bestBlock := blockchain.GetBestBlock()
		if bestBlock == nil {
			t.Fatal("Failed to get best block")
		}

		// Get block by hash
		blockByHash := blockchain.GetBlockByHash(bestBlock.Header.Hash)
		if blockByHash == nil {
			t.Fatal("Failed to get block by hash")
		}

		// Get block by height
		blockByHeight := blockchain.GetBlockByHeight(bestBlock.Header.Height)
		if blockByHeight == nil {
			t.Fatal("Failed to get block by height")
		}

		// Get block count
		blockCount := blockchain.GetBlockCount()
		if blockCount != 1 {
			t.Fatalf("Expected block count 1, got %d", blockCount)
		}

		// Get blocks
		blocks := blockchain.GetBlocks(0, 10)
		if len(blocks) != 1 {
			t.Fatalf("Expected 1 block, got %d", len(blocks))
		}

		// Get blocks by type
		goldenBlocks := blockchain.GetBlocksByType(block.GoldenBlock)
		if len(goldenBlocks) != 1 {
			t.Fatalf("Expected 1 golden block, got %d", len(goldenBlocks))
		}
	})
}
