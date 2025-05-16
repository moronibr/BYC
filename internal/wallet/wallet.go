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

	// Create transaction
	tx, err := bc.CreateTransaction(w.Address, to, amount, coinType)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	return &tx, nil
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

	// TODO: Implement the conversion logic
	// This would involve creating a special transaction that converts
	// the required coins into an Ephraim coin

	return nil
}

// CreateManassehCoin creates a Manasseh coin from Silver Block coins
func (w *Wallet) CreateManassehCoin(bc *blockchain.Blockchain) error {
	// Check if we have enough coins to create a Manasseh coin
	balances := w.GetAllBalances(bc)
	if balances[blockchain.Onti] < 1 {
		return fmt.Errorf("insufficient Onti coins to create Manasseh coin")
	}

	// TODO: Implement the conversion logic
	// This would involve creating a special transaction that converts
	// the required coins into a Manasseh coin

	return nil
}

// CreateJosephCoin creates a Joseph coin from Ephraim and Manasseh coins
func (w *Wallet) CreateJosephCoin(bc *blockchain.Blockchain) error {
	// Check if we have both Ephraim and Manasseh coins
	balances := w.GetAllBalances(bc)
	if balances[blockchain.Ephraim] < 1 || balances[blockchain.Manasseh] < 1 {
		return fmt.Errorf("insufficient Ephraim or Manasseh coins to create Joseph coin")
	}

	// TODO: Implement the conversion logic
	// This would involve creating a special transaction that combines
	// an Ephraim coin and a Manasseh coin into a Joseph coin

	return nil
}
