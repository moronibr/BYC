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
	"math/big"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mr-tron/base58"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func init() {
	mathrand.Seed(time.Now().UnixNano())
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAddress    = errors.New("invalid address")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrWalletLocked      = errors.New("wallet is locked")
)

// AddressType represents the type of address
type AddressType int

const (
	// AddressTypeP2PKH represents a Pay-to-Public-Key-Hash address
	AddressTypeP2PKH AddressType = iota
	// AddressTypeP2SH represents a Pay-to-Script-Hash address
	AddressTypeP2SH
	// AddressTypeSegWit represents a Segregated Witness address
	AddressTypeSegWit
	// AddressTypeTaproot represents a Taproot address
	AddressTypeTaproot
	// AddressTypeMultiSig represents a Multi-Signature address
	AddressTypeMultiSig
)

// Address represents a cryptocurrency address
type Address struct {
	Type     AddressType
	Hash     []byte
	Script   []byte             // For P2SH addresses
	Version  byte               // For SegWit addresses
	Keys     []*ecdsa.PublicKey // For MultiSig addresses
	TapTweak []byte             // For Taproot addresses
}

// String returns the base58-encoded address string
func (a *Address) String() string {
	// Create payload based on address type
	var payload []byte
	switch a.Type {
	case AddressTypeP2PKH:
		payload = append([]byte{0x00}, a.Hash...)
	case AddressTypeP2SH:
		payload = append([]byte{0x05}, a.Hash...)
	case AddressTypeSegWit:
		payload = append([]byte{0x06}, a.Version)
		payload = append(payload, a.Hash...)
	case AddressTypeTaproot:
		payload = append([]byte{0x07}, a.Hash...)
	case AddressTypeMultiSig:
		// Create redeem script for multi-sig
		script := make([]byte, 0)
		script = append(script, byte(len(a.Keys))) // Number of public keys
		for _, key := range a.Keys {
			pubKeyBytes := elliptic.Marshal(key.Curve, key.X, key.Y)
			script = append(script, byte(len(pubKeyBytes)))
			script = append(script, pubKeyBytes...)
		}
		script = append(script, byte(len(a.Keys)/2+1)) // Required signatures
		script = append(script, 0xae)                  // OP_CHECKMULTISIG
		a.Script = script
		scriptHash := sha256.Sum256(script)
		payload = append([]byte{0x05}, scriptHash[:]...)
	default:
		return ""
	}

	// Calculate checksum
	hash := sha256.Sum256(payload)
	hash = sha256.Sum256(hash[:])
	checksum := hash[:4]

	// Combine payload and checksum
	combined := append(payload, checksum...)

	// Encode to base58
	return base58.Encode(combined)
}

// Account represents a wallet account
type Account struct {
	Index     uint32
	Address   *Address
	Balance   uint64
	UTXOs     map[string]*UTXO // key: txHash:outputIndex
	PublicKey *ecdsa.PublicKey
}

// KeyStore represents a secure storage for private keys
type KeyStore struct {
	mu   sync.RWMutex
	keys map[string]*ecdsa.PrivateKey // key: account index
}

// NewKeyStore creates a new key store
func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys: make(map[string]*ecdsa.PrivateKey),
	}
}

// StoreKey stores a private key securely
func (ks *KeyStore) StoreKey(index uint32, key *ecdsa.PrivateKey) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.keys[fmt.Sprintf("%d", index)] = key
}

// GetKey retrieves a private key
func (ks *KeyStore) GetKey(index uint32) (*ecdsa.PrivateKey, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	key, exists := ks.keys[fmt.Sprintf("%d", index)]
	return key, exists
}

// HardwareWalletDeviceInfo represents information about a hardware wallet device
type HardwareWalletDeviceInfo struct {
	ID              string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	IsInitialized   bool
	IsLocked        bool
	BatteryLevel    int // 0-100, -1 if not applicable
	HasPassphrase   bool
	HasPin          bool
}

