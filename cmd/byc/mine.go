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
		lastUpdate := time.Now()
		lastShares := int64(0)
		for {
			select {
			case <-done:
				return
			case status := <-statusChan:
				now := time.Now()
				elapsed := now.Sub(lastUpdate).Seconds()
				sharesPerSecond := float64(status.Shares-lastShares) / elapsed

				// Clear previous line
				fmt.Print("\033[2K\r")

				// Display mining status with better formatting
				fmt.Printf("Mining Status | Hash Rate: %d H/s | Shares: %d (%.1f/s) | Blocks: %d | Difficulty: %d",
					status.HashRate,
					status.Shares,
					sharesPerSecond,
					status.BlocksFound,
					status.Difficulty)

				// Update last values
				lastUpdate = now
				lastShares = status.Shares
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

	// Show final statistics
	stats := miner.GetMiningStats()
	fmt.Println("\nMining Statistics:")
	fmt.Println("-----------------")
	fmt.Printf("Total Blocks Found: %d\n", stats["blocks"])
	fmt.Printf("Total Shares: %d\n", stats["shares"])
	fmt.Printf("Final Hash Rate: %d H/s\n", stats["hash_rate"])
	fmt.Printf("Mining Address: %s\n", stats["address"])
	fmt.Printf("Total Rewards: %.2f %s\n", stats["rewards"], coinType)

	fmt.Println("\nReturning to main menu...")
	time.Sleep(1 * time.Second) // Give user time to read the message
}
