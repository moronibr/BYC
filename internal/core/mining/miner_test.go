package mining

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/types"
	"github.com/youngchain/internal/core/utxo"
)

// MockBlockStore is a mock implementation of storage.BlockStore
type MockBlockStore struct {
	mock.Mock
}

func (m *MockBlockStore) PutBlock(block *block.Block) error {
	args := m.Called(block)
	return args.Error(0)
}

func (m *MockBlockStore) GetBlock(height uint64) (*block.Block, error) {
	args := m.Called(height)
	return args.Get(0).(*block.Block), args.Error(1)
}

func (m *MockBlockStore) GetLastBlock() (*block.Block, error) {
	args := m.Called()
	return args.Get(0).(*block.Block), args.Error(1)
}

func (m *MockBlockStore) DeleteBlock(height uint64) error {
	args := m.Called(height)
	return args.Error(0)
}

// MockUTXOSet is a mock implementation of utxo.UTXOSetInterface
type MockUTXOSet struct {
	mock.Mock
}

func (m *MockUTXOSet) GetUTXO(txHash []byte, outIndex uint32) (*utxo.UTXO, error) {
	args := m.Called(txHash, outIndex)
	return args.Get(0).(*utxo.UTXO), args.Error(1)
}

func (m *MockUTXOSet) Update(block *block.Block) error {
	args := m.Called(block)
	return args.Error(0)
}

func (m *MockUTXOSet) AddUTXO(utxo *utxo.UTXO) error {
	args := m.Called(utxo)
	return args.Error(0)
}

func (m *MockUTXOSet) GetBalance(address []byte) (uint64, error) {
	args := m.Called(address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockUTXOSet) RemoveUTXO(txHash []byte, outIndex uint32) error {
	args := m.Called(txHash, outIndex)
	return args.Error(0)
}

func (m *MockUTXOSet) UpdateWithBlock(block *block.Block) error {
	args := m.Called(block)
	return args.Error(0)
}

func TestNewMiner(t *testing.T) {
	txPool := transaction.NewTxPool(1000, 1000, nil)
	blockStore := &MockBlockStore{}
	utxoSet := &MockUTXOSet{}
	miningAddress := "miner_address"

	miner := NewMiner(txPool, blockStore, utxoSet, miningAddress)

	assert.NotNil(t, miner)
	assert.Equal(t, txPool, miner.txPool)
	assert.Equal(t, blockStore, miner.blockStore)
	assert.Equal(t, utxoSet, miner.utxoSet)
	assert.Equal(t, miningAddress, miner.miningAddress)
	assert.NotNil(t, miner.target)
	assert.NotNil(t, miner.stopChan)
	assert.False(t, miner.isMining)
}

func TestMiner_StartStopMining(t *testing.T) {
	txPool := transaction.NewTxPool(1000, 1000, nil)
	blockStore := &MockBlockStore{}
	utxoSet := &MockUTXOSet{}
	miningAddress := "miner_address"

	miner := NewMiner(txPool, blockStore, utxoSet, miningAddress)

	// Test starting mining
	miner.StartMining()
	assert.True(t, miner.isMining)

	// Test stopping mining
	miner.StopMining()
	assert.False(t, miner.isMining)
}

func TestMiner_CreateBlock(t *testing.T) {
	txPool := transaction.NewTxPool(1000, 1000, nil)
	blockStore := &MockBlockStore{}
	utxoSet := &MockUTXOSet{}
	miningAddress := "miner_address"

	// Setup mock expectations
	lastBlock := &block.Block{
		Header: &common.Header{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    TargetBits,
			Height:        1,
			Hash:          []byte("block_hash"),
		},
		Transactions: []*common.Transaction{},
	}
	blockStore.On("GetLastBlock").Return(lastBlock, nil)

	miner := NewMiner(txPool, blockStore, utxoSet, miningAddress)

	// Test block creation
	newBlock, err := miner.createBlock()
	assert.NoError(t, err)
	assert.NotNil(t, newBlock)
	assert.Equal(t, lastBlock.Header.Height+1, newBlock.Header.Height)
	assert.Equal(t, lastBlock.Header.Hash, newBlock.Header.PrevBlockHash)
	assert.NotEmpty(t, newBlock.Transactions) // Should have at least coinbase transaction
}

func TestMiner_MineBlock(t *testing.T) {
	txPool := transaction.NewTxPool(1000, 1000, nil)
	blockStore := &MockBlockStore{}
	utxoSet := &MockUTXOSet{}
	miningAddress := "miner_address"

	miner := NewMiner(txPool, blockStore, utxoSet, miningAddress)

	// Create a test block
	testBlock := &block.Block{
		Header: &common.Header{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    TargetBits,
			Height:        1,
		},
		Transactions: []*common.Transaction{},
	}

	// Test mining
	nonce, err := miner.mineBlock(testBlock)
	assert.NoError(t, err)
	assert.NotZero(t, nonce)
}

func TestCreateCoinbaseTx(t *testing.T) {
	minerAddress := []byte("miner_address")
	blockHeight := uint64(1)

	coinbaseTx := createCoinbaseTx(minerAddress, blockHeight)

	assert.NotNil(t, coinbaseTx)
	assert.Equal(t, string(minerAddress), coinbaseTx.To())
	assert.Equal(t, BlockReward, coinbaseTx.Amount())
	assert.NotNil(t, coinbaseTx.GetTransaction().Inputs[0])
	assert.Equal(t, uint32(0xffffffff), coinbaseTx.GetTransaction().Inputs[0].PreviousTxIndex)
	assert.Equal(t, uint32(0xffffffff), coinbaseTx.GetTransaction().Inputs[0].Sequence)
}

func TestCalculateBlockReward(t *testing.T) {
	tests := []struct {
		name     string
		height   uint64
		expected uint64
	}{
		{
			name:     "initial reward",
			height:   0,
			expected: BlockReward,
		},
		{
			name:     "first halving",
			height:   HalvingInterval,
			expected: BlockReward / 2,
		},
		{
			name:     "second halving",
			height:   HalvingInterval * 2,
			expected: BlockReward / 4,
		},
		{
			name:     "max halvings",
			height:   HalvingInterval * 64,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := calculateBlockReward(tt.height)
			assert.Equal(t, tt.expected, reward)
		})
	}
}

