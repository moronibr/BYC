package pow

import (
	"bytes"
	"math/big"
	"testing"
	"time"
)

// MockBlock implements the Block interface for testing
type MockBlock struct {
	prevHash  []byte
	data      []byte
	timestamp int64
	nonce     int64
	hash      []byte
	height    int64
}

func (b *MockBlock) GetHash() []byte      { return b.hash }
func (b *MockBlock) GetPrevHash() []byte  { return b.prevHash }
func (b *MockBlock) GetTimestamp() int64  { return b.timestamp }
func (b *MockBlock) GetData() []byte      { return b.data }
func (b *MockBlock) GetNonce() int64      { return b.nonce }
func (b *MockBlock) GetHeight() int64     { return b.height }
func (b *MockBlock) SetNonce(nonce int64) { b.nonce = nonce }
func (b *MockBlock) SetHash(hash []byte)  { b.hash = hash }

func TestNewProofOfWork(t *testing.T) {
	block := &MockBlock{
		prevHash:  []byte("prev"),
		data:      []byte("data"),
		timestamp: time.Now().Unix(),
		height:    1,
	}

	pow := NewProofOfWork(block)

	if pow.block != block {
		t.Error("block not set correctly")
	}

	if pow.target == nil {
		t.Error("target not set")
	}
}

func TestRun(t *testing.T) {
	block := &MockBlock{
		prevHash:  []byte("prev"),
		data:      []byte("data"),
		timestamp: time.Now().Unix(),
		height:    1,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	if nonce == 0 {
		t.Error("nonce should not be 0")
	}

	if len(hash) == 0 {
		t.Error("hash should not be empty")
	}

	// Verify the hash meets the target
	var hashInt big.Int
	hashInt.SetBytes(hash)
	if hashInt.Cmp(pow.target) != -1 {
		t.Error("hash does not meet target")
	}
}

func TestRunParallel(t *testing.T) {
	block := &MockBlock{
		prevHash:  []byte("prev"),
		data:      []byte("data"),
		timestamp: time.Now().Unix(),
		height:    1,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.RunParallel(4) // Test with 4 workers

	if nonce == 0 {
		t.Error("nonce should not be 0")
	}

	if len(hash) == 0 {
		t.Error("hash should not be empty")
	}

	// Verify the hash meets the target
	var hashInt big.Int
	hashInt.SetBytes(hash)
	if hashInt.Cmp(pow.target) != -1 {
		t.Error("hash does not meet target")
	}
}

func TestValidate(t *testing.T) {
	block := &MockBlock{
		prevHash:  []byte("prev"),
		data:      []byte("data"),
		timestamp: time.Now().Unix(),
		height:    1,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.SetNonce(nonce)
	block.SetHash(hash)

	if !pow.Validate() {
		t.Error("validation should pass")
	}

	// Test with invalid nonce
	block.SetNonce(nonce + 1)
	if pow.Validate() {
		t.Error("validation should fail with invalid nonce")
	}
}

func TestPrepareData(t *testing.T) {
	block := &MockBlock{
		prevHash:  []byte("prev"),
		data:      []byte("data"),
		timestamp: time.Now().Unix(),
		height:    1,
	}

	pow := NewProofOfWork(block)
	data := pow.prepareData(1)

	expected := bytes.Join(
		[][]byte{
			block.GetPrevHash(),
			block.GetData(),
			IntToHex(block.GetTimestamp()),
			IntToHex(int64(TargetBits)),
			IntToHex(1),
		},
		[]byte{},
	)

	if !bytes.Equal(data, expected) {
		t.Error("prepared data does not match expected")
	}
}

func TestCalculateNextDifficulty(t *testing.T) {
	// Test with too few blocks
	blocks := make([]Block, DifficultyAdjustmentInterval-1)
	for i := range blocks {
		blocks[i] = &MockBlock{
			timestamp: time.Now().Unix() + int64(i*TargetTimePerBlock),
			height:    int64(i + 1),
		}
	}
	if diff := CalculateNextDifficulty(blocks); diff != TargetBits {
		t.Errorf("expected difficulty %d, got %d", TargetBits, diff)
	}

	// Test with blocks mined too quickly
	blocks = make([]Block, DifficultyAdjustmentInterval)
	for i := range blocks {
		blocks[i] = &MockBlock{
			timestamp: time.Now().Unix() + int64(i*TargetTimePerBlock/4), // 4x faster
			height:    int64(i + 1),
		}
	}
	if diff := CalculateNextDifficulty(blocks); diff <= TargetBits {
		t.Error("difficulty should increase for fast mining")
	}

	// Test with blocks mined too slowly
	for i := range blocks {
		blocks[i] = &MockBlock{
			timestamp: time.Now().Unix() + int64(i*TargetTimePerBlock*4), // 4x slower
			height:    int64(i + 1),
		}
	}
	if diff := CalculateNextDifficulty(blocks); diff >= TargetBits {
		t.Error("difficulty should decrease for slow mining")
	}

	// Test difficulty bounds
	blocks = make([]Block, DifficultyAdjustmentInterval)
	for i := range blocks {
		blocks[i] = &MockBlock{
			timestamp: time.Now().Unix() + int64(i*TargetTimePerBlock*10), // Very slow
			height:    int64(i + 1),
		}
	}
	if diff := CalculateNextDifficulty(blocks); diff < MinDifficultyBits {
		t.Errorf("difficulty %d below minimum %d", diff, MinDifficultyBits)
	}

	for i := range blocks {
		blocks[i] = &MockBlock{
			timestamp: time.Now().Unix() + int64(i*TargetTimePerBlock/10), // Very fast
			height:    int64(i + 1),
		}
	}
	if diff := CalculateNextDifficulty(blocks); diff > MaxDifficultyBits {
		t.Errorf("difficulty %d above maximum %d", diff, MaxDifficultyBits)
	}
}
