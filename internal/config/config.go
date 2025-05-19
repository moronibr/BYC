package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the configuration for the blockchain node
type Config struct {
	// Network settings
	Network struct {
		ListenAddress  string   `json:"listen_address"`
		MaxPeers       int      `json:"max_peers"`
		BootstrapPeers []string `json:"bootstrap_peers"`
	} `json:"network"`

	// Mining settings
	Mining struct {
		Enabled   bool   `json:"enabled"`
		Threads   int    `json:"threads"`
		CoinType  string `json:"coin_type"`
		BlockType string `json:"block_type"`
	} `json:"mining"`

	// Blockchain settings
	Blockchain struct {
		Difficulty    int `json:"difficulty"`
		BlockTime     int `json:"block_time"`
		MaxBlockSize  int `json:"max_block_size"`
		MaxTxPerBlock int `json:"max_tx_per_block"`
	} `json:"blockchain"`

	// Database settings
	Database struct {
		Type    string `json:"type"` // "leveldb" or "badger"
		Path    string `json:"path"`
		MaxSize int64  `json:"max_size"`
	} `json:"database"`

	// Security settings
	Security struct {
		RateLimit      int      `json:"rate_limit"`
		MaxConnections int      `json:"max_connections"`
		Whitelist      []string `json:"whitelist"`
	} `json:"security"`

	// Logging settings
	Logging struct {
		Level      string `json:"level"`
		File       string `json:"file"`
		MaxSize    int    `json:"max_size"`
		MaxBackups int    `json:"max_backups"`
		MaxAge     int    `json:"max_age"`
	} `json:"logging"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	config := &Config{}

	// Set default network settings
	config.Network.ListenAddress = "localhost:3000"
	config.Network.MaxPeers = 50
	config.Network.BootstrapPeers = []string{}

	// Set default mining settings
	config.Mining.Enabled = false
	config.Mining.Threads = 1
	config.Mining.CoinType = "leah"
	config.Mining.BlockType = "golden"

	// Set default blockchain settings
	config.Blockchain.Difficulty = 4
	config.Blockchain.BlockTime = 60
	config.Blockchain.MaxBlockSize = 1024 * 1024 // 1MB
	config.Blockchain.MaxTxPerBlock = 1000

	// Set default database settings
	config.Database.Type = "leveldb"
	config.Database.Path = "data/blockchain"
	config.Database.MaxSize = 1024 * 1024 * 1024 // 1GB

	// Set default security settings
	config.Security.RateLimit = 100
	config.Security.MaxConnections = 1000
	config.Security.Whitelist = []string{}

	// Set default logging settings
	config.Logging.Level = "info"
	config.Logging.File = "logs/byc.log"
	config.Logging.MaxSize = 100 // MB
	config.Logging.MaxBackups = 3
	config.Logging.MaxAge = 28 // days

	return config
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	// Read config file if it exists
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, fmt.Errorf("failed to decode config file: %v", err)
		}
	} else {
		// Create default config file
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create config file: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(config); err != nil {
			return nil, fmt.Errorf("failed to write default config: %v", err)
		}
	}

	return config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}
