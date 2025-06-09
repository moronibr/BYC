package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"byc/internal/blockchain"
	"byc/internal/logger"
	"byc/internal/mining"
	"byc/internal/wallet"

	"golang.org/x/term"
)

func displayMenu() {
	fmt.Println("\n=== BYC CLI Menu ===")
	fmt.Println("1. Network Operations")
	fmt.Println("2. Wallet Operations")
	fmt.Println("3. Dashboard")
	fmt.Println("4. Mining")
	fmt.Println("5. View Genesis Block")
	fmt.Println("6. Peer Management")
	fmt.Println("7. Backup & Restore")
	fmt.Println("8. System Maintenance")
	fmt.Println("9. Advanced Wallet")
	fmt.Println("10. Version Management")
	fmt.Println("11. Exit")
	fmt.Print("\nEnter your choice (1-11): ")
}

func main() {
	// Initialize logger first
	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	bc := blockchain.NewBlockchain()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n=== BYC CLI Menu ===")
		fmt.Println("1. Network Operations")
		fmt.Println("2. Wallet Operations")
		fmt.Println("3. Dashboard")
		fmt.Println("4. Mining")
		fmt.Println("5. View Genesis Block")
		fmt.Println("6. Peer Management")
		fmt.Println("7. Backup & Restore")
		fmt.Println("8. System Maintenance")
		fmt.Println("9. Advanced Wallet")
		fmt.Println("10. Version Management")
		fmt.Println("11. Exit")

		fmt.Print("\nEnter your choice (1-11): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			handleNetworkMenu(bc)
		case "2":
			handleWalletMenu(bc)
		case "3":
			handleDashboardMenu(bc)
		case "4":
			handleMiningMenu()
		case "5":
			bc.DisplayGenesisBlock()
		case "6":
			handlePeerMenu(bc)
		case "7":
			handleBackupMenu(bc)
		case "8":
			handleMaintenanceMenu(bc)
		case "9":
			handleAdvancedWalletMenu(bc)
		case "10":
			handleVersionMenu(bc)
		case "11":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func startMining(bc *blockchain.Blockchain, blockType, coinType, address *string) {
	// Validate block type
	var bt blockchain.BlockType
	switch *blockType {
	case "golden":
		bt = blockchain.GoldenBlock
	case "silver":
		bt = blockchain.SilverBlock
	default:
		fmt.Printf("Invalid block type: %s\n", *blockType)
		return
	}

	// Validate coin type
	var ct blockchain.CoinType
	switch *coinType {
	case "leah":
		ct = blockchain.Leah
	case "shiblum":
		ct = blockchain.Shiblum
	case "shiblon":
		ct = blockchain.Shiblon
	default:
		fmt.Printf("Invalid coin type: %s\n", *coinType)
		return
	}

	// Create miner
	miner, err := mining.NewMiner(bc, bt, ct, *address)
	if err != nil {
		fmt.Printf("Error creating miner: %v\n", err)
		return
	}

	// Create context for mining
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start mining
	go miner.Start(ctx)

	fmt.Println("Mining started. Press Ctrl+C to stop.")
	<-sigChan
	fmt.Println("\nShutting down mining...")
	cancel()
	miner.Stop()
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

func handleNetworkMenu(bc *blockchain.Blockchain) {
	fmt.Println("\n=== Network Operations ===")
	fmt.Println("1. Start Node")
	fmt.Println("2. Monitor Network")
	fmt.Println("3. Node Management")
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

	switch choice {
	case 1:
		fmt.Print("Enter node address (default: localhost:3000): ")
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)
		if address == "" {
			address = "localhost:3000"
		} else if !strings.Contains(address, ":") {
			// If only port is provided, add localhost
			address = "localhost:" + address
		}

		fmt.Print("Enter peer address (optional, format: host:port): ")
		peer, _ := reader.ReadString('\n')
		peer = strings.TrimSpace(peer)
		if peer != "" && !strings.Contains(peer, ":") {
			fmt.Println("Invalid peer address format. Please use host:port format.")
			return
		}

		fmt.Printf("Starting node on %s...\n", address)
		runNode(bc, address, peer)

	case 2:
		fmt.Print("Enter monitoring interval in seconds (default: 5): ")
		interval, _ := reader.ReadString('\n')
		interval = strings.TrimSpace(interval)
		if interval == "" {
			interval = "5"
		}

		node, err := getNode()
		if err != nil {
			fmt.Printf("Error getting node: %v\n", err)
			return
		}

		fmt.Printf("Starting node with monitoring...\n")
		// Start monitoring in a goroutine
		go func() {
			for {
				peers := node.GetPeers()
				fmt.Printf("\nConnected peers: %d\n", len(peers))
				for _, peer := range peers {
					fmt.Printf("- %s (Last seen: %s)\n", peer.Address, peer.LastSeen.Format(time.RFC3339))
				}
				time.Sleep(time.Duration(5) * time.Second)
			}
		}()

	case 3:
		// Node Management
		handleNodeManagement(bc)

	case 4:
		return
	default:
		fmt.Println("Invalid choice")
	}
}

func handleWalletMenu(bc *blockchain.Blockchain) {
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

	switch choice {
	case 1:
		runWallet(bc, "create")
	case 2:
		runWallet(bc, "balance")
	case 3:
		runWallet(bc, "send")
	case 4:
		return
	default:
		fmt.Println("Invalid choice")
	}
}

func handleDashboardMenu(bc *blockchain.Blockchain) {
	fmt.Println("\n=== Dashboard ===")

	// Get mining wallet info
	walletsDir := "wallets"
	walletFile := filepath.Join(walletsDir, "mining_wallet.json")

	// Network Status
	fmt.Println("\nNetwork Status:")
	fmt.Printf("Blockchain Height: %d\n", bc.GetCurrentHeight())

	// Safely get latest block
	latestBlock := bc.GetLatestBlock()
	if latestBlock != nil {
		fmt.Printf("Latest Block: %x\n", latestBlock.Hash)
	} else {
		fmt.Println("Latest Block: None (Genesis block not created)")
	}

	// Mining Stats
	fmt.Println("\nMining Statistics:")
	if _, err := os.Stat(walletFile); err == nil {
		data, err := os.ReadFile(walletFile)
		if err == nil {
			var walletInfo struct {
				Address string
				Rewards map[string]float64
			}
			if err := json.Unmarshal(data, &walletInfo); err == nil {
				fmt.Printf("Mining Address: %s\n", walletInfo.Address)
				fmt.Println("\nMining Rewards:")
				for coinType, amount := range walletInfo.Rewards {
					fmt.Printf("%s: %.2f\n", coinType, amount)
				}
			}
		}
	} else {
		fmt.Println("No mining wallet found")
	}

	// Recent Blocks
	fmt.Println("\nRecent Blocks:")
	blocks := bc.Blocks
	if len(blocks) > 0 {
		// Show last 5 blocks
		start := len(blocks) - 5
		if start < 0 {
			start = 0
		}
		for i := len(blocks) - 1; i >= start; i-- {
			block := blocks[i]
			fmt.Printf("Block %d: %x\n", i, block.Hash)
			fmt.Printf("  Timestamp: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 15:04:05"))
			fmt.Printf("  Transactions: %d\n", len(block.Transactions))
			fmt.Printf("  Block Type: %s\n", block.BlockType)
			fmt.Println()
		}
	} else {
		fmt.Println("No blocks found")
	}

	fmt.Println("\nPress Enter to return to main menu...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func handleMiningMenu() {
	// Start mining
	fmt.Println("Starting mining...")
	fmt.Println("Please enter the following information:")

	// Get block type
	fmt.Println("\nSelect block type:")
	fmt.Println("1. Golden")
	fmt.Println("2. Silver")
	fmt.Print("Enter choice (1-2): ")
	var blockChoice int
	fmt.Scan(&blockChoice)

	var block string
	switch blockChoice {
	case 1:
		block = "golden"
	case 2:
		block = "silver"
	default:
		fmt.Println("Invalid choice")
		return
	}

	// Get coin type
	fmt.Println("\nSelect coin type:")
	fmt.Println("1. Leah")
	fmt.Println("2. Shiblum")
	fmt.Println("3. Shiblon")
	fmt.Print("Enter choice (1-3): ")
	var coinChoice int
	fmt.Scan(&coinChoice)

	var coin string
	switch coinChoice {
	case 1:
		coin = "leah"
	case 2:
		coin = "shiblum"
	case 3:
		coin = "shiblon"
	default:
		fmt.Println("Invalid choice")
		return
	}

	// Get node address
	fmt.Print("\nEnter node address (default: localhost:3001): ")
	var nodeAddress string
	fmt.Scan(&nodeAddress)
	if nodeAddress == "" {
		nodeAddress = "localhost:3001"
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Set up terminal for reading keypresses
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

	// Create blockchain instance
	bc := blockchain.NewBlockchain()

	// Create miner
	miner, err := mining.NewMiner(bc, blockType, coinType, nodeAddress)
	if err != nil {
		log.Fatalf("Failed to create miner: %v", err)
	}

	// Start mining
	miner.Start(ctx)
	fmt.Println("Mining in progress. Press Esc or 'q' to stop.")
	<-ctx.Done() // Wait until user cancels (Esc, 'q', or Ctrl+C)
	miner.Stop()
}

func runNode(bc *blockchain.Blockchain, address, peerAddress string) {
	// Get or create node instance
	node, err := getNode()
	if err != nil {
		log.Fatalf("Failed to get node: %v", err)
	}

	// Connect to peer if specified
	if peerAddress != "" {
		if err := node.ConnectToPeer(peerAddress); err != nil {
			log.Printf("Failed to connect to peer: %v", err)
		}
	}
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
		// Get the mining wallet
		walletsDir := "wallets"
		walletFile := filepath.Join(walletsDir, "mining_wallet.json")

		if _, err := os.Stat(walletFile); err != nil {
			fmt.Println("No wallet found. Please mine some coins first.")
			return
		}

		// Read wallet file
		data, err := os.ReadFile(walletFile)
		if err != nil {
			fmt.Printf("Error reading wallet file: %v\n", err)
			return
		}

		var walletInfo struct {
			Address string
			Rewards map[string]float64
		}
		if err := json.Unmarshal(data, &walletInfo); err != nil {
			fmt.Printf("Error parsing wallet file: %v\n", err)
			return
		}

		fmt.Println("\n=== Wallet Balance ===")
		fmt.Printf("Address: %s\n", walletInfo.Address)
		fmt.Println("\nRewards:")
		for coinType, amount := range walletInfo.Rewards {
			fmt.Printf("%s: %.2f\n", coinType, amount)
		}
		fmt.Println("=====================\n")

	case "send":
		// TODO: Implement sending coins
		fmt.Println("Sending coins not implemented yet")

	default:
		fmt.Printf("Unknown wallet action: %s\n", action)
		os.Exit(1)
	}
}

func handleBackupMenu(bc *blockchain.Blockchain) {
	fmt.Println("\n=== Backup & Restore ===")
	fmt.Println("1. Create Backup")
	fmt.Println("2. Restore from Backup")
	fmt.Println("3. List Backups")
	fmt.Println("4. Delete Backup")
	fmt.Println("5. Back to Main Menu")
	fmt.Print("\nEnter your choice (1-5): ")

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
		fmt.Print("Enter backup name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if err := bc.CreateBackup(name); err != nil {
			fmt.Printf("Error creating backup: %v\n", err)
		} else {
			fmt.Println("Backup created successfully")
		}
	case 2:
		fmt.Print("Enter backup name to restore: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if err := bc.RestoreBackup(name); err != nil {
			fmt.Printf("Error restoring backup: %v\n", err)
		} else {
			fmt.Println("Backup restored successfully")
		}
	case 3:
		backups := bc.ListBackups()
		fmt.Println("\nAvailable Backups:")
		for _, backup := range backups {
			fmt.Printf("- %s\n", backup)
		}
	case 4:
		fmt.Print("Enter backup name to delete: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if err := bc.DeleteBackup(name); err != nil {
			fmt.Printf("Error deleting backup: %v\n", err)
		} else {
			fmt.Println("Backup deleted successfully")
		}
	case 5:
		return
	default:
		fmt.Println("Invalid choice")
	}
}

func handleMaintenanceMenu(bc *blockchain.Blockchain) {
	fmt.Println("\n=== System Maintenance ===")
	fmt.Println("1. Check System Health")
	fmt.Println("2. Run Maintenance Tasks")
	fmt.Println("3. View Maintenance Log")
	fmt.Println("4. Configure Maintenance")
	fmt.Println("5. Back to Main Menu")
	fmt.Print("\nEnter your choice (1-5): ")

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
		health := bc.CheckSystemHealth()
		fmt.Println("\nSystem Health Status:")
		fmt.Printf("Status: %s\n", health.Status)
		fmt.Printf("Last Check: %s\n", health.LastCheck.Format(time.RFC3339))
		fmt.Println("\nComponents:")
		for name, comp := range health.Components {
			fmt.Printf("- %s: %s\n", name, comp.Status)
		}
	case 2:
		fmt.Println("Running maintenance tasks...")
		if err := bc.RunMaintenance(); err != nil {
			fmt.Printf("Error running maintenance: %v\n", err)
		} else {
			fmt.Println("Maintenance completed successfully")
		}
	case 3:
		logs := bc.GetMaintenanceLog()
		fmt.Println("\nMaintenance Log:")
		for _, log := range logs {
			fmt.Printf("[%s] %s\n", log.Timestamp.Format(time.RFC3339), log.Message)
		}
	case 4:
		fmt.Println("\nMaintenance Configuration:")
		fmt.Println("1. Set Schedule")
		fmt.Println("2. Configure Tasks")
		fmt.Println("3. Set Alerts")
		fmt.Print("\nEnter your choice (1-3): ")
		configChoice, _ := reader.ReadString('\n')
		configChoice = strings.TrimSpace(configChoice)
		switch configChoice {
		case "1":
			fmt.Print("Enter schedule (e.g., 'daily', 'weekly'): ")
			schedule, _ := reader.ReadString('\n')
			schedule = strings.TrimSpace(schedule)
			bc.SetMaintenanceSchedule(schedule)
		case "2":
			fmt.Println("Available Tasks:")
			tasks := bc.GetMaintenanceTasks()
			for _, task := range tasks {
				fmt.Printf("- %s: %s\n", task.Name, task.Description)
			}
		case "3":
			fmt.Print("Enter alert email: ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)
			bc.SetMaintenanceAlert(email)
		}
	case 5:
		return
	default:
		fmt.Println("Invalid choice")
	}
}

func handleAdvancedWalletMenu(bc *blockchain.Blockchain) {
	fmt.Println("\n=== Advanced Wallet ===")
	fmt.Println("1. Create Ephraim Coin")
	fmt.Println("2. Create Manasseh Coin")
	fmt.Println("3. Create Joseph Coin")
	fmt.Println("4. View Special Coins")
	fmt.Println("5. Back to Main Menu")
	fmt.Print("\nEnter your choice (1-5): ")

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
		if err := bc.CreateEphraimCoin(); err != nil {
			fmt.Printf("Error creating Ephraim coin: %v\n", err)
		} else {
			fmt.Println("Ephraim coin created successfully")
		}
	case 2:
		if err := bc.CreateManassehCoin(); err != nil {
			fmt.Printf("Error creating Manasseh coin: %v\n", err)
		} else {
			fmt.Println("Manasseh coin created successfully")
		}
	case 3:
		if err := bc.CreateJosephCoin(); err != nil {
			fmt.Printf("Error creating Joseph coin: %v\n", err)
		} else {
			fmt.Println("Joseph coin created successfully")
		}
	case 4:
		coins := bc.GetSpecialCoins()
		fmt.Println("\nSpecial Coins:")
		for _, coin := range coins {
			fmt.Printf("- %s: %d\n", coin.Type, coin.Amount)
		}
	case 5:
		return
	default:
		fmt.Println("Invalid choice")
	}
}

func handleVersionMenu(bc *blockchain.Blockchain) {
	fmt.Println("\n=== Version Management ===")
	fmt.Println("1. Check Current Version")
	fmt.Println("2. View Version History")
	fmt.Println("3. Upgrade Version")
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

	switch choice {
	case 1:
		version := bc.GetCurrentVersion()
		fmt.Printf("\nCurrent Version: %s\n", version)
	case 2:
		history := bc.GetVersionHistory()
		fmt.Println("\nVersion History:")
		for _, version := range history {
			fmt.Printf("- %s (%s)\n", version.Number, version.Date.Format("2006-01-02"))
		}
	case 3:
		fmt.Print("Enter version to upgrade to: ")
		version, _ := reader.ReadString('\n')
		version = strings.TrimSpace(version)
		if err := bc.UpgradeVersion(version); err != nil {
			fmt.Printf("Error upgrading version: %v\n", err)
		} else {
			fmt.Println("Version upgraded successfully")
		}
	case 4:
		return
	default:
		fmt.Println("Invalid choice")
	}
}
