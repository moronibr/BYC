package tests

import (
	"testing"
	"time"

	"github.com/byc/internal/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockchain(t *testing.T) {
	bc := blockchain.NewBlockchain()
	assert.NotNil(t, bc)
	assert.NotEmpty(t, bc.GoldenBlocks)
	assert.NotEmpty(t, bc.SilverBlocks)
	assert.Equal(t, 1, len(bc.GoldenBlocks))
	assert.Equal(t, 1, len(bc.SilverBlocks))
}

func TestAddBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// Create a test block
	block := blockchain.Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		BlockType:    blockchain.GoldenBlock,
		Difficulty:   4,
	}

	err := bc.AddBlock(block)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bc.GoldenBlocks))
}

func TestGetBalance(t *testing.T) {
	bc := blockchain.NewBlockchain()
	address := "test_address"

	// Test initial balance
	balance := bc.GetBalance(address, blockchain.Leah)
	assert.Equal(t, 0.0, balance)

	// Create and add a transaction
	tx := blockchain.Transaction{
		ID: []byte("test_tx"),
		Outputs: []blockchain.TxOutput{
			{
				Value:         10.0,
				CoinType:      blockchain.Leah,
				PublicKeyHash: []byte(address),
				Address:       address,
			},
		},
		Timestamp: time.Now(),
		BlockType: blockchain.GoldenBlock,
	}

	block := blockchain.Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{tx},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		BlockType:    blockchain.GoldenBlock,
		Difficulty:   4,
	}

	err := bc.AddBlock(block)
	assert.NoError(t, err)

	// Test updated balance
	balance = bc.GetBalance(address, blockchain.Leah)
	assert.Equal(t, 10.0, balance)
}

func TestMineBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// Create test transactions
	txs := []blockchain.Transaction{
		{
			ID: []byte("test_tx1"),
			Outputs: []blockchain.TxOutput{
				{
					Value:         5.0,
					CoinType:      blockchain.Leah,
					PublicKeyHash: []byte("test_address1"),
					Address:       "test_address1",
				},
			},
			Timestamp: time.Now(),
			BlockType: blockchain.GoldenBlock,
		},
	}

	// Mine a new block
	block, err := bc.MineBlock(txs, blockchain.GoldenBlock, blockchain.Leah)
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, blockchain.GoldenBlock, block.BlockType)
	assert.Equal(t, 1, len(block.Transactions))
}

func TestGetBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// Get genesis block
	block, err := bc.GetBlock(bc.GoldenBlocks[0].Hash)
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, blockchain.GoldenBlock, block.BlockType)

	// Test non-existent block
	nonExistentHash := []byte("non_existent")
	block, err = bc.GetBlock(nonExistentHash)
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestGetTransaction(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// Create and add a transaction
	tx := blockchain.Transaction{
		ID: []byte("test_tx"),
		Outputs: []blockchain.TxOutput{
			{
				Value:         10.0,
				CoinType:      blockchain.Leah,
				PublicKeyHash: []byte("test_address"),
				Address:       "test_address",
			},
		},
		Timestamp: time.Now(),
		BlockType: blockchain.GoldenBlock,
	}

	block := blockchain.Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{tx},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		BlockType:    blockchain.GoldenBlock,
		Difficulty:   4,
	}

	err := bc.AddBlock(block)
	assert.NoError(t, err)

	// Get the transaction
	retrievedTx, err := bc.GetTransaction(tx.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedTx)
	assert.Equal(t, tx.ID, retrievedTx.ID)

	// Test non-existent transaction
	nonExistentID := []byte("non_existent")
	retrievedTx, err = bc.GetTransaction(nonExistentID)
	assert.Error(t, err)
	assert.Nil(t, retrievedTx)
}
