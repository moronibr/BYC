package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/wallet"
)

// TestCrossChainTransfer tests cross-chain transfer functionality
func TestCrossChainTransfer(t *testing.T) {
	// Initialize components
	goldenChain := block.NewBlockchain()
	silverChain := block.NewBlockchain()
	txPool := transaction.NewTxPool(1000)

	// Create wallets for both chains
	goldenWallet, err := wallet.NewWallet(coin.Golden)
	assert.NoError(t, err)
	silverWallet, err := wallet.NewWallet(coin.Silver)
	assert.NoError(t, err)

	// Test cases
	t.Run("Golden to Silver Transfer", func(t *testing.T) {
		// Create transfer transaction
		tx := transaction.NewTransaction(1, coin.Golden)
		tx.AddInput(goldenWallet.Address, 1000000)
		tx.AddOutput(silverWallet.Address, 500000) // 1:2 conversion rate
		tx.AddOutput(goldenWallet.Address, 500000) // Change

		// Sign transaction
		err := goldenWallet.SignTransaction(tx)
		assert.NoError(t, err)

		// Validate transaction
		valid, err := tx.Validate()
		assert.NoError(t, err)
		assert.True(t, valid)

		// Add transaction to pool
		err = txPool.AddTransaction(tx)
		assert.NoError(t, err)

		// Create and mine block
		newBlock := block.NewBlock(goldenChain.GetBestBlock().Header.Hash, uint64(time.Now().Unix()))
		newBlock.AddTransaction(tx)
		goldenChain.AddBlock(newBlock)

		// Verify balances
		goldenBalance, err := goldenWallet.GetBalance(txPool)
		assert.NoError(t, err)
		assert.Equal(t, uint64(500000), goldenBalance)

		silverBalance, err := silverWallet.GetBalance(txPool)
		assert.NoError(t, err)
		assert.Equal(t, uint64(500000), silverBalance)
	})

	t.Run("Silver to Golden Transfer", func(t *testing.T) {
		// Create transfer transaction
		tx := transaction.NewTransaction(1, coin.Silver)
		tx.AddInput(silverWallet.Address, 500000)
		tx.AddOutput(goldenWallet.Address, 250000) // 2:1 conversion rate
		tx.AddOutput(silverWallet.Address, 250000) // Change

		// Sign transaction
		err := silverWallet.SignTransaction(tx)
		assert.NoError(t, err)

		// Validate transaction
		valid, err := tx.Validate()
		assert.NoError(t, err)
		assert.True(t, valid)

		// Add transaction to pool
		err = txPool.AddTransaction(tx)
		assert.NoError(t, err)

		// Create and mine block
		newBlock := block.NewBlock(silverChain.GetBestBlock().Header.Hash, uint64(time.Now().Unix()))
		newBlock.AddTransaction(tx)
		silverChain.AddBlock(newBlock)

		// Verify balances
		goldenBalance, err := goldenWallet.GetBalance(txPool)
		assert.NoError(t, err)
		assert.Equal(t, uint64(750000), goldenBalance)

		silverBalance, err := silverWallet.GetBalance(txPool)
		assert.NoError(t, err)
		assert.Equal(t, uint64(250000), silverBalance)
	})

	t.Run("Invalid Cross-Chain Transfer", func(t *testing.T) {
		// Create invalid transfer transaction
		tx := transaction.NewTransaction(1, coin.Golden)
		tx.AddInput(goldenWallet.Address, 1000000)
		tx.AddOutput(silverWallet.Address, 2000000) // Invalid amount

		// Sign transaction
		err := goldenWallet.SignTransaction(tx)
		assert.NoError(t, err)

		// Validate transaction
		valid, err := tx.Validate()
		assert.Error(t, err)
		assert.False(t, valid)
	})

	t.Run("Double Cross-Chain Transfer", func(t *testing.T) {
		// Create first transfer
		tx1 := transaction.NewTransaction(1, coin.Golden)
		tx1.AddInput(goldenWallet.Address, 1000000)
		tx1.AddOutput(silverWallet.Address, 500000)
		tx1.AddOutput(goldenWallet.Address, 500000)

		// Sign and validate first transaction
		err := goldenWallet.SignTransaction(tx1)
		assert.NoError(t, err)
		valid, err := tx1.Validate()
		assert.NoError(t, err)
		assert.True(t, valid)

		// Create second transfer with same input
		tx2 := transaction.NewTransaction(1, coin.Golden)
		tx2.AddInput(goldenWallet.Address, 1000000)
		tx2.AddOutput(silverWallet.Address, 500000)
		tx2.AddOutput(goldenWallet.Address, 500000)

		// Sign and validate second transaction
		err = goldenWallet.SignTransaction(tx2)
		assert.NoError(t, err)
		valid, err = tx2.Validate()
		assert.Error(t, err)
		assert.False(t, valid)
	})

	t.Run("Cross-Chain Transfer with Fees", func(t *testing.T) {
		// Create transfer transaction with fee
		tx := transaction.NewTransaction(1, coin.Golden)
		tx.AddInput(goldenWallet.Address, 1000000)
		tx.AddOutput(silverWallet.Address, 450000) // 1:2 conversion rate with 10% fee
		tx.AddOutput(goldenWallet.Address, 450000) // Change
		tx.Fee = 100000

		// Sign transaction
		err := goldenWallet.SignTransaction(tx)
		assert.NoError(t, err)

		// Validate transaction
		valid, err := tx.Validate()
		assert.NoError(t, err)
		assert.True(t, valid)

		// Add transaction to pool
		err = txPool.AddTransaction(tx)
		assert.NoError(t, err)

		// Create and mine block
		newBlock := block.NewBlock(goldenChain.GetBestBlock().Header.Hash, uint64(time.Now().Unix()))
		newBlock.AddTransaction(tx)
		goldenChain.AddBlock(newBlock)

		// Verify balances
		goldenBalance, err := goldenWallet.GetBalance(txPool)
		assert.NoError(t, err)
		assert.Equal(t, uint64(450000), goldenBalance)

		silverBalance, err := silverWallet.GetBalance(txPool)
		assert.NoError(t, err)
		assert.Equal(t, uint64(450000), silverBalance)
	})
}
