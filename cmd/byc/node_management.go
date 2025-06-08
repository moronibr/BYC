package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
)

// handleNodeManagement handles node management operations
func handleNodeManagement(bc *blockchain.Blockchain) {
	for {
		fmt.Println("\n=== Node Management ===")
		fmt.Println("1. Stop Node")
		fmt.Println("2. Reset Node")
		fmt.Println("3. Backup Node Data")
		fmt.Println("4. View Node Status")
		fmt.Println("5. Back to Network Menu")
		fmt.Print("\nEnter your choice (1-5): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid choice")
			continue
		}

		switch choice {
		case 1:
			stopNodeOperation()
		case 2:
			resetNodeOperation()
		case 3:
			backupNodeOperation(bc)
		case 4:
			viewNodeStatus(bc)
		case 5:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

// stopNodeOperation handles stopping the node
func stopNodeOperation() {
	fmt.Print("Are you sure you want to stop the node? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		if err := stopNode(); err != nil {
			fmt.Printf("Error stopping node: %v\n", err)
		} else {
			fmt.Println("Node stopped successfully")
		}
	} else {
		fmt.Println("Node stop cancelled")
	}
}

// resetNodeOperation handles resetting the node
func resetNodeOperation() {
	fmt.Print("Are you sure you want to reset the node? This will disconnect all peers and clear the node state. (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		if err := stopNode(); err != nil {
			fmt.Printf("Error stopping node: %v\n", err)
			return
		}
		// Clear the singleton instance
		nodeMutex.Lock()
		nodeInstance = nil
		nodeMutex.Unlock()
		fmt.Println("Node reset successfully")
	} else {
		fmt.Println("Node reset cancelled")
	}
}

// backupNodeOperation handles backing up node data
func backupNodeOperation(bc *blockchain.Blockchain) {
	fmt.Print("Enter backup directory path (default: ./backups): ")
	reader := bufio.NewReader(os.Stdin)
	backupDir, _ := reader.ReadString('\n')
	backupDir = strings.TrimSpace(backupDir)
	if backupDir == "" {
		backupDir = "./backups"
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Printf("Error creating backup directory: %v\n", err)
		return
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("node_backup_%s.json", timestamp))

	node, err := getNode()
	if err != nil {
		fmt.Printf("Error getting node: %v\n", err)
		return
	}

	// Get node data
	nodeData := struct {
		Address     string
		Peers       []string
		BlockHeight int64
		LastSeen    time.Time
		Version     string
		Status      string
	}{
		Address:     node.GetAddress(),
		Peers:       node.GetPeerAddresses(),
		BlockHeight: bc.GetCurrentHeight(),
		LastSeen:    time.Now(),
		Version:     "1.0.0", // TODO: Get actual version from node
		Status:      "active",
	}

	// Save to file
	data, err := json.MarshalIndent(nodeData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling node data: %v\n", err)
		return
	}

	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		fmt.Printf("Error saving backup: %v\n", err)
		return
	}

	fmt.Printf("Node data backed up successfully to %s\n", backupFile)
}

// viewNodeStatus displays current node status
func viewNodeStatus(bc *blockchain.Blockchain) {
	node, err := getNode()
	if err != nil {
		fmt.Printf("Error getting node: %v\n", err)
		return
	}

	fmt.Println("\n=== Node Status ===")
	fmt.Printf("Address: %s\n", node.GetAddress())
	fmt.Printf("Connected Peers: %d\n", len(node.GetPeers()))
	fmt.Printf("Blockchain Height: %d\n", bc.GetCurrentHeight())

	// Display connected peers
	peers := node.GetPeers()
	if len(peers) > 0 {
		fmt.Println("\nConnected Peers:")
		for _, peer := range peers {
			fmt.Printf("- %s (Last seen: %s)\n",
				peer.Address,
				peer.LastSeen.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println("\nPress Enter to continue...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
