package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/scrypt"

	"github.com/youngchain/internal/crypto"
)

var (
	ErrInvalidAddress     = errors.New("invalid address")
	ErrWalletExists       = errors.New("wallet already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrWalletNotEncrypted = errors.New("wallet is not encrypted")
)

const (
	// BIP32 constants
	HardenedKeyStart = uint32(0x80000000)
	DefaultPath      = "m/44'/0'/0'/0/"
)

// HDWallet represents a Hierarchical Deterministic wallet
type HDWallet struct {
	sync.RWMutex
	Mnemonic  string
	MasterKey *ecdsa.PrivateKey
	Accounts  map[string]*Account
	Encrypted bool
	Salt      []byte
	IV        []byte
}

// Account represents a wallet account with multiple addresses
type Account struct {
	Index     uint32
	Addresses map[string]*Address
	Balance   map[string]uint64 // Balance per coin type
}

// Address represents a single address in an account
type Address struct {
	Index      uint32
	PublicKey  []byte
	PrivateKey *ecdsa.PrivateKey
	Used       bool
}

// WalletManager manages multiple HD wallets
type WalletManager struct {
	sync.RWMutex
	wallets    map[string]*HDWallet
	walletFile string
}

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
	Balance    uint64
	mu         sync.RWMutex
}

// NewWalletManager creates a new wallet manager
func NewWalletManager(walletFile string) *WalletManager {
	return &WalletManager{
		wallets:    make(map[string]*HDWallet),
		walletFile: walletFile,
	}
}

// NewWallet creates a new wallet with a fresh key pair
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	publicKey := &privateKey.PublicKey
	address := generateAddress(publicKey)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
		Balance:    0,
	}, nil
}

// CreateWallet creates a new HD wallet
func (wm *WalletManager) CreateWallet(password string) (*HDWallet, error) {
	// Generate mnemonic
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entropy: %v", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mnemonic: %v", err)
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Generate master key
	masterKey, err := generateMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %v", err)
	}

	// Create wallet
	wallet := &HDWallet{
		Mnemonic:  mnemonic,
		MasterKey: masterKey,
		Accounts:  make(map[string]*Account),
	}

	// Create first account
	if err := wallet.createAccount(0); err != nil {
		return nil, err
	}

	// Encrypt wallet if password provided
	if password != "" {
		if err := wallet.Encrypt(password); err != nil {
			return nil, err
		}
	}

	// Add to manager
	wm.Lock()
	defer wm.Unlock()

	// Generate wallet ID from master key
	walletID := generateWalletID(masterKey)
	if _, exists := wm.wallets[walletID]; exists {
		return nil, ErrWalletExists
	}

	wm.wallets[walletID] = wallet

	// Save wallets
	if err := wm.saveWallets(); err != nil {
		delete(wm.wallets, walletID)
		return nil, err
	}

	return wallet, nil
}

// createAccount creates a new account in the wallet
func (w *HDWallet) createAccount(index uint32) error {
	w.Lock()
	defer w.Unlock()

	// Generate account key
	accountKey, err := w.deriveKey(fmt.Sprintf("%s%d'", DefaultPath, index))
	if err != nil {
		return err
	}

	// Create account
	account := &Account{
		Index:     index,
		Addresses: make(map[string]*Address),
		Balance:   make(map[string]uint64),
	}

	// Generate first address
	if err := account.generateAddress(accountKey, 0); err != nil {
		return err
	}

	// Add account to wallet
	w.Accounts[fmt.Sprintf("%d", index)] = account

	return nil
}

// generateAddress generates a new address for an account
func (a *Account) generateAddress(accountKey *ecdsa.PrivateKey, index uint32) error {
	// Derive address key
	path := fmt.Sprintf("%d", index)
	addressKey, err := deriveChildKey(accountKey, path)
	if err != nil {
		return err
	}

	// Generate address
	address := &Address{
		Index:      index,
		PublicKey:  append(addressKey.PublicKey.X.Bytes(), addressKey.PublicKey.Y.Bytes()...),
		PrivateKey: addressKey,
	}

	// Add address to account
	addressStr := generateAddress(address.PublicKey)
	a.Addresses[addressStr] = address

	return nil
}