// HardwareWalletDevice represents a hardware wallet device
type HardwareWalletDevice interface {
	// Device Management
	GetDeviceInfo() (*HardwareWalletDeviceInfo, error)
	Initialize(passphrase string) error
	Wipe() error
	SetPin(pin string) error
	ChangePin(oldPin, newPin string) error
	Unlock(pin string) error
	Lock() error

	// Firmware Management
	GetFirmwareVersion() (string, error)
	UpdateFirmware(firmwareData []byte) error
	VerifyFirmware(firmwareData []byte) error
	GetFirmwareUpdateStatus() (string, error)

	// Device Features
	GetBatteryLevel() (int, error)
	SetPassphrase(passphrase string) error
	RemovePassphrase() error
	GetBackupStatus() (bool, error)
	CreateBackup() error
	RestoreBackup(backupData []byte) error
}

// HardwareWallet represents a hardware wallet interface
type HardwareWallet interface {
	// Device Management
	Connect() error
	Disconnect() error
	GetConnectedDevices() ([]HardwareWalletDevice, error)
	GetDeviceByID(id string) (HardwareWalletDevice, error)

	// Existing methods
	GetPublicKey(path string) (*ecdsa.PublicKey, error)
	SignTransaction(tx *Transaction, path string) error
	GetAddress(path string) (*Address, error)
}

// HardwareWalletManager represents a manager for hardware wallet devices
type HardwareWalletManager struct {
	devices map[string]HardwareWalletDevice
	mu      sync.RWMutex
}

// NewHardwareWalletManager creates a new hardware wallet manager
func NewHardwareWalletManager() *HardwareWalletManager {
	return &HardwareWalletManager{
		devices: make(map[string]HardwareWalletDevice),
	}
}

// GetConnectedDevices returns a list of connected devices
func (m *HardwareWalletManager) GetConnectedDevices() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]string, 0, len(m.devices))
	for id := range m.devices {
		devices = append(devices, id)
	}
	return devices, nil
}

// GetDeviceStatus returns the status of a device
func (m *HardwareWalletManager) GetDeviceStatus(deviceID string) (map[string]interface{}, error) {
	m.mu.RLock()
	device, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	info, err := device.GetDeviceInfo()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":               info.ID,
		"model":            info.Model,
		"firmware_version": info.FirmwareVersion,
		"is_initialized":   info.IsInitialized,
		"is_locked":        info.IsLocked,
		"battery_level":    info.BatteryLevel,
		"has_passphrase":   info.HasPassphrase,
		"has_pin":          info.HasPin,
	}, nil
}

// GetDeviceCapabilities returns the capabilities of a device
func (m *HardwareWalletManager) GetDeviceCapabilities(deviceID string) (map[string]interface{}, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would get capabilities from the device
	return map[string]interface{}{
		"supports_bip32":           true,
		"supports_bip39":           true,
		"supports_passphrase":      true,
		"supports_pin":             true,
		"supports_firmware_update": true,
		"supports_backup":          true,
		"supports_restore":         true,
	}, nil
}

// GetDeviceMetrics returns metrics for a device
func (m *HardwareWalletManager) GetDeviceMetrics(deviceID string, start, end int64) (map[string]interface{}, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would get metrics from the device
	return map[string]interface{}{
		"battery_level": 100,
		"is_connected":  true,
		"last_seen":     time.Now().Unix(),
	}, nil
}

// GetDeviceLogs returns logs for a device
func (m *HardwareWalletManager) GetDeviceLogs(deviceID string, start, end int64, level string) ([]map[string]interface{}, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would fetch logs from the device
	return []map[string]interface{}{
		{
			"timestamp": time.Now().Unix(),
			"level":     level,
			"message":   "Device connected",
		},
	}, nil
}

// GetDeviceEvents returns events for a device
func (m *HardwareWalletManager) GetDeviceEvents(deviceID string, start, end int64, eventType string) ([]map[string]interface{}, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would fetch events from the device
	return []map[string]interface{}{
		{
			"timestamp": time.Now().Unix(),
			"type":      eventType,
			"data":      "Device event",
		},
	}, nil
}

// ExportPublicKey exports a public key from a device
func (m *HardwareWalletManager) ExportPublicKey(deviceID, path string) (string, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would export the public key from the device
	return "03" + hex.EncodeToString(make([]byte, 32)), nil
}

