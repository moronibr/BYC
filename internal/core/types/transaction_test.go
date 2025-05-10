package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	from := []byte("from_address")
	to := []byte("to_address")
	amount := uint64(100)
	data := []byte("transaction_data")

	tx := NewTransaction(from, to, amount, data)

	assert.NotNil(t, tx)
	assert.Equal(t, uint32(1), tx.Version)
	assert.Equal(t, string(from), tx.Inputs[0].Address)
	assert.Equal(t, string(to), tx.Outputs[0].Address)
	assert.Equal(t, amount, tx.Outputs[0].Value)
	assert.Equal(t, data, tx.Data)
	assert.NotNil(t, tx.Hash)
}

func TestTransaction_CalculateHash(t *testing.T) {
	tx := NewTransaction(
		[]byte("from"),
		[]byte("to"),
		100,
		[]byte("data"),
	)

	hash := tx.CalculateHash()
	assert.NotNil(t, hash)
	assert.Len(t, hash, 32) // SHA-256 hash length

	// Hash should be deterministic
	hash2 := tx.CalculateHash()
	assert.Equal(t, hash, hash2)

	// Hash should change when transaction data changes
	tx.Outputs[0].Value = 200
	hash3 := tx.CalculateHash()
	assert.NotEqual(t, hash, hash3)
}

func TestTransaction_Copy(t *testing.T) {
	from := []byte("from")
	to := []byte("to")
	amount := uint64(100)
	data := []byte("data")

	tx := NewTransaction(from, to, amount, data)
	txCopy := tx.Copy()

	assert.NotNil(t, txCopy)
	assert.Equal(t, tx.Version, txCopy.Version)
	assert.Equal(t, tx.Inputs[0].Address, txCopy.Inputs[0].Address)
	assert.Equal(t, tx.Outputs[0].Address, txCopy.Outputs[0].Address)
	assert.Equal(t, tx.Outputs[0].Value, txCopy.Outputs[0].Value)
	assert.Equal(t, tx.Data, txCopy.Data)
	assert.Equal(t, tx.Hash, txCopy.Hash)

	// Verify deep copy
	txCopy.Outputs[0].Value = 200
	assert.NotEqual(t, tx.Outputs[0].Value, txCopy.Outputs[0].Value)
}

func TestTransaction_Size(t *testing.T) {
	tx := NewTransaction(
		[]byte("from"),
		[]byte("to"),
		100,
		[]byte("data"),
	)

	size := tx.Size()
	assert.Greater(t, size, 0)

	// Size should increase with larger data
	tx.Data = []byte("larger_data")
	newSize := tx.Size()
	assert.Greater(t, newSize, size)
}
