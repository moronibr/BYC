package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"byc/internal/blockchain"
	"byc/internal/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWallet tests wallet creation
func TestNewWallet(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.NotNil(t, wallet.PrivateKey)
	assert.NotNil(t, wallet.PublicKey)
	assert.NotEmpty(t, wallet.Address)
	assert.NotNil(t, wallet.balances)
	assert.NotNil(t, wallet.Transactions)
	assert.NotNil(t, wallet.MultiSigWallets)
	assert.NotNil(t, wallet.AddressBook)
}

// TestHDWallet tests HD wallet functionality
func TestHDWallet(t *testing.T) {
	// Test creation
	wallet, err := NewHDWallet()
	require.NoError(t, err)
	assert.NotNil(t, wallet.HDWallet)
	assert.NotEmpty(t, wallet.HDWallet.Mnemonic)
	assert.NotEmpty(t, wallet.HDWallet.Seed)
	assert.NotEmpty(t, wallet.HDWallet.MasterKey)

	// Test mnemonic retrieval
	mnemonic, err := wallet.GetMnemonic()
	require.NoError(t, err)
	assert.Equal(t, wallet.HDWallet.Mnemonic, mnemonic)

	// Test restoration
	restoredWallet, err := RestoreFromMnemonic(mnemonic)
	require.NoError(t, err)
	assert.Equal(t, wallet.HDWallet.Mnemonic, restoredWallet.HDWallet.Mnemonic)
}

// TestEncryption tests wallet encryption and decryption
func TestEncryption(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Test encryption
	password := "test-password"
	err = wallet.EncryptWallet(password)
	require.NoError(t, err)
	assert.True(t, wallet.Encrypted)
	assert.Nil(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.EncryptedKey)
	assert.NotEmpty(t, wallet.Salt)
	assert.NotEmpty(t, wallet.IV)

	// Test decryption
	err = wallet.DecryptWallet(password)
	require.NoError(t, err)
	assert.False(t, wallet.Encrypted)
	assert.NotNil(t, wallet.PrivateKey)
	assert.Empty(t, wallet.EncryptedKey)

	// Test invalid password
	err = wallet.EncryptWallet(password)
	require.NoError(t, err)
	err = wallet.DecryptWallet("wrong-password")
	assert.Equal(t, ErrInvalidPassword, err)
}

// TestBackupAndRestore tests wallet backup and restore functionality
func TestBackupAndRestore(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Add some test data
	wallet.AddToAddressBook("Test", "test-address", "Test address")
	wallet.balances[blockchain.Leah] = 100

	// Create backup
	backupPath := filepath.Join(t.TempDir(), "wallet.backup")
	err = wallet.Backup(backupPath)
	require.NoError(t, err)

	// Restore from backup
	restoredWallet, err := Restore(backupPath)
	require.NoError(t, err)
	assert.Equal(t, wallet.Address, restoredWallet.Address)
	assert.Equal(t, wallet.balances, restoredWallet.balances)
	assert.Equal(t, wallet.AddressBook, restoredWallet.AddressBook)

	// Test invalid backup
	err = os.WriteFile(backupPath, []byte("invalid data"), 0600)
	require.NoError(t, err)
	_, err = Restore(backupPath)
	assert.Equal(t, ErrInvalidBackup, err)
}

// TestTransactionCreation tests transaction creation and validation
func TestTransactionCreation(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Create a test blockchain
	bc := blockchain.NewBlockchain()

	// Test invalid amount
	_, err = wallet.CreateTransaction("recipient", -1, blockchain.Leah, bc)
	assert.Equal(t, ErrInvalidAmount, err)

	// Test invalid address
	_, err = wallet.CreateTransaction("invalid-address", 1, blockchain.Leah, bc)
	assert.Equal(t, ErrInvalidAddress, err)

	// Test insufficient funds
	_, err = wallet.CreateTransaction("recipient", 1000, blockchain.Leah, bc)
	assert.Equal(t, ErrInsufficientFunds, err)
}