// VerifyMessage verifies a message signature
func (m *HardwareWalletManager) VerifyMessage(deviceID, message, signature, address string) (bool, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would verify the message signature
	return true, nil
}

// SignMessage signs a message
func (m *HardwareWalletManager) SignMessage(deviceID, message, path string) (string, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would sign the message using the device
	return hex.EncodeToString(make([]byte, 64)), nil
}

// GetXPub gets an extended public key
func (m *HardwareWalletManager) GetXPub(deviceID, path string) (string, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would get the extended public key from the device
	return "xpub" + hex.EncodeToString(make([]byte, 32)), nil
}

// GetAddress gets an address
func (m *HardwareWalletManager) GetAddress(deviceID, path, addressType string) (string, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would get the address from the device
	return "1" + hex.EncodeToString(make([]byte, 20)), nil
}

// GetAddressInfo gets information about an address
func (m *HardwareWalletManager) GetAddressInfo(deviceID, address string) (map[string]interface{}, error) {
	m.mu.RLock()
	_, exists := m.devices[deviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// In a real implementation, this would get information about the address
	return map[string]interface{}{
		"address": address,
		"type":    "p2pkh",
		"path":    "m/44'/0'/0'/0/0",
	}, nil
}

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	mu sync.RWMutex

	// Key management
	masterKey *bip32.Key
	accounts  map[uint32]*Account
	keyStore  *KeyStore
	encrypted bool
	password  []byte

	// Hardware wallet support
	hardwareWallet    HardwareWallet
	useHardwareWallet bool

	// Configuration
	accountGapLimit uint32
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash      string
	OutputIndex uint32
	Value       uint64
	Script      []byte
	Height      uint64
	IsCoinbase  bool
	PublicKey   *ecdsa.PublicKey
}

// Transaction represents a blockchain transaction
type Transaction struct {
	Inputs  []*TransactionInput
	Outputs []*TransactionOutput
	Hash    string
}

// TransactionInput represents a transaction input
type TransactionInput struct {
	TxHash      string
	OutputIndex uint32
	Signature   []byte
}

// TransactionOutput represents a transaction output
type TransactionOutput struct {
	Address string
	Value   uint64
}

// UTXOSelectionStrategy defines the strategy for selecting UTXOs
type UTXOSelectionStrategy int

const (
	// StrategyLargestFirst selects the largest UTXOs first
	StrategyLargestFirst UTXOSelectionStrategy = iota
	// StrategySmallestFirst selects the smallest UTXOs first
	StrategySmallestFirst
	// StrategyRandom selects UTXOs randomly
	StrategyRandom
	// StrategyOptimal selects UTXOs to minimize the number of inputs while keeping change minimal
	StrategyOptimal
)

// UTXOSelectionOptions contains options for UTXO selection
type UTXOSelectionOptions struct {
	Strategy         UTXOSelectionStrategy
	MaxInputs        int
	MinChange        uint64
	MaxChange        uint64
	ExcludeCoinbase  bool
	MinConfirmations uint64
	MaxDustAmount    uint64
	PrivacyMode      bool // When true, tries to avoid linking transactions
}

// DefaultUTXOSelectionOptions returns default options for UTXO selection
func DefaultUTXOSelectionOptions() *UTXOSelectionOptions {
	return &UTXOSelectionOptions{
		Strategy:         StrategyOptimal,
		MaxInputs:        10,
		MinChange:        1000,    // Minimum change amount to create a new output
		MaxChange:        1000000, // Maximum change amount before trying to optimize
		ExcludeCoinbase:  false,
		MinConfirmations: 1,
		MaxDustAmount:    546, // Standard dust threshold
		PrivacyMode:      false,
	}
}

