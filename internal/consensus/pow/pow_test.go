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
	// Test with nil block
	pow, err := NewProofOfWork(nil, nil)
	if err != ErrInvalidBlock {
		t.Errorf("Expected ErrInvalidBlock, got %v", err)
	}
	if pow != nil {
		t.Error("Expected nil pow for nil block")
	}

	// Test with valid block
	block := &MockBlock{data: []byte("test")}
	pow, err = NewProofOfWork(block, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pow == nil {
		t.Error("Expected non-nil pow for valid block")
	}

	// Test with custom config
	config := &Config{
		MiningTimeout:   time.Minute,
		MaxRetries:      2,
		RecoveryEnabled: false,
		LoggingEnabled:  false,
	}
	pow, err = NewProofOfWork(block, config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pow.config != config {
		t.Error("Expected custom config to be set")
	}
}

func TestRun(t *testing.T) {
	block := &MockBlock{data: []byte("test")}
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}

	nonce, hash, err := pow.Run()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if nonce == 0 {
		t.Error("Expected non-zero nonce")
	}
	if len(hash) == 0 {
		t.Error("Expected non-empty hash")
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

	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}

	nonce, hash, err := pow.RunParallel(4) // Test with 4 workers
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

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
	block := &MockBlock{data: []byte("test")}
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}

	// Test valid block
	nonce, hash, err := pow.Run()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	block.SetNonce(nonce)
	block.SetHash(hash)
	if !pow.Validate() {
		t.Error("Expected valid block")
	}

	// Test invalid block
	block.SetNonce(0)
	if pow.Validate() {
		t.Error("Expected invalid block")
	}
}

func TestPrepareData(t *testing.T) {
	block := &MockBlock{data: []byte("test")}
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}

	data, err := pow.prepareData(1)
	if err != nil {
		t.Fatalf("Failed to prepare data: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	// Verify data contains block data
	if !bytes.Contains(data, block.GetData()) {
		t.Error("Data should contain block data")
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

func TestRunAdaptive(t *testing.T) {
	block := &MockBlock{data: []byte("test")}
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}

	// Test successful mining
	nonce, hash, err := pow.RunAdaptive()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if nonce == 0 {
		t.Error("Expected non-zero nonce")
	}
	if len(hash) == 0 {
		t.Error("Expected non-empty hash")
	}

	// Test resource exhaustion
	pow.rm.SetResourceLimits(0, 0) // Set limits to force resource exhaustion
	_, _, err = pow.RunAdaptive()
	if err == nil {
		t.Error("Expected error for resource exhaustion")
	}

	// Test timeout
	config := &Config{
		MiningTimeout:   time.Nanosecond, // Set very short timeout
		MaxRetries:      1,
		RecoveryEnabled: true,
		LoggingEnabled:  true,
	}
	pow, err = NewProofOfWork(block, config)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}
	_, _, err = pow.RunAdaptive()
	if err == nil {
		t.Error("Expected error for timeout")
	}
}

func TestResourceManager(t *testing.T) {
	rm := NewResourceManager()

	// Test initial state
	workerCount := rm.GetOptimalWorkerCount()
	if workerCount < 1 {
		t.Errorf("Expected at least 1 worker, got %d", workerCount)
	}

	// Test resource limits
	rm.SetResourceLimits(50, 80)
	cpuCount, memUsage := rm.GetSystemMetrics()
	if cpuCount > 50 {
		t.Errorf("CPU count exceeds limit: %d > 50", cpuCount)
	}
	if memUsage > 0.8 {
		t.Errorf("Memory usage exceeds limit: %f > 0.8", memUsage)
	}

	// Test worker load updates
	rm.UpdateWorkerLoad(4, 100.0)
	workerCount = rm.GetOptimalWorkerCount()
	if workerCount > 4 {
		t.Errorf("Worker count exceeds limit: %d > 4", workerCount)
	}
}

func TestShutdown(t *testing.T) {
	block := &MockBlock{data: []byte("test")}
	pow, err := NewProofOfWork(block, nil)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}

	// Test normal shutdown
	err = pow.Shutdown()
	if err != nil {
		t.Errorf("Unexpected error during shutdown: %v", err)
	}

	// Test double shutdown
	err = pow.Shutdown()
	if err != nil {
		t.Errorf("Unexpected error during second shutdown: %v", err)
	}

	// Test mining after shutdown
	_, _, err = pow.RunAdaptive()
	if err == nil {
		t.Error("Expected error when mining after shutdown")
	}
}

func TestCircuitBreaker(t *testing.T) {
	block := &MockBlock{data: []byte("test")}
	config := &Config{
		MiningTimeout:   time.Millisecond * 100,
		MaxRetries:      1,
		RecoveryEnabled: true,
		LoggingEnabled:  false,
	}
	pow, err := NewProofOfWork(block, config)
	if err != nil {
		t.Fatalf("Failed to create PoW: %v", err)
	}

	// Force resource exhaustion to trigger failures
	pow.rm.SetResourceLimits(0, 0)

	// Test circuit breaker opening
	for i := 0; i < pow.circuitBreakerConfig.FailureThreshold; i++ {
		_, _, err = pow.RunAdaptive()
		if err == nil {
			t.Errorf("Expected error on attempt %d", i+1)
		}
	}

	// Verify circuit breaker is open
	_, _, err = pow.RunAdaptive()
	if err == nil {
		t.Error("Expected error when circuit breaker is open")
	}

	// Wait for reset timeout
	time.Sleep(pow.circuitBreakerConfig.ResetTimeout)

	// Verify circuit breaker is in half-open state
	_, _, err = pow.RunAdaptive()
	if err == nil {
		t.Error("Expected error when circuit breaker is half-open")
	}

	// Reset resource limits
	pow.rm.SetResourceLimits(100, 100)

	// Wait for reset timeout again
	time.Sleep(pow.circuitBreakerConfig.ResetTimeout)

	// Verify circuit breaker closes after successful operation
	_, _, err = pow.RunAdaptive()
	if err != nil {
		t.Errorf("Unexpected error after circuit breaker reset: %v", err)
	}
}
