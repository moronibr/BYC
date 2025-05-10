package consensus

import (
	"testing"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/types"
)

func TestMineBlock(t *testing.T) {
	// Create consensus instance
	consensus := NewConsensus()

	// Create test block
	block := block.NewBlock([]byte("test"), uint64(consensus.GetDifficulty()))
	block.Header.Height = 1
	block.Header.Timestamp = time.Now()

	// Mine block
	err := consensus.MineBlock(block)
	if err != nil {
		t.Fatalf("Failed to mine block: %v", err)
	}

	// Verify block hash
	if len(block.Header.Hash) == 0 {
		t.Fatal("Block hash is empty")
	}

	// Verify block validation
	err = consensus.ValidateBlock(block)
	if err != nil {
		t.Fatalf("Block validation failed: %v", err)
	}
}

func TestDifficultyAdjustment(t *testing.T) {
	// Create consensus instance
	consensus := NewConsensus()

	// Create test blocks
	blocks := make([]*block.Block, DifficultyAdjustmentInterval)
	for i := 0; i < DifficultyAdjustmentInterval; i++ {
		blocks[i] = block.NewBlock([]byte("test"), uint64(consensus.GetDifficulty()))
		blocks[i].Header.Height = uint64(i + 1)
		blocks[i].Header.Timestamp = time.Now().Add(time.Duration(i) * TargetBlockTime)
	}

	// Adjust difficulty for each block
	for _, block := range blocks {
		err := consensus.AdjustDifficulty(block)
		if err != nil {
			t.Fatalf("Failed to adjust difficulty: %v", err)
		}
	}

	// Verify difficulty adjustment
	newDifficulty := consensus.GetDifficulty()
	if newDifficulty == MinDifficulty {
		t.Fatal("Difficulty was not adjusted")
	}
}

func TestMiningReward(t *testing.T) {
	// Create consensus instance
	consensus := NewConsensus()

	// Create test block
	block := block.NewBlock([]byte("test"), uint64(consensus.GetDifficulty()))
	block.Header.Height = 1
	block.Header.Timestamp = time.Now()

	// Add mining reward
	err := consensus.addMiningReward(block)
	if err != nil {
		t.Fatalf("Failed to add mining reward: %v", err)
	}

	// Verify mining reward transaction
	if len(block.Transactions) == 0 {
		t.Fatal("No mining reward transaction")
	}

	rewardTx := block.Transactions[0]
	if len(rewardTx.From) != 0 {
		t.Fatal("Mining reward transaction has inputs")
	}

	if len(rewardTx.To) == 0 {
		t.Fatal("Mining reward transaction has no outputs")
	}

	expectedReward := consensus.calculateBlockReward(block.Header.Height)
	if rewardTx.Amount != expectedReward {
		t.Fatalf("Invalid mining reward amount: got %d, want %d", rewardTx.Amount, expectedReward)
	}
}

func TestBlockTimeValidation(t *testing.T) {
	// Create consensus instance
	consensus := NewConsensus()

	// Create test block
	block := block.NewBlock([]byte("test"), uint64(consensus.GetDifficulty()))
	block.Header.Height = 1

	// Test future block
	block.Header.Timestamp = time.Now().Add(MaxFutureBlockTime * time.Second)
	err := consensus.validateBlockTime(block)
	if err == nil {
		t.Fatal("Future block time validation should fail")
	}

	// Test old block
	block.Header.Timestamp = time.Now().Add(-MaxFutureBlockTime * time.Second)
	err = consensus.validateBlockTime(block)
	if err == nil {
		t.Fatal("Old block time validation should fail")
	}

	// Test valid block time
	block.Header.Timestamp = time.Now()
	err = consensus.validateBlockTime(block)
	if err != nil {
		t.Fatalf("Valid block time validation failed: %v", err)
	}
}

func TestTransactionValidation(t *testing.T) {
	// Create consensus instance
	consensus := NewConsensus()

	// Create test transaction
	tx := types.NewTransaction(
		[]byte("test_from"),
		[]byte("test_to"),
		1000,
		[]byte("test_data"),
	)

	// Validate transaction
	err := consensus.ValidateTransaction(tx)
	if err != nil {
		t.Fatalf("Transaction validation failed: %v", err)
	}

	// Test oversized transaction
	tx.Data = make([]byte, 1000001)
	err = consensus.ValidateTransaction(tx)
	if err == nil {
		t.Fatal("Oversized transaction validation should fail")
	}

	// Test low fee transaction
	tx.Data = []byte("test_data")
	tx.Amount = 0
	err = consensus.ValidateTransaction(tx)
	if err == nil {
		t.Fatal("Low fee transaction validation should fail")
	}
}

func TestChainSelection(t *testing.T) {
	// Create consensus instance
	consensus := NewConsensus()

	// Create test chains
	chains := make([]*block.Block, 3)
	for i := 0; i < 3; i++ {
		chains[i] = block.NewBlock([]byte("test"), uint64(consensus.GetDifficulty()))
		chains[i].Header.Height = uint64(i + 1)
		chains[i].Header.Timestamp = time.Now()
	}

	// Select best chain
	bestChain := consensus.SelectBestChain(chains)
	if bestChain == nil {
		t.Fatal("Failed to select best chain")
	}

	// Verify chain selection
	if bestChain.Header.Height != 3 {
		t.Fatalf("Invalid best chain height: got %d, want %d", bestChain.Header.Height, 3)
	}
}
