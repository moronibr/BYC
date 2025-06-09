package tests

import (
	"testing"

	"byc/internal/blockchain"
	"byc/internal/wallet"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	w, err := wallet.NewWallet()
	assert.NoError(t, err)
	assert.NotNil(t, w)
	assert.NotNil(t, w.PrivateKey)
	assert.NotNil(t, w.PublicKey)
	assert.NotEmpty(t, w.Address)
}

func TestCreateTransaction(t *testing.T) {
	// Create wallets
	sender, err := wallet.NewWallet()
	assert.NoError(t, err)

	recipient, err := wallet.NewWallet()
	assert.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Create transaction
	tx, err := sender.CreateTransaction(recipient.Address, 10.0, blockchain.Leah, bc)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, recipient.Address, tx.Outputs[0].Address)
	assert.Equal(t, 10.0, tx.Outputs[0].Value)
	assert.Equal(t, blockchain.Leah, tx.Outputs[0].CoinType)
}

func TestGetBalance(t *testing.T) {
	// Create wallets
	w, err := wallet.NewWallet()
	assert.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Test initial balance
	balance := w.GetBalance(blockchain.Leah, bc)
	assert.Equal(t, 0.0, balance)

	// Test all coin types
	balances := w.GetAllBalances(bc)
	assert.NotNil(t, balances)
	assert.Equal(t, 0.0, balances[blockchain.Leah])
	assert.Equal(t, 0.0, balances[blockchain.Shiblum])
	assert.Equal(t, 0.0, balances[blockchain.Senum])
}

func TestCreateSpecialCoins(t *testing.T) {
	// Create wallet
	w, err := wallet.NewWallet()
	assert.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Test creating Ephraim coin
	err = w.CreateEphraimCoin(bc)
	assert.Error(t, err) // Should fail due to insufficient funds

	// Test creating Manasseh coin
	err = w.CreateManassehCoin(bc)
	assert.Error(t, err) // Should fail due to insufficient funds

	// Test creating Joseph coin
	err = w.CreateJosephCoin(bc)
	assert.Error(t, err) // Should fail due to insufficient funds
}

func TestTransactionVerification(t *testing.T) {
	// Create wallets
	sender, err := wallet.NewWallet()
	assert.NoError(t, err)

	recipient, err := wallet.NewWallet()
	assert.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Create transaction
	tx, err := sender.CreateTransaction(recipient.Address, 10.0, blockchain.Leah, bc)
	assert.NoError(t, err)

	// Verify transaction
	assert.True(t, tx.Verify())
}

func TestInvalidTransactions(t *testing.T) {
	// Create wallets
	sender, err := wallet.NewWallet()
	assert.NoError(t, err)

	recipient, err := wallet.NewWallet()
	assert.NoError(t, err)

	// Create blockchain
	bc := blockchain.NewBlockchain()

	// Test negative amount
	_, err = sender.CreateTransaction(recipient.Address, -10.0, blockchain.Leah, bc)
	assert.Error(t, err)

	// Test invalid coin type
	_, err = sender.CreateTransaction(recipient.Address, 10.0, "INVALID", bc)
	assert.Error(t, err)
}
