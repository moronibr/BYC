package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/monitoring"
	"github.com/moroni/BYC/internal/network"
)

type SystemMetrics struct {
	CPU     float64
	Memory  float64
	Network float64
	Time    time.Time
}

func handleDashboard(cmd *flag.FlagSet) {
	fmt.Println("\n=== System Dashboard ===")
	fmt.Println("Press Ctrl+C to return to the main menu")
	fmt.Println("----------------------------------------")

	// Create a channel to handle graceful shutdown
	done := make(chan bool)

	// Initialize blockchain and node
	bc := blockchain.NewBlockchain()
	node, err := network.NewNode(&network.Config{})
	if err != nil {
		fmt.Printf("Error initializing node: %v\n", err)
		return
	}

	// Create monitoring instance with initialized components
	monitor := monitoring.NewMonitor(bc, node, "")

	// Start the dashboard in a goroutine
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				showDashboard(monitor)
				time.Sleep(5 * time.Second)
				fmt.Println("\n---")
			}
		}
	}()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for Ctrl+C
	<-sigChan

	// Cleanup
	done <- true
	fmt.Println("\nReturning to main menu...")
	time.Sleep(1 * time.Second) // Give user time to read the message
}

func showDashboard(monitor *monitoring.Monitor) {
	// Get system metrics
	metrics := getSystemMetrics()
	health := monitor.GetHealth()

	fmt.Println("System Metrics:")
	fmt.Println("---------------")
	fmt.Printf("CPU Usage: %.1f%%\n", metrics.CPU)
	fmt.Printf("Memory Usage: %.1f%%\n", metrics.Memory)
	fmt.Printf("Network Traffic: %.1f MB/s\n", metrics.Network)
	fmt.Printf("Last Update: %s\n", metrics.Time.Format("15:04:05"))

	fmt.Println("\nNetwork Status:")
	fmt.Println("---------------")

	fmt.Printf("Status: %s\n", health["status"])

	details := health["details"].(map[string]interface{})
	network := details["network"].(map[string]interface{})
	blockchain := details["blockchain"].(map[string]interface{})
	system := details["system"].(map[string]interface{})

	fmt.Printf("Peers: %d\n", network["peers"])
	fmt.Printf("Last Sync: %s\n", network["last_sync_time"])

	// Handle block counts with proper type assertions
	goldenBlocks := int(blockchain["golden_blocks"].(int))
	silverBlocks := int(blockchain["silver_blocks"].(int))
	fmt.Printf("Block Height: %d\n", goldenBlocks+silverBlocks)

	fmt.Println("\nBlockchain Status:")
	fmt.Println("-----------------")
	fmt.Printf("Golden Blocks: %d\n", goldenBlocks)
	fmt.Printf("Silver Blocks: %d\n", silverBlocks)
	fmt.Printf("Sync Status: %v\n", blockchain["is_synced"])

	fmt.Println("\nSystem Health:")
	fmt.Println("-------------")
	fmt.Printf("Memory Usage: %.1f MB\n", float64(system["memory_usage_bytes"].(int64))/1024/1024)
	fmt.Printf("CPU Usage: %.1f%%\n", system["cpu_usage_percent"].(float64))
	fmt.Printf("Disk Usage: %.1f GB\n", float64(system["disk_usage_bytes"].(int64))/1024/1024/1024)
}

func getSystemMetrics() SystemMetrics {
	// TODO: Implement actual system metrics collection
	// This is a placeholder that returns dummy data
	return SystemMetrics{
		CPU:     45.0,
		Memory:  2.3,
		Network: 1.2,
		Time:    time.Now(),
	}
}
