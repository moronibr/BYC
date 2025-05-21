package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/mining"
	"github.com/moroni/BYC/internal/network"
	"github.com/moroni/BYC/internal/wallet"
	"golang.org/x/term"
)

func main() {
	for {
		showMenu()
		choice := getUserChoice()

		switch choice {
		case 1:
			handleNetworkMenu()
		case 2:
			handleWalletMenu()
		case 3:
			handleDashboardMenu()
		case 4:
			handleMiningMenu()
		case 5:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func showMenu() {
	fmt.Println("\n=== BYC CLI Menu ===")
	fmt.Println("1. Network Operations")
	fmt.Println("2. Wallet Operations")
	fmt.Println("3. Dashboard")
	fmt.Println("4. Mining")
	fmt.Println("5. Exit")
	fmt.Print("\nEnter your choice (1-5): ")
}

func getUserChoice() int {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number between 1 and 5.")
			fmt.Print("Enter your choice (1-5): ")
			continue
		}
		if choice < 1 || choice > 5 {
			fmt.Println("Invalid choice. Please enter a number between 1 and 5.")
			fmt.Print("Enter your choice (1-5): ")
			continue
		}
		return choice
	}
}

func handleNetworkMenu() {
	fmt.Println("\n=== Network Operations ===")
	fmt.Println("1. Start Node")
	fmt.Println("2. Monitor Network")
	fmt.Println("3. Back to Main Menu")
	fmt.Print("\nEnter your choice (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid choice")
		return
	}

	cmd := flag.NewFlagSet("network", flag.ExitOnError)
	cmd.String("address", "localhost:3000", "Node address")
	cmd.String("peer", "", "Peer address to connect to")
	cmd.Bool("monitor", false, "Monitor network continuously")
	cmd.Int("interval", 5, "Monitoring interval in seconds")

	switch choice {
	case 1:
		fmt.Print("Enter node address (default: localhost:3000): ")
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)
		if address != "" {
			// Validate address format
			if !strings.Contains(address, ":") {
				address = "localhost:" + address
			}
			cmd.Set("address", address)
		}

		fmt.Print("Enter peer address (optional, format: host:port): ")
		peer, _ := reader.ReadString('\n')
		peer = strings.TrimSpace(peer)
		if peer != "" {
			// Validate peer address format
			if !strings.Contains(peer, ":") {
				fmt.Println("Invalid peer address format. Please use host:port format.")
				return
			}
			cmd.Set("peer", peer)
		}

		handleNetwork(cmd)

	case 2:
		fmt.Print("Enter monitoring interval in seconds (default: 5): ")
		interval, _ := reader.ReadString('\n')
		interval = strings.TrimSpace(interval)
		if interval != "" {
			intervalNum, err := strconv.Atoi(interval)
			if err != nil || intervalNum <= 0 {
				fmt.Println("Invalid interval. Using default value of 5 seconds.")
			} else {
				cmd.Set("interval", interval)
			}
		}
		cmd.Set("monitor", "true")
		handleNetwork(cmd)

	case 3:
		return
	}
}

func handleWalletMenu() {
	fmt.Println("\n=== Wallet Operations ===")
	fmt.Println("1. Create New Wallet")
	fmt.Println("2. Check Balance")
	fmt.Println("3. Send Coins")
	fmt.Println("4. Back to Main Menu")
	fmt.Print("\nEnter your choice (1-4): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid choice")
		return
	}

	cmd := flag.NewFlagSet("wallet", flag.ExitOnError)
	switch choice {
	case 1:
		cmd.String("action", "create", "")
		handleWallet(cmd)
	case 2:
		cmd.String("action", "balance", "")
		handleWallet(cmd)
	case 3:
		cmd.String("action", "send", "")
		handleWallet(cmd)
	case 4:
		return
	}
}

func handleDashboardMenu() {
	fmt.Println("\n=== Dashboard ===")
	fmt.Println("Loading dashboard...")
	cmd := flag.NewFlagSet("dashboard", flag.ExitOnError)
	handleDashboard(cmd)
}

