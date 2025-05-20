package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/wallet"
)

func handleWallet(cmd *flag.FlagSet) {
	// Get the action from the flag
	action := cmd.Lookup("action").Value.String()

	switch action {
	case "create":
		createWallet()
	case "balance":
		showBalance()
	case "send":
		handleSendCoins()
	default:
		fmt.Println("Please specify an action: create, balance, or send")
		os.Exit(1)
	}
}

func createWallet() {
	// Create a new wallet
	w, err := wallet.NewWallet()
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}

	fmt.Println("\n=== New Wallet Created ===")
	fmt.Printf("Address: %s\n", w.Address)
	fmt.Println("Please save this address securely!")
	fmt.Println("===========================\n")
}

func showBalance() {
	// Get the mining wallet
	walletsDir := "wallets"
	walletFile := filepath.Join(walletsDir, "mining_wallet.json")

	if _, err := os.Stat(walletFile); err != nil {
		fmt.Println("No wallet found. Please mine some coins first.")
		return
	}

	// Read wallet file
	data, err := os.ReadFile(walletFile)
	if err != nil {
		fmt.Printf("Error reading wallet file: %v\n", err)
		return
	}

	var walletInfo struct {
		Address string
		Rewards map[string]float64
	}
	if err := json.Unmarshal(data, &walletInfo); err != nil {
		fmt.Printf("Error parsing wallet file: %v\n", err)
		return
	}

	fmt.Println("\n=== Wallet Balance ===")
	fmt.Printf("Address: %s\n", walletInfo.Address)
	fmt.Println("\nRewards:")
	for coinType, amount := range walletInfo.Rewards {
		fmt.Printf("%s: %.2f\n", coinType, amount)
	}
	fmt.Println("=====================\n")
}

func handleSendCoins() {
	// Create a new blockchain instance
	bc := blockchain.NewBlockchain()

	// Get the mining wallet
	walletsDir := "wallets"
	walletFile := filepath.Join(walletsDir, "mining_wallet.json")

	if _, err := os.Stat(walletFile); err != nil {
		fmt.Println("No wallet found. Please mine some coins first.")
		return
	}

	// Read wallet file
	data, err := os.ReadFile(walletFile)
	if err != nil {
		fmt.Printf("Error reading wallet file: %v\n", err)
		return
	}

	var walletInfo struct {
		Address string
		Rewards map[string]float64
	}
	if err := json.Unmarshal(data, &walletInfo); err != nil {
		fmt.Printf("Error parsing wallet file: %v\n", err)
		return
	}

	// Get recipient address
	fmt.Print("Enter recipient address: ")
	reader := bufio.NewReader(os.Stdin)
	recipient, _ := reader.ReadString('\n')
	recipient = strings.TrimSpace(recipient)

	// Get amount
	fmt.Print("Enter amount to send: ")
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		fmt.Println("Invalid amount")
		return
	}

	// Get coin type
	fmt.Println("\nSelect coin type:")
	fmt.Println("1. Leah")
	fmt.Println("2. Shiblum")
	fmt.Println("3. Shiblon")
	fmt.Print("Enter choice (1-3): ")
	coinChoice, _ := reader.ReadString('\n')
	coinChoice = strings.TrimSpace(coinChoice)

	var coinType blockchain.CoinType
	switch coinChoice {
	case "1":
		coinType = blockchain.Leah
	case "2":
		coinType = blockchain.Shiblum
	case "3":
		coinType = blockchain.Shiblon
	default:
		fmt.Println("Invalid coin type")
		return
	}

	// Check if we have enough balance
	if walletInfo.Rewards[string(coinType)] < amount {
		fmt.Printf("Insufficient balance. You have %.2f %s\n", walletInfo.Rewards[string(coinType)], coinType)
		return
	}

	// Create and send transaction
	w := &wallet.Wallet{Address: walletInfo.Address}
	tx, err := w.CreateTransaction(recipient, amount, coinType, bc)
	if err != nil {
		fmt.Printf("Error creating transaction: %v\n", err)
		return
	}

	// Add transaction to blockchain
	if err := bc.AddTransaction(tx); err != nil {
		fmt.Printf("Error adding transaction: %v\n", err)
		return
	}

	fmt.Printf("\nTransaction sent successfully!\n")
	fmt.Printf("Amount: %.2f %s\n", amount, coinType)
	fmt.Printf("To: %s\n", recipient)
	fmt.Printf("Transaction ID: %x\n", tx.ID)
}
