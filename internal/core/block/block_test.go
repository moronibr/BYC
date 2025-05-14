package block

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/types"
)

func TestNewBlock(t *testing.T) {
	prevHash := []byte("previous block hash")
	height := uint64(1)

	block := NewBlock(prevHash, height)

	assert.NotNil(t, block)
	assert.Equal(t, uint32(1), block.Header.Version)
	assert.Equal(t, prevHash, block.Header.PrevBlockHash)
	assert.Equal(t, height, block.Header.Height)
	assert.Equal(t, uint32(0x1d00ffff), block.Header.Difficulty)
	assert.NotZero(t, block.Header.Timestamp)
	assert.Empty(t, block.Transactions)
}

func TestBlock_AddTransaction(t *testing.T) {
	block := NewBlock([]byte("prev"), 1)
	tx := &common.Transaction{
		Tx: types.NewTransaction([]byte("from"), []byte("to"), 100, []byte("data")),
	}

	err := block.AddTransaction(tx)
	assert.NoError(t, err)
	assert.Len(t, block.Transactions, 1)
	assert.Equal(t, tx, block.Transactions[0])
}

func TestBlock_CalculateHash(t *testing.T) {
	block := NewBlock([]byte("prev"), 1)
	block.Header.Nonce = 12345

	hash := block.CalculateHash()
	assert.NotNil(t, hash)
	assert.Len(t, hash, 32) // SHA-256 hash length
}

