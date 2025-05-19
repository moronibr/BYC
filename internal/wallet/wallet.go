package wallet

import (
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
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/crypto"
	"github.com/byc/internal/logger"
	"github.com/byc/internal/network"
	"github.com/tyler-smith/go-bip39"
	"go.uber.org/zap"
	"golang.org/x/crypto/scrypt"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidAddress    = errors.New("invalid address")
	ErrInvalidBackup     = errors.New("invalid backup data")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrInvalidMnemonic   = errors.New("invalid mnemonic")
	ErrWalletEncrypted   = errors.New("wallet is encrypted")
	ErrWalletDecrypted   = errors.New("wallet is not encrypted")
)

// TransactionRecord represents a transaction in the wallet's history
type TransactionRecord struct {
	TxID        string
	Type        string // "send", "receive", "convert"
	Amount      float64
	CoinType    blockchain.CoinType
	From        string
	To          string
	Timestamp   time.Time
	BlockHeight int64
	Status      string // "pending", "confirmed", "failed"
}

// MultiSigWallet represents a multi-signature wallet
type MultiSigWallet struct {
	Address    string
	PublicKeys [][]byte
	Threshold  int
	Signatures map[string][]byte // txID -> signature
	mu         sync.RWMutex
}

// HDWallet represents a hierarchical deterministic wallet
type HDWallet struct {
	Mnemonic  string
	Seed      []byte
	MasterKey []byte
	ChildKeys map[uint32][]byte
	mu        sync.RWMutex
}

// WatchOnlyWallet represents a watch-only wallet
type WatchOnlyWallet struct {
	Address   string
	PublicKey *ecdsa.PublicKey
	Balances  map[blockchain.CoinType]float64
	mu        sync.RWMutex
}

// AddressBookEntry represents an entry in the address book
type AddressBookEntry struct {
	Name        string
	Address     string
	Description string
	CreatedAt   time.Time
}

// Wallet represents a user's wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
	balances   map[blockchain.CoinType]float64
	mu         sync.RWMutex
	logger     *zap.Logger

	// New fields
	Transactions    []TransactionRecord
	MultiSigWallets map[string]*MultiSigWallet
	HDWallet        *HDWallet
	WatchOnly       bool
	AddressBook     map[string]*AddressBookEntry
	Encrypted       bool
	Salt            []byte
	IV              []byte
	EncryptedKey    []byte
	rateLimiter     *RateLimiter
}

// WalletBackup represents the backup data for a wallet
type WalletBackup struct {
	PrivateKey      []byte
	PublicKey       []byte
	Address         string
	Transactions    []TransactionRecord
	MultiSigWallets map[string]*MultiSigWallet
	HDWallet        *HDWallet
	AddressBook     map[string]*AddressBookEntry
	Salt            []byte
	IV              []byte
}

// NewWallet creates a new wallet
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, &SecurityError{
			Operation: "generate_key",
			Reason:    err.Error(),
		}
	}

	publicKey := &privateKey.PublicKey
	address := generateAddress(publicKey)

	return &Wallet{
		PrivateKey:      privateKey,
		PublicKey:       publicKey,
		Address:         address,
		balances:        make(map[blockchain.CoinType]float64),
		Transactions:    make([]TransactionRecord, 0),
		MultiSigWallets: make(map[string]*MultiSigWallet),
		AddressBook:     make(map[string]*AddressBookEntry),
		logger:          logger.NewLogger("wallet"),
		rateLimiter:     NewRateLimiter(),
	}, nil
}

// NewHDWallet creates a new HD wallet
func NewHDWallet() (*Wallet, error) {
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

	// Create master key
	masterKey := sha256.Sum256(seed)

	// Create wallet
	wallet, err := NewWallet()
	if err != nil {
		return nil, err
	}

	wallet.HDWallet = &HDWallet{
		Mnemonic:  mnemonic,
		Seed:      seed,
		MasterKey: masterKey[:],
		ChildKeys: make(map[uint32][]byte),
	}

	return wallet, nil
}

// NewWatchOnlyWallet creates a new watch-only wallet
func NewWatchOnlyWallet(publicKey *ecdsa.PublicKey) *Wallet {
	address := generateAddress(publicKey)
	return &Wallet{
		PublicKey:   publicKey,
		Address:     address,
		balances:    make(map[blockchain.CoinType]float64),
		WatchOnly:   true,
		AddressBook: make(map[string]*AddressBookEntry),
		logger:      logger.NewLogger("wallet"),
	}
}

