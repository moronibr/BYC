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

	"byc/internal/blockchain"
	"byc/internal/mining"
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

	// Clear screen and show header
	fmt.Print("\033[H\033[2J")
	fmt.Println("=== BYC Mining Dashboard ===")
	fmt.Printf("Mining %s coins using %s blocks\n", coinType, blockType)
	fmt.Printf("Connected to node at %s\n", nodeAddress)
	fmt.Println("===========================")
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

	// Track start time
	startTime := time.Now()

	// Display mining status with enhanced formatting
	go func() {
		lastUpdate := time.Now()
		lastShares := int64(0)
		lastBlocks := int64(0)

		for {
			select {
			case <-done:
				return
			case status := <-statusChan:
				now := time.Now()
				elapsed := now.Sub(lastUpdate).Seconds()
				runtime := now.Sub(startTime)

				// Calculate rates
				sharesPerSecond := float64(status.Shares-lastShares) / elapsed
				blocksPerHour := float64(status.BlocksFound-lastBlocks) / (runtime.Hours())

				// Clear previous lines
				fmt.Print("\033[H\033[2J")
				fmt.Println("=== BYC Mining Dashboard ===")
				fmt.Printf("Runtime: %s\n", formatDuration(runtime))
				fmt.Println("===========================")

				// Mining Status
				fmt.Println("\nMining Status:")
				fmt.Println("-------------")
				fmt.Printf("Status: %s\n", getMiningStatus(status))
				fmt.Printf("Hash Rate: %s\n", formatHashRate(status.HashRate))
				fmt.Printf("Difficulty: %d\n", status.Difficulty)
				fmt.Printf("Current Block: %d\n", status.CurrentBlock)

				// Performance Metrics
				fmt.Println("\nPerformance Metrics:")
				fmt.Println("-------------------")
				fmt.Printf("Shares: %d (%.1f/s)\n", status.Shares, sharesPerSecond)
				fmt.Printf("Blocks Found: %d (%.2f/hour)\n", status.BlocksFound, blocksPerHour)
				fmt.Printf("Efficiency: %.1f%%\n", calculateEfficiency(status))

				// Rewards
				fmt.Println("\nRewards:")
				fmt.Println("--------")
				fmt.Printf("Current Block Reward: %.2f %s\n", status.CurrentReward, coinType)
				fmt.Printf("Total Rewards: %.2f %s\n", status.TotalRewards, coinType)
				fmt.Printf("Estimated Daily: %.2f %s\n", calculateDailyEstimate(status), coinType)

				// Network Status
				fmt.Println("\nNetwork Status:")
				fmt.Println("--------------")
				fmt.Printf("Connected Peers: %d\n", status.ConnectedPeers)
				fmt.Printf("Network Hash Rate: %s\n", formatHashRate(status.NetworkHashRate))
				fmt.Printf("Block Time: %.1f seconds\n", status.AverageBlockTime)

				// Update last values
				lastUpdate = now
				lastShares = status.Shares
				lastBlocks = status.BlocksFound
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
	fmt.Printf("Total Runtime: %s\n", formatDuration(time.Since(startTime)))
	fmt.Printf("Total Blocks Found: %d\n", stats["blocks"])
	fmt.Printf("Total Shares: %d\n", stats["shares"])
	fmt.Printf("Average Hash Rate: %s\n", formatHashRate(stats["hash_rate"].(int64)))
	fmt.Printf("Mining Address: %s\n", stats["address"])
	fmt.Printf("Total Rewards: %.2f %s\n", stats["rewards"], coinType)

	fmt.Println("\nReturning to main menu...")
	time.Sleep(1 * time.Second)
}

// Helper functions for formatting
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

func formatHashRate(rate int64) string {
	if rate >= 1000000000 {
		return fmt.Sprintf("%.2f GH/s", float64(rate)/1000000000)
	} else if rate >= 1000000 {
		return fmt.Sprintf("%.2f MH/s", float64(rate)/1000000)
	} else if rate >= 1000 {
		return fmt.Sprintf("%.2f KH/s", float64(rate)/1000)
	}
	return fmt.Sprintf("%d H/s", rate)
}

func getMiningStatus(status mining.Status) string {
	if status.IsRunning {
		return "\033[32mRunning\033[0m"
	}
	return "\033[31mStopped\033[0m"
}

func calculateEfficiency(status mining.Status) float64 {
	if status.NetworkHashRate == 0 {
		return 0
	}
	return float64(status.HashRate) / float64(status.NetworkHashRate) * 100
}

func calculateDailyEstimate(status mining.Status) float64 {
	if status.AverageBlockTime == 0 {
		return 0
	}
	blocksPerDay := 86400 / status.AverageBlockTime
	return status.CurrentReward * blocksPerDay
}
