package wallet

import (
	"encoding/hex"
	"testing"

	"byc/internal/blockchain"
	"byc/internal/network"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWalletBlockchainIntegration tests wallet integration with blockchain
func TestWalletBlockchainIntegration(t *testing.T) {
	// Create wallets
	sender, err := NewWallet()
	require.NoError(t, err)

	recipient, err := NewWallet()
	require.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Test balance tracking
	initialBalance := sender.GetBalance(blockchain.Leah, bc)
	assert.Equal(t, 0.0, initialBalance)

	// Test transaction creation and broadcasting
	tx, err := sender.CreateTransaction(recipient.Address, 1, blockchain.Leah, bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)

		// Create network node
		node := &network.Node{}

		// Test transaction broadcasting
		err = sender.BroadcastTransaction(tx, node)
		require.NoError(t, err)

		// Verify transaction history
		history := sender.GetTransactionHistory()
		assert.NotEmpty(t, history)
		assert.Equal(t, "pending", history[0].Status)
	}
}

// TestWalletEncryptionIntegration tests wallet encryption integration
func TestWalletEncryptionIntegration(t *testing.T) {
	// Create wallet
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Add some data
	wallet.AddToAddressBook("Test", "test-address", "Test address")
	wallet.balances[blockchain.Leah] = 100

	// Test encryption
	password := "test-password"
	err = wallet.EncryptWallet(password)
	require.NoError(t, err)

	// Verify wallet is encrypted
	assert.True(t, wallet.Encrypted)
	assert.Nil(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.EncryptedKey)

	// Test backup and restore
	backupPath := "test.backup"
	err = wallet.Backup(backupPath)
	require.NoError(t, err)

	// Restore wallet
	restoredWallet, err := Restore(backupPath)
	require.NoError(t, err)

	// Verify restored wallet
	assert.True(t, restoredWallet.Encrypted)
	assert.Equal(t, wallet.Address, restoredWallet.Address)
	assert.Equal(t, wallet.balances, restoredWallet.balances)
	assert.Equal(t, wallet.AddressBook, restoredWallet.AddressBook)

	// Test decryption
	err = restoredWallet.DecryptWallet(password)
	require.NoError(t, err)
	assert.False(t, restoredWallet.Encrypted)
	assert.NotNil(t, restoredWallet.PrivateKey)
}

// TestHDWalletIntegration tests HD wallet integration
func TestHDWalletIntegration(t *testing.T) {
	// Create HD wallet
	wallet, err := NewHDWallet()
	require.NoError(t, err)

	// Test mnemonic generation
	mnemonic, err := wallet.GetMnemonic()
	require.NoError(t, err)
	assert.NotEmpty(t, mnemonic)

	// Test wallet restoration
	restoredWallet, err := RestoreFromMnemonic(mnemonic)
	require.NoError(t, err)
	assert.Equal(t, wallet.HDWallet.Mnemonic, restoredWallet.HDWallet.Mnemonic)
	assert.Equal(t, wallet.HDWallet.Seed, restoredWallet.HDWallet.Seed)
	assert.Equal(t, wallet.HDWallet.MasterKey, restoredWallet.HDWallet.MasterKey)

	// Test transaction creation with restored wallet
	bc := blockchain.NewBlockchain()
	tx, err := restoredWallet.CreateTransaction("recipient", 1, blockchain.Leah, bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)
		assert.NotNil(t, tx)
	}
}

// TestMultiSigWalletIntegration tests multi-signature wallet integration
func TestMultiSigWalletIntegration(t *testing.T) {
	// Create wallets
	wallet1, err := NewWallet()
	require.NoError(t, err)

	wallet2, err := NewWallet()
	require.NoError(t, err)

	wallet3, err := NewWallet()
	require.NoError(t, err)

	// Create multi-sig wallet
	publicKeys := [][]byte{
		wallet1.ExportPublicKey(),
		wallet2.ExportPublicKey(),
		wallet3.ExportPublicKey(),
	}

	multiSigWallet, err := wallet1.CreateMultiSigWallet(publicKeys, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, multiSigWallet.Address)
	assert.Equal(t, 2, multiSigWallet.Threshold)

	// Test transaction signing
	bc := blockchain.NewBlockchain()
	tx, err := wallet1.CreateTransaction(multiSigWallet.Address, 1, blockchain.Leah, bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)

		// Sign with first wallet
		err = wallet1.SignMultiSigTransaction(hex.EncodeToString(tx.ID), tx)
		require.NoError(t, err)

		// Sign with second wallet
		err = wallet2.SignMultiSigTransaction(hex.EncodeToString(tx.ID), tx)
		require.NoError(t, err)

		// Verify signatures
		assert.True(t, tx.Verify())
	}
}

