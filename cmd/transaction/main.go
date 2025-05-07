package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/network"
	"github.com/youngchain/internal/storage"
	"go.etcd.io/bbolt"
)

var (
	dbPath      = flag.String("db", "transactions.db", "Path to the database file")
	networkAddr = flag.String("network", "", "Network address to connect to")
)

func main() {
	flag.Parse()

	// Open database
	db, err := bbolt.Open(*dbPath, 0600, nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize stores
	txStore := storage.NewTransactionStore(db)
	historyStore := storage.NewTransactionHistoryStore(db)
	feeCalculator := transaction.NewFeeCalculator()

	// Initialize network
	broadcaster := network.NewTransactionBroadcaster()
	if *networkAddr != "" {
		if err := broadcaster.AddPeer(*networkAddr); err != nil {
			log.Printf("Failed to connect to network: %v", err)
		}
	}

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "send":
		if len(os.Args) < 5 {
			printUsage()
			return
		}
		amount, err := strconv.ParseUint(os.Args[2], 10, 64)
		if err != nil {
			log.Fatalf("Invalid amount: %v", err)
		}
		toAddress := os.Args[3]
		coinType := coin.CoinType(os.Args[4])
		sendTransaction(txStore, historyStore, feeCalculator, broadcaster, amount, toAddress, coinType)

	case "balance":
		if len(os.Args) < 3 {
			printUsage()
			return
		}
		address := os.Args[2]
		coinType := coin.CoinType(os.Args[3])
		showBalance(txStore, address, coinType)

	case "history":
		if len(os.Args) < 3 {
			printUsage()
			return
		}
		address := os.Args[2]
		coinType := coin.CoinType(os.Args[3])
		showHistory(historyStore, address, coinType)

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  send <amount> <to_address> <coin_type>")
	fmt.Println("  balance <address> <coin_type>")
	fmt.Println("  history <address> <coin_type>")
}

func sendTransaction(txStore *storage.TransactionStore, historyStore *storage.TransactionHistoryStore,
	feeCalculator *transaction.FeeCalculator, broadcaster *network.TransactionBroadcaster,
	amount uint64, toAddress string, coinType coin.CoinType) {

	// Create transaction
	inputs := []*transaction.Input{
		// TODO: Select appropriate UTXOs
	}
	outputs := []*transaction.Output{
		{
			Value:   amount,
			Address: toAddress,
		},
	}

	// Calculate fee
	tx := transaction.CreateTransaction(inputs, outputs, 0, coinType)
	fee := feeCalculator.CalculateFee(tx)
	tx.Fee = fee

	// Save transaction
	if err := txStore.SaveTransaction(tx); err != nil {
		log.Fatalf("Failed to save transaction: %v", err)
	}

	// Add to history
	history := &storage.TransactionHistory{
		TxHash:    tx.CalculateHash(),
		Address:   toAddress,
		Amount:    int64(amount),
		Timestamp: time.Now().Unix(),
		CoinType:  coinType,
		Status:    "pending",
	}
	if err := historyStore.AddHistory(history); err != nil {
		log.Printf("Failed to add history: %v", err)
	}

	// Broadcast transaction
	if err := broadcaster.BroadcastTransaction(tx); err != nil {
		log.Printf("Failed to broadcast transaction: %v", err)
	}

	fmt.Printf("Transaction sent: %x\n", tx.CalculateHash())
}

func showBalance(txStore *storage.TransactionStore, address string, coinType coin.CoinType) {
	balance, err := txStore.GetBalance(address, coinType)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Printf("Balance: %d %s\n", balance, coinType)
}

func showHistory(historyStore *storage.TransactionHistoryStore, address string, coinType coin.CoinType) {
	history, err := historyStore.GetHistory(address, coinType)
	if err != nil {
		log.Fatalf("Failed to get history: %v", err)
	}

	fmt.Printf("Transaction history for %s (%s):\n", address, coinType)
	for _, entry := range history {
		fmt.Printf("Tx: %x\n", entry.TxHash)
		fmt.Printf("  Amount: %d\n", entry.Amount)
		fmt.Printf("  Time: %s\n", time.Unix(entry.Timestamp, 0))
		fmt.Printf("  Status: %s\n", entry.Status)
		if len(entry.BlockHash) > 0 {
			fmt.Printf("  Block: %x\n", entry.BlockHash)
		}
		fmt.Println()
	}
}
