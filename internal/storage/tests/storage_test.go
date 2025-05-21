package tests

import (
	"os"
	"testing"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/storage"
	"github.com/stretchr/testify/assert"
)

func setupTestStorage(t *testing.T) (*storage.LevelDBStorage, string) {
	// Create temporary directory for test
	dir, err := os.MkdirTemp("", "blockchain_test")
	assert.NoError(t, err)

	// Create storage
	s, err := storage.NewLevelDBStorage(dir)
	assert.NoError(t, err)
	assert.NotNil(t, s)

	return s, dir
}

func cleanupTestStorage(s *storage.LevelDBStorage, dir string) {
	s.Close()
	os.RemoveAll(dir)
}

func TestPutAndGetBlock(t *testing.T) {
	s, dir := setupTestStorage(t)
	defer cleanupTestStorage(s, dir)

	// Create test block
	block := &blockchain.Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
		PrevHash:     []byte("prev_hash"),
		Hash:         []byte("test_hash"),
		BlockType:    blockchain.GoldenBlock,
		Difficulty:   4,
	}

	// Store block
	err := s.PutBlock(block)
	assert.NoError(t, err)

	// Retrieve block
	retrievedBlock, err := s.GetBlock(block.Hash)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedBlock)
	assert.Equal(t, block.Hash, retrievedBlock.Hash)
	assert.Equal(t, block.BlockType, retrievedBlock.BlockType)
}

func TestPutAndGetTransaction(t *testing.T) {
	s, dir := setupTestStorage(t)
	defer cleanupTestStorage(s, dir)

	// Create test transaction
	tx := &blockchain.Transaction{
		ID:        []byte("test_tx"),
		Inputs:    []blockchain.TxInput{},
		Outputs:   []blockchain.TxOutput{},
		Timestamp: time.Now(),
		BlockType: blockchain.GoldenBlock,
	}

	// Store transaction
	err := s.PutTransaction(tx)
	assert.NoError(t, err)

	// Retrieve transaction
	retrievedTx, err := s.GetTransaction(tx.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedTx)
	assert.Equal(t, tx.ID, retrievedTx.ID)
}

func TestPutAndGetUTXO(t *testing.T) {
	s, dir := setupTestStorage(t)
	defer cleanupTestStorage(s, dir)

	// Create test UTXO
	utxo := &blockchain.UTXO{
		TxID:          []byte("test_tx"),
		OutputIndex:   0,
		Amount:        10.0,
		Address:       "test_address",
		PublicKeyHash: []byte("test_pubkey"),
		CoinType:      blockchain.Leah,
	}

	// Store UTXO
	err := s.PutUTXO(utxo)
	assert.NoError(t, err)

	// Retrieve UTXO
	retrievedUTXO, err := s.GetUTXO(string(utxo.TxID), utxo.OutputIndex)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUTXO)
	assert.Equal(t, utxo.TxID, retrievedUTXO.TxID)
	assert.Equal(t, utxo.Amount, retrievedUTXO.Amount)
}

func TestGetTransactions(t *testing.T) {
	s, dir := setupTestStorage(t)
	defer cleanupTestStorage(s, dir)

	address := "test_address"

	// Create and store test transactions
	tx1 := &blockchain.Transaction{
		ID: []byte("tx1"),
		Inputs: []blockchain.TxInput{
			{
				PublicKey: []byte(address),
			},
		},
		Timestamp: time.Now(),
		BlockType: blockchain.GoldenBlock,
	}

	tx2 := &blockchain.Transaction{
		ID: []byte("tx2"),
		Outputs: []blockchain.TxOutput{
			{
				Address: address,
			},
		},
		Timestamp: time.Now(),
		BlockType: blockchain.GoldenBlock,
	}

	err := s.PutTransaction(tx1)
	assert.NoError(t, err)
	err = s.PutTransaction(tx2)
	assert.NoError(t, err)

	// Get transactions for address
	txs, err := s.GetTransactions(address)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(txs))
}

func TestGetUTXOs(t *testing.T) {
	s, dir := setupTestStorage(t)
	defer cleanupTestStorage(s, dir)

	address := "test_address"

	// Create and store test UTXOs
	utxo1 := &blockchain.UTXO{
		TxID:          []byte("tx1"),
		OutputIndex:   0,
		Amount:        10.0,
		Address:       address,
		PublicKeyHash: []byte("pubkey1"),
		CoinType:      blockchain.Leah,
	}

	utxo2 := &blockchain.UTXO{
		TxID:          []byte("tx2"),
		OutputIndex:   0,
		Amount:        20.0,
		Address:       address,
		PublicKeyHash: []byte("pubkey2"),
		CoinType:      blockchain.Shiblum,
	}

	err := s.PutUTXO(utxo1)
	assert.NoError(t, err)
	err = s.PutUTXO(utxo2)
	assert.NoError(t, err)

	// Get UTXOs for address
	utxos, err := s.GetUTXOs(address)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(utxos))
}

func TestDeleteOperations(t *testing.T) {
	s, dir := setupTestStorage(t)
	defer cleanupTestStorage(s, dir)

	// Create test data
	block := &blockchain.Block{
		Hash:      []byte("test_hash"),
		BlockType: blockchain.GoldenBlock,
	}
	tx := &blockchain.Transaction{
		ID: []byte("test_tx"),
	}
	utxo := &blockchain.UTXO{
		TxID:        []byte("test_tx"),
		OutputIndex: 0,
	}

	// Store data
	err := s.PutBlock(block)
	assert.NoError(t, err)
	err = s.PutTransaction(tx)
	assert.NoError(t, err)
	err = s.PutUTXO(utxo)
	assert.NoError(t, err)

	// Delete data
	err = s.DeleteBlock(block.Hash)
	assert.NoError(t, err)
	err = s.DeleteTransaction(tx.ID)
	assert.NoError(t, err)
	err = s.DeleteUTXO(string(utxo.TxID), utxo.OutputIndex)
	assert.NoError(t, err)

	// Verify deletion
	retrievedBlock, err := s.GetBlock(block.Hash)
	assert.NoError(t, err)
	assert.Nil(t, retrievedBlock)

	retrievedTx, err := s.GetTransaction(tx.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrievedTx)

	retrievedUTXO, err := s.GetUTXO(string(utxo.TxID), utxo.OutputIndex)
	assert.NoError(t, err)
	assert.Nil(t, retrievedUTXO)
}
