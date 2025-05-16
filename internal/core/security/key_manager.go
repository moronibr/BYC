package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// KeyManager manages cryptographic keys securely
type KeyManager struct {
	mu       sync.RWMutex
	keys     map[string]*rsa.PrivateKey
	keyPath  string
	rotation time.Duration
}

// NewKeyManager creates a new key manager
func NewKeyManager(keyPath string, rotation time.Duration) *KeyManager {
	return &KeyManager{
		keys:     make(map[string]*rsa.PrivateKey),
		keyPath:  keyPath,
		rotation: rotation,
	}
}

// GenerateKey generates a new RSA key pair
func (km *KeyManager) GenerateKey(id string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	km.keys[id] = key
	return km.saveKey(id, key)
}

// GetKey retrieves a key by ID
func (km *KeyManager) GetKey(id string) (*rsa.PrivateKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	key, exists := km.keys[id]
	if !exists {
		return nil, errors.New("key not found")
	}
	return key, nil
}

// RotateKey rotates a key by generating a new one and replacing the old one
func (km *KeyManager) RotateKey(id string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to rotate key: %v", err)
	}

	km.keys[id] = key
	return km.saveKey(id, key)
}

// saveKey saves a key to disk
func (km *KeyManager) saveKey(id string, key *rsa.PrivateKey) error {
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	keyFile := fmt.Sprintf("%s/%s.pem", km.keyPath, id)
	return os.WriteFile(keyFile, keyPEM, 0600)
}

// loadKey loads a key from disk
func (km *KeyManager) loadKey(id string) (*rsa.PrivateKey, error) {
	keyFile := fmt.Sprintf("%s/%s.pem", km.keyPath, id)
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %v", err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %v", err)
	}

	return key, nil
}

// StartRotation starts a background process to rotate keys periodically
func (km *KeyManager) StartRotation() {
	go func() {
		ticker := time.NewTicker(km.rotation)
		defer ticker.Stop()

		for range ticker.C {
			for id := range km.keys {
				if err := km.RotateKey(id); err != nil {
					fmt.Printf("Failed to rotate key %s: %v\n", id, err)
				}
			}
		}
	}()
}
