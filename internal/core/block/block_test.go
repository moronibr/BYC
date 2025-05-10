package block

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/youngchain/internal/core/common"
)

func TestNewBlock(t *testing.T) {
	prevHash := []byte("previous_hash")
	difficulty := uint64(1)

	block := NewBlock(prevHash, difficulty)

	assert.NotNil(t, block)
	assert.Equal(t, uint32(1), block.Header.Version)
	assert.Equal(t, prevHash, block.Header.PrevBlockHash)
	assert.Equal(t, difficulty, block.Header.Difficulty)
	assert.Empty(t, block.Transactions)
}

func TestBlock_AddTransaction(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1)
	tx := common.NewTransaction(
		[]byte("from"),
		[]byte("to"),
		100,
		[]byte("data"),
	)

	block.AddTransaction(tx)

	assert.Len(t, block.Transactions, 1)
	assert.Equal(t, tx, block.Transactions[0])
}

func TestBlock_CalculateMerkleRoot(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1)

	// Test empty transactions
	root := block.CalculateMerkleRoot()
	assert.NotNil(t, root)

	// Test with one transaction
	tx := common.NewTransaction(
		[]byte("from"),
		[]byte("to"),
		100,
		[]byte("data"),
	)
	block.AddTransaction(tx)
	root = block.CalculateMerkleRoot()
	assert.NotNil(t, root)

	// Test with multiple transactions
	tx2 := common.NewTransaction(
		[]byte("from2"),
		[]byte("to2"),
		200,
		[]byte("data2"),
	)
	block.AddTransaction(tx2)
	root = block.CalculateMerkleRoot()
	assert.NotNil(t, root)
}

func TestBlock_CalculateHash(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1)
	hash := block.CalculateHash()

	assert.NotNil(t, hash)
	assert.Len(t, hash, 32) // SHA-256 hash length
}

func TestBlock_Copy(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1)
	tx := common.NewTransaction(
		[]byte("from"),
		[]byte("to"),
		100,
		[]byte("data"),
	)
	block.AddTransaction(tx)

	blockCopy := block.Copy()

	assert.NotNil(t, blockCopy)
	assert.Equal(t, block.Header.Version, blockCopy.Header.Version)
	assert.Equal(t, block.Header.PrevBlockHash, blockCopy.Header.PrevBlockHash)
	assert.Equal(t, block.Header.Difficulty, blockCopy.Header.Difficulty)
	assert.Len(t, blockCopy.Transactions, 1)

	// Verify deep copy
	underlyingTx := blockCopy.Transactions[0].GetTransaction()
	underlyingTx.Outputs[0].Value = 200
	assert.NotEqual(t, block.Transactions[0].Amount(), blockCopy.Transactions[0].Amount())
}