// TestMultiSigWallet tests multi-signature wallet functionality
func TestMultiSigWallet(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Generate test public keys
	var publicKeys [][]byte
	for i := 0; i < 3; i++ {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		publicKeys = append(publicKeys, crypto.PublicKeyToBytes(&privateKey.PublicKey))
	}

	// Test creation
	multiSigWallet, err := wallet.CreateMultiSigWallet(publicKeys, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, multiSigWallet.Address)
	assert.Equal(t, 2, multiSigWallet.Threshold)
	assert.Equal(t, publicKeys, multiSigWallet.PublicKeys)

	// Test invalid threshold
	_, err = wallet.CreateMultiSigWallet(publicKeys, 4)
	assert.Error(t, err)
}

// TestAddressBook tests address book functionality
func TestAddressBook(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Test adding address
	err = wallet.AddToAddressBook("Test", "test-address", "Test address")
	assert.NoError(t, err)

	// Test invalid address
	err = wallet.AddToAddressBook("Test", "invalid-address", "Test address")
	assert.Equal(t, ErrInvalidAddress, err)

	// Test getting address book
	book := wallet.GetAddressBook()
	assert.NotEmpty(t, book)
	assert.Equal(t, "Test", book["test-address"].Name)
}

// TestSpecialCoins tests special coin conversion functionality
func TestSpecialCoins(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)
	bc := blockchain.NewBlockchain()

	// Test insufficient funds
	err = wallet.CreateEphraimCoin(bc)
	assert.Error(t, err)
	err = wallet.CreateManassehCoin(bc)
	assert.Error(t, err)
	err = wallet.CreateJosephCoin(bc)
	assert.Error(t, err)
}

// BenchmarkWalletCreation benchmarks wallet creation
func BenchmarkWalletCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := NewWallet()
		require.NoError(b, err)
	}
}

// BenchmarkEncryption benchmarks wallet encryption
func BenchmarkEncryption(b *testing.B) {
	wallet, err := NewWallet()
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := wallet.EncryptWallet("test-password")
		require.NoError(b, err)
		err = wallet.DecryptWallet("test-password")
		require.NoError(b, err)
	}
}

// BenchmarkTransactionCreation benchmarks transaction creation
func BenchmarkTransactionCreation(b *testing.B) {
	wallet, err := NewWallet()
	require.NoError(b, err)
	bc := blockchain.NewBlockchain()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.CreateTransaction("recipient", 1, blockchain.Leah, bc)
		if err != nil && err != ErrInsufficientFunds {
			require.NoError(b, err)
		}
	}
}

// TestErrorTypes tests specific error types
func TestErrorTypes(t *testing.T) {
	assert.Equal(t, "insufficient funds", ErrInsufficientFunds.Error())
	assert.Equal(t, "invalid amount", ErrInvalidAmount.Error())
	assert.Equal(t, "invalid address", ErrInvalidAddress.Error())
	assert.Equal(t, "invalid backup data", ErrInvalidBackup.Error())
	assert.Equal(t, "invalid password", ErrInvalidPassword.Error())
	assert.Equal(t, "invalid signature", ErrInvalidSignature.Error())
	assert.Equal(t, "invalid mnemonic", ErrInvalidMnemonic.Error())
}

// TestRecoveryMechanisms tests recovery mechanisms
func TestRecoveryMechanisms(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Test recovery from backup
	backupPath := filepath.Join(t.TempDir(), "wallet.backup")
	err = wallet.Backup(backupPath)
	require.NoError(t, err)

	// Simulate wallet corruption
	wallet.PrivateKey = nil
	wallet.PublicKey = nil

	// Restore from backup
	restoredWallet, err := Restore(backupPath)
	require.NoError(t, err)
	assert.NotNil(t, restoredWallet.PrivateKey)
	assert.NotNil(t, restoredWallet.PublicKey)

	// Test recovery from mnemonic
	hdWallet, err := NewHDWallet()
	require.NoError(t, err)
	mnemonic, err := hdWallet.GetMnemonic()
	require.NoError(t, err)

	// Simulate wallet loss
	hdWallet.PrivateKey = nil
	hdWallet.PublicKey = nil

	// Restore from mnemonic
	restoredHDWallet, err := RestoreFromMnemonic(mnemonic)
	require.NoError(t, err)
	assert.NotNil(t, restoredHDWallet.PrivateKey)
	assert.NotNil(t, restoredHDWallet.PublicKey)
}