// AddTransactionToHistory adds a transaction to the wallet's history
func (w *Wallet) AddTransactionToHistory(tx *blockchain.Transaction, status string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	record := TransactionRecord{
		TxID:        hex.EncodeToString(tx.ID),
		Type:        "send",
		Amount:      tx.Outputs[0].Value,
		CoinType:    tx.Outputs[0].CoinType,
		From:        tx.Inputs[0].Address,
		To:          tx.Outputs[0].Address,
		Timestamp:   time.Now(),
		BlockHeight: 0, // Will be updated when confirmed
		Status:      status,
	}

	w.Transactions = append(w.Transactions, record)
}

// GetTransactionHistory returns the wallet's transaction history
func (w *Wallet) GetTransactionHistory() []TransactionRecord {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Return a copy of the transactions
	transactions := make([]TransactionRecord, len(w.Transactions))
	copy(transactions, w.Transactions)
	return transactions
}

// CreateMultiSigWallet creates a new multi-signature wallet
func (w *Wallet) CreateMultiSigWallet(publicKeys [][]byte, threshold int) (*MultiSigWallet, error) {
	if threshold > len(publicKeys) {
		return nil, fmt.Errorf("threshold cannot be greater than number of public keys")
	}

	// Create multi-sig address
	address := generateMultiSigAddress(publicKeys)

	wallet := &MultiSigWallet{
		Address:    address,
		PublicKeys: publicKeys,
		Threshold:  threshold,
		Signatures: make(map[string][]byte),
	}

	w.MultiSigWallets[address] = wallet
	return wallet, nil
}

// SignMultiSigTransaction signs a multi-signature transaction
func (w *Wallet) SignMultiSigTransaction(txID string, tx *blockchain.Transaction) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Find the multi-sig wallet
	wallet, exists := w.MultiSigWallets[tx.Inputs[0].Address]
	if !exists {
		return fmt.Errorf("multi-sig wallet not found")
	}

	// Sign the transaction
	signature, err := w.SignMessage(tx.ID)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	wallet.Signatures[txID] = signature
	return nil
}

// EstimateTransactionFee estimates the fee for a transaction
func (w *Wallet) EstimateTransactionFee(amount float64, coinType blockchain.CoinType) float64 {
	// Base fee
	baseFee := 0.001

	// Size-based fee
	sizeFee := float64(len(w.Address)) * 0.0001

	// Priority fee based on amount
	priorityFee := amount * 0.01

	return baseFee + sizeFee + priorityFee
}

// AddToAddressBook adds an address to the address book
func (w *Wallet) AddToAddressBook(name, address, description string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !isValidAddress(address) {
		return ErrInvalidAddress
	}

	w.AddressBook[address] = &AddressBookEntry{
		Name:        name,
		Address:     address,
		Description: description,
		CreatedAt:   time.Now(),
	}

	return nil
}

// GetAddressBook returns the address book
func (w *Wallet) GetAddressBook() map[string]*AddressBookEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Return a copy of the address book
	book := make(map[string]*AddressBookEntry)
	for addr, entry := range w.AddressBook {
		book[addr] = entry
	}
	return book
}

// BroadcastTransaction broadcasts a transaction to the network
func (w *Wallet) BroadcastTransaction(tx *blockchain.Transaction, node *network.Node) error {
	// Add transaction to history
	w.AddTransactionToHistory(tx, "pending")

	// Serialize transaction
	txBytes, err := json.Marshal(tx)
	if err != nil {
		w.AddTransactionToHistory(tx, "failed")
		return fmt.Errorf("failed to serialize transaction: %v", err)
	}

	// Create and broadcast message
	msg := &network.Message{
		Type:    network.TxMsg,
		Payload: txBytes,
	}
	if err := node.BroadcastMessage(msg); err != nil {
		w.AddTransactionToHistory(tx, "failed")
		return fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	return nil
}

// EncryptWallet encrypts the wallet with a password
func (w *Wallet) EncryptWallet(password string) error {
	// Check rate limit
	if err := w.rateLimiter.CheckRateLimit("encrypt_wallet"); err != nil {
		return err
	}

	// Generate salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return &EncryptionError{
			Operation: "generate_salt",
			Reason:    err.Error(),
		}
	}

	// Generate key from password
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return &EncryptionError{
			Operation: "derive_key",
			Reason:    err.Error(),
		}
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return &EncryptionError{
			Operation: "generate_iv",
			Reason:    err.Error(),
		}
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return &EncryptionError{
			Operation: "create_cipher",
			Reason:    err.Error(),
		}
	}

	// Convert private key to bytes
	privateKeyBytes := crypto.PrivateKeyToBytes(w.PrivateKey)
	if privateKeyBytes == nil {
		return &EncryptionError{
			Operation: "convert_private_key",
			Reason:    "failed to convert private key to bytes",
		}
	}

	// Encrypt private key
	encryptedPrivateKey := make([]byte, len(privateKeyBytes))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encryptedPrivateKey, privateKeyBytes)

	// Store encrypted private key and clear original
	w.EncryptedKey = encryptedPrivateKey
	w.PrivateKey = nil // Clear private key
	w.Salt = salt
	w.IV = iv
	w.Encrypted = true

	// Log encryption
	w.logger.Info("Wallet encrypted",
		zap.String("address", w.Address),
	)

	return nil
}