func handleMiningMenu() {
	// Create a new blockchain instance
	bc := blockchain.NewBlockchain()

	fmt.Println("\n=== Mining Operations ===")
	fmt.Println("1. Start Mining")
	fmt.Println("2. Back to Main Menu")
	fmt.Print("\nEnter your choice (1-2): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid choice")
		return
	}

	switch choice {
	case 1:
		fmt.Println("\nSelect coin type:")
		fmt.Println("1. Leah")
		fmt.Println("2. Shiblum")
		fmt.Println("3. Shiblon")
		fmt.Print("\nEnter your choice (1-3): ")

		coinInput, _ := reader.ReadString('\n')
		coinInput = strings.TrimSpace(coinInput)
		coinChoice, err := strconv.Atoi(coinInput)
		if err != nil {
			fmt.Println("Invalid coin choice")
			return
		}

		var coin string
		switch coinChoice {
		case 1:
			coin = "leah"
		case 2:
			coin = "shiblum"
		case 3:
			coin = "shiblon"
		default:
			fmt.Println("Invalid coin choice")
			return
		}

		fmt.Println("\nSelect block type:")
		fmt.Println("1. Golden")
		fmt.Println("2. Silver")
		fmt.Print("\nEnter your choice (1-2): ")

		blockInput, _ := reader.ReadString('\n')
		blockInput = strings.TrimSpace(blockInput)
		blockChoice, err := strconv.Atoi(blockInput)
		if err != nil {
			fmt.Println("Invalid block choice")
			return
		}

		var block string
		switch blockChoice {
		case 1:
			block = "golden"
		case 2:
			block = "silver"
		default:
			fmt.Println("Invalid block choice")
			return
		}

		fmt.Print("\nEnter node address (default: localhost:3000): ")
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)
		if address != "" {
			// Validate address format
			if !strings.Contains(address, ":") {
				address = "localhost:" + address
			}
		} else {
			address = "localhost:3000"
		}

		fmt.Println("Press Esc or 'q' to stop mining and return to the main menu")

		// Set up context and signal handling
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Listen for Ctrl+C
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("\nStopping miner...")
			cancel()
		}()

		// Listen for Esc or 'q' key
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err == nil {
			defer term.Restore(int(os.Stdin.Fd()), oldState)
			go func() {
				buf := make([]byte, 1)
				for {
					os.Stdin.Read(buf)
					if buf[0] == 27 || buf[0] == 'q' || buf[0] == 'Q' { // 27 is Esc
						fmt.Println("\nStopping miner...")
						cancel()
						return
					}
				}
			}()
		}

		// Convert block and coin to correct types
		var blockType blockchain.BlockType
		var coinType blockchain.CoinType
		switch block {
		case "golden":
			blockType = blockchain.GoldenBlock
		case "silver":
			blockType = blockchain.SilverBlock
		default:
			fmt.Println("Invalid block type")
			return
		}
		switch coin {
		case "leah":
			coinType = blockchain.Leah
		case "shiblum":
			coinType = blockchain.Shiblum
		case "shiblon":
			coinType = blockchain.Shiblon
		default:
			fmt.Println("Invalid coin type")
			return
		}

		// Use the blockchain instance bc
		miner, err := mining.NewMiner(bc, blockType, coinType, "localhost:3000")
		if err != nil {
			log.Fatalf("Failed to create miner: %v", err)
		}
		miner.Start(ctx)
		fmt.Println("Mining in progress. Press Esc or 'q' to stop.")
		<-ctx.Done() // Wait until user cancels (Esc, 'q', or Ctrl+C)
		miner.Stop()
		return
	case 2:
		return
	default:
		fmt.Println("Invalid choice")
	}
}

func runNode(bc *blockchain.Blockchain, address, peerAddress string) {
	// Create node
	node, err := network.NewNode(&network.Config{
		Address:        address,
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	})
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	// Connect to peer if specified
	if peerAddress != "" {
		node.ConnectToPeer(peerAddress)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func runMining(bc *blockchain.Blockchain, coinType, blockType string) {
	// Convert string to CoinType
	var coin blockchain.CoinType
	switch coinType {
	case "leah":
		coin = blockchain.Leah
	case "shiblum":
		coin = blockchain.Shiblum
	case "shiblon":
		coin = blockchain.Shiblon
	default:
		log.Fatalf("Invalid coin type: %s", coinType)
	}

	// Convert string to BlockType
	var block blockchain.BlockType
	switch blockType {
	case "golden":
		block = blockchain.GoldenBlock
	case "silver":
		block = blockchain.SilverBlock
	default:
		log.Fatalf("Invalid block type: %s", blockType)
	}

	// Create miner
	miner, err := mining.NewMiner(bc, block, coin, "localhost:3000")
	if err != nil {
		log.Fatalf("Failed to create miner: %v", err)
	}

	// Start mining
	ctx := context.Background()
	miner.Start(ctx)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop mining
	miner.Stop()
}

func runWallet(bc *blockchain.Blockchain, action string) {
	switch action {
	case "create":
		// Create new wallet
		w, err := wallet.NewWallet()
		if err != nil {
			log.Fatalf("Failed to create wallet: %v", err)
		}
		fmt.Printf("Created new wallet with address: %s\n", w.Address)

	case "balance":
		// TODO: Implement balance checking
		fmt.Println("Balance checking not implemented yet")

	case "send":
		// TODO: Implement sending coins
		fmt.Println("Sending coins not implemented yet")

	default:
		fmt.Printf("Unknown wallet action: %s\n", action)
		os.Exit(1)
	}
}
