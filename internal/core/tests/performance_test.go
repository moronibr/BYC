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

// BenchmarkBlockCreation benchmarks block creation
func BenchmarkBlockCreation(b *testing.B) {
	// Create blockchain
	blockchain := block.NewBlockchain()
	if blockchain == nil {
		b.Fatal("Failed to create blockchain")
	}

	// Create transaction pool
	txPool := transaction.NewTxPool(1000, 0.001, nil)
	if txPool == nil {
		b.Fatal("Failed to create transaction pool")
	}

	// Create consensus
	consensusConfig := &consensus.ConsensusConfig{
		TargetBits: 24,
		MaxNonce:   1000000,
	}
	consensus := consensus.NewConsensus(consensusConfig)
	if consensus == nil {
		b.Fatal("Failed to create consensus")
	}

	// Create miner
	minerConfig := &mining.MinerConfig{
		MiningAddress: "test_address",
		TargetBits:    24,
		MaxNonce:      1000000,
	}
	miner := mining.NewMiner(minerConfig, blockchain, txPool, nil)
	if miner == nil {
		b.Fatal("Failed to create miner")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a new block
		newBlock := block.NewBlock(block.GoldenBlock, 1, []byte("prev_hash"), time.Now())
		if newBlock == nil {
			b.Fatal("Failed to create new block")
		}

		// Add transactions to block
		tx := transaction.NewTransaction(1, "Leah")
		if tx == nil {
			b.Fatal("Failed to create transaction")
		}

		if err := newBlock.AddTransaction(tx); err != nil {
			b.Fatalf("Failed to add transaction to block: %v", err)
		}

		// Mine block
		if err := miner.MineBlock(newBlock); err != nil {
			b.Fatalf("Failed to mine block: %v", err)
		}

		// Verify block
		if !newBlock.IsValid() {
			b.Fatal("Block validation failed")
		}

		// Add block to blockchain
		if err := blockchain.AddBlock(newBlock); err != nil {
			b.Fatalf("Failed to add block to blockchain: %v", err)
		}
	}
}

// BenchmarkTransactionProcessing benchmarks transaction processing
func BenchmarkTransactionProcessing(b *testing.B) {
	// Create transaction pool
	txPool := transaction.NewTxPool(1000, 0.001, nil)
	if txPool == nil {
		b.Fatal("Failed to create transaction pool")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create transaction
		tx := transaction.NewTransaction(1, "Leah")
		if tx == nil {
			b.Fatal("Failed to create transaction")
		}

		// Add transaction to pool
		if err := txPool.AddTransaction(tx); err != nil {
			b.Fatalf("Failed to add transaction to pool: %v", err)
		}

		// Get best transactions
		bestTxs := txPool.GetBest(10)
		if len(bestTxs) != 1 {
			b.Fatal("Failed to get best transactions")
		}
	}
}

// BenchmarkNetworkCommunication benchmarks network communication
func BenchmarkNetworkCommunication(b *testing.B) {
	// Create network node
	nodeConfig := &network.NodeConfig{
		ListenPort:       8333,
		MaxPeers:         10,
		HandshakeTimeout: 30 * time.Second,
		PingInterval:     2 * time.Minute,
	}
	node := network.NewNode(nodeConfig)
	if node == nil {
		b.Fatal("Failed to create network node")
	}

	// Start node
	if err := node.Start(); err != nil {
		b.Fatalf("Failed to start node: %v", err)
	}
	defer node.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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
			b.Fatalf("Failed to broadcast message: %v", err)
		}
	}
}

// BenchmarkConsensus benchmarks consensus operations
func BenchmarkConsensus(b *testing.B) {
	// Create consensus
	consensusConfig := &consensus.ConsensusConfig{
		TargetBits: 24,
		MaxNonce:   1000000,
	}
	consensus := consensus.NewConsensus(consensusConfig)
	if consensus == nil {
		b.Fatal("Failed to create consensus")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create test block
		testBlock := block.NewBlock(block.GoldenBlock, 1, []byte("prev_hash"), time.Now())
		if testBlock == nil {
			b.Fatal("Failed to create test block")
		}

		// Validate block
		if err := consensus.ValidateBlock(testBlock); err != nil {
			b.Fatalf("Block validation failed: %v", err)
		}

		// Validate block header
		if err := consensus.ValidateBlockHeader(testBlock.Header); err != nil {
			b.Fatalf("Block header validation failed: %v", err)
		}

		// Validate block size
		if err := consensus.ValidateBlockSize(testBlock); err != nil {
			b.Fatalf("Block size validation failed: %v", err)
		}

		// Validate transactions
		if err := consensus.ValidateTransactions(testBlock.Transactions); err != nil {
			b.Fatalf("Transaction validation failed: %v", err)
		}

		// Validate proof of work
		if err := consensus.ValidateProofOfWork(testBlock); err != nil {
			b.Fatalf("Proof of work validation failed: %v", err)
		}
	}
}

// BenchmarkBlockchainOperations benchmarks blockchain operations
func BenchmarkBlockchainOperations(b *testing.B) {
	// Create blockchain
	blockchain := block.NewBlockchain()
	if blockchain == nil {
		b.Fatal("Failed to create blockchain")
	}

	// Add test block
	testBlock := block.NewBlock(block.GoldenBlock, 1, []byte("prev_hash"), time.Now())
	if testBlock == nil {
		b.Fatal("Failed to create test block")
	}

	if err := blockchain.AddBlock(testBlock); err != nil {
		b.Fatalf("Failed to add block to blockchain: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Get best block
		bestBlock := blockchain.GetBestBlock()
		if bestBlock == nil {
			b.Fatal("Failed to get best block")
		}

		// Get block by hash
		blockByHash := blockchain.GetBlockByHash(bestBlock.Header.Hash)
		if blockByHash == nil {
			b.Fatal("Failed to get block by hash")
		}

		// Get block by height
		blockByHeight := blockchain.GetBlockByHeight(bestBlock.Header.Height)
		if blockByHeight == nil {
			b.Fatal("Failed to get block by height")
		}

		// Get block count
		blockCount := blockchain.GetBlockCount()
		if blockCount != 1 {
			b.Fatalf("Expected block count 1, got %d", blockCount)
		}

		// Get blocks
		blocks := blockchain.GetBlocks(0, 10)
		if len(blocks) != 1 {
			b.Fatalf("Expected 1 block, got %d", len(blocks))
		}

		// Get blocks by type
		goldenBlocks := blockchain.GetBlocksByType(block.GoldenBlock)
		if len(goldenBlocks) != 1 {
			b.Fatalf("Expected 1 golden block, got %d", len(goldenBlocks))
		}
	}
}
