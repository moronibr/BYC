package blockchain

import (
	"testing"
	"time"
)

func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchain()

	// Test genesis blocks
	if len(bc.GoldenBlocks) != 1 {
		t.Errorf("Expected 1 golden genesis block, got %d", len(bc.GoldenBlocks))
	}
	if len(bc.SilverBlocks) != 1 {
		t.Errorf("Expected 1 silver genesis block, got %d", len(bc.SilverBlocks))
	}

	// Test genesis block properties
	goldenGenesis := bc.GoldenBlocks[0]
	if goldenGenesis.BlockType != GoldenBlock {
		t.Errorf("Expected golden genesis block type, got %s", goldenGenesis.BlockType)
	}
	if len(goldenGenesis.PrevHash) != 0 {
		t.Error("Expected empty previous hash for genesis block")
	}

	silverGenesis := bc.SilverBlocks[0]
	if silverGenesis.BlockType != SilverBlock {
		t.Errorf("Expected silver genesis block type, got %s", silverGenesis.BlockType)
	}
	if len(silverGenesis.PrevHash) != 0 {
		t.Error("Expected empty previous hash for genesis block")
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

	// Test mining block with invalid coin type
	_, err := bc.MineBlock([]Transaction{}, "INVALID", Leah)
	if err == nil {
		t.Error("Expected error for invalid block type")
	}

	// Test mining block with valid parameters
	block, err := bc.MineBlock([]Transaction{}, GoldenBlock, Leah)
	if err != nil {
		t.Errorf("Failed to mine block: %v", err)
	}
	if block.BlockType != GoldenBlock {
		t.Errorf("Expected golden block type, got %s", block.BlockType)
	}
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockchain()

	// Create a valid block
	block := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		BlockType:    GoldenBlock,
		Difficulty:   1,
	}

	// Mine the block
	block.Hash = calculateHash(block)
	for !bc.isValidProof(block) {
		block.Nonce++
		block.Hash = calculateHash(block)
	}

	// Test adding valid block
	err := bc.AddBlock(block)
	if err != nil {
		t.Errorf("Failed to add valid block: %v", err)
	}

	// Test adding invalid block (wrong previous hash)
	invalidBlock := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     []byte("invalid"),
		BlockType:    GoldenBlock,
		Difficulty:   1,
	}
	err = bc.AddBlock(invalidBlock)
	if err == nil {
		t.Error("Expected error when adding block with invalid previous hash")
	}
}

func TestValidateBlock(t *testing.T) {
	bc := NewBlockchain()

	// Test block with future timestamp
	futureBlock := Block{
		Timestamp:    time.Now().Unix() + 1000,
		Transactions: []Transaction{},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		BlockType:    GoldenBlock,
		Difficulty:   1,
	}
	err := bc.validateBlock(futureBlock)
	if err == nil {
		t.Error("Expected error for block with future timestamp")
	}

	// Test block with no transactions
	emptyBlock := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		BlockType:    GoldenBlock,
		Difficulty:   1,
	}
	err = bc.validateBlock(emptyBlock)
	if err == nil {
		t.Error("Expected error for block with no transactions")
	}

	// Test block with invalid proof of work
	invalidPowBlock := Block{
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     bc.GoldenBlocks[0].Hash,
		BlockType:    GoldenBlock,
		Difficulty:   1,
		Hash:         []byte("invalid"),
	}
	err = bc.validateBlock(invalidPowBlock)
	if err == nil {
		t.Error("Expected error for block with invalid proof of work")
	}
}

func TestGetBalance(t *testing.T) {
	bc := NewBlockchain()
	address := "test_address"

	// Test initial balance
	balance := bc.GetBalance(address, Leah)
	if balance != 0 {
		t.Errorf("Expected initial balance of 0, got %f", balance)
	}

	// TODO: Add tests for balance after transactions
}

func TestCreateTransaction(t *testing.T) {
	bc := NewBlockchain()

	// Test creating transaction with invalid amount
	_, err := bc.CreateTransaction("from", "to", -1, Leah)
	if err == nil {
		t.Error("Expected error for negative amount")
	}

	// Test creating transaction with zero amount
	_, err = bc.CreateTransaction("from", "to", 0, Leah)
	if err == nil {
		t.Error("Expected error for zero amount")
	}

	// Test creating transaction with invalid coin type
	_, err = bc.CreateTransaction("from", "to", 1, "INVALID")
	if err == nil {
		t.Error("Expected error for invalid coin type")
	}
}

func TestBlockSize(t *testing.T) {
	bc := NewBlockchain()

	// Create a block that exceeds MaxBlockSize
	largeBlock := Block{
		Timestamp: time.Now().Unix(),
		Transactions: []Transaction{
			{
				ID: make([]byte, MaxBlockSize), // Create a transaction that's too large
			},
		},
		PrevHash:   bc.GoldenBlocks[0].Hash,
		BlockType:  GoldenBlock,
		Difficulty: 1,
	}

	err := bc.validateBlock(largeBlock)
	if err == nil {
		t.Error("Expected error for block exceeding maximum size")
	}
}

func TestConcurrentBlockOperations(t *testing.T) {
	bc := NewBlockchain()
	done := make(chan bool)

	// Test concurrent block additions
	for i := 0; i < 10; i++ {
		go func() {
			block := Block{
				Timestamp:    time.Now().Unix(),
				Transactions: []Transaction{},
				PrevHash:     bc.GoldenBlocks[0].Hash,
				BlockType:    GoldenBlock,
				Difficulty:   1,
			}
			bc.AddBlock(block)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify blockchain state
	if len(bc.GoldenBlocks) <= 1 {
		t.Error("Expected multiple blocks after concurrent additions")
	}
}
