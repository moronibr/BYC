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
}

func (b *MockBlock) GetHash() []byte      { return b.hash }
func (b *MockBlock) GetPrevHash() []byte  { return b.prevHash }
func (b *MockBlock) GetTimestamp() int64  { return b.timestamp }
func (b *MockBlock) GetData() []byte      { return b.data }
func (b *MockBlock) GetNonce() int64      { return b.nonce }
func (b *MockBlock) SetNonce(nonce int64) { b.nonce = nonce }
func (b *MockBlock) SetHash(hash []byte)  { b.hash = hash }

func TestNewProofOfWork(t *testing.T) {
	block := &MockBlock{
		prevHash:  []byte("prev"),
		data:      []byte("data"),
		timestamp: time.Now().Unix(),
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

func TestValidate(t *testing.T) {
	block := &MockBlock{
		prevHash:  []byte("prev"),
		data:      []byte("data"),
		timestamp: time.Now().Unix(),
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
