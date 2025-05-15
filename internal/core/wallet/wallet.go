package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
	CoinType   coin.Type
	mu         sync.RWMutex
}

// NewWallet creates a new wallet
func NewWallet(coinType coin.Type) (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	publicKey := &privateKey.PublicKey
	address := generateAddress(publicKey)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
		CoinType:   coinType,
	}, nil
}

// LoadWallet loads a wallet from a file
func LoadWallet(filename string) (*Wallet, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %v", err)
	}

	var wallet Wallet
	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %v", err)
	}

	return &wallet, nil
}

// SaveWallet saves a wallet to a file
func (w *Wallet) SaveWallet(filename string) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	data, err := json.Marshal(w)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %v", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %v", err)
	}

	return nil
}

// SignTransaction signs a transaction
func (w *Wallet) SignTransaction(tx *types.Transaction) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Create signature hash
	if err := tx.CalculateHash(); err != nil {
		return fmt.Errorf("failed to calculate hash: %v", err)
	}

	// Sign hash
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, tx.GetHash())
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Add signature to transaction
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature

	return nil
}

// CreateTransaction creates a new transaction
func (w *Wallet) CreateTransaction(to string, amount uint64, fee uint64) (*types.Transaction, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	tx := types.NewTransaction(1, w.CoinType)
	tx.Fee = int64(fee)

	// Add output
	tx.AddOutput(int64(amount), []byte(to))

	// Sign transaction
	if err := w.SignTransaction(tx); err != nil {
		return nil, err
	}

	return tx, nil
}

// GetBalance gets the wallet balance
func (w *Wallet) GetBalance(utxoSet UTXOSetInterface) (uint64, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var balance uint64
	utxos, err := utxoSet.GetUTXOsByAddress(w.Address)
	if err != nil {
		return 0, fmt.Errorf("failed to get UTXOs: %v", err)
	}

	for _, utxo := range utxos {
		if utxo.CoinType == w.CoinType && !utxo.Spent {
			balance += uint64(utxo.Value)
		}
	}

	return balance, nil
}

// generateAddress generates a wallet address from a public key
func generateAddress(publicKey *ecdsa.PublicKey) string {
	// Convert public key to bytes
	publicKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// Hash public key
	hash := sha256.Sum256(publicKeyBytes)

	// Convert hash to hex string
	return hex.EncodeToString(hash[:])
}

// UTXOSetInterface defines the interface for UTXO set operations
type UTXOSetInterface interface {
	GetUTXOsByAddress(address string) ([]*types.UTXO, error)
}
