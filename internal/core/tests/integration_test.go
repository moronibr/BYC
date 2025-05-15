package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/consensus"
	"github.com/youngchain/internal/core/mining"
	"github.com/youngchain/internal/core/network"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/wallet"
)

// TestBlockchainIntegration tests the integration of blockchain components
func TestBlockchainIntegration(t *testing.T) {
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
	nodeConfig := network.Config{
		ListenPort: 8333,
		MaxPeers:   10,
	}
	node := network.NewNode(nodeConfig)

	// Test block creation and mining
	t.Run("Block Creation and Mining", func(t *testing.T) {
		// Create a new block
		newBlock := block.NewBlock(blockchain.GetBestBlock().Header.Hash, time.Now().Unix())
		assert.NotNil(t, newBlock)

		// Add transactions to the block
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)
		newBlock.AddTransaction(tx)

		// Mine the block
		minedBlock, err := miner.MineBlock(newBlock)
		assert.NoError(t, err)
		assert.NotNil(t, minedBlock)
		assert.True(t, minedBlock.Header.Hash != nil)

		// Add the block to the blockchain
		err = blockchain.AddBlock(minedBlock)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), blockchain.GetBlockCount())
	})

	// Test transaction handling
	t.Run("Transaction Handling", func(t *testing.T) {
		// Create a new transaction
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)

		// Add transaction to the pool
		err := txPool.AddTransaction(tx)
		assert.NoError(t, err)
		assert.Equal(t, 1, txPool.Size())

		// Get transaction from pool
		poolTx := txPool.GetTransaction(tx.Hash)
		assert.NotNil(t, poolTx)
		assert.Equal(t, tx.Hash, poolTx.Hash)
	})

	// Test network communication
	t.Run("Network Communication", func(t *testing.T) {
		// Start the node
		err := node.Start()
		assert.NoError(t, err)
		defer node.Stop()

		// Create a test message
		msg := network.NewMessage(network.MsgBlock, []byte("test"))
		assert.NotNil(t, msg)

		// Broadcast the message
		err = node.Broadcast(msg)
		assert.NoError(t, err)
	})

	// Test consensus validation
	t.Run("Consensus Validation", func(t *testing.T) {
		// Create a new block
		newBlock := block.NewBlock(blockchain.GetBestBlock().Header.Hash, time.Now().Unix())
		assert.NotNil(t, newBlock)

		// Validate the block
		valid, err := consensus.ValidateBlock(newBlock)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test blockchain operations
	t.Run("Blockchain Operations", func(t *testing.T) {
		// Get best block
		bestBlock := blockchain.GetBestBlock()
		assert.NotNil(t, bestBlock)

		// Get block by hash
		blockByHash := blockchain.GetBlockByHash(bestBlock.Header.Hash)
		assert.NotNil(t, blockByHash)
		assert.Equal(t, bestBlock.Header.Hash, blockByHash.Header.Hash)

		// Get block by height
		blockByHeight := blockchain.GetBlockByHeight(0)
		assert.NotNil(t, blockByHeight)
		assert.Equal(t, uint64(0), blockByHeight.Header.Height)

		// Get block count
		count := blockchain.GetBlockCount()
		assert.Equal(t, uint64(1), count)
	})

	// Test wallet operations
	t.Run("Wallet Operations", func(t *testing.T) {
		// Create a new wallet
		w, err := wallet.NewWallet("golden")
		assert.NoError(t, err)
		assert.NotNil(t, w)

		// Save wallet
		err = w.SaveWallet("test_wallet.dat")
		assert.NoError(t, err)

		// Load wallet
		loadedWallet, err := wallet.LoadWallet("test_wallet.dat")
		assert.NoError(t, err)
		assert.NotNil(t, loadedWallet)
		assert.Equal(t, w.Address, loadedWallet.Address)

		// Get balance
		balance, err := loadedWallet.GetBalance(txPool)
		assert.NoError(t, err)
		assert.NotNil(t, balance)
	})

	// Test transaction validation
	t.Run("Transaction Validation", func(t *testing.T) {
		// Create a new transaction
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)

		// Validate transaction
		valid, err := tx.Validate()
		assert.NoError(t, err)
		assert.True(t, valid)

		// Calculate transaction hash
		tx.CalculateHash()
		assert.NotNil(t, tx.Hash)

		// Get transaction size
		size := tx.Size()
		assert.Greater(t, size, 0)

		// Get transaction weight
		weight := tx.Weight()
		assert.Greater(t, weight, 0)
	})

	// Test block validation
	t.Run("Block Validation", func(t *testing.T) {
		// Create a new block
		newBlock := block.NewBlock(blockchain.GetBestBlock().Header.Hash, time.Now().Unix())
		assert.NotNil(t, newBlock)

		// Add transactions to the block
		tx := transaction.NewTransaction(1, "golden")
		tx.AddOutput("BYC1...", 1000000)
		newBlock.AddTransaction(tx)

		// Validate block
		valid, err := newBlock.Validate()
		assert.NoError(t, err)
		assert.True(t, valid)

		// Calculate block hash
		newBlock.CalculateHash()
		assert.NotNil(t, newBlock.Header.Hash)

		// Get block size
		size := newBlock.Size()
		assert.Greater(t, size, 0)
	})

	// Test network node operations
	t.Run("Network Node Operations", func(t *testing.T) {
		// Start the node
		err := node.Start()
		assert.NoError(t, err)
		defer node.Stop()

		// Get node status
		status := node.GetStatus()
		assert.NotNil(t, status)
		assert.True(t, status.IsRunning)

		// Get connected peers
		peers := node.GetPeers()
		assert.NotNil(t, peers)

		// Get node address
		addr := node.GetAddress()
		assert.NotEmpty(t, addr)
	})

	// Test consensus operations
	t.Run("Consensus Operations", func(t *testing.T) {
		// Get current difficulty
		difficulty := consensus.GetDifficulty()
		assert.Greater(t, difficulty, 0)

		// Get target bits
		targetBits := consensus.GetTargetBits()
		assert.Equal(t, consensusConfig.TargetBits, targetBits)

		// Get block interval
		interval := consensus.GetBlockInterval()
		assert.Equal(t, consensusConfig.BlockInterval, interval)
	})

	// Test mining operations
	t.Run("Mining Operations", func(t *testing.T) {
		// Get mining status
		status := miner.GetStatus()
		assert.NotNil(t, status)

		// Start mining
		err := miner.Start()
		assert.NoError(t, err)
		defer miner.Stop()

		// Check if mining is running
		assert.True(t, miner.IsRunning())

		// Get mining rate
		rate := miner.GetHashRate()
		assert.Greater(t, rate, 0)
	})
}