// DecryptWallet decrypts the wallet with a password
func (w *Wallet) DecryptWallet(password string) error {
	if !w.Encrypted {
		return nil
	}

	// Generate key from password
	key, err := scrypt.Key([]byte(password), w.Salt, 32768, 8, 1, 32)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %v", err)
	}

	// Decrypt private key
	stream := cipher.NewCFBDecrypter(block, w.IV)
	privateKeyBytes := make([]byte, len(w.EncryptedKey))
	stream.XORKeyStream(privateKeyBytes, w.EncryptedKey)

	// Restore private key
	privateKey, err := crypto.BytesToPrivateKey(privateKeyBytes)
	if err != nil {
		return ErrInvalidPassword
	}

	w.PrivateKey = privateKey
	w.EncryptedKey = nil // Clear encrypted key
	w.Encrypted = false
	return nil
}

// GetMnemonic returns the wallet's mnemonic phrase
func (w *Wallet) GetMnemonic() (string, error) {
	if w.HDWallet == nil {
		return "", fmt.Errorf("not an HD wallet")
	}
	return w.HDWallet.Mnemonic, nil
}

// RestoreFromMnemonic restores a wallet from a mnemonic phrase
func RestoreFromMnemonic(mnemonic string) (*Wallet, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, ErrInvalidMnemonic
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Create master key
	masterKey := sha256.Sum256(seed)

	// Create wallet
	wallet, err := NewWallet()
	if err != nil {
		return nil, err
	}

	wallet.HDWallet = &HDWallet{
		Mnemonic:  mnemonic,
		Seed:      seed,
		MasterKey: masterKey[:],
		ChildKeys: make(map[uint32][]byte),
	}

	return wallet, nil
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

// generateMultiSigAddress generates a multi-signature address from public keys
func generateMultiSigAddress(publicKeys [][]byte) string {
	// Concatenate all public keys
	var combined []byte
	for _, key := range publicKeys {
		combined = append(combined, key...)
	}

	// Hash the combined public keys
	hash := sha256.Sum256(combined)

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
	// Check rate limit
	if err := w.rateLimiter.CheckRateLimit("create_transaction"); err != nil {
		return nil, err
	}

	if amount <= 0 {
		return nil, &InvalidAmountError{
			Amount: amount,
			Reason: "amount must be greater than 0",
		}
	}

	// Validate recipient address
	if !isValidAddress(to) {
		return nil, &InvalidAddressError{
			Address: to,
			Reason:  "invalid address format",
		}
	}

	// Get UTXOs for the sender
	utxos, err := bc.UTXOSet.GetUTXOs(w.Address)
	if err != nil {
		return nil, &TransactionError{
			Operation: "get_utxos",
			Reason:    err.Error(),
		}
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
		return nil, &InsufficientFundsError{
			Required:  amount,
			Available: totalInput,
			CoinType:  coinType.String(),
		}
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
		return nil, &TransactionError{
			Operation: "sign_transaction",
			Reason:    err.Error(),
			TxID:      hex.EncodeToString(tx.ID),
		}
	}

	// Log transaction creation
	w.logger.Info("Transaction created",
		zap.String("tx_id", hex.EncodeToString(tx.ID)),
		zap.Float64("amount", amount),
		zap.String("coin_type", coinType.String()),
		zap.String("to", to),
	)

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
	// Check rate limit
	if err := w.rateLimiter.CheckRateLimit("backup_wallet"); err != nil {
		return err
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	// Create backup data
	backup := WalletBackup{
		PrivateKey:      crypto.PrivateKeyToBytes(w.PrivateKey),
		PublicKey:       crypto.PublicKeyToBytes(w.PublicKey),
		Address:         w.Address,
		Transactions:    w.Transactions,
		MultiSigWallets: w.MultiSigWallets,
		HDWallet:        w.HDWallet,
		AddressBook:     w.AddressBook,
		Salt:            w.Salt,
		IV:              w.IV,
	}

	// Marshal backup data
	data, err := json.Marshal(backup)
	if err != nil {
		return &BackupError{
			Operation: "marshal_backup",
			Path:      path,
			Reason:    err.Error(),
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return &BackupError{
			Operation: "create_directory",
			Path:      path,
			Reason:    err.Error(),
		}
	}

	// Write backup file
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return &BackupError{
			Operation: "write_backup",
			Path:      path,
			Reason:    err.Error(),
		}
	}

	// Log backup
	w.logger.Info("Wallet backup created",
		zap.String("path", path),
		zap.String("address", w.Address),
	)

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
		PrivateKey:      privateKey,
		PublicKey:       publicKey,
		Address:         backup.Address,
		balances:        make(map[blockchain.CoinType]float64),
		Transactions:    backup.Transactions,
		MultiSigWallets: backup.MultiSigWallets,
		HDWallet:        backup.HDWallet,
		AddressBook:     backup.AddressBook,
		Salt:            backup.Salt,
		IV:              backup.IV,
		logger:          logger.NewLogger("wallet"),
	}

	// Verify address
	if wallet.Address != generateAddress(publicKey) {
		return nil, ErrInvalidBackup
	}

	return wallet, nil
}

// ExportPublicKey returns the wallet's public key in bytes
func (w *Wallet) ExportPublicKey() []byte {
	w.mu.RLock()
	defer w.mu.RUnlock()
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

// CreateEphraimCoin creates an Ephraim coin from Golden Block coins
func (w *Wallet) CreateEphraimCoin(bc *blockchain.Blockchain) error {
	// Check if we have enough coins to create an Ephraim coin
	balances := w.GetAllBalances(bc)
	if balances[blockchain.Limnah] < 1 {
		return fmt.Errorf("insufficient Limnah coins to create Ephraim coin")
	}

	// Create a transaction to convert Limnah to Ephraim
	tx, err := w.CreateTransaction(w.Address, 1, blockchain.Ephraim, bc)
	if err != nil {
		return fmt.Errorf("failed to create conversion transaction: %v", err)
	}

	// Add transaction to the blockchain
	if err := bc.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add conversion transaction: %v", err)
	}

	return nil
}

// CreateManassehCoin creates a Manasseh coin from Silver Block coins
func (w *Wallet) CreateManassehCoin(bc *blockchain.Blockchain) error {
	// Check if we have enough coins to create a Manasseh coin
	balances := w.GetAllBalances(bc)
	if balances[blockchain.Onti] < 1 {
		return fmt.Errorf("insufficient Onti coins to create Manasseh coin")
	}

	// Create a transaction to convert Onti to Manasseh
	tx, err := w.CreateTransaction(w.Address, 1, blockchain.Manasseh, bc)
	if err != nil {
		return fmt.Errorf("failed to create conversion transaction: %v", err)
	}

	// Add transaction to the blockchain
	if err := bc.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add conversion transaction: %v", err)
	}

	return nil
}

// CreateJosephCoin creates a Joseph coin from Ephraim and Manasseh coins
func (w *Wallet) CreateJosephCoin(bc *blockchain.Blockchain) error {
	// Check if we have both Ephraim and Manasseh coins
	balances := w.GetAllBalances(bc)
	if balances[blockchain.Ephraim] < 1 || balances[blockchain.Manasseh] < 1 {
		return fmt.Errorf("insufficient Ephraim or Manasseh coins to create Joseph coin")
	}

	// Create a transaction to combine Ephraim and Manasseh into Joseph
	tx, err := w.CreateTransaction(w.Address, 1, blockchain.Joseph, bc)
	if err != nil {
		return fmt.Errorf("failed to create conversion transaction: %v", err)
	}

	// Add transaction to the blockchain
	if err := bc.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add conversion transaction: %v", err)
	}

	return nil
}