func TestCalculateMerkleRoot(t *testing.T) {
	// Test empty transactions
	root := calculateMerkleRoot([]*types.Transaction{})
	assert.Nil(t, root)

	// Test single transaction
	tx1 := types.NewTransaction([]byte("from1"), []byte("to1"), 100, []byte("data1"))
	root = calculateMerkleRoot([]*types.Transaction{tx1})
	assert.NotNil(t, root)
	assert.Len(t, root, 32)

	// Test multiple transactions
	tx2 := types.NewTransaction([]byte("from2"), []byte("to2"), 200, []byte("data2"))
	root = calculateMerkleRoot([]*types.Transaction{tx1, tx2})
	assert.NotNil(t, root)
	assert.Len(t, root, 32)
}

func TestMiner_UpdateUTXOSet(t *testing.T) {
	txPool := transaction.NewTxPool(1000, 1000, nil)
	blockStore := &MockBlockStore{}
	utxoSet := &MockUTXOSet{}
	miningAddress := "miner_address"

	// Setup mock expectations
	testBlock := &block.Block{
		Header: &common.Header{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    TargetBits,
			Height:        1,
		},
		Transactions: []*common.Transaction{},
	}
	utxoSet.On("Update", testBlock).Return(nil)

	miner := NewMiner(txPool, blockStore, utxoSet, miningAddress)

	// Test UTXO set update
	err := miner.updateUTXOSet(testBlock)
	assert.NoError(t, err)
	utxoSet.AssertCalled(t, "Update", testBlock)
}

func TestMiner_AdjustDifficulty(t *testing.T) {
	txPool := transaction.NewTxPool(1000, 1000, nil)
	blockStore := &MockBlockStore{}
	utxoSet := &MockUTXOSet{}
	miningAddress := "miner_address"

	// Setup mock expectations
	lastBlock := &block.Block{
		Header: &common.Header{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now().Add(-TargetTimePerBlock),
			Difficulty:    TargetBits,
			Height:        DifficultyAdjustmentInterval,
		},
	}
	blockStore.On("GetLastBlock").Return(lastBlock, nil)

	miner := NewMiner(txPool, blockStore, utxoSet, miningAddress)

	// Test difficulty adjustment
	err := miner.adjustDifficulty()
	assert.NoError(t, err)
}
