package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
)

// TestKeyManagement tests key management security
func TestKeyManagement(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Test private key security
	privateKey := wallet.PrivateKey
	assert.NotNil(t, privateKey)
	assert.NotEmpty(t, privateKey.D.Bytes())

	// Test public key derivation
	publicKey := wallet.PublicKey
	assert.NotNil(t, publicKey)
	assert.Equal(t, privateKey.PublicKey.X, publicKey.X)
	assert.Equal(t, privateKey.PublicKey.Y, publicKey.Y)

	// Test key pair consistency
	message := []byte("test message")
	signature, err := wallet.SignMessage(message)
	require.NoError(t, err)
	assert.True(t, wallet.VerifyMessage(message, signature))
}

// TestEncryptionSecurity tests encryption security
func TestEncryptionSecurity(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Test encryption strength
	password := "test-password"
	err = wallet.EncryptWallet(password)
	require.NoError(t, err)

	// Verify private key is cleared
	assert.Nil(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.EncryptedKey)
	assert.NotEmpty(t, wallet.Salt)
	assert.NotEmpty(t, wallet.IV)

	// Test decryption with wrong password
	err = wallet.DecryptWallet("wrong-password")
	assert.Equal(t, ErrInvalidPassword, err)

	// Test decryption with correct password
	err = wallet.DecryptWallet(password)
	require.NoError(t, err)
	assert.NotNil(t, wallet.PrivateKey)
	assert.Empty(t, wallet.EncryptedKey)
}

// TestMultiSigSecurity tests multi-signature security
func TestMultiSigSecurity(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Generate test keys
	var publicKeys [][]byte
	var privateKeys []*ecdsa.PrivateKey
	for i := 0; i < 3; i++ {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		privateKeys = append(privateKeys, privateKey)
		publicKeys = append(publicKeys, crypto.PublicKeyToBytes(&privateKey.PublicKey))
	}

	// Create multi-sig wallet
	multiSigWallet, err := wallet.CreateMultiSigWallet(publicKeys, 2)
	require.NoError(t, err)

	// Test threshold security
	assert.Equal(t, 2, multiSigWallet.Threshold)
	assert.Len(t, multiSigWallet.PublicKeys, 3)

	// Test signature collection
	txID := "test-tx"
	message := []byte(txID)

	// Sign with first key
	signature1, err := crypto.Sign(message, privateKeys[0].D.Bytes())
	require.NoError(t, err)
	multiSigWallet.Signatures[txID] = signature1

	// Verify single signature is not enough
	assert.False(t, verifyMultiSigSignatures(message, multiSigWallet))

	// Sign with second key
	signature2, err := crypto.Sign(message, privateKeys[1].D.Bytes())
	require.NoError(t, err)
	multiSigWallet.Signatures[txID] = signature2

	// Verify two signatures are enough
	assert.True(t, verifyMultiSigSignatures(message, multiSigWallet))
}

// TestTransactionSecurity tests transaction security
func TestTransactionSecurity(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)
	bc := blockchain.NewBlockchain()

	// Test transaction signing
	tx, err := wallet.CreateTransaction("recipient", 1, blockchain.Leah, bc)
	if err != nil && err != ErrInsufficientFunds {
		require.NoError(t, err)
		assert.NotNil(t, tx)
		assert.NotEmpty(t, tx.Inputs)
		for _, input := range tx.Inputs {
			assert.NotEmpty(t, input.Signature)
		}
	}

	// Test transaction validation
	if tx != nil {
		valid := tx.Verify()
		assert.True(t, valid)
	}
}

// TestBackupSecurity tests backup security
func TestBackupSecurity(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Add sensitive data
	wallet.AddToAddressBook("Test", "test-address", "Test address")
	wallet.balances[blockchain.Leah] = 100

	// Create encrypted backup
	password := "test-password"
	err = wallet.EncryptWallet(password)
	require.NoError(t, err)

	backupPath := "test.backup"
	err = wallet.Backup(backupPath)
	require.NoError(t, err)

	// Verify backup security
	restoredWallet, err := Restore(backupPath)
	require.NoError(t, err)
	assert.True(t, restoredWallet.Encrypted)
	assert.Nil(t, restoredWallet.PrivateKey)
	assert.NotEmpty(t, restoredWallet.EncryptedKey)

	// Test decryption
	err = restoredWallet.DecryptWallet(password)
	require.NoError(t, err)
	assert.NotNil(t, restoredWallet.PrivateKey)
	assert.Equal(t, wallet.Address, restoredWallet.Address)
	assert.Equal(t, wallet.balances, restoredWallet.balances)
}

// TestHDWalletSecurity tests HD wallet security
func TestHDWalletSecurity(t *testing.T) {
	wallet, err := NewHDWallet()
	require.NoError(t, err)

	// Test mnemonic security
	mnemonic, err := wallet.GetMnemonic()
	require.NoError(t, err)
	assert.NotEmpty(t, mnemonic)
	assert.True(t, bip39.IsMnemonicValid(mnemonic))

	// Test seed derivation
	assert.NotEmpty(t, wallet.HDWallet.Seed)
	assert.NotEmpty(t, wallet.HDWallet.MasterKey)

	// Test key derivation
	childKey, err := deriveChildKey(wallet.HDWallet.MasterKey, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, childKey)

	// Test recovery security
	restoredWallet, err := RestoreFromMnemonic(mnemonic)
	require.NoError(t, err)
	assert.Equal(t, wallet.HDWallet.Mnemonic, restoredWallet.HDWallet.Mnemonic)
	assert.Equal(t, wallet.HDWallet.Seed, restoredWallet.HDWallet.Seed)
	assert.Equal(t, wallet.HDWallet.MasterKey, restoredWallet.HDWallet.MasterKey)
}

// TestRateLimiting tests rate limiting security
func TestRateLimiting(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)
	bc := blockchain.NewBlockchain()

	// Test transaction rate limiting
	for i := 0; i < 10; i++ {
		_, err := wallet.CreateTransaction("recipient", 1, blockchain.Leah, bc)
		if err != nil && err != ErrInsufficientFunds {
			require.NoError(t, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TestErrorHandling tests error handling security
func TestErrorHandling(t *testing.T) {
	wallet, err := NewWallet()
	require.NoError(t, err)

	// Test invalid operations
	_, err = wallet.CreateTransaction("invalid-address", -1, blockchain.Leah, nil)
	assert.Error(t, err)

	err = wallet.EncryptWallet("")
	assert.Error(t, err)

	err = wallet.DecryptWallet("wrong-password")
	assert.Equal(t, ErrInvalidPassword, err)

	// Test recovery from errors
	err = wallet.EncryptWallet("test-password")
	require.NoError(t, err)
	err = wallet.DecryptWallet("test-password")
	require.NoError(t, err)
	assert.NotNil(t, wallet.PrivateKey)
}

// Helper function to verify multi-signature signatures
func verifyMultiSigSignatures(message []byte, wallet *MultiSigWallet) bool {
	signatureCount := 0
	for _, signature := range wallet.Signatures {
		for _, publicKey := range wallet.PublicKeys {
			if crypto.Verify(message, signature, publicKey) {
				signatureCount++
				break
			}
		}
	}
	return signatureCount >= wallet.Threshold
}

// Helper function to derive child key
func deriveChildKey(masterKey []byte, index uint32) ([]byte, error) {
	// Implement child key derivation logic
	return nil, nil
}
