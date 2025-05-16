package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/mining"
	"github.com/byc/internal/network"
	"github.com/byc/internal/wallet"
)

func main() {
	// Parse command line flags
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)
	walletCmd := flag.NewFlagSet("wallet", flag.ExitOnError)

	// Node command flags
	nodeAddress := nodeCmd.String("address", "localhost:3000", "Node address")
	peerAddress := nodeCmd.String("peer", "", "Peer address to connect to")

	// Mine command flags
	mineCoin := mineCmd.String("coin", "", "Coin type to mine (leah, shiblum, shiblon)")
	mineBlock := mineCmd.String("block", "", "Block type to mine (golden, silver)")

	// Wallet command flags
	walletAction := walletCmd.String("action", "", "Wallet action (create, balance, send)")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'node', 'mine', or 'wallet' subcommands")
		os.Exit(1)
	}

	// Create blockchain instance
	bc := blockchain.NewBlockchain()

	switch os.Args[1] {
	case "node":
		nodeCmd.Parse(os.Args[2:])
		runNode(bc, *nodeAddress, *peerAddress)
	case "mine":
		mineCmd.Parse(os.Args[2:])
		runMining(bc, *mineCoin, *mineBlock)
	case "wallet":
		walletCmd.Parse(os.Args[2:])
		runWallet(bc, *walletAction)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runNode(bc *blockchain.Blockchain, address, peerAddress string) {
	// Create and start node
	node := network.NewNode(address, bc)
	if err := node.Start(); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// Connect to peer if specified
	if peerAddress != "" {
		if err := node.Connect(peerAddress); err != nil {
			log.Printf("Failed to connect to peer: %v", err)
		}
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func runMining(bc *blockchain.Blockchain, coinType, blockType string) {
	// Convert string to CoinType
	var coin blockchain.CoinType
	switch coinType {
	case "leah":
		coin = blockchain.Leah
	case "shiblum":
		coin = blockchain.Shiblum
	case "shiblon":
		coin = blockchain.Shiblon
	default:
		log.Fatalf("Invalid coin type: %s", coinType)
	}

	// Convert string to BlockType
	var block blockchain.BlockType
	switch blockType {
	case "golden":
		block = blockchain.GoldenBlock
	case "silver":
		block = blockchain.SilverBlock
	default:
		log.Fatalf("Invalid block type: %s", blockType)
	}

	// Create miner
	miner, err := mining.NewMiner(bc, block, coin, "localhost:3000")
	if err != nil {
		log.Fatalf("Failed to create miner: %v", err)
	}

	// Start mining
	ctx := context.Background()
	miner.Start(ctx)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop mining
	miner.Stop()
}

func runWallet(bc *blockchain.Blockchain, action string) {
	switch action {
	case "create":
		// Create new wallet
		w, err := wallet.NewWallet()
		if err != nil {
			log.Fatalf("Failed to create wallet: %v", err)
		}
		fmt.Printf("Created new wallet with address: %s\n", w.Address)

	case "balance":
		// TODO: Implement balance checking
		fmt.Println("Balance checking not implemented yet")

	case "send":
		// TODO: Implement sending coins
		fmt.Println("Sending coins not implemented yet")

	default:
		fmt.Printf("Unknown wallet action: %s\n", action)
		os.Exit(1)
	}
}
