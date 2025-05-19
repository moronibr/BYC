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
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/crypto"
	"github.com/byc/internal/logger"
	"go.uber.org/zap"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidAddress    = errors.New("invalid address")
	ErrInvalidBackup     = errors.New("invalid backup data")
)

// Wallet represents a user's wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
	balances   map[blockchain.CoinType]float64
	mu         sync.RWMutex
	logger     *zap.Logger
}

// WalletBackup represents the backup data for a wallet
type WalletBackup struct {
	PrivateKey []byte
	PublicKey  []byte
	Address    string
}

// NewWallet creates a new wallet
func NewWallet() (*Wallet, error) {
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
		balances:   make(map[blockchain.CoinType]float64),
		logger:     logger.NewLogger("wallet"),
	}, nil
}

// generateAddress generates a wallet address from a public key
func generateAddress(publicKey *ecdsa.PublicKey) string {
	// Convert public key to bytes
	publicKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// Hash the public key
	hash := sha256.Sum256(publicKeyBytes)

	// Return the hex-encoded hash as the address
	return hex.EncodeToString(hash[:])
}

// GetBalance returns the balance for a specific coin type
func (w *Wallet) GetBalance(coinType blockchain.CoinType, bc *blockchain.Blockchain) float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Update balance from blockchain
	balance := bc.GetBalance(w.Address, coinType)
	w.balances[coinType] = balance

	return balance
}

// GetAllBalances returns balances for all coin types
func (w *Wallet) GetAllBalances(bc *blockchain.Blockchain) map[blockchain.CoinType]float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	balances := make(map[blockchain.CoinType]float64)

	// Update balances for all coin types
	for _, coinType := range []blockchain.CoinType{
		blockchain.Leah, blockchain.Shiblum, blockchain.Shiblon,
		blockchain.Senine, blockchain.Seon, blockchain.Shum,
		blockchain.Limnah, blockchain.Antion, blockchain.Senum,
		blockchain.Amnor, blockchain.Ezrom, blockchain.Onti,
	} {
		balances[coinType] = bc.GetBalance(w.Address, coinType)
	}

	w.balances = balances
	return balances
}

// CreateTransaction creates a new transaction
func (w *Wallet) CreateTransaction(to string, amount float64, coinType blockchain.CoinType, bc *blockchain.Blockchain) (*blockchain.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Validate recipient address
	if !isValidAddress(to) {
		return nil, ErrInvalidAddress
	}

	// Check if the coin can be transferred between blocks
	if !blockchain.CanTransferBetweenBlocks(coinType) {
		blockType := blockchain.GetBlockType(coinType)
		if blockType == "" {
			return nil, fmt.Errorf("invalid coin type")
		}
	}

	// Get UTXOs for the sender
	utxos, err := bc.UTXOSet.GetUTXOs(w.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
	}

	// Find UTXOs with the specified coin type
	var inputs []blockchain.TxInput
	var totalInput float64
	for _, utxo := range utxos {
		if utxo.CoinType == coinType {
			input := blockchain.TxInput{
				TxID:        []byte(utxo.TxID),
				OutputIndex: utxo.OutputIndex,
				Amount:      utxo.Amount,
				PublicKey:   []byte(w.Address),
			}
			inputs = append(inputs, input)
			totalInput += utxo.Amount

			if totalInput >= amount {
				break
			}
		}
	}

	if totalInput < amount {
		return nil, ErrInsufficientFunds
	}

	// Create outputs
	outputs := []blockchain.TxOutput{
		{
			Value:         amount,
			CoinType:      coinType,
			PublicKeyHash: []byte(to),
			Address:       to,
		},
	}

	// Add change output if needed
	if totalInput > amount {
		outputs = append(outputs, blockchain.TxOutput{
			Value:         totalInput - amount,
			CoinType:      coinType,
			PublicKeyHash: []byte(w.Address),
			Address:       w.Address,
		})
	}

	// Create transaction
	tx := blockchain.NewTransaction(w.Address, to, amount, coinType, inputs, outputs)

	// Sign transaction
	if err := tx.Sign(w.PrivateKey.D.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return tx, nil
}

// isValidAddress validates a wallet address
func isValidAddress(address string) bool {
	// Check if the address is a valid hex string
	_, err := hex.DecodeString(address)
	if err != nil {
		return false
	}

	// Check if the address has the correct length (32 bytes = 64 hex characters)
	return len(address) == 64
}

// Backup creates a backup of the wallet
func (w *Wallet) Backup(path string) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Create backup data
	backup := WalletBackup{
		PrivateKey: crypto.PrivateKeyToBytes(w.PrivateKey),
		PublicKey:  crypto.PublicKeyToBytes(w.PublicKey),
		Address:    w.Address,
	}

	// Marshal backup data
	data, err := json.Marshal(backup)
	if err != nil {
		return fmt.Errorf("failed to marshal backup data: %v", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Write backup file
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %v", err)
	}

	w.logger.Info("Wallet backup created", zap.String("path", path))
	return nil
}

// Restore restores a wallet from backup
func Restore(path string) (*Wallet, error) {
	// Read backup file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %v", err)
	}

	// Unmarshal backup data
	var backup WalletBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, ErrInvalidBackup
	}

	// Restore private key
	privateKey, err := crypto.BytesToPrivateKey(backup.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to restore private key: %v", err)
	}

	// Restore public key
	publicKey, err := crypto.BytesToPublicKey(backup.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to restore public key: %v", err)
	}

	// Create wallet
	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    backup.Address,
		balances:   make(map[blockchain.CoinType]float64),
		logger:     logger.NewLogger("wallet"),
	}

	// Verify address
	if wallet.Address != generateAddress(publicKey) {
		return nil, ErrInvalidBackup
	}

	return wallet, nil
}

// ExportPublicKey exports the wallet's public key
func (w *Wallet) ExportPublicKey() []byte {
	return crypto.PublicKeyToBytes(w.PublicKey)
}

// ImportPublicKey imports a public key
func ImportPublicKey(publicKeyBytes []byte) (*ecdsa.PublicKey, error) {
	return crypto.BytesToPublicKey(publicKeyBytes)
}

// SignMessage signs a message with the wallet's private key
func (w *Wallet) SignMessage(message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	return crypto.Sign(hash[:], w.PrivateKey.D.Bytes())
}

// VerifyMessage verifies a message signature
func (w *Wallet) VerifyMessage(message, signature []byte) bool {
	hash := sha256.Sum256(message)
	return crypto.Verify(hash[:], signature, crypto.PublicKeyToBytes(w.PublicKey))
}
