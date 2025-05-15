package tests

import (
	"testing"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/consensus"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/core/network"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/wallet"
)

// BenchmarkBlockCreation benchmarks block creation
func BenchmarkBlockCreation(b *testing.B) {
	// Initialize components
	blockchain := block.NewBlockchain()
	txPool := transaction.NewTxPool(1000)
	consensusConfig := consensus.Config{
		TargetBits:    20,
		MaxNonce:      1000000,
		BlockInterval: 10 * time.Second,
	}
	consensus := consensus.NewConsensus(consensusConfig)
	minerConfig := mining.Config{
		Blockchain: blockchain,
		TxPool:     txPool,
		Consensus:  consensus,
	}
	miner := mining.NewMiner(minerConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new block
		newBlock := block.NewBlock(blockchain.GetBestBlock().Header.Hash, uint64(time.Now().Unix()))

		// Add transactions to the block
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)
		newBlock.AddTransaction(tx)

		// Mine the block
		minedBlock, _ := miner.MineBlock(newBlock)

		// Add the block to the blockchain
		blockchain.AddBlock(minedBlock)
	}
}

// BenchmarkTransactionProcessing benchmarks transaction processing
func BenchmarkTransactionProcessing(b *testing.B) {
	// Initialize components
	txPool := transaction.NewTxPool(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new transaction
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)

		// Add transaction to the pool
		txPool.AddTransaction(tx)

		// Get transaction from pool
		txPool.GetTransaction(tx.Hash)
	}
}

// BenchmarkNetworkCommunication benchmarks network communication
func BenchmarkNetworkCommunication(b *testing.B) {
	// Initialize components
	nodeConfig := network.Config{
		ListenPort: 8333,
		MaxPeers:   10,
	}
	node, _ := network.NewNode(nodeConfig)
	node.Start()
	defer node.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a test message
		msg := network.NewMessage(network.MsgBlock, []byte("test"))

		// Broadcast the message
		node.Broadcast(msg)
	}
}

// BenchmarkConsensus benchmarks consensus operations
func BenchmarkConsensus(b *testing.B) {
	// Initialize components
	blockchain := block.NewBlockchain()
	consensusConfig := consensus.Config{
		TargetBits:    20,
		MaxNonce:      1000000,
		BlockInterval: 10 * time.Second,
	}
	consensus := consensus.NewConsensus(consensusConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new block
		newBlock := block.NewBlock(blockchain.GetBestBlock().Header.Hash, uint64(time.Now().Unix()))

		// Validate the block
		consensus.ValidateBlock(newBlock)
	}
}

// BenchmarkBlockchainOperations benchmarks blockchain operations
func BenchmarkBlockchainOperations(b *testing.B) {
	// Initialize components
	blockchain := block.NewBlockchain()

	// Add some blocks for testing
	for i := 0; i < 100; i++ {
		newBlock := block.NewBlock(blockchain.GetBestBlock().Header.Hash, uint64(time.Now().Unix()))
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)
		newBlock.AddTransaction(tx)
		blockchain.AddBlock(newBlock)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get best block
		blockchain.GetBestBlock()

		// Get block by hash
		blockchain.GetBlockByHash(blockchain.GetBestBlock().Header.Hash)

		// Get block by height
		blockchain.GetBlockByHeight(uint64(i % 100))

		// Get block count
		blockchain.GetBlockCount()
	}
}

// BenchmarkWalletOperations benchmarks wallet operations
func BenchmarkWalletOperations(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new wallet
		wallet.NewWallet("golden")

		// Save wallet
		w, _ := wallet.NewWallet("golden")
		w.SaveWallet("test_wallet.dat")

		// Load wallet
		wallet.LoadWallet("test_wallet.dat")
	}
}

// BenchmarkTransactionValidation benchmarks transaction validation
func BenchmarkTransactionValidation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new transaction
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)

		// Validate transaction
		tx.Validate()

		// Calculate transaction hash
		tx.CalculateHash()

		// Get transaction size
		tx.Size()

		// Get transaction weight
		tx.Weight()
	}
}

// BenchmarkBlockValidation benchmarks block validation
func BenchmarkBlockValidation(b *testing.B) {
	// Initialize components
	blockchain := block.NewBlockchain()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new block
		newBlock := block.NewBlock(blockchain.GetBestBlock().Header.Hash, uint64(time.Now().Unix()))

		// Add transactions to the block
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)
		newBlock.AddTransaction(tx)

		// Validate block
		newBlock.Validate()

		// Calculate block hash
		newBlock.CalculateHash()

		// Get block size
		newBlock.Size()
	}
}

// BenchmarkNetworkNodeOperations benchmarks network node operations
func BenchmarkNetworkNodeOperations(b *testing.B) {
	// Initialize components
	nodeConfig := network.Config{
		ListenPort: 8333,
		MaxPeers:   10,
	}
	node, _ := network.NewNode(nodeConfig)
	node.Start()
	defer node.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get node status
		node.GetStatus()

		// Get connected peers
		node.GetPeers()

		// Get node address
		node.GetAddress()
	}
}

// BenchmarkConsensusOperations benchmarks consensus operations
func BenchmarkConsensusOperations(b *testing.B) {
	// Initialize components
	consensusConfig := consensus.Config{
		TargetBits:    20,
		MaxNonce:      1000000,
		BlockInterval: 10 * time.Second,
	}
	consensus := consensus.NewConsensus(consensusConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get current difficulty
		consensus.GetDifficulty()

		// Get target bits
		consensus.GetTargetBits()

		// Get block interval
		consensus.GetBlockInterval()
	}
}

// BenchmarkMiningOperations benchmarks mining operations
func BenchmarkMiningOperations(b *testing.B) {
	// Initialize components
	blockchain := block.NewBlockchain()
	txPool := transaction.NewTxPool(1000)
	consensusConfig := consensus.Config{
		TargetBits:    20,
		MaxNonce:      1000000,
		BlockInterval: 10 * time.Second,
	}
	consensus := consensus.NewConsensus(consensusConfig)
	minerConfig := mining.Config{
		Blockchain: blockchain,
		TxPool:     txPool,
		Consensus:  consensus,
	}
	miner := mining.NewMiner(minerConfig)
	miner.Start()
	defer miner.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get mining status
		miner.GetStatus()

		// Check if mining is running
		miner.IsRunning()

		// Get mining rate
		miner.GetHashRate()
	}
}
