package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// Key derivation parameters
	saltSize     = 32
	iterations   = 2048
	keySize      = 32
	ivSize       = 16
	checksumSize = 4
)

// EncryptedWallet represents an encrypted wallet
type EncryptedWallet struct {
	// Encrypted master key
	encryptedKey []byte
	// Salt for key derivation
	salt []byte
	// IV for AES encryption
	iv []byte
	// Checksum of the master key
	checksum []byte
}

// NewEncryptedWallet creates a new encrypted wallet
func NewEncryptedWallet(masterKey *bip32.Key, password string) (*EncryptedWallet, error) {
	// Generate salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	// Generate IV
	iv := make([]byte, ivSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %v", err)
	}

	// Derive encryption key
	key := pbkdf2.Key([]byte(password), salt, iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Encrypt master key
	encryptedKey := gcm.Seal(nil, iv, masterKey.Key, nil)

	// Calculate checksum
	checksum := sha256.Sum256(masterKey.Key)

	return &EncryptedWallet{
		encryptedKey: encryptedKey,
		salt:         salt,
		iv:           iv,
		checksum:     checksum[:checksumSize],
	}, nil
}

// Decrypt decrypts the wallet
func (w *EncryptedWallet) Decrypt(password string) (*bip32.Key, error) {
	// Derive encryption key
	key := pbkdf2.Key([]byte(password), w.salt, iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Decrypt master key
	masterKey, err := gcm.Open(nil, w.iv, w.encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key: %v", err)
	}

	// Verify checksum
	checksum := sha256.Sum256(masterKey)
	if !bytes.Equal(checksum[:checksumSize], w.checksum) {
		return nil, fmt.Errorf("invalid checksum")
	}

	// Create BIP32 key
	return &bip32.Key{
		Key: masterKey,
	}, nil
}

// GenerateMnemonic generates a new mnemonic phrase
func GenerateMnemonic() (string, error) {
	// Generate entropy
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %v", err)
	}

	// Generate mnemonic
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %v", err)
	}

	return mnemonic, nil
}

// ValidateMnemonic validates a mnemonic phrase
func ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// DeriveFromMnemonic derives a master key from a mnemonic
func DeriveFromMnemonic(mnemonic string) (*bip32.Key, error) {
	// Generate seed
	seed := bip39.NewSeed(mnemonic, "")

	// Generate master key
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %v", err)
	}

	return masterKey, nil
}

// DeriveChild derives a child key
func DeriveChild(masterKey *bip32.Key, path string) (*bip32.Key, error) {
	// Parse path
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path")
	}

	// Start with master key
	key := masterKey

	// Derive each part
	for _, part := range parts[1:] {
		// Parse index
		index := uint32(0)
		if strings.HasSuffix(part, "'") {
			// Hardened derivation
			index = 0x80000000
			part = part[:len(part)-1]
		}
		i, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path component: %s", part)
		}
		index += uint32(i)

		// Derive child
		child, err := key.NewChildKey(index)
		if err != nil {
			return nil, fmt.Errorf("failed to derive child key: %v", err)
		}
		key = child
	}

	return key, nil
}

// BackupWallet creates a backup of the wallet
func BackupWallet(wallet *EncryptedWallet) (string, error) {
	// Serialize wallet
	data, err := json.Marshal(wallet)
	if err != nil {
		return "", fmt.Errorf("failed to serialize wallet: %v", err)
	}

	// Encode as hex
	return hex.EncodeToString(data), nil
}

// RestoreWallet restores a wallet from backup
func RestoreWallet(backup string) (*EncryptedWallet, error) {
	// Decode hex
	data, err := hex.DecodeString(backup)
	if err != nil {
		return nil, fmt.Errorf("failed to decode backup: %v", err)
	}

	// Deserialize wallet
	var wallet EncryptedWallet
	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("failed to deserialize wallet: %v", err)
	}

	return &wallet, nil
}
