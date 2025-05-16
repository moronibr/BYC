package wallet

import (
	"testing"

	"github.com/byc/internal/blockchain"
)

func TestNewWallet(t *testing.T) {
	w, err := NewWallet()
	if err != nil {
		t.Errorf("NewWallet failed: %v", err)
	}

	if w.PrivateKey == nil {
		t.Error("Expected private key to be set")
	}
	if w.PublicKey == nil {
		t.Error("Expected public key to be set")
	}
	if w.Address == "" {
		t.Error("Expected address to be set")
	}
}

func TestCreateTransaction(t *testing.T) {
	w, _ := NewWallet()
	bc := blockchain.NewBlockchain()

	// Test creating transaction with Antion (transferable between blocks)
	tx, err := w.CreateTransaction("recipient", 1.0, blockchain.Antion, bc)
	if err != nil {
		t.Errorf("CreateTransaction failed: %v", err)
	}
	if tx.BlockType != "" {
		t.Errorf("Expected empty block type for Antion, got %s", tx.BlockType)
	}

	// Test creating transaction with Senine (Golden Block only)
	tx, err = w.CreateTransaction("recipient", 1.0, blockchain.Senine, bc)
	if err != nil {
		t.Errorf("CreateTransaction failed: %v", err)
	}
	if tx.BlockType != blockchain.GoldenBlock {
		t.Errorf("Expected GoldenBlock type, got %s", tx.BlockType)
	}

	// Test creating transaction with invalid amount
	_, err = w.CreateTransaction("recipient", -1.0, blockchain.Leah, bc)
	if err == nil {
		t.Error("Expected error when creating transaction with negative amount")
	}
}

func TestGetBalance(t *testing.T) {
	w, _ := NewWallet()
	bc := blockchain.NewBlockchain()

	// Test getting balance for all coin types
	balances := w.GetAllBalances(bc)

	// Check that all coin types are present
	expectedCoins := []blockchain.CoinType{
		blockchain.Leah, blockchain.Shiblum, blockchain.Shiblon,
		blockchain.Senine, blockchain.Seon, blockchain.Shum,
		blockchain.Limnah, blockchain.Antion, blockchain.Senum,
		blockchain.Amnor, blockchain.Ezrom, blockchain.Onti,
		blockchain.Ephraim, blockchain.Manasseh, blockchain.Joseph,
	}

	for _, coin := range expectedCoins {
		if _, exists := balances[coin]; !exists {
			t.Errorf("Expected balance for %s to be present", coin)
		}
	}
}

func TestCreateEphraimCoin(t *testing.T) {
	w, _ := NewWallet()
	bc := blockchain.NewBlockchain()

	// Test creating Ephraim coin without sufficient Limnah
	err := w.CreateEphraimCoin(bc)
	if err == nil {
		t.Error("Expected error when creating Ephraim coin without sufficient Limnah")
	}
}

func TestCreateManassehCoin(t *testing.T) {
	w, _ := NewWallet()
	bc := blockchain.NewBlockchain()

	// Test creating Manasseh coin without sufficient Onti
	err := w.CreateManassehCoin(bc)
	if err == nil {
		t.Error("Expected error when creating Manasseh coin without sufficient Onti")
	}
}

func TestCreateJosephCoin(t *testing.T) {
	w, _ := NewWallet()
	bc := blockchain.NewBlockchain()

	// Test creating Joseph coin without sufficient Ephraim and Manasseh
	err := w.CreateJosephCoin(bc)
	if err == nil {
		t.Error("Expected error when creating Joseph coin without sufficient Ephraim and Manasseh")
	}
}
