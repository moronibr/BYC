package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"time"

	"golang.org/x/crypto/argon2"
)

// EncryptionConfig holds encryption parameters
type EncryptionConfig struct {
	// Argon2 parameters
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	// PBKDF2 parameters
	Iterations int
	SaltLen    int
	// AES parameters
	KeySize int
}

// DefaultEncryptionConfig returns default encryption parameters
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		Time:       1,
		Memory:     64 * 1024, // 64MB
		Threads:    4,
		KeyLen:     32,
		Iterations: 100000,
		SaltLen:    32,
		KeySize:    32,
	}
}

// EncryptedWallet represents an encrypted wallet
type EncryptedWallet struct {
	// Encrypted data
	Data []byte
	// Salt used for key derivation
	Salt []byte
	// IV for AES encryption
	IV []byte
	// Hash of the master key for verification
	KeyHash []byte
	// Timestamp of encryption
	Timestamp int64
	// Version of encryption
	Version int
}

// EncryptWallet encrypts a wallet with a password
func EncryptWallet(wallet *Wallet, password string, config *EncryptionConfig) (*EncryptedWallet, error) {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	// Generate salt
	salt := make([]byte, config.SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, &EncryptionError{
			Operation: "generate_salt",
			Reason:    err.Error(),
		}
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, &EncryptionError{
			Operation: "generate_iv",
			Reason:    err.Error(),
		}
	}

	// Derive encryption key using Argon2
	key := argon2.Key(
		[]byte(password),
		salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &EncryptionError{
			Operation: "create_cipher",
			Reason:    err.Error(),
		}
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &EncryptionError{
			Operation: "create_gcm",
			Reason:    err.Error(),
		}
	}

	// Serialize wallet
	walletData, err := wallet.Serialize()
	if err != nil {
		return nil, &EncryptionError{
			Operation: "serialize_wallet",
			Reason:    err.Error(),
		}
	}

	// Encrypt wallet data
	encryptedData := gcm.Seal(nil, iv, walletData, nil)

	// Generate key hash for verification
	keyHash := sha256.Sum256(key)

	return &EncryptedWallet{
		Data:      encryptedData,
		Salt:      salt,
		IV:        iv,
		KeyHash:   keyHash[:],
		Timestamp: time.Now().Unix(),
		Version:   1,
	}, nil
}

// DecryptWallet decrypts a wallet with a password
func DecryptWallet(encrypted *EncryptedWallet, password string, config *EncryptionConfig) (*Wallet, error) {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	// Derive encryption key using Argon2
	key := argon2.Key(
		[]byte(password),
		encrypted.Salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	// Verify key hash
	keyHash := sha256.Sum256(key)
	if !bytes.Equal(keyHash[:], encrypted.KeyHash) {
		return nil, ErrInvalidPassword
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &EncryptionError{
			Operation: "create_cipher",
			Reason:    err.Error(),
		}
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &EncryptionError{
			Operation: "create_gcm",
			Reason:    err.Error(),
		}
	}

	// Decrypt wallet data
	walletData, err := gcm.Open(nil, encrypted.IV, encrypted.Data, nil)
	if err != nil {
		return nil, &EncryptionError{
			Operation: "decrypt_data",
			Reason:    err.Error(),
		}
	}

	// Deserialize wallet
	wallet := &Wallet{}
	if err := wallet.Deserialize(walletData); err != nil {
		return nil, &EncryptionError{
			Operation: "deserialize_wallet",
			Reason:    err.Error(),
		}
	}

	return wallet, nil
}

// ChangePassword changes the encryption password
func ChangePassword(encrypted *EncryptedWallet, oldPassword, newPassword string, config *EncryptionConfig) (*EncryptedWallet, error) {
	// Decrypt with old password
	wallet, err := DecryptWallet(encrypted, oldPassword, config)
	if err != nil {
		return nil, err
	}

	// Encrypt with new password
	return EncryptWallet(wallet, newPassword, config)
}

// VerifyPassword verifies if a password is correct
func VerifyPassword(encrypted *EncryptedWallet, password string, config *EncryptionConfig) bool {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	// Derive encryption key using Argon2
	key := argon2.Key(
		[]byte(password),
		encrypted.Salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	// Verify key hash
	keyHash := sha256.Sum256(key)
	return bytes.Equal(keyHash[:], encrypted.KeyHash)
}

// BackupWallet creates an encrypted backup of a wallet
func BackupWallet(wallet *Wallet, password string, config *EncryptionConfig) (*EncryptedWallet, error) {
	// Add backup metadata
	wallet.BackupTime = time.Now().Unix()
	wallet.BackupVersion = 1

	// Encrypt wallet
	return EncryptWallet(wallet, password, config)
}

// RestoreWallet restores a wallet from an encrypted backup
func RestoreWallet(encrypted *EncryptedWallet, password string, config *EncryptionConfig) (*Wallet, error) {
	// Decrypt wallet
	wallet, err := DecryptWallet(encrypted, password, config)
	if err != nil {
		return nil, err
	}

	// Verify backup metadata
	if wallet.BackupTime == 0 || wallet.BackupVersion == 0 {
		return nil, ErrInvalidBackup
	}

	return wallet, nil
}
