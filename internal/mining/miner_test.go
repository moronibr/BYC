package mining

import (
	"context"
	"testing"
	"time"

	"github.com/byc/internal/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestNewMiner(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// Test creating miner for mineable coin
	miner, err := NewMiner(bc, blockchain.GoldenBlock, blockchain.Leah, "localhost:3000")
	if err != nil {
		t.Errorf("NewMiner failed: %v", err)
	}
	if miner.BlockType != blockchain.GoldenBlock {
		t.Errorf("Expected GoldenBlock type, got %s", miner.BlockType)
	}
	if miner.CoinType != blockchain.Leah {
		t.Errorf("Expected Leah coin type, got %s", miner.CoinType)
	}

	// Test creating miner for non-mineable coin
	_, err = NewMiner(bc, blockchain.GoldenBlock, blockchain.Senine, "localhost:3000")
	if err == nil {
		t.Error("Expected error when creating miner for non-mineable coin")
	}
}

func TestMinerStartStop(t *testing.T) {
	bc := blockchain.NewBlockchain()
	miner, _ := NewMiner(bc, blockchain.GoldenBlock, blockchain.Leah, "localhost:3000")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start mining
	miner.Start(ctx)

	// Wait for a short time to ensure mining has started
	time.Sleep(100 * time.Millisecond)

	// Stop mining
	miner.Stop()
}

func TestGetMiningDifficulty(t *testing.T) {
	bc := blockchain.NewBlockchain()

	tests := []struct {
		coinType   blockchain.CoinType
		difficulty int
	}{
		{blockchain.Leah, bc.Difficulty * 1},
		{blockchain.Shiblum, bc.Difficulty * 2},
		{blockchain.Shiblon, bc.Difficulty * 4},
	}

	for _, tt := range tests {
		miner, _ := NewMiner(bc, blockchain.GoldenBlock, tt.coinType, "localhost:3000")
		got := miner.GetMiningDifficulty()
		if got != tt.difficulty {
			t.Errorf("GetMiningDifficulty(%s) = %d; want %d", tt.coinType, got, tt.difficulty)
		}
	}
}

func TestGetMiningStats(t *testing.T) {
	bc := blockchain.NewBlockchain()
	miner, _ := NewMiner(bc, blockchain.GoldenBlock, blockchain.Leah, "localhost:3000")

	stats := miner.GetMiningStats()

	// Check required fields
	requiredFields := []string{"block_type", "coin_type", "difficulty", "address", "is_mining"}
	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected %s in mining stats", field)
		}
	}

	// Check values
	if stats["block_type"] != blockchain.GoldenBlock {
		t.Errorf("Expected block_type to be %s, got %s", blockchain.GoldenBlock, stats["block_type"])
	}
	if stats["coin_type"] != blockchain.Leah {
		t.Errorf("Expected coin_type to be %s, got %s", blockchain.Leah, stats["coin_type"])
	}
	if stats["is_mining"] != true {
		t.Error("Expected is_mining to be true")
	}
}

func TestMinerRewards(t *testing.T) {
	// Create a new blockchain
	bc := blockchain.NewBlockchain()

	// Test cases for different coin types
	testCases := []struct {
		name     string
		coinType blockchain.CoinType
		expected float64
	}{
		{"Leah Reward", blockchain.Leah, 50.0},
		{"Shiblum Reward", blockchain.Shiblum, 25.0},
		{"Shiblon Reward", blockchain.Shiblon, 12.5},
		{"Ephraim Reward", blockchain.Ephraim, 0.5},
		{"Manasseh Reward", blockchain.Manasseh, 0.5},
		{"Other Coin Reward", blockchain.Senine, 1.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new miner
			miner, err := NewMiner(bc, blockchain.GoldenBlock, tc.coinType, "test_address")
			assert.NoError(t, err)

			// Start mining
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			miner.Start(ctx)

			// Wait for mining to complete
			time.Sleep(1 * time.Second)
			miner.Stop()

			// Check the reward
			status := miner.GetStatus()
			assert.Equal(t, tc.expected, status.Rewards[tc.coinType], "Reward should match expected value")
		})
	}
}

func TestBlockTime(t *testing.T) {
	bc := blockchain.NewBlockchain()
	miner, err := NewMiner(bc, blockchain.GoldenBlock, blockchain.Leah, "test_address")
	assert.NoError(t, err)

	// Start mining
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	miner.Start(ctx)

	// Record start time
	startTime := time.Now()

	// Wait for first block
	time.Sleep(11 * time.Minute)

	// Get mining stats
	stats := miner.GetMiningStats()
	blocksFound := stats["blocks"].(int64)

	// Stop mining
	miner.Stop()

	// Verify block time
	elapsedTime := time.Since(startTime)
	expectedBlocks := int64(elapsedTime.Minutes() / 10) // Should be roughly 1 block per 10 minutes
	assert.GreaterOrEqual(t, blocksFound, expectedBlocks-1, "Should have mined approximately one block per 10 minutes")
}

func TestSupplyLimits(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// Test Ephraim supply limit
	t.Run("Ephraim Supply Limit", func(t *testing.T) {
		miner, err := NewMiner(bc, blockchain.GoldenBlock, blockchain.Ephraim, "test_address")
		assert.NoError(t, err)

		// Start mining
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		miner.Start(ctx)

		// Wait for mining to complete
		time.Sleep(1 * time.Second)
		miner.Stop()

		// Check if supply is within limits
		supply := bc.GetTotalSupply(blockchain.Ephraim)
		assert.LessOrEqual(t, supply, blockchain.MaxEphraimSupply, "Ephraim supply should not exceed maximum")
	})

	// Test Manasseh supply limit
	t.Run("Manasseh Supply Limit", func(t *testing.T) {
		miner, err := NewMiner(bc, blockchain.SilverBlock, blockchain.Manasseh, "test_address")
		assert.NoError(t, err)

		// Start mining
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		miner.Start(ctx)

		// Wait for mining to complete
		time.Sleep(1 * time.Second)
		miner.Stop()

		// Check if supply is within limits
		supply := bc.GetTotalSupply(blockchain.Manasseh)
		assert.LessOrEqual(t, supply, blockchain.MaxManassehSupply, "Manasseh supply should not exceed maximum")
	})
}

func TestMiningPool(t *testing.T) {
	// Create a new mining pool
	pool := NewMiningPool("pool_address")
	assert.NotNil(t, pool)

	// Create a test miner
	bc := blockchain.NewBlockchain()
	miner, err := NewMiner(bc, blockchain.GoldenBlock, blockchain.Leah, "test_miner")
	assert.NoError(t, err)

	// Add miner to pool
	pool.AddMiner(miner)
	assert.Equal(t, 1, len(pool.Miners), "Pool should have one miner")

	// Get pool stats
	stats := pool.GetPoolStats()
	assert.Equal(t, 1, stats["total_miners"], "Pool stats should show one miner")

	// Remove miner from pool
	pool.RemoveMiner("test_miner")
	assert.Equal(t, 0, len(pool.Miners), "Pool should be empty after removing miner")
}
