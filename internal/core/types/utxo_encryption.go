package types

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const (
	// DefaultKeySize is the default key size in bytes
	DefaultKeySize = 32 // 256 bits
	// DefaultNonceSize is the default nonce size in bytes
	DefaultNonceSize = 12 // 96 bits
	// DefaultTagSize is the default authentication tag size in bytes
	DefaultTagSize = 16 // 128 bits
)

// EncryptionMethod represents the encryption method to use
type EncryptionMethod byte

const (
	// EncryptionMethodNone indicates no encryption
	EncryptionMethodNone EncryptionMethod = iota
	// EncryptionMethodAESGCM indicates AES-GCM encryption
	EncryptionMethodAESGCM
)

// UTXOEncryption handles encryption of the UTXO set
type UTXOEncryption struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Encryption state
	method    EncryptionMethod
	key       []byte
	nonceSize int
	tagSize   int
}

// NewUTXOEncryption creates a new UTXO encryption handler
func NewUTXOEncryption(utxoSet *UTXOSet) *UTXOEncryption {
	return &UTXOEncryption{
		utxoSet:   utxoSet,
		method:    EncryptionMethodNone,
		key:       make([]byte, DefaultKeySize),
		nonceSize: DefaultNonceSize,
		tagSize:   DefaultTagSize,
	}
}

// SetKey sets the encryption key
func (ue *UTXOEncryption) SetKey(key []byte) error {
	if len(key) != DefaultKeySize {
		return fmt.Errorf("invalid key size: expected %d, got %d", DefaultKeySize, len(key))
	}
	ue.mu.Lock()
	ue.key = make([]byte, len(key))
	copy(ue.key, key)
	ue.mu.Unlock()
	return nil
}

// GenerateKey generates a new encryption key
func (ue *UTXOEncryption) GenerateKey() error {
	key := make([]byte, DefaultKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}
	return ue.SetKey(key)
}

// Encrypt encrypts the UTXO set data
func (ue *UTXOEncryption) Encrypt(data []byte) ([]byte, error) {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	// Check if encryption is needed
	if ue.method == EncryptionMethodNone {
		return data, nil
	}

	// Create buffer for encrypted data
	var buf bytes.Buffer

	// Write encryption header
	header := make([]byte, 5)
	header[0] = byte(ue.method)
	binary.LittleEndian.PutUint32(header[1:5], uint32(len(data)))
	if _, err := buf.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write encryption header: %v", err)
	}

	// Encrypt data based on method
	switch ue.method {
	case EncryptionMethodAESGCM:
		// Create cipher
		block, err := aes.NewCipher(ue.key)
		if err != nil {
			return nil, fmt.Errorf("failed to create cipher: %v", err)
		}

		// Create GCM
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCM: %v", err)
		}

		// Generate nonce
		nonce := make([]byte, ue.nonceSize)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, fmt.Errorf("failed to generate nonce: %v", err)
		}

		// Write nonce
		if _, err := buf.Write(nonce); err != nil {
			return nil, fmt.Errorf("failed to write nonce: %v", err)
		}

		// Encrypt data
		ciphertext := gcm.Seal(nil, nonce, data, nil)
		if _, err := buf.Write(ciphertext); err != nil {
			return nil, fmt.Errorf("failed to write ciphertext: %v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported encryption method: %d", ue.method)
	}

	return buf.Bytes(), nil
}

// Decrypt decrypts the UTXO set data
func (ue *UTXOEncryption) Decrypt(data []byte) ([]byte, error) {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	// Check if data is encrypted
	if len(data) < 5 {
		return data, nil
	}

	// Read encryption header
	method := EncryptionMethod(data[0])
	originalSize := binary.LittleEndian.Uint32(data[1:5])
	encryptedData := data[5:]

	// Check if decryption is needed
	if method == EncryptionMethodNone {
		return encryptedData, nil
	}

	// Decrypt data based on method
	switch method {
	case EncryptionMethodAESGCM:
		// Check data size
		if len(encryptedData) < ue.nonceSize {
			return nil, fmt.Errorf("encrypted data too short")
		}

		// Extract nonce and ciphertext
		nonce := encryptedData[:ue.nonceSize]
		ciphertext := encryptedData[ue.nonceSize:]

		// Create cipher
		block, err := aes.NewCipher(ue.key)
		if err != nil {
			return nil, fmt.Errorf("failed to create cipher: %v", err)
		}

		// Create GCM
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCM: %v", err)
		}

		// Decrypt data
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt data: %v", err)
		}

		// Verify size
		if uint32(len(plaintext)) != originalSize {
			return nil, fmt.Errorf("decrypted size mismatch: expected %d, got %d", originalSize, len(plaintext))
		}

		return plaintext, nil

	default:
		return nil, fmt.Errorf("unsupported encryption method: %d", method)
	}
}

// GetEncryptionStats returns statistics about the encryption
func (ue *UTXOEncryption) GetEncryptionStats(data []byte) (*EncryptionStats, error) {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	// Encrypt data
	encrypted, err := ue.Encrypt(data)
	if err != nil {
		return nil, err
	}

	// Calculate hash
	hash := sha256.Sum256(data)

	return &EncryptionStats{
		OriginalSize:  int64(len(data)),
		EncryptedSize: int64(len(encrypted)),
		Method:        ue.method,
		KeySize:       len(ue.key),
		NonceSize:     ue.nonceSize,
		TagSize:       ue.tagSize,
		Hash:          hash[:],
	}, nil
}

// SetEncryptionMethod sets the encryption method
func (ue *UTXOEncryption) SetEncryptionMethod(method EncryptionMethod) {
	ue.mu.Lock()
	ue.method = method
	ue.mu.Unlock()
}

// SetNonceSize sets the nonce size
func (ue *UTXOEncryption) SetNonceSize(size int) {
	ue.mu.Lock()
	ue.nonceSize = size
	ue.mu.Unlock()
}

// SetTagSize sets the authentication tag size
func (ue *UTXOEncryption) SetTagSize(size int) {
	ue.mu.Lock()
	ue.tagSize = size
	ue.mu.Unlock()
}

// EncryptionStats holds statistics about the encryption
type EncryptionStats struct {
	// OriginalSize is the size of the original data in bytes
	OriginalSize int64
	// EncryptedSize is the size of the encrypted data in bytes
	EncryptedSize int64
	// Method is the encryption method used
	Method EncryptionMethod
	// KeySize is the size of the encryption key in bytes
	KeySize int
	// NonceSize is the size of the nonce in bytes
	NonceSize int
	// TagSize is the size of the authentication tag in bytes
	TagSize int
	// Hash is the SHA-256 hash of the original data
	Hash []byte
}

// String returns a string representation of the encryption statistics
func (es *EncryptionStats) String() string {
	return fmt.Sprintf(
		"Original Size: %d bytes\n"+
			"Encrypted Size: %d bytes\n"+
			"Method: %d\n"+
			"Key Size: %d bytes\n"+
			"Nonce Size: %d bytes\n"+
			"Tag Size: %d bytes",
		es.OriginalSize, es.EncryptedSize,
		es.Method, es.KeySize, es.NonceSize,
		es.TagSize,
	)
}