// Encrypt encrypts the wallet with a password
func (w *HDWallet) Encrypt(password string) error {
	if w.Encrypted {
		return nil
	}

	// Generate salt and IV
	salt := make([]byte, 32)
	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Generate key from password
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Encrypt master key
	masterKeyBytes := w.MasterKey.D.Bytes()
	encrypted := make([]byte, len(masterKeyBytes))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(encrypted, masterKeyBytes)

	// Update wallet
	w.Encrypted = true
	w.Salt = salt
	w.IV = iv
	w.MasterKey.D.SetBytes(encrypted)

	return nil
}

// Decrypt decrypts the wallet with a password
func (w *HDWallet) Decrypt(password string) error {
	if !w.Encrypted {
		return ErrWalletNotEncrypted
	}

	// Generate key from password
	key, err := scrypt.Key([]byte(password), w.Salt, 32768, 8, 1, 32)
	if err != nil {
		return err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Decrypt master key
	encrypted := w.MasterKey.D.Bytes()
	decrypted := make([]byte, len(encrypted))
	stream := cipher.NewCTR(block, w.IV)
	stream.XORKeyStream(decrypted, encrypted)

	// Update wallet
	w.MasterKey.D.SetBytes(decrypted)
	w.Encrypted = false

	return nil
}

// generateWalletID generates a unique ID for the wallet
func generateWalletID(key *ecdsa.PrivateKey) string {
	hash := sha256.Sum256(key.D.Bytes())
	return hex.EncodeToString(hash[:])
}

// generateMasterKey generates a master key from a seed
func generateMasterKey(seed []byte) (*ecdsa.PrivateKey, error) {
	hash := sha256.Sum256(seed)
	return ecdsa.GenerateKey(elliptic.P256(), bytes.NewReader(hash[:]))
}

// deriveKey derives a child key from a parent key
func deriveChildKey(parent *ecdsa.PrivateKey, path string) (*ecdsa.PrivateKey, error) {
	// Implementation of BIP32 key derivation
	// This is a simplified version - in production, use a proper BIP32 library
	hash := sha256.Sum256(append(parent.D.Bytes(), []byte(path)...))
	return ecdsa.GenerateKey(elliptic.P256(), bytes.NewReader(hash[:]))
}

// deriveKey derives a child key from the master key
func (w *HDWallet) deriveKey(path string) (*ecdsa.PrivateKey, error) {
	return deriveChildKey(w.MasterKey, path)
}

// GetWallet gets a wallet by address
func (wm *WalletManager) GetWallet(address string) (*HDWallet, error) {
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

// generateAddress creates a unique address from a public key
func generateAddress(publicKey *ecdsa.PublicKey) string {
	// Convert public key to bytes
	pubKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// Hash the public key
	hash := sha256.Sum256(pubKeyBytes)

	// Take the last 20 bytes as the address
	address := hex.EncodeToString(hash[12:])

	return "0x" + address
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
func (w *Wallet) SignTransaction(tx *Transaction) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.PrivateKey == nil {
		return errors.New("wallet has no private key")
	}

	// Create the transaction hash
	txHash := tx.Hash()

	// Sign the hash
	signature, err := crypto.Sign(txHash, w.PrivateKey)
	if err != nil {
		return err
	}

	tx.Signature = signature
	return nil
}

// UpdateBalance updates the wallet's balance
func (w *Wallet) UpdateBalance(amount uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Balance = amount
}

// GetBalance returns the current wallet balance
func (w *Wallet) GetBalance() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Balance
}

// Transaction represents a cryptocurrency transaction
type Transaction struct {
	From      string
	To        string
	Amount    uint64
	Nonce     uint64
	Signature []byte
	Hash      []byte
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount, nonce uint64) *Transaction {
	tx := &Transaction{
		From:   from,
		To:     to,
		Amount: amount,
		Nonce:  nonce,
	}
	tx.Hash = tx.calculateHash()
	return tx
}

// calculateHash computes the transaction hash
func (tx *Transaction) calculateHash() []byte {
	data := []byte(tx.From + tx.To + string(tx.Amount) + string(tx.Nonce))
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifySignature verifies the transaction signature
func (tx *Transaction) VerifySignature(publicKey *ecdsa.PublicKey) bool {
	return crypto.Verify(tx.Hash, tx.Signature, publicKey)
}
