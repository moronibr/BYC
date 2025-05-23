package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moroni/BYC/internal/api"
	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/network"
)

func main() {
	address := flag.String("address", "localhost:3000", "Node address")
	peer := flag.String("peer", "", "Peer address to connect to")
	flag.Parse()

	// Create blockchain instance
	bc := blockchain.NewBlockchain()

	// Create node
	node, err := network.NewNode(&network.Config{
		Address:        *address,
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	})
	if err != nil {
		fmt.Printf("Failed to create node: %v\n", err)
		os.Exit(1)
	}

	// Create API server config
	config := api.NewConfig(*address, blockchain.GoldenBlock, []string{})

	// Create API server with the node instance
	server := api.NewServer(bc, config)

	// Start the API server
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start API server: %v\n", err)
		os.Exit(1)
	}

	if *peer != "" {
		node.ConnectToPeer(*peer)
	}

	fmt.Printf("Node running at %s. Press Ctrl+C to exit.\n", *address)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Shutting down node...")
}
