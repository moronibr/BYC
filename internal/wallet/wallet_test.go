package wallet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/byc/internal/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	wallet, err := NewWallet()
	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.NotNil(t, wallet.PrivateKey)
	assert.NotNil(t, wallet.PublicKey)
	assert.NotEmpty(t, wallet.Address)
}

func TestGetBalance(t *testing.T) {
	wallet, _ := NewWallet()
	bc := blockchain.NewBlockchain()

	// Test balance for each coin type
	for _, coinType := range []blockchain.CoinType{
		blockchain.Leah, blockchain.Shiblum, blockchain.Shiblon,
		blockchain.Senine, blockchain.Seon, blockchain.Shum,
		blockchain.Limnah, blockchain.Antion, blockchain.Senum,
		blockchain.Amnor, blockchain.Ezrom, blockchain.Onti,
	} {
		balance := wallet.GetBalance(coinType, bc)
		assert.Equal(t, 0.0, balance)
	}
}

func TestCreateTransaction(t *testing.T) {
	wallet, _ := NewWallet()
	bc := blockchain.NewBlockchain()

	// Test invalid amount
	_, err := wallet.CreateTransaction("recipient", -1, blockchain.Leah, bc)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)

	// Test invalid address
	_, err = wallet.CreateTransaction("invalid", 1, blockchain.Leah, bc)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAddress, err)

	// Test insufficient funds
	_, err = wallet.CreateTransaction(wallet.Address, 1, blockchain.Leah, bc)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientFunds, err)
}

func TestBackupAndRestore(t *testing.T) {
	wallet, _ := NewWallet()
	tempDir := t.TempDir()
	backupPath := filepath.Join(tempDir, "wallet.backup")

	// Test backup
	err := wallet.Backup(backupPath)
	assert.NoError(t, err)
	assert.FileExists(t, backupPath)

	// Test restore
	restoredWallet, err := Restore(backupPath)
	assert.NoError(t, err)
	assert.NotNil(t, restoredWallet)
	assert.Equal(t, wallet.Address, restoredWallet.Address)

	// Test invalid backup
	err = os.WriteFile(backupPath, []byte("invalid"), 0600)
	assert.NoError(t, err)
	_, err = Restore(backupPath)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidBackup, err)
}

func TestSignAndVerifyMessage(t *testing.T) {
	wallet, _ := NewWallet()
	message := []byte("test message")

	// Test signing
	signature, err := wallet.SignMessage(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// Test verification
	valid := wallet.VerifyMessage(message, signature)
	assert.True(t, valid)

	// Test invalid message
	invalid := wallet.VerifyMessage([]byte("wrong message"), signature)
	assert.False(t, invalid)

	// Test invalid signature
	invalid = wallet.VerifyMessage(message, []byte("invalid signature"))
	assert.False(t, invalid)
}

func TestExportAndImportPublicKey(t *testing.T) {
	wallet, _ := NewWallet()

	// Test export
	publicKeyBytes := wallet.ExportPublicKey()
	assert.NotEmpty(t, publicKeyBytes)

	// Test import
	publicKey, err := ImportPublicKey(publicKeyBytes)
	assert.NoError(t, err)
	assert.NotNil(t, publicKey)
	assert.Equal(t, wallet.PublicKey.X, publicKey.X)
	assert.Equal(t, wallet.PublicKey.Y, publicKey.Y)

	// Test invalid public key
	_, err = ImportPublicKey([]byte("invalid"))
	assert.Error(t, err)
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