func TestBlock_Validate(t *testing.T) {
	tests := []struct {
		name    string
		block   *Block
		wantErr bool
	}{
		{
			name: "valid block",
			block: &Block{
				Header: &common.Header{
					Version:       1,
					PrevBlockHash: make([]byte, 32),
					MerkleRoot:    make([]byte, 32),
					Timestamp:     time.Now(),
					Difficulty:    0x1d00ffff,
					Hash:          make([]byte, 32),
				},
				Transactions: []*common.Transaction{},
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			block: &Block{
				Header: &common.Header{
					Version: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid prev block hash",
			block: &Block{
				Header: &common.Header{
					Version:       1,
					PrevBlockHash: make([]byte, 16),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid merkle root",
			block: &Block{
				Header: &common.Header{
					Version:       1,
					PrevBlockHash: make([]byte, 32),
					MerkleRoot:    make([]byte, 16),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid timestamp",
			block: &Block{
				Header: &common.Header{
					Version:       1,
					PrevBlockHash: make([]byte, 32),
					MerkleRoot:    make([]byte, 32),
					Timestamp:     time.Time{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty",
			block: &Block{
				Header: &common.Header{
					Version:       1,
					PrevBlockHash: make([]byte, 32),
					MerkleRoot:    make([]byte, 32),
					Timestamp:     time.Now(),
					Difficulty:    0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.block.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBlock_Clone(t *testing.T) {
	original := NewBlock([]byte("prev"), 1)
	original.Header.Nonce = 12345
	original.AddTransaction(&common.Transaction{
		Tx: types.NewTransaction([]byte("from"), []byte("to"), 100, []byte("data")),
	})

	clone := original.Clone()

	assert.NotEqual(t, original, clone)
	assert.Equal(t, original.Header.Version, clone.Header.Version)
	assert.Equal(t, original.Header.Nonce, clone.Header.Nonce)
	assert.Equal(t, original.Header.Height, clone.Header.Height)
	assert.True(t, bytes.Equal(original.Header.PrevBlockHash, clone.Header.PrevBlockHash))
	assert.Len(t, clone.Transactions, len(original.Transactions))
}

func TestBlock_CalculateMerkleRoot(t *testing.T) {
	block := NewBlock([]byte("prev"), 1)

	// Test empty transactions
	root := block.CalculateMerkleRoot()
	assert.Nil(t, root)

	// Test single transaction
	tx1 := &common.Transaction{
		Tx: types.NewTransaction([]byte("from1"), []byte("to1"), 100, []byte("data1")),
	}
	block.AddTransaction(tx1)
	root = block.CalculateMerkleRoot()
	assert.NotNil(t, root)
	assert.Len(t, root, 32)

	// Test multiple transactions
	tx2 := &common.Transaction{
		Tx: types.NewTransaction([]byte("from2"), []byte("to2"), 200, []byte("data2")),
	}
	block.AddTransaction(tx2)
	root = block.CalculateMerkleRoot()
	assert.NotNil(t, root)
	assert.Len(t, root, 32)
}

func TestBlock_GetBlockSize(t *testing.T) {
	block := NewBlock([]byte("prev"), 1)

	// Test empty block
	size := block.GetBlockSize()
	assert.Greater(t, size, 0)

	// Test block with transaction
	tx := &common.Transaction{
		Tx: types.NewTransaction([]byte("from"), []byte("to"), 100, []byte("data")),
	}
	block.AddTransaction(tx)
	size = block.GetBlockSize()
	assert.Greater(t, size, 0)
}

func TestBlock_GetBlockWeight(t *testing.T) {
	block := NewBlock([]byte("prev"), 1)

	// Test empty block
	weight := block.GetBlockWeight()
	assert.Greater(t, weight, 0)

	// Test block with transaction
	tx := &common.Transaction{
		Tx: types.NewTransaction([]byte("from"), []byte("to"), 100, []byte("data")),
	}
	block.AddTransaction(tx)
	weight = block.GetBlockWeight()
	assert.Greater(t, weight, 0)
}

func TestBlock_ValidateBlockWeight(t *testing.T) {
	block := NewBlock([]byte("prev"), 1)

	// Test valid weight
	err := block.ValidateBlockWeight()
	assert.NoError(t, err)

	// Test invalid weight (too large)
	block.Weight = 4000001 // Assuming max weight is 4000000
	err = block.ValidateBlockWeight()
	assert.Error(t, err)
}

func TestBlock_CanAddTransaction(t *testing.T) {
	block := NewBlock([]byte("prev"), 1)
	tx := &common.Transaction{
		Tx: types.NewTransaction([]byte("from"), []byte("to"), 100, []byte("data")),
	}

	// Test can add transaction
	canAdd := block.CanAddTransaction(tx)
	assert.True(t, canAdd)

	// Test cannot add transaction (block full)
	block.Weight = 4000000 // Assuming max weight is 4000000
	canAdd = block.CanAddTransaction(tx)
	assert.False(t, canAdd)
}

func TestBlock_MarshalUnmarshalJSON(t *testing.T) {
	original := NewBlock([]byte("prev"), 1)
	original.Header.Nonce = 12345
	original.AddTransaction(&common.Transaction{
		Tx: types.NewTransaction([]byte("from"), []byte("to"), 100, []byte("data")),
	})

	// Marshal to JSON
	jsonData, err := original.MarshalJSON()
	assert.NoError(t, err)
	assert.NotNil(t, jsonData)

	// Unmarshal from JSON
	var unmarshaled Block
	err = unmarshaled.UnmarshalJSON(jsonData)
	assert.NoError(t, err)

	// Compare original and unmarshaled
	assert.Equal(t, original.Header.Version, unmarshaled.Header.Version)
	assert.Equal(t, original.Header.Nonce, unmarshaled.Header.Nonce)
	assert.Equal(t, original.Header.Height, unmarshaled.Header.Height)
	assert.True(t, bytes.Equal(original.Header.PrevBlockHash, unmarshaled.Header.PrevBlockHash))
	assert.Len(t, unmarshaled.Transactions, len(original.Transactions))
}

func TestGetInitialDifficulty(t *testing.T) {
	tests := []struct {
		name      string
		blockType BlockType
		want      uint32
	}{
		{
			name:      "golden block",
			blockType: GoldenBlock,
			want:      0x1d00ffff,
		},
		{
			name:      "silver block",
			blockType: SilverBlock,
			want:      0x1d00ffff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetInitialDifficulty(tt.blockType)
			assert.Equal(t, tt.want, got)
		})
	}
}