// NewWallet creates a new wallet instance
func NewWallet(useHardwareWallet bool, hw HardwareWallet) (*Wallet, error) {
	var wallet *Wallet
	var err error

	if useHardwareWallet {
		if hw == nil {
			return nil, fmt.Errorf("hardware wallet interface required when useHardwareWallet is true")
		}
		wallet = &Wallet{
			accounts:          make(map[uint32]*Account),
			keyStore:          NewKeyStore(),
			accountGapLimit:   20,
			hardwareWallet:    hw,
			useHardwareWallet: true,
		}
	} else {
		// Generate mnemonic
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return nil, fmt.Errorf("failed to generate entropy: %w", err)
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
		}

		// Generate seed
		seed := bip39.NewSeed(mnemonic, "")

		// Generate master key
		masterKey, err := bip32.NewMasterKey(seed)
		if err != nil {
			return nil, fmt.Errorf("failed to generate master key: %w", err)
		}

		wallet = &Wallet{
			masterKey:       masterKey,
			accounts:        make(map[uint32]*Account),
			keyStore:        NewKeyStore(),
			accountGapLimit: 20,
		}
	}

	// Create first account
	account, err := wallet.createAccount(0)
	if err != nil {
		return nil, fmt.Errorf("failed to create first account: %w", err)
	}
	wallet.accounts[0] = account

	return wallet, nil
}

// LoadWallet loads a wallet from a file
func LoadWallet(filename string, password []byte) (*Wallet, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	var encryptedWallet struct {
		Encrypted bool   `json:"encrypted"`
		Data      []byte `json:"data"`
	}

	if err := json.Unmarshal(data, &encryptedWallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}

	var walletData []byte
	if encryptedWallet.Encrypted {
		if len(password) == 0 {
			return nil, ErrInvalidPassword
		}

		// Decrypt wallet data
		key := sha256.Sum256(password)
		block, err := aes.NewCipher(key[:])
		if err != nil {
			return nil, fmt.Errorf("failed to create cipher: %w", err)
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCM: %w", err)
		}

		nonceSize := gcm.NonceSize()
		if len(encryptedWallet.Data) < nonceSize {
			return nil, fmt.Errorf("invalid encrypted data")
		}

		nonce := encryptedWallet.Data[:nonceSize]
		ciphertext := encryptedWallet.Data[nonceSize:]

		walletData, err = gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, ErrInvalidPassword
		}
	} else {
		walletData = encryptedWallet.Data
	}

	var wallet Wallet
	if err := json.Unmarshal(walletData, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet data: %w", err)
	}

	if encryptedWallet.Encrypted {
		wallet.encrypted = true
		wallet.password = password
	}

	return &wallet, nil
}

// SaveWallet saves a wallet to a file
func (w *Wallet) SaveWallet(filename string) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	walletData, err := json.Marshal(w)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}

	var encryptedWallet struct {
		Encrypted bool   `json:"encrypted"`
		Data      []byte `json:"data"`
	}

	if w.encrypted {
		// Encrypt wallet data
		key := sha256.Sum256(w.password)
		block, err := aes.NewCipher(key[:])
		if err != nil {
			return fmt.Errorf("failed to create cipher: %w", err)
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return fmt.Errorf("failed to create GCM: %w", err)
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return fmt.Errorf("failed to generate nonce: %w", err)
		}

		encryptedWallet.Encrypted = true
		encryptedWallet.Data = gcm.Seal(nonce, nonce, walletData, nil)
	} else {
		encryptedWallet.Encrypted = false
		encryptedWallet.Data = walletData
	}

	data, err := json.Marshal(encryptedWallet)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted wallet: %w", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

// CreateAccount creates a new account
func (w *Wallet) CreateAccount() (*Account, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.encrypted {
		return nil, ErrWalletLocked
	}

	// Find next available account index
	var nextIndex uint32
	for {
		if _, exists := w.accounts[nextIndex]; !exists {
			break
		}
		nextIndex++
	}

	return w.createAccount(nextIndex)
}

// GetAccount returns an account by index
func (w *Wallet) GetAccount(index uint32) (*Account, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	account, exists := w.accounts[index]
	if !exists {
		return nil, fmt.Errorf("account %d not found", index)
	}

	return account, nil
}

// GetAccounts returns all accounts
func (w *Wallet) GetAccounts() []*Account {
	w.mu.RLock()
	defer w.mu.RUnlock()

	accounts := make([]*Account, 0, len(w.accounts))
	for _, account := range w.accounts {
		accounts = append(accounts, account)
	}

	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Index < accounts[j].Index
	})

	return accounts
}

