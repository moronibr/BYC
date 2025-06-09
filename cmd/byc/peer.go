package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"byc/internal/blockchain"
)

func handlePeerMenu(bc *blockchain.Blockchain) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n=== Peer Management ===")
		fmt.Println("1. List Peers")
		fmt.Println("2. Connect to Peer")
		fmt.Println("3. Disconnect from Peer")
		fmt.Println("4. Back to Main Menu")
		fmt.Print("Enter your choice: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		switch input {
		case "1":
			listPeers()
		case "2":
			connectToPeer(reader)
		case "3":
			disconnectFromPeer(reader)
		case "4":
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func listPeers() {
	node, err := getNode()
	if err != nil {
		fmt.Println("No peers are currently connected")
		return
	}

	peers := node.GetPeers()
	if len(peers) == 0 {
		fmt.Println("No peers are currently connected")
		return
	}

	fmt.Println("\nConnected Peers:")
	for _, peer := range peers {
		fmt.Printf("- %s (Last seen: %s)\n", peer.Address, peer.LastSeen.Format("2006-01-02 15:04:05"))
	}
}

func connectToPeer(reader *bufio.Reader) {
	node, err := getNode()
	if err != nil {
		fmt.Println("Please start a node first")
		return
	}

	fmt.Print("Enter peer address (e.g., localhost:3001): ")
	address, _ := reader.ReadString('\n')
	address = strings.TrimSpace(address)

	if err := node.ConnectToPeer(address); err != nil {
		fmt.Printf("Failed to connect to peer: %v\n", err)
		return
	}

	fmt.Printf("Successfully connected to peer at %s\n", address)
}

func disconnectFromPeer(reader *bufio.Reader) {
	node, err := getNode()
	if err != nil {
		fmt.Println("No peers are currently connected")
		return
	}

	fmt.Print("Enter peer address to disconnect: ")
	address, _ := reader.ReadString('\n')
	address = strings.TrimSpace(address)

	if err := node.DisconnectPeer(address); err != nil {
		fmt.Printf("Failed to disconnect from peer: %v\n", err)
		return
	}

	fmt.Printf("Successfully disconnected from peer at %s\n", address)
}
