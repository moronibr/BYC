package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"byc/internal/api"
	"byc/internal/blockchain"
	"byc/internal/config"
	"byc/internal/logger"
	"byc/internal/network"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Command line flags
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create blockchain instance
	bc := blockchain.NewBlockchain()

	// Create node with P2P address
	node, err := network.NewNode(&network.Config{
		Address:        cfg.P2P.Address,
		BlockType:      cfg.Blockchain.BlockType,
		BootstrapPeers: cfg.P2P.BootstrapPeers,
	})
	if err != nil {
		fmt.Printf("Failed to create node: %v\n", err)
		os.Exit(1)
	}

	// Create API server config
	apiConfig := api.NewConfig(cfg.API.Address, cfg.Blockchain.BlockType, cfg.P2P.BootstrapPeers)

	// Create API server
	server := api.NewServer(bc, apiConfig)

	// Start the API server
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start API server: %v\n", err)
		os.Exit(1)
	}

	// Start mining if configured
	if cfg.Mining.Enabled && cfg.Mining.AutoStart {
		coinType := blockchain.CoinType(cfg.Mining.CoinType)
		if err := node.StartMining(coinType); err != nil {
			fmt.Printf("Failed to start mining: %v\n", err)
		} else {
			fmt.Printf("Started mining %s coins\n", cfg.Mining.CoinType)
		}
	}

	fmt.Printf("Node running at:\n")
	fmt.Printf("  API: %s\n", cfg.API.Address)
	fmt.Printf("  P2P: %s\n", cfg.P2P.Address)
	fmt.Printf("Press Ctrl+C to exit.\n")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down node...")

	// Graceful shutdown
	if err := server.Stop(); err != nil {
		fmt.Printf("Error during server shutdown: %v\n", err)
	}
	if err := node.Stop(); err != nil {
		fmt.Printf("Error during node shutdown: %v\n", err)
	}
}