// UpdateAccountBalance updates an account's balance and UTXO set
func (w *Wallet) UpdateAccountBalance(index uint32, utxos []*UTXO) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	account, exists := w.accounts[index]
	if !exists {
		return fmt.Errorf("account %d not found", index)
	}

	// Clear existing UTXOs and balance
	account.UTXOs = make(map[string]*UTXO)
	account.Balance = 0

	// Add new UTXOs
	for _, utxo := range utxos {
		key := fmt.Sprintf("%s:%d", utxo.TxHash, utxo.OutputIndex)
		account.UTXOs[key] = utxo
		account.Balance += utxo.Value
	}

	return nil
}

// CreateTransaction creates a new transaction
func (w *Wallet) CreateTransaction(fromIndex uint32, toAddress string, amount uint64, fee uint64) (*Transaction, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.encrypted {
		return nil, ErrWalletLocked
	}

	// Get source account
	account, exists := w.accounts[fromIndex]
	if !exists {
		return nil, fmt.Errorf("account %d not found", fromIndex)
	}

	// Validate address
	if !w.validateAddress(toAddress) {
		return nil, ErrInvalidAddress
	}

	// Check if we have enough funds
	if account.Balance < amount+fee {
		return nil, ErrInsufficientFunds
	}

	// Select UTXOs to spend
	selectedUTXOs, change := w.selectUTXOs(account.UTXOs, amount+fee, nil)
	if change < 0 {
		return nil, ErrInsufficientFunds
	}

	// Create transaction
	tx := &Transaction{
		Inputs:  make([]*TransactionInput, 0, len(selectedUTXOs)),
		Outputs: make([]*TransactionOutput, 0, 2),
	}

	// Add inputs
	for _, utxo := range selectedUTXOs {
		tx.Inputs = append(tx.Inputs, &TransactionInput{
			TxHash:      utxo.TxHash,
			OutputIndex: utxo.OutputIndex,
		})
	}

	// Add output for recipient
	tx.Outputs = append(tx.Outputs, &TransactionOutput{
		Address: toAddress,
		Value:   amount,
	})

	// Add change output if needed
	if change > 0 {
		tx.Outputs = append(tx.Outputs, &TransactionOutput{
			Address: account.Address.String(),
			Value:   uint64(change),
		})
	}

	// Sign transaction
	if err := w.signTransaction(tx, account); err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Calculate transaction hash
	tx.Hash = w.calculateTransactionHash(tx)

	return tx, nil
}

// Helper functions

func (w *Wallet) createAccount(index uint32) (*Account, error) {
	path := fmt.Sprintf("m/44'/0'/%d'/0/0", index)

	if w.useHardwareWallet {
		// Get public key from hardware wallet
		publicKey, err := w.hardwareWallet.GetPublicKey(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key from hardware wallet: %w", err)
		}

		// Get address from hardware wallet
		address, err := w.hardwareWallet.GetAddress(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get address from hardware wallet: %w", err)
		}

		// Create account
		account := &Account{
			Index:     index,
			Address:   address,
			UTXOs:     make(map[string]*UTXO),
			PublicKey: publicKey,
		}

		return account, nil
	}

	// Derive account key using BIP32
	accountKey, err := w.deriveKey(path)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account key: %w", err)
	}

	// Convert to ECDSA private key
	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(accountKey.PublicKey().Key),
			Y:     new(big.Int).SetBytes(accountKey.PublicKey().Key[32:]),
		},
		D: new(big.Int).SetBytes(accountKey.Key),
	}

	// Store private key securely
	w.keyStore.StoreKey(index, privateKey)

	// Create account
	account := &Account{
		Index:     index,
		UTXOs:     make(map[string]*UTXO),
		PublicKey: &privateKey.PublicKey,
	}

	// Generate address
	if err := w.generateAddress(account); err != nil {
		return nil, fmt.Errorf("failed to generate address: %w", err)
	}

	return account, nil
}

