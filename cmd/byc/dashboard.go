package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"byc/internal/blockchain"
	"byc/internal/monitoring"
	"byc/internal/network"
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

	// Clear screen and show header
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println("=== BYC System Dashboard ===")
	fmt.Printf("Last Update: %s\n", metrics.Time.Format("2006-01-02 15:04:05"))
	fmt.Println("===========================")

	// System Metrics with color coding
	fmt.Println("\nSystem Metrics:")
	fmt.Println("---------------")
	cpuColor := getColorForMetric(metrics.CPU, 80, 60)
	memColor := getColorForMetric(metrics.Memory, 80, 60)
	netColor := getColorForMetric(metrics.Network, 80, 60)
	fmt.Printf("CPU Usage: %s%.1f%%%s\n", cpuColor, metrics.CPU, "\033[0m")
	fmt.Printf("Memory Usage: %s%.1f%%%s\n", memColor, metrics.Memory, "\033[0m")
	fmt.Printf("Network Traffic: %s%.1f MB/s%s\n", netColor, metrics.Network, "\033[0m")

	// Network Status with detailed peer information
	fmt.Println("\nNetwork Status:")
	fmt.Println("---------------")
	statusColor := getStatusColor(health["status"].(string))
	fmt.Printf("Status: %s%s%s\n", statusColor, health["status"], "\033[0m")

	details := health["details"].(map[string]interface{})
	network := details["network"].(map[string]interface{})
	blockchain := details["blockchain"].(map[string]interface{})
	system := details["system"].(map[string]interface{})

	// Enhanced peer information
	peers := network["peers"].(int)
	fmt.Printf("Connected Peers: %d\n", peers)
	if peers > 0 {
		fmt.Println("\nTop Peers:")
		peerList := network["peer_list"].([]interface{})
		for i, p := range peerList {
			if i >= 5 { // Show only top 5 peers
				break
			}
			peer := p.(map[string]interface{})
			fmt.Printf("- %s (Latency: %dms, Blocks: %d)\n",
				peer["address"],
				peer["latency"],
				peer["blocks"])
		}
	}

	// Enhanced blockchain status
	fmt.Println("\nBlockchain Status:")
	fmt.Println("-----------------")
	goldenBlocks := int(blockchain["golden_blocks"].(int))
	silverBlocks := int(blockchain["silver_blocks"].(int))
	totalBlocks := goldenBlocks + silverBlocks

	fmt.Printf("Total Blocks: %d\n", totalBlocks)
	fmt.Printf("Golden Blocks: %d (%.1f%%)\n", goldenBlocks, float64(goldenBlocks)/float64(totalBlocks)*100)
	fmt.Printf("Silver Blocks: %d (%.1f%%)\n", silverBlocks, float64(silverBlocks)/float64(totalBlocks)*100)

	// Show recent block information
	if recentBlocks, ok := blockchain["recent_blocks"].([]interface{}); ok {
		fmt.Println("\nRecent Blocks:")
		for i, b := range recentBlocks {
			if i >= 3 { // Show only last 3 blocks
				break
			}
			block := b.(map[string]interface{})
			fmt.Printf("- Block %d: %s (Type: %s, Txs: %d)\n",
				block["height"],
				block["hash"],
				block["type"],
				block["transactions"])
		}
	}

	// Enhanced system health
	fmt.Println("\nSystem Health:")
	fmt.Println("-------------")
	memUsage := float64(system["memory_usage_bytes"].(int64)) / 1024 / 1024
	cpuUsage := system["cpu_usage_percent"].(float64)
	diskUsage := float64(system["disk_usage_bytes"].(int64)) / 1024 / 1024 / 1024

	fmt.Printf("Memory: %s%.1f MB%s\n", getColorForMetric(cpuUsage, 80, 60), memUsage, "\033[0m")
	fmt.Printf("CPU: %s%.1f%%%s\n", getColorForMetric(cpuUsage, 80, 60), cpuUsage, "\033[0m")
	fmt.Printf("Disk: %s%.1f GB%s\n", getColorForMetric(diskUsage, 80, 60), diskUsage, "\033[0m")

	// Show any active alerts
	if alerts, ok := system["alerts"].([]interface{}); ok && len(alerts) > 0 {
		fmt.Println("\nActive Alerts:")
		for _, alert := range alerts {
			a := alert.(map[string]interface{})
			fmt.Printf("⚠️  %s: %s\n", a["type"], a["message"])
		}
	}

	// Show mining status if active
	if mining, ok := system["mining"].(map[string]interface{}); ok {
		fmt.Println("\nMining Status:")
		fmt.Println("-------------")
		fmt.Printf("Status: %s\n", mining["status"])
		fmt.Printf("Hash Rate: %d H/s\n", mining["hash_rate"])
		fmt.Printf("Current Block: %d\n", mining["current_block"])
		fmt.Printf("Rewards: %.2f %s\n", mining["rewards"], mining["coin_type"])
	}

	fmt.Println("\nPress Ctrl+C to return to main menu...")
}

// Helper function to get color based on metric value
func getColorForMetric(value float64, warningThreshold, criticalThreshold float64) string {
	if value >= criticalThreshold {
		return "\033[31m" // Red
	} else if value >= warningThreshold {
		return "\033[33m" // Yellow
	}
	return "\033[32m" // Green
}

// Helper function to get color based on status
func getStatusColor(status string) string {
	switch status {
	case "healthy":
		return "\033[32m" // Green
	case "degraded":
		return "\033[33m" // Yellow
	case "unhealthy":
		return "\033[31m" // Red
	default:
		return "\033[0m" // Default
	}
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
