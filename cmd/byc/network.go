package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"byc/internal/blockchain"
)

func handleNetworkOperations(bc *blockchain.Blockchain) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n=== Network Operations ===")
		fmt.Println("1. Start Node")
		fmt.Println("2. Monitor Network")
		fmt.Println("3. Node Management")
		fmt.Println("4. Back to Main Menu")
		fmt.Print("Enter your choice: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		switch input {
		case "1":
			startNode(reader)
		case "2":
			monitorNetwork()
		case "3":
			handleNodeManagement(bc)
		case "4":
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func monitorNetwork() {
	node, err := getNode()
	if err != nil {
		fmt.Println("Please start a node first")
		return
	}

	fmt.Println("\nMonitoring network...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("-------------------")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			peers := node.GetPeers()
			fmt.Printf("\nActive Peers: %d\n", len(peers))
			for _, peer := range peers {
				fmt.Printf("- %s (Last seen: %s)\n",
					peer.Address,
					peer.LastSeen.Format("2006-01-02 15:04:05"))
			}
		}
	}
}