func (w *Wallet) deriveKey(path string) (*bip32.Key, error) {
	parts := strings.Split(path, "/")
	key := w.masterKey

	for _, part := range parts[1:] {
		var index uint32
		if strings.HasSuffix(part, "'") {
			// Hardened derivation
			index = bip32.FirstHardenedChild
			part = strings.TrimSuffix(part, "'")
		}
		i, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}
		index += uint32(i)
		key, err = key.NewChildKey(index)
		if err != nil {
			return nil, fmt.Errorf("failed to derive key: %w", err)
		}
	}

	return key, nil
}

func (w *Wallet) generateAddress(account *Account) error {
	// Convert public key to bytes
	pubKeyBytes := elliptic.Marshal(account.PublicKey.Curve, account.PublicKey.X, account.PublicKey.Y)

	// Create address based on type
	address := &Address{
		Type: AddressTypeP2PKH, // Default to P2PKH
	}

	// Hash public key
	hash := sha256.Sum256(pubKeyBytes)
	address.Hash = hash[:]

	switch address.Type {
	case AddressTypeP2SH:
		// Create a P2WPKH redeem script
		redeemScript := make([]byte, 0, 22)
		redeemScript = append(redeemScript, 0x00) // Witness version
		redeemScript = append(redeemScript, 0x14) // Hash length
		redeemScript = append(redeemScript, hash[:]...)
		address.Script = redeemScript

		// Hash the redeem script
		scriptHash := sha256.Sum256(redeemScript)
		address.Hash = scriptHash[:]

	case AddressTypeSegWit:
		address.Version = 0x00 // Witness version 0
		address.Hash = hash[:]

	case AddressTypeTaproot:
		// Generate Taproot internal key
		internalKey := account.PublicKey

		// Generate TapTweak (in a real implementation, this would be derived from a merkle tree)
		tapTweak := make([]byte, 32)
		if _, err := rand.Read(tapTweak); err != nil {
			return fmt.Errorf("failed to generate tap tweak: %w", err)
		}
		address.TapTweak = tapTweak

		// Calculate Taproot output key
		// In a real implementation, this would use proper Taproot key tweaking
		tapHash := sha256.Sum256(append(internalKey.X.Bytes(), tapTweak...))
		address.Hash = tapHash[:]

	case AddressTypeMultiSig:
		// For demonstration, we'll use the account's public key multiple times
		// In a real implementation, these would be different public keys
		address.Keys = []*ecdsa.PublicKey{
			account.PublicKey,
			account.PublicKey,
			account.PublicKey,
		}
	}

	account.Address = address
	return nil
}

func (w *Wallet) validateAddress(address string) bool {
	// Decode address
	addr, err := w.decodeAddress(address)
	if err != nil {
		return false
	}

	// Validate based on type
	switch addr.Type {
	case AddressTypeP2PKH:
		return len(addr.Hash) == 20 // RIPEMD160 hash length
	case AddressTypeP2SH:
		return len(addr.Hash) == 20 && len(addr.Script) > 0
	case AddressTypeSegWit:
		return len(addr.Hash) == 20 && addr.Version == 0x00
	case AddressTypeTaproot:
		return len(addr.Hash) == 32 && len(addr.TapTweak) == 32
	case AddressTypeMultiSig:
		return len(addr.Keys) > 0
	default:
		return false
	}
}

