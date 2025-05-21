package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/mining"
)

func handleMining(cmd *flag.FlagSet) {
	// Get values from flags
	coinType := cmd.Lookup("coin").Value.String()
	blockType := cmd.Lookup("block").Value.String()
	nodeAddress := cmd.Lookup("address").Value.String()

	// Validate coin type
	var coin blockchain.CoinType
	switch coinType {
	case "leah":
		coin = blockchain.Leah
	case "shiblum":
		coin = blockchain.Shiblum
	case "shiblon":
		coin = blockchain.Shiblon
	default:
		fmt.Printf("Invalid coin type: %s\n", coinType)
		fmt.Println("Valid coin types: leah, shiblum, shiblon")
		os.Exit(1)
	}

	// Validate block type
	var block blockchain.BlockType
	switch blockType {
	case "golden":
		block = blockchain.GoldenBlock
	case "silver":
		block = blockchain.SilverBlock
	default:
		fmt.Printf("Invalid block type: %s\n", blockType)
		fmt.Println("Valid block types: golden, silver")
		os.Exit(1)
	}

	// Create blockchain instance
	bc := blockchain.NewBlockchain()

	// Create miner
	miner, err := mining.NewMiner(bc, block, coin, nodeAddress)
	if err != nil {
		log.Fatalf("Failed to create miner: %v", err)
	}

	fmt.Printf("\nStarting mining for %s coins using %s blocks...\n", coinType, blockType)
	fmt.Printf("Connected to node at %s\n", nodeAddress)
	fmt.Println("Press Ctrl+C to stop mining and return to the main menu")
	fmt.Println("--------------------------------------------------------")

	// Create channels for mining status updates
	statusChan := make(chan mining.Status)
	done := make(chan bool)

	// Start mining status updates
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				status := miner.GetStatus()
				statusChan <- status
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Start mining
	ctx := context.Background()
	miner.Start(ctx)

	// Display mining status
	go func() {
		for {
			select {
			case <-done:
				return
			case status := <-statusChan:
				fmt.Printf("\rHash Rate: %d H/s | Shares: %d | Blocks Found: %d | Difficulty: %d",
					status.HashRate,
					status.Shares,
					status.BlocksFound,
					status.Difficulty)
			}
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	done <- true
	fmt.Println("\n\nStopping miner...")
	miner.Stop()
	fmt.Println("Returning to main menu...")
	time.Sleep(1 * time.Second) // Give user time to read the message
}
