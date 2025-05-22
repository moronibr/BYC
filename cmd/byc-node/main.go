package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/mining"
)

func displayMenu() {
	fmt.Println("\n=== BYC CLI Menu ===")
	fmt.Println("1. Network Operations")
	fmt.Println("2. Wallet Operations")
	fmt.Println("3. Dashboard")
	fmt.Println("4. Mining")
	fmt.Println("5. View Genesis Block")
	fmt.Println("6. Exit")
	fmt.Print("\nEnter your choice (1-6): ")
}

func main() {
	// Parse command line flags
	blockType := flag.String("block", "golden", "Block type to mine (golden/silver)")
	coinType := flag.String("coin", "leah", "Coin type to mine")
	address := flag.String("address", "", "Mining address")
	viewGenesis := flag.Bool("view-genesis", false, "View Genesis block information")
	saveGenesis := flag.String("save-genesis", "", "Save Genesis block information to file")
	flag.Parse()

	// Create new blockchain instance
	bc := blockchain.NewBlockchain()

	// If view-genesis flag is set, display Genesis block
	if *viewGenesis {
		bc.DisplayGenesisBlock()
	}

	// If save-genesis flag is set, save Genesis block info to file
	if *saveGenesis != "" {
		if err := bc.SaveGenesisBlockInfo(*saveGenesis); err != nil {
			fmt.Printf("Error saving Genesis block info: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Genesis block information saved to %s\n", *saveGenesis)
	}

	// Exit if only viewing or saving Genesis block
	if *viewGenesis || *saveGenesis != "" {
		return
	}

	// Interactive menu loop
	for {
		displayMenu()
		var choice int
		fmt.Scan(&choice)

		switch choice {
		case 1:
			// Network Operations
			fmt.Println("Network Operations selected")
		case 2:
			// Wallet Operations
			fmt.Println("Wallet Operations selected")
		case 3:
			// Dashboard
			fmt.Println("Dashboard selected")
		case 4:
			// Mining
			startMining(bc, blockType, coinType, address)
		case 5:
			// View Genesis Block
			bc.DisplayGenesisBlock()
			fmt.Print("\nPress Enter to continue...")
			fmt.Scanln()
		case 6:
			// Exit
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func startMining(bc *blockchain.Blockchain, blockType, coinType, address *string) {
	// Validate block type
	var bt blockchain.BlockType
	switch *blockType {
	case "golden":
		bt = blockchain.GoldenBlock
	case "silver":
		bt = blockchain.SilverBlock
	default:
		fmt.Printf("Invalid block type: %s\n", *blockType)
		return
	}

	// Validate coin type
	var ct blockchain.CoinType
	switch *coinType {
	case "leah":
		ct = blockchain.Leah
	case "shiblum":
		ct = blockchain.Shiblum
	case "shiblon":
		ct = blockchain.Shiblon
	default:
		fmt.Printf("Invalid coin type: %s\n", *coinType)
		return
	}

	// Create miner
	miner, err := mining.NewMiner(bc, bt, ct, *address)
	if err != nil {
		fmt.Printf("Error creating miner: %v\n", err)
		return
	}

	// Create context for mining
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start mining
	go miner.Start(ctx)

	fmt.Println("Mining started. Press Ctrl+C to stop.")
	<-sigChan
	fmt.Println("\nShutting down mining...")
	cancel()
	miner.Stop()
}