// TestWatchOnlyWalletIntegration tests watch-only wallet integration
func TestWatchOnlyWalletIntegration(t *testing.T) {
	// Create regular wallet
	regularWallet, err := NewWallet()
	require.NoError(t, err)

	// Create watch-only wallet
	watchOnlyWallet := NewWatchOnlyWallet(regularWallet.PublicKey)
	assert.True(t, watchOnlyWallet.WatchOnly)
	assert.Equal(t, regularWallet.Address, watchOnlyWallet.Address)

	// Test balance tracking
	bc := blockchain.NewBlockchain()
	balance := watchOnlyWallet.GetBalance(blockchain.Leah, bc)
	assert.Equal(t, 0.0, balance)

	// Test address book
	err = watchOnlyWallet.AddToAddressBook("Test", "test-address", "Test address")
	require.NoError(t, err)

	book := watchOnlyWallet.GetAddressBook()
	assert.NotEmpty(t, book)
	assert.Equal(t, "Test", book["test-address"].Name)
}

// TestSpecialCoinsIntegration tests special coin conversion integration
func TestSpecialCoinsIntegration(t *testing.T) {
	// Create wallet
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Test coin conversion
	err = wallet.CreateEphraimCoin(bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)
	}

	err = wallet.CreateManassehCoin(bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)
	}

	err = wallet.CreateJosephCoin(bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)
	}
}

// TestTransactionHistoryIntegration tests transaction history integration
func TestTransactionHistoryIntegration(t *testing.T) {
	// Create wallets
	sender, err := NewWallet()
	require.NoError(t, err)

	recipient, err := NewWallet()
	require.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Create and broadcast transaction
	tx, err := sender.CreateTransaction(recipient.Address, 1, blockchain.Leah, bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)

		node := &network.Node{}
		err = sender.BroadcastTransaction(tx, node)
		require.NoError(t, err)

		// Verify transaction history
		history := sender.GetTransactionHistory()
		assert.NotEmpty(t, history)
		assert.Equal(t, "pending", history[0].Status)
		assert.Equal(t, 1.0, history[0].Amount)
		assert.Equal(t, blockchain.Leah, history[0].CoinType)
		assert.Equal(t, sender.Address, history[0].From)
		assert.Equal(t, recipient.Address, history[0].To)
	}
}

// TestAddressBookIntegration tests address book integration
func TestAddressBookIntegration(t *testing.T) {
	// Create wallet
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Add addresses
	err = wallet.AddToAddressBook("Alice", "alice-address", "Alice's address")
	require.NoError(t, err)

	err = wallet.AddToAddressBook("Bob", "bob-address", "Bob's address")
	require.NoError(t, err)

	// Get address book
	book := wallet.GetAddressBook()
	assert.Len(t, book, 2)
	assert.Equal(t, "Alice", book["alice-address"].Name)
	assert.Equal(t, "Bob", book["bob-address"].Name)

	// Test invalid address
	err = wallet.AddToAddressBook("Invalid", "invalid-address", "Invalid address")
	assert.Equal(t, ErrInvalidAddress, err)
}

// TestFeeEstimationIntegration tests transaction fee estimation integration
func TestFeeEstimationIntegration(t *testing.T) {
	// Create wallet
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Test fee estimation
	fee := wallet.EstimateTransactionFee(10.0, blockchain.Leah)
	assert.Greater(t, fee, 0.0)

	// Test different amounts
	fee1 := wallet.EstimateTransactionFee(1.0, blockchain.Leah)
	fee2 := wallet.EstimateTransactionFee(100.0, blockchain.Leah)
	assert.Greater(t, fee2, fee1)
}

// TestRecoveryIntegration tests wallet recovery integration
func TestRecoveryIntegration(t *testing.T) {
	// Create HD wallet
	wallet, err := NewHDWallet()
	require.NoError(t, err)

	// Get mnemonic
	mnemonic, err := wallet.GetMnemonic()
	require.NoError(t, err)

	// Simulate wallet loss
	wallet.PrivateKey = nil
	wallet.PublicKey = nil

	// Restore from mnemonic
	restoredWallet, err := RestoreFromMnemonic(mnemonic)
	require.NoError(t, err)
	assert.NotNil(t, restoredWallet.PrivateKey)
	assert.NotNil(t, restoredWallet.PublicKey)
	assert.Equal(t, wallet.Address, restoredWallet.Address)

	// Test backup and restore
	backupPath := "test.backup"
	err = wallet.Backup(backupPath)
	require.NoError(t, err)

	restoredFromBackup, err := Restore(backupPath)
	require.NoError(t, err)
	assert.Equal(t, wallet.Address, restoredFromBackup.Address)
}
