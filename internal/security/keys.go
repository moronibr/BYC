package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	// Key rotation periods
	KeyRotationPeriod = 24 * time.Hour
	MaxKeyAge         = 7 * 24 * time.Hour

	// Encryption parameters
	SaltLength  = 16
	KeyLength   = 32
	Iterations  = 3
	Memory      = 64 * 1024
	Threads     = 4
	NonceLength = 12
	TagLength   = 16
)

// KeyManager handles key generation, rotation, and encryption
type KeyManager struct {
	mu            sync.RWMutex
	currentKey    []byte
	previousKey   []byte
	keyCreatedAt  time.Time
	keyRotatedAt  time.Time
	encryptionKey []byte
}

// NewKeyManager creates a new key manager
func NewKeyManager() (*KeyManager, error) {
	km := &KeyManager{}
	if err := km.rotateKeys(); err != nil {
		return nil, err
	}
	return km, nil
}

// rotateKeys generates new encryption keys
func (km *KeyManager) rotateKeys() error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Generate new key
	newKey := make([]byte, KeyLength)
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return fmt.Errorf("failed to generate new key: %v", err)
	}

	// Store previous key and update current key
	km.previousKey = km.currentKey
	km.currentKey = newKey
	km.keyCreatedAt = time.Now()
	km.keyRotatedAt = time.Now()

	return nil
}

// shouldRotateKeys checks if keys need to be rotated
func (km *KeyManager) shouldRotateKeys() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()

	return time.Since(km.keyRotatedAt) >= KeyRotationPeriod ||
		time.Since(km.keyCreatedAt) >= MaxKeyAge
}

// Encrypt encrypts data using AES-GCM
func (km *KeyManager) Encrypt(data []byte) ([]byte, error) {
	if km.shouldRotateKeys() {
		if err := km.rotateKeys(); err != nil {
			return nil, err
		}
	}

	km.mu.RLock()
	defer km.mu.RUnlock()

	block, err := aes.NewCipher(km.currentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-GCM
func (km *KeyManager) Decrypt(data []byte) ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Try current key first
	plaintext, err := km.decryptWithKey(data, km.currentKey)
	if err == nil {
		return plaintext, nil
	}

	// Try previous key if current key fails
	if km.previousKey != nil {
		plaintext, err = km.decryptWithKey(data, km.previousKey)
		if err == nil {
			return plaintext, nil
		}
	}

	return nil, errors.New("decryption failed with all available keys")
}

// decryptWithKey attempts to decrypt data with a specific key
func (km *KeyManager) decryptWithKey(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// DeriveKey derives an encryption key from a password
func DeriveKey(password string, salt []byte) []byte {
	return argon2.Key(
		[]byte(password),
		salt,
		Iterations,
		Memory,
		Threads,
		KeyLength,
	)
}

// HashPassword hashes a password using Argon2
func HashPassword(password string) (string, error) {
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	hash := argon2.Key(
		[]byte(password),
		salt,
		Iterations,
		Memory,
		Threads,
		KeyLength,
	)

	// Combine salt and hash
	combined := make([]byte, SaltLength+KeyLength)
	copy(combined, salt)
	copy(combined[SaltLength:], hash)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	decoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}

	if len(decoded) != SaltLength+KeyLength {
		return false
	}

	salt := decoded[:SaltLength]
	expectedHash := decoded[SaltLength:]

	actualHash := argon2.Key(
		[]byte(password),
		salt,
		Iterations,
		Memory,
		Threads,
		KeyLength,
	)

	return sha256.Sum256(actualHash) == sha256.Sum256(expectedHash)
}
