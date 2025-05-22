package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/mining"
)

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

	// Validate block type
	var bt blockchain.BlockType
	switch *blockType {
	case "golden":
		bt = blockchain.GoldenBlock
	case "silver":
		bt = blockchain.SilverBlock
	default:
		fmt.Printf("Invalid block type: %s\n", *blockType)
		os.Exit(1)
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
	// Add other coin types as needed
	default:
		fmt.Printf("Invalid coin type: %s\n", *coinType)
		os.Exit(1)
	}

	// Create miner
	miner, err := mining.NewMiner(bc, bt, ct, *address)
	if err != nil {
		fmt.Printf("Error creating miner: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start mining
	go miner.Start()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down...")
	miner.Stop()
}
