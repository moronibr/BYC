package blockchain

import (
	"testing"
	"time"
)

func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchain()

	// Test genesis blocks
	if len(bc.GoldenBlocks) != 1 {
		t.Errorf("Expected 1 block in GoldenBlocks, got %d", len(bc.GoldenBlocks))
	}
	if len(bc.SilverBlocks) != 1 {
		t.Errorf("Expected 1 block in SilverBlocks, got %d", len(bc.SilverBlocks))
	}

	// Test genesis block types
	if bc.GoldenBlocks[0].BlockType != GoldenBlock {
		t.Errorf("Expected GoldenBlock type, got %s", bc.GoldenBlocks[0].BlockType)
	}
	if bc.SilverBlocks[0].BlockType != SilverBlock {
		t.Errorf("Expected SilverBlock type, got %s", bc.SilverBlocks[0].BlockType)
	}
}

func TestMiningDifficulty(t *testing.T) {
	tests := []struct {
		coinType   CoinType
		difficulty int
	}{
		{Leah, 1},
		{Shiblum, 2},
		{Shiblon, 4},
		{Senine, 0},
		{Antion, 0},
	}

	for _, tt := range tests {
		got := MiningDifficulty(tt.coinType)
		if got != tt.difficulty {
			t.Errorf("MiningDifficulty(%s) = %d; want %d", tt.coinType, got, tt.difficulty)
		}
	}
}

func TestIsMineable(t *testing.T) {
	tests := []struct {
		coinType CoinType
		want     bool
	}{
		{Leah, true},
		{Shiblum, true},
		{Shiblon, true},
		{Senine, false},
		{Antion, false},
	}

	for _, tt := range tests {
		got := IsMineable(tt.coinType)
		if got != tt.want {
			t.Errorf("IsMineable(%s) = %v; want %v", tt.coinType, got, tt.want)
		}
	}
}

func TestCanTransferBetweenBlocks(t *testing.T) {
	tests := []struct {
		coinType CoinType
		want     bool
	}{
		{Antion, true},
		{Leah, false},
		{Shiblum, false},
		{Shiblon, false},
	}

	for _, tt := range tests {
		got := CanTransferBetweenBlocks(tt.coinType)
		if got != tt.want {
			t.Errorf("CanTransferBetweenBlocks(%s) = %v; want %v", tt.coinType, got, tt.want)
		}
	}
}

func TestGetBlockType(t *testing.T) {
	tests := []struct {
		coinType CoinType
		want     BlockType
	}{
		{Senine, GoldenBlock},
		{Seon, GoldenBlock},
		{Shum, GoldenBlock},
		{Limnah, GoldenBlock},
		{Senum, SilverBlock},
		{Amnor, SilverBlock},
		{Ezrom, SilverBlock},
		{Onti, SilverBlock},
		{Leah, ""},
		{Shiblum, ""},
		{Shiblon, ""},
		{Antion, ""},
		{Ephraim, GoldenBlock},
		{Manasseh, SilverBlock},
		{Joseph, ""},
	}

	for _, tt := range tests {
		got := GetBlockType(tt.coinType)
		if got != tt.want {
			t.Errorf("GetBlockType(%s) = %s; want %s", tt.coinType, got, tt.want)
		}
	}
}

func TestMineBlock(t *testing.T) {
	bc := NewBlockchain()

	// Test mining Leah coins
	block, err := bc.MineBlock([]Transaction{}, GoldenBlock, Leah)
	if err != nil {
		t.Errorf("MineBlock failed: %v", err)
	}
	if block.BlockType != GoldenBlock {
		t.Errorf("Expected GoldenBlock type, got %s", block.BlockType)
	}
	if block.Difficulty != bc.Difficulty*MiningDifficulty(Leah) {
		t.Errorf("Expected difficulty %d, got %d", bc.Difficulty*MiningDifficulty(Leah), block.Difficulty)
	}

	// Test mining non-mineable coin
	_, err = bc.MineBlock([]Transaction{}, GoldenBlock, Senine)
	if err == nil {
		t.Error("Expected error when mining non-mineable coin")
	}
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockchain()

	// Create a valid block
	block := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		Nonce:        0,
		BlockType:    GoldenBlock,
		Difficulty:   bc.Difficulty,
	}

	// Test adding valid block
	err := bc.AddBlock(block)
	if err == nil {
		t.Error("Expected error when adding block with invalid proof")
	}

	// Test adding block with invalid type
	block.BlockType = "INVALID"
	err = bc.AddBlock(block)
	if err == nil {
		t.Error("Expected error when adding block with invalid type")
	}
}
