package mining

import (
	"context"
	"testing"
	"time"

	"github.com/byc/internal/blockchain"
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

func TestMiningPool(t *testing.T) {
	pool := NewMiningPool("localhost:3000")
	bc := blockchain.NewBlockchain()

	// Create and add miners
	miner1, _ := NewMiner(bc, blockchain.GoldenBlock, blockchain.Leah, "localhost:3001")
	miner2, _ := NewMiner(bc, blockchain.SilverBlock, blockchain.Shiblum, "localhost:3002")

	pool.AddMiner(miner1)
	pool.AddMiner(miner2)

	// Test pool stats
	stats := pool.GetPoolStats()
	if stats["total_miners"] != 2 {
		t.Errorf("Expected 2 miners in pool, got %d", stats["total_miners"])
	}

	// Test getting miner
	gotMiner := pool.GetMiner("localhost:3001")
	if gotMiner != miner1 {
		t.Error("Expected to get miner1")
	}

	// Test removing miner
	pool.RemoveMiner("localhost:3001")
	stats = pool.GetPoolStats()
	if stats["total_miners"] != 1 {
		t.Errorf("Expected 1 miner in pool after removal, got %d", stats["total_miners"])
	}
}
