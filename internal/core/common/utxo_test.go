package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUTXOSet(t *testing.T) {
	utxoSet := NewUTXOSet()
	assert.NotNil(t, utxoSet)
	assert.NotNil(t, utxoSet.utxos)
	assert.Empty(t, utxoSet.utxos)
}

func TestUTXOSet_GetUTXO(t *testing.T) {
	utxoSet := NewUTXOSet()
	txHash := []byte("tx_hash")
	outIndex := uint32(0)

	// Test non-existent UTXO
	utxo, exists := utxoSet.GetUTXO(txHash, outIndex)
	assert.Nil(t, utxo)
	assert.False(t, exists)

	// Add UTXO
	utxo = &UTXO{
		TxHash:      txHash,
		OutIndex:    outIndex,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   time.Now(),
	}
	utxoSet.utxos[utxoKey(txHash, outIndex)] = utxo

	// Test existing UTXO
	retrieved, exists := utxoSet.GetUTXO(txHash, outIndex)
	assert.NotNil(t, retrieved)
	assert.True(t, exists)
	assert.Equal(t, utxo, retrieved)
}

func TestUTXOKey(t *testing.T) {
	txHash := []byte("tx_hash")
	outIndex := uint32(0)

	key := utxoKey(txHash, outIndex)
	assert.Equal(t, "tx_hash:0", key)

	// Test with different index
	outIndex = 1
	key = utxoKey(txHash, outIndex)
	assert.Equal(t, "tx_hash:1", key)
}

func TestUTXO_Spend(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   createdAt,
	}

	// Test spending
	utxo.IsSpent = true
	utxo.SpentAt = time.Now()

	assert.True(t, utxo.IsSpent)
	assert.NotZero(t, utxo.SpentAt)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, uint64(100), utxo.Amount)
	assert.Equal(t, []byte("script"), utxo.ScriptPub)
	assert.Equal(t, uint64(1), utxo.BlockHeight)
	assert.False(t, utxo.IsCoinbase)
	assert.False(t, utxo.IsSegWit)
	assert.True(t, utxo.IsConfirmed)
	assert.Equal(t, createdAt, utxo.CreatedAt)
}

func TestUTXO_Confirm(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: false,
		CreatedAt:   createdAt,
	}

	// Test confirmation
	utxo.IsConfirmed = true

	assert.True(t, utxo.IsConfirmed)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, uint64(100), utxo.Amount)
	assert.Equal(t, []byte("script"), utxo.ScriptPub)
	assert.Equal(t, uint64(1), utxo.BlockHeight)
	assert.False(t, utxo.IsCoinbase)
	assert.False(t, utxo.IsSegWit)
	assert.False(t, utxo.IsSpent)
	assert.Equal(t, createdAt, utxo.CreatedAt)
}

func TestUTXO_Coinbase(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   createdAt,
	}

	// Test coinbase
	utxo.IsCoinbase = true

	assert.True(t, utxo.IsCoinbase)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, uint64(100), utxo.Amount)
	assert.Equal(t, []byte("script"), utxo.ScriptPub)
	assert.Equal(t, uint64(1), utxo.BlockHeight)
	assert.False(t, utxo.IsSegWit)
	assert.False(t, utxo.IsSpent)
	assert.True(t, utxo.IsConfirmed)
	assert.Equal(t, createdAt, utxo.CreatedAt)
}

func TestUTXO_SegWit(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   createdAt,
	}

	// Test SegWit
	utxo.IsSegWit = true

	assert.True(t, utxo.IsSegWit)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, uint64(100), utxo.Amount)
	assert.Equal(t, []byte("script"), utxo.ScriptPub)
	assert.Equal(t, uint64(1), utxo.BlockHeight)
	assert.False(t, utxo.IsCoinbase)
	assert.False(t, utxo.IsSpent)
	assert.True(t, utxo.IsConfirmed)
	assert.Equal(t, createdAt, utxo.CreatedAt)
}

func TestUTXO_BlockHeight(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   createdAt,
	}

	// Test block height
	newHeight := uint64(2)
	utxo.BlockHeight = newHeight

	assert.Equal(t, newHeight, utxo.BlockHeight)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, uint64(100), utxo.Amount)
	assert.Equal(t, []byte("script"), utxo.ScriptPub)
	assert.False(t, utxo.IsCoinbase)
	assert.False(t, utxo.IsSegWit)
	assert.False(t, utxo.IsSpent)
	assert.True(t, utxo.IsConfirmed)
	assert.Equal(t, createdAt, utxo.CreatedAt)
}

func TestUTXO_Amount(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   createdAt,
	}

	// Test amount
	newAmount := uint64(200)
	utxo.Amount = newAmount

	assert.Equal(t, newAmount, utxo.Amount)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, []byte("script"), utxo.ScriptPub)
	assert.Equal(t, uint64(1), utxo.BlockHeight)
	assert.False(t, utxo.IsCoinbase)
	assert.False(t, utxo.IsSegWit)
	assert.False(t, utxo.IsSpent)
	assert.True(t, utxo.IsConfirmed)
	assert.Equal(t, createdAt, utxo.CreatedAt)
}

func TestUTXO_ScriptPub(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   createdAt,
	}

	// Test script pub key
	newScript := []byte("new_script")
	utxo.ScriptPub = newScript

	assert.Equal(t, newScript, utxo.ScriptPub)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, uint64(100), utxo.Amount)
	assert.Equal(t, uint64(1), utxo.BlockHeight)
	assert.False(t, utxo.IsCoinbase)
	assert.False(t, utxo.IsSegWit)
	assert.False(t, utxo.IsSpent)
	assert.True(t, utxo.IsConfirmed)
	assert.Equal(t, createdAt, utxo.CreatedAt)
}

func TestUTXO_Timestamps(t *testing.T) {
	createdAt := time.Now()
	utxo := &UTXO{
		TxHash:      []byte("tx_hash"),
		OutIndex:    0,
		Amount:      100,
		ScriptPub:   []byte("script"),
		BlockHeight: 1,
		IsCoinbase:  false,
		IsSegWit:    false,
		IsSpent:     false,
		IsConfirmed: true,
		CreatedAt:   createdAt,
	}

	// Test created at
	newCreatedAt := time.Now()
	utxo.CreatedAt = newCreatedAt
	assert.Equal(t, newCreatedAt, utxo.CreatedAt)
	assert.Equal(t, []byte("tx_hash"), utxo.TxHash)
	assert.Equal(t, uint32(0), utxo.OutIndex)
	assert.Equal(t, uint64(100), utxo.Amount)
	assert.Equal(t, []byte("script"), utxo.ScriptPub)
	assert.Equal(t, uint64(1), utxo.BlockHeight)
	assert.False(t, utxo.IsCoinbase)
	assert.False(t, utxo.IsSegWit)
	assert.False(t, utxo.IsSpent)
	assert.True(t, utxo.IsConfirmed)

	// Test spent at
	newSpentAt := time.Now()
	utxo.SpentAt = newSpentAt
	assert.Equal(t, newSpentAt, utxo.SpentAt)
}
