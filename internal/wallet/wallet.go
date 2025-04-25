package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"

	"github.com/youngchain/internal/core/block"
)

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

// WalletManager manages multiple wallets
type WalletManager struct {
	wallets map[string]*Wallet
	mu      sync.RWMutex
}

// NewWalletManager creates a new wallet manager
func NewWalletManager() *WalletManager {
	return &WalletManager{
		wallets: make(map[string]*Wallet),
	}
}

// CreateWallet creates a new wallet
func (wm *WalletManager) CreateWallet() (*Wallet, error) {
	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Create wallet
	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		Address:    generateAddress(&privateKey.PublicKey),
	}

	// Store wallet
	wm.mu.Lock()
	wm.wallets[wallet.Address] = wallet
	wm.mu.Unlock()

	return wallet, nil
}

// GetWallet returns a wallet by address
func (wm *WalletManager) GetWallet(address string) (*Wallet, error) {
	wm.mu.RLock()
	wallet, exists := wm.wallets[address]
	wm.mu.RUnlock()

	if !exists {
		return nil, errors.New("wallet not found")
	}
	return wallet, nil
}

// SignTransaction signs a transaction with the wallet's private key
func (w *Wallet) SignTransaction(tx *block.Transaction) error {
	// Calculate transaction hash
	hash := tx.CalculateHash()

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash)
	if err != nil {
		return err
	}

	// Combine r and s into signature
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])

	// Set transaction signature
	tx.Signature = signature
	return nil
}

// VerifyTransaction verifies a transaction's signature
func (w *Wallet) VerifyTransaction(tx *block.Transaction) bool {
	if len(tx.Signature) != 64 {
		return false
	}

	// Split signature into r and s
	r := new(big.Int).SetBytes(tx.Signature[:32])
	s := new(big.Int).SetBytes(tx.Signature[32:])

	// Calculate transaction hash
	hash := tx.CalculateHash()

	// Verify signature
	return ecdsa.Verify(w.PublicKey, hash, r, s)
}

// generateAddress generates a wallet address from a public key
func generateAddress(pubKey *ecdsa.PublicKey) string {
	// Serialize public key
	pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)

	// Hash public key
	hash := sha256.Sum256(pubKeyBytes)

	// Return hex-encoded hash as address
	return hex.EncodeToString(hash[:])
}
