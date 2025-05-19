package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/byc/internal/crypto"
)

// CoinType represents the different types of coins in the system
type CoinType string

const (
	// Golden Block Coins
	Leah    CoinType = "LEAH"
	Shiblum CoinType = "SHIBLUM"
	Shiblon CoinType = "SHIBLON"
	Senine  CoinType = "SENINE"
	Seon    CoinType = "SEON"
	Shum    CoinType = "SHUM"
	Limnah  CoinType = "LIMNAH"
	Antion  CoinType = "ANTION"

	// Silver Block Coins
	Senum CoinType = "SENUM"
	Amnor CoinType = "AMNOR"
	Ezrom CoinType = "EZROM"
	Onti  CoinType = "ONTI"

	// Special Coins
	Ephraim  CoinType = "EPHRAIM"
	Manasseh CoinType = "MANASSEH"
	Joseph   CoinType = "JOSEPH"
)

// BlockType represents which blockchain a block belongs to
type BlockType string

const (
	GoldenBlock BlockType = "GOLDEN"
	SilverBlock BlockType = "SILVER"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp    int64
	Transactions []Transaction
	PrevHash     []byte
	Hash         []byte
	Nonce        int64
	BlockType    BlockType
	Difficulty   int
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	ID        []byte
	Inputs    []TxInput
	Outputs   []TxOutput
	Timestamp time.Time
	BlockType BlockType
}

// TxInput represents a transaction input
type TxInput struct {
	TxID        []byte
	OutputIndex int
	Amount      float64
	Signature   []byte
	PublicKey   []byte
	Address     string
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value         float64
	CoinType      CoinType
	PublicKeyHash []byte
	Address       string
}

// Wallet represents a user's wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

// Node represents a node in the P2P network
type Node struct {
	Address    string
	Peers      []string
	BlockType  BlockType
	IsMining   bool
	MiningCoin CoinType
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount float64, coinType CoinType, inputs []TxInput, outputs []TxOutput) *Transaction {
	tx := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now(),
		BlockType: GetBlockType(coinType),
	}

	// Calculate transaction ID
	tx.ID = tx.CalculateHash()

	return tx
}

// CalculateHash calculates the hash of a transaction
func (tx *Transaction) CalculateHash() []byte {
	// Create a copy of the transaction without signatures
	txCopy := *tx
	txCopy.Inputs = make([]TxInput, len(tx.Inputs))
	copy(txCopy.Inputs, tx.Inputs)

	// Clear signatures and public keys
	for i := range txCopy.Inputs {
		txCopy.Inputs[i].Signature = nil
		txCopy.Inputs[i].PublicKey = nil
	}

	// Convert the transaction to bytes
	data, err := json.Marshal(txCopy)
	if err != nil {
		return nil
	}

	// Calculate the hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// Sign signs a transaction with the given private key
func (tx *Transaction) Sign(privateKey []byte) error {
	txCopy := tx.TrimmedCopy()

	for i, input := range txCopy.Inputs {
		// Set the public key for this input
		txCopy.Inputs[i].PublicKey = input.PublicKey

		// Calculate the hash of the transaction
		hash := txCopy.CalculateHash()

		// Sign the hash with the private key
		signature, err := crypto.Sign(hash, privateKey)
		if err != nil {
			return err
		}

		// Set the signature for this input
		tx.Inputs[i].Signature = signature
	}

	return nil
}

// Verify verifies the signature of a transaction
func (tx *Transaction) Verify() bool {
	txCopy := tx.TrimmedCopy()

	for i, input := range tx.Inputs {
		// Set the public key for this input
		txCopy.Inputs[i].PublicKey = input.PublicKey

		// Calculate the hash of the transaction
		hash := txCopy.CalculateHash()

		// Verify the signature
		if !crypto.Verify(hash, input.Signature, input.PublicKey) {
			return false
		}
	}

	return true
}

// IsCoinbase checks if a transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TxID) == 0 && tx.Inputs[0].OutputIndex == -1
}