func (w *Wallet) decodeAddress(address string) (*Address, error) {
	// Decode base58 address
	decoded, err := base58.Decode(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	// Check checksum
	if len(decoded) < 5 {
		return nil, fmt.Errorf("address too short")
	}

	// Verify checksum
	checksum := decoded[len(decoded)-4:]
	payload := decoded[:len(decoded)-4]
	hash := sha256.Sum256(payload)
	hash = sha256.Sum256(hash[:])
	if !bytes.Equal(checksum, hash[:4]) {
		return nil, fmt.Errorf("invalid checksum")
	}

	// Create address based on version
	addr := &Address{}
	switch decoded[0] {
	case 0x00: // P2PKH
		addr.Type = AddressTypeP2PKH
		addr.Hash = payload[1:]
	case 0x05: // P2SH
		addr.Type = AddressTypeP2SH
		addr.Hash = payload[1:]
	case 0x06: // SegWit
		addr.Type = AddressTypeSegWit
		addr.Version = payload[1]
		addr.Hash = payload[2:]
	case 0x07: // Taproot
		addr.Type = AddressTypeTaproot
		addr.Hash = payload[1:]
		addr.TapTweak = payload[1:]
	case 0x0a: // MultiSig
		addr.Type = AddressTypeMultiSig
		addr.Keys = make([]*ecdsa.PublicKey, len(payload)/65)
		for i := 0; i < len(payload); i += 65 {
			pubKey := &ecdsa.PublicKey{
				Curve: elliptic.P256(),
				X:     new(big.Int).SetBytes(payload[i : i+32]),
				Y:     new(big.Int).SetBytes(payload[i+32 : i+64]),
			}
			addr.Keys[i/65] = pubKey
		}
	default:
		return nil, fmt.Errorf("unsupported address type")
	}

	return addr, nil
}

func (w *Wallet) selectUTXOs(utxos map[string]*UTXO, amount uint64, options *UTXOSelectionOptions) ([]*UTXO, int64) {
	if options == nil {
		options = DefaultUTXOSelectionOptions()
	}

	// Filter UTXOs based on options
	filteredUTXOs := make([]*UTXO, 0, len(utxos))
	for _, utxo := range utxos {
		if options.ExcludeCoinbase && utxo.IsCoinbase {
			continue
		}
		if utxo.Value <= options.MaxDustAmount {
			continue
		}
		if utxo.Height < options.MinConfirmations {
			continue
		}
		filteredUTXOs = append(filteredUTXOs, utxo)
	}

	switch options.Strategy {
	case StrategyLargestFirst:
		return w.selectLargestFirst(filteredUTXOs, amount, options)
	case StrategySmallestFirst:
		return w.selectSmallestFirst(filteredUTXOs, amount, options)
	case StrategyRandom:
		return w.selectRandom(filteredUTXOs, amount, options)
	case StrategyOptimal:
		return w.selectOptimal(filteredUTXOs, amount, options)
	default:
		return w.selectOptimal(filteredUTXOs, amount, options)
	}
}

func (w *Wallet) selectLargestFirst(utxos []*UTXO, amount uint64, options *UTXOSelectionOptions) ([]*UTXO, int64) {
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Value > utxos[j].Value
	})

	var selected []*UTXO
	var total uint64

	for _, utxo := range utxos {
		if len(selected) >= options.MaxInputs {
			break
		}
		selected = append(selected, utxo)
		total += utxo.Value
		if total >= amount {
			return selected, int64(total - amount)
		}
	}

	return nil, -1
}

func (w *Wallet) selectSmallestFirst(utxos []*UTXO, amount uint64, options *UTXOSelectionOptions) ([]*UTXO, int64) {
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Value < utxos[j].Value
	})

	var selected []*UTXO
	var total uint64

	for _, utxo := range utxos {
		if len(selected) >= options.MaxInputs {
			break
		}
		selected = append(selected, utxo)
		total += utxo.Value
		if total >= amount {
			return selected, int64(total - amount)
		}
	}

	return nil, -1
}

func (w *Wallet) selectRandom(utxos []*UTXO, amount uint64, options *UTXOSelectionOptions) ([]*UTXO, int64) {
	// Shuffle UTXOs
	mathrand.Shuffle(len(utxos), func(i, j int) {
		utxos[i], utxos[j] = utxos[j], utxos[i]
	})

	var selected []*UTXO
	var total uint64

	for _, utxo := range utxos {
		if len(selected) >= options.MaxInputs {
			break
		}
		selected = append(selected, utxo)
		total += utxo.Value
		if total >= amount {
			return selected, int64(total - amount)
		}
	}

	return nil, -1
}

