package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ripemd160"

	"github.com/youngchain/internal/core/block"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
	ErrWalletExists   = errors.New("wallet already exists")
)

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	sync.RWMutex
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	Address    string
}

// WalletManager manages multiple wallets
type WalletManager struct {
	sync.RWMutex
	wallets    map[string]*Wallet
	walletFile string
}

// NewWalletManager creates a new wallet manager
func NewWalletManager(walletFile string) *WalletManager {
	return &WalletManager{
		wallets:    make(map[string]*Wallet),
		walletFile: walletFile,
	}
}

// CreateWallet creates a new wallet
func (wm *WalletManager) CreateWallet() (*Wallet, error) {
	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create wallet
	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...),
	}

	// Generate address
	wallet.Address = generateAddress(wallet.PublicKey)

	// Add to manager
	wm.Lock()
	defer wm.Unlock()

	if _, exists := wm.wallets[wallet.Address]; exists {
		return nil, ErrWalletExists
	}

	wm.wallets[wallet.Address] = wallet

	// Save wallets
	if err := wm.saveWallets(); err != nil {
		delete(wm.wallets, wallet.Address)
		return nil, err
	}

	return wallet, nil
}

// GetWallet gets a wallet by address
func (wm *WalletManager) GetWallet(address string) (*Wallet, error) {
	wm.RLock()
	defer wm.RUnlock()

	wallet, exists := wm.wallets[address]
	if !exists {
		return nil, ErrInvalidAddress
	}

	return wallet, nil
}

// LoadWallets loads wallets from file
func (wm *WalletManager) LoadWallets() error {
	wm.Lock()
	defer wm.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(wm.walletFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create wallet directory: %v", err)
	}

	// Read wallet file
	data, err := os.ReadFile(wm.walletFile)
	if err != nil {
		if os.IsNotExist(err) {
			return wm.saveWallets()
		}
		return fmt.Errorf("failed to read wallet file: %v", err)
	}

	// Parse wallets
	if err := json.Unmarshal(data, &wm.wallets); err != nil {
		return fmt.Errorf("failed to parse wallet file: %v", err)
	}

	return nil
}

// SaveWallets saves wallets to file
func (wm *WalletManager) saveWallets() error {
	// Marshal wallets
	data, err := json.MarshalIndent(wm.wallets, "", "   ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallets: %v", err)
	}

	// Write to file
	if err := os.WriteFile(wm.walletFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %v", err)
	}

	return nil
}

// ListWallets lists all wallet addresses
func (wm *WalletManager) ListWallets() []string {
	wm.RLock()
	defer wm.RUnlock()

	addresses := make([]string, 0, len(wm.wallets))
	for address := range wm.wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

// generateAddress generates a wallet address from a public key
func generateAddress(publicKey []byte) string {
	// SHA256 hash
	sha256Hash := sha256.Sum256(publicKey)

	// RIPEMD160 hash
	ripemd160Hasher := ripemd160.New()
	_, err := ripemd160Hasher.Write(sha256Hash[:])
	if err != nil {
		panic(err) // Should never happen
	}
	ripemd160Hash := ripemd160Hasher.Sum(nil)

	// Add version byte
	versionedHash := append([]byte{0x00}, ripemd160Hash...)

	// Double SHA256 for checksum
	firstHash := sha256.Sum256(versionedHash)
	secondHash := sha256.Sum256(firstHash[:])
	checksum := secondHash[:4]

	// Append checksum
	finalHash := append(versionedHash, checksum...)

	// Return base58 encoded address
	return hex.EncodeToString(finalHash)
}

// ValidateAddress validates a wallet address
func ValidateAddress(address string) bool {
	// Decode address
	decoded, err := hex.DecodeString(address)
	if err != nil {
		return false
	}

	// Check length
	if len(decoded) != 25 {
		return false
	}

	// Extract checksum
	checksum := decoded[len(decoded)-4:]
	versionedHash := decoded[:len(decoded)-4]

	// Calculate checksum
	firstHash := sha256.Sum256(versionedHash)
	secondHash := sha256.Sum256(firstHash[:])
	targetChecksum := secondHash[:4]

	// Compare checksums
	return string(checksum) == string(targetChecksum)
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

	// Create public key from private key
	publicKey := &w.PrivateKey.PublicKey

	// Verify signature
	return ecdsa.Verify(publicKey, hash, r, s)
}
