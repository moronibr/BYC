package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/youngchain/internal/core/coin"
)

func TestNewTransaction(t *testing.T) {
	from := []byte("from_address")
	to := []byte("to_address")
	amount := uint64(100)
	data := []byte("transaction_data")

	tx := NewTransaction(from, to, amount, data)

	assert.NotNil(t, tx)
	assert.Equal(t, uint32(1), tx.Version)
	assert.NotZero(t, tx.Timestamp)
	assert.Equal(t, data, tx.Data)
	assert.Len(t, tx.Inputs, 1)
	assert.Len(t, tx.Outputs, 1)
	assert.Equal(t, string(from), tx.Inputs[0].Address)
	assert.Equal(t, string(to), tx.Outputs[0].Address)
	assert.Equal(t, amount, tx.Outputs[0].Value)
}

func TestTransaction_CalculateHash(t *testing.T) {
	tx := NewTransaction([]byte("from"), []byte("to"), 100, []byte("data"))

	// Calculate hash
	tx.CalculateHash()

	assert.NotNil(t, tx.Hash)
	assert.Len(t, tx.Hash, 32) // SHA-256 hash length
}

func TestTransaction_AddInput(t *testing.T) {
	tx := NewTransaction(nil, nil, 0, nil)
	input := &TxInput{
		PreviousTxHash:  []byte("prev_hash"),
		PreviousTxIndex: 0,
		ScriptSig:       []byte("script"),
		Sequence:        1,
		Address:         "address",
	}

	tx.AddInput(input)

	assert.Len(t, tx.Inputs, 1)
	assert.Equal(t, input, tx.Inputs[0])
}

func TestTransaction_AddOutput(t *testing.T) {
	tx := NewTransaction(nil, nil, 0, nil)
	output := &TxOutput{
		Value:        100,
		ScriptPubKey: []byte("script"),
		Address:      "address",
	}

	tx.AddOutput(output)

	assert.Len(t, tx.Outputs, 1)
	assert.Equal(t, output, tx.Outputs[0])
}

func TestTransaction_EncodeDecode(t *testing.T) {
	original := NewTransaction([]byte("from"), []byte("to"), 100, []byte("data"))
	original.Fee = 10
	original.CoinType = coin.Leah
	original.LockTime = 12345

	// Encode
	encoded, err := original.Encode()
	assert.NoError(t, err)
	assert.NotNil(t, encoded)

	// Decode
	var decoded Transaction
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	// Compare
	assert.Equal(t, original.Version, decoded.Version)
	assert.Equal(t, original.Fee, decoded.Fee)
	assert.Equal(t, original.CoinType, decoded.CoinType)
	assert.Equal(t, original.LockTime, decoded.LockTime)
	assert.Equal(t, original.Data, decoded.Data)
	assert.Len(t, decoded.Inputs, len(original.Inputs))
	assert.Len(t, decoded.Outputs, len(original.Outputs))
}

func TestTransaction_Copy(t *testing.T) {
	original := NewTransaction([]byte("from"), []byte("to"), 100, []byte("data"))
	original.Fee = 10
	original.CoinType = coin.Leah
	original.LockTime = 12345

	copy := original.Copy()

	assert.NotEqual(t, original, copy)
	assert.Equal(t, original.Version, copy.Version)
	assert.Equal(t, original.Fee, copy.Fee)
	assert.Equal(t, original.CoinType, copy.CoinType)
	assert.Equal(t, original.LockTime, copy.LockTime)
	assert.True(t, bytes.Equal(original.Data, copy.Data))
	assert.Len(t, copy.Inputs, len(original.Inputs))
	assert.Len(t, copy.Outputs, len(original.Outputs))
}

func TestTransaction_Size(t *testing.T) {
	tx := NewTransaction([]byte("from"), []byte("to"), 100, []byte("data"))

	size := tx.Size()
	assert.Greater(t, size, 0)
}

func TestTransaction_SerializeDeserialize(t *testing.T) {
	original := NewTransaction([]byte("from"), []byte("to"), 100, []byte("data"))
	original.Fee = 10
	original.CoinType = coin.Leah
	original.LockTime = 12345

	// Serialize
	serialized := original.Serialize()
	assert.NotNil(t, serialized)

	// Deserialize
	var deserialized Transaction
	err := deserialized.Deserialize(bytes.NewBuffer(serialized))
	assert.NoError(t, err)

	// Compare
	assert.Equal(t, original.Version, deserialized.Version)
	assert.Equal(t, original.Fee, deserialized.Fee)
	assert.Equal(t, original.CoinType, deserialized.CoinType)
	assert.Equal(t, original.LockTime, deserialized.LockTime)
	assert.True(t, bytes.Equal(original.Data, deserialized.Data))
	assert.Len(t, deserialized.Inputs, len(original.Inputs))
	assert.Len(t, deserialized.Outputs, len(original.Outputs))
}

func TestTransaction_IsCoinbase(t *testing.T) {
	// Regular transaction
	tx := NewTransaction([]byte("from"), []byte("to"), 100, []byte("data"))
	assert.False(t, tx.IsCoinbase())

	// Coinbase transaction (no inputs)
	tx = NewTransaction(nil, []byte("to"), 100, []byte("data"))
	assert.True(t, tx.IsCoinbase())
}

func TestTransaction_MarshalUnmarshalJSON(t *testing.T) {
	original := NewTransaction([]byte("from"), []byte("to"), 100, []byte("data"))
	original.Fee = 10
	original.CoinType = coin.Leah
	original.LockTime = 12345

	// Marshal to JSON
	jsonData, err := original.MarshalJSON()
	assert.NoError(t, err)
	assert.NotNil(t, jsonData)

	// Unmarshal from JSON
	var unmarshaled Transaction
	err = unmarshaled.UnmarshalJSON(jsonData)
	assert.NoError(t, err)

	// Compare
	assert.Equal(t, original.Version, unmarshaled.Version)
	assert.Equal(t, original.Fee, unmarshaled.Fee)
	assert.Equal(t, original.CoinType, unmarshaled.CoinType)
	assert.Equal(t, original.LockTime, unmarshaled.LockTime)
	assert.True(t, bytes.Equal(original.Data, unmarshaled.Data))
	assert.Len(t, unmarshaled.Inputs, len(original.Inputs))
	assert.Len(t, unmarshaled.Outputs, len(original.Outputs))
}