func (w *Wallet) selectOptimal(utxos []*UTXO, amount uint64, options *UTXOSelectionOptions) ([]*UTXO, int64) {
	// Sort UTXOs by value
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Value < utxos[j].Value
	})

	// Try to find a single UTXO that covers the amount with minimal change
	for _, utxo := range utxos {
		if utxo.Value >= amount && utxo.Value <= amount+options.MaxChange {
			return []*UTXO{utxo}, int64(utxo.Value - amount)
		}
	}

	// If no single UTXO is optimal, use dynamic programming to find the best combination
	dp := make([]struct {
		utxos []*UTXO
		total uint64
	}, amount+options.MaxChange+1)

	for _, utxo := range utxos {
		for i := amount + options.MaxChange; i >= utxo.Value; i-- {
			if dp[i-utxo.Value].total > 0 || i == utxo.Value {
				newTotal := dp[i-utxo.Value].total + utxo.Value
				if dp[i].total == 0 || newTotal < dp[i].total {
					newUTXOs := make([]*UTXO, len(dp[i-utxo.Value].utxos))
					copy(newUTXOs, dp[i-utxo.Value].utxos)
					newUTXOs = append(newUTXOs, utxo)
					dp[i] = struct {
						utxos []*UTXO
						total uint64
					}{newUTXOs, newTotal}
				}
			}
		}
	}

	// Find the best combination
	var bestUTXOs []*UTXO
	var bestChange int64 = -1

	for i := amount; i <= amount+options.MaxChange; i++ {
		if dp[i].total > 0 && len(dp[i].utxos) <= options.MaxInputs {
			change := int64(dp[i].total - amount)
			if change >= int64(options.MinChange) {
				bestUTXOs = dp[i].utxos
				bestChange = change
				break
			}
		}
	}

	return bestUTXOs, bestChange
}

func (w *Wallet) signTransaction(tx *Transaction, account *Account) error {
	path := fmt.Sprintf("m/44'/0'/%d'/0/0", account.Index)

	if w.useHardwareWallet {
		// Sign transaction using hardware wallet
		return w.hardwareWallet.SignTransaction(tx, path)
	}

	// Get private key
	privateKey, exists := w.keyStore.GetKey(account.Index)
	if !exists {
		return fmt.Errorf("private key not found for account %d", account.Index)
	}

	// Create signature hash
	hash := w.calculateTransactionHash(tx)

	// Sign each input
	for _, input := range tx.Inputs {
		// Get the UTXO being spent
		utxoKey := fmt.Sprintf("%s:%d", input.TxHash, input.OutputIndex)
		_, exists := account.UTXOs[utxoKey]
		if !exists {
			return fmt.Errorf("UTXO not found: %s", utxoKey)
		}

		// Sign the hash
		signature, err := ecdsa.SignASN1(rand.Reader, privateKey, []byte(hash))
		if err != nil {
			return fmt.Errorf("failed to sign transaction: %w", err)
		}

		input.Signature = signature
	}

	return nil
}

func (w *Wallet) calculateTransactionHash(tx *Transaction) string {
	// Create a deterministic string representation of the transaction
	var data strings.Builder

	// Add inputs
	for _, input := range tx.Inputs {
		data.WriteString(input.TxHash)
		data.WriteString(fmt.Sprintf("%d", input.OutputIndex))
	}

	// Add outputs
	for _, output := range tx.Outputs {
		data.WriteString(output.Address)
		data.WriteString(fmt.Sprintf("%d", output.Value))
	}

	// Hash the data
	hash := sha256.Sum256([]byte(data.String()))
	return hex.EncodeToString(hash[:])
}

// VerifyTransaction verifies a transaction's signatures
func (w *Wallet) VerifyTransaction(tx *Transaction, utxos map[string]*UTXO) error {
	// Calculate transaction hash
	hash := w.calculateTransactionHash(tx)

	// Verify each input
	for _, input := range tx.Inputs {
		// Get the UTXO being spent
		utxoKey := fmt.Sprintf("%s:%d", input.TxHash, input.OutputIndex)
		utxo, exists := utxos[utxoKey]
		if !exists {
			return fmt.Errorf("UTXO not found: %s", utxoKey)
		}

		// Verify signature
		valid := ecdsa.VerifyASN1(utxo.PublicKey, []byte(hash), input.Signature)
		if !valid {
			return ErrInvalidSignature
		}
	}

	return nil
}