// GetTotalInput returns the total input amount
func (tx *Transaction) GetTotalInput() float64 {
	var total float64
	for _, input := range tx.Inputs {
		total += input.Amount
	}
	return total
}

// GetTotalOutput returns the total output amount
func (tx *Transaction) GetTotalOutput() float64 {
	var total float64
	for _, output := range tx.Outputs {
		total += output.Value
	}
	return total
}

// GetFee returns the transaction fee
func (tx *Transaction) GetFee() float64 {
	return tx.GetTotalInput() - tx.GetTotalOutput()
}

// Validate validates a transaction against the UTXO set
func (tx *Transaction) Validate(utxoSet *UTXOSet) bool {
	// Verify the transaction signature
	if !tx.Verify() {
		return false
	}

	// Check if the transaction is a coinbase transaction
	if tx.IsCoinbase() {
		return true
	}

	// Get the total input amount
	totalInput := tx.GetTotalInput()

	// Get the total output amount
	totalOutput := tx.GetTotalOutput()

	// Check if the input amount is sufficient
	if totalInput < totalOutput {
		return false
	}

	// Check if all inputs are valid UTXOs
	for _, input := range tx.Inputs {
		utxo := utxoSet.GetUTXO(input.TxID, input.OutputIndex)
		if len(utxo.TxID) == 0 {
			return false
		}

		// Check if the input belongs to the sender
		if !bytes.Equal(utxo.PublicKeyHash, crypto.HashPublicKey(input.PublicKey)) {
			return false
		}
	}

	return true
}

// MiningDifficulty returns the difficulty multiplier for a given coin type
func MiningDifficulty(coinType CoinType) int {
	switch coinType {
	case Leah:
		return 1
	case Shiblum:
		return 2
	case Shiblon:
		return 4
	default:
		return 0 // Not mineable
	}
}

// IsMineable checks if a coin type is mineable
func IsMineable(coinType CoinType) bool {
	switch coinType {
	case Leah, Shiblum, Shiblon, Senine, Seon, Shum, Limnah, Antion, Senum, Amnor, Ezrom, Onti:
		return true
	default:
		return false
	}
}

// CanTransferBetweenBlocks checks if a coin type can be transferred between blocks
func CanTransferBetweenBlocks(coinType CoinType) bool {
	switch coinType {
	case Antion, Senum, Amnor, Ezrom, Onti:
		return true
	default:
		return false
	}
}

// GetBlockType returns the block type for a coin type
func GetBlockType(coinType CoinType) BlockType {
	switch coinType {
	case Leah, Shiblum, Shiblon, Senine, Seon, Shum, Limnah, Antion:
		return GoldenBlock
	case Senum, Amnor, Ezrom, Onti:
		return SilverBlock
	default:
		return ""
	}
}

// TrimmedCopy creates a copy of the transaction without signatures
func (tx *Transaction) TrimmedCopy() *Transaction {
	txCopy := *tx
	txCopy.Inputs = make([]TxInput, len(tx.Inputs))
	copy(txCopy.Inputs, tx.Inputs)

	// Clear signatures and public keys
	for i := range txCopy.Inputs {
		txCopy.Inputs[i].Signature = nil
		txCopy.Inputs[i].PublicKey = nil
	}

	return &txCopy
}

// String returns the string representation of the coin type
func (c CoinType) String() string {
	switch c {
	case Leah:
		return "Leah"
	case Shiblum:
		return "Shiblum"
	case Shiblon:
		return "Shiblon"
	case Senine:
		return "Senine"
	case Seon:
		return "Seon"
	case Shum:
		return "Shum"
	case Limnah:
		return "Limnah"
	case Antion:
		return "Antion"
	case Senum:
		return "Senum"
	case Amnor:
		return "Amnor"
	case Ezrom:
		return "Ezrom"
	case Onti:
		return "Onti"
	case Ephraim:
		return "Ephraim"
	case Manasseh:
		return "Manasseh"
	case Joseph:
		return "Joseph"
	default:
		return "Unknown"
	}
}
