package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/byc/internal/blockchain"
)

// Wallet represents a user's wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
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

// CreateTransaction creates a new transaction
func (w *Wallet) CreateTransaction(to string, amount float64, coinType blockchain.CoinType, bc *blockchain.Blockchain) (*blockchain.Transaction, error) {
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
				PubKey:      []byte(w.Address),
			}
			inputs = append(inputs, input)
			totalInput += utxo.Amount

			if totalInput >= amount {
				break
			}
		}
	}

	if totalInput < amount {
		return nil, fmt.Errorf("insufficient funds")
	}

	// Create outputs
	outputs := []blockchain.TxOutput{
		{
			Value:      amount,
			CoinType:   coinType,
			PubKeyHash: []byte(to),
			Address:    to,
		},
	}

	// Add change output if needed
	if totalInput > amount {
		outputs = append(outputs, blockchain.TxOutput{
			Value:      totalInput - amount,
			CoinType:   coinType,
			PubKeyHash: []byte(w.Address),
			Address:    w.Address,
		})
	}

	// Create transaction
	tx := blockchain.NewTransaction(w.Address, to, amount, coinType, inputs, outputs)

	// Sign transaction
	if err := tx.Sign(w.PrivateKey); err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return tx, nil
}

// GetBalance returns the balance for a specific coin type
func (w *Wallet) GetBalance(coinType blockchain.CoinType, bc *blockchain.Blockchain) float64 {
	return bc.GetBalance(w.Address, coinType)
}

// GetAllBalances returns balances for all coin types
func (w *Wallet) GetAllBalances(bc *blockchain.Blockchain) map[blockchain.CoinType]float64 {
	balances := make(map[blockchain.CoinType]float64)

	// Golden Block coins
	balances[blockchain.Leah] = w.GetBalance(blockchain.Leah, bc)
	balances[blockchain.Shiblum] = w.GetBalance(blockchain.Shiblum, bc)
	balances[blockchain.Shiblon] = w.GetBalance(blockchain.Shiblon, bc)
	balances[blockchain.Senine] = w.GetBalance(blockchain.Senine, bc)
	balances[blockchain.Seon] = w.GetBalance(blockchain.Seon, bc)
	balances[blockchain.Shum] = w.GetBalance(blockchain.Shum, bc)
	balances[blockchain.Limnah] = w.GetBalance(blockchain.Limnah, bc)
	balances[blockchain.Antion] = w.GetBalance(blockchain.Antion, bc)

	// Silver Block coins
	balances[blockchain.Senum] = w.GetBalance(blockchain.Senum, bc)
	balances[blockchain.Amnor] = w.GetBalance(blockchain.Amnor, bc)
	balances[blockchain.Ezrom] = w.GetBalance(blockchain.Ezrom, bc)
	balances[blockchain.Onti] = w.GetBalance(blockchain.Onti, bc)

	// Special coins
	balances[blockchain.Ephraim] = w.GetBalance(blockchain.Ephraim, bc)
	balances[blockchain.Manasseh] = w.GetBalance(blockchain.Manasseh, bc)
	balances[blockchain.Joseph] = w.GetBalance(blockchain.Joseph, bc)

	return balances
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
