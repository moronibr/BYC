package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// Network configuration
	ListenAddr     string   `json:"listen_addr"`
	BootstrapNodes []string `json:"bootstrap_nodes"`
	MaxPeers       int      `json:"max_peers"`

	// Database configuration
	DBPath string `json:"db_path"`

	// Mining configuration
	MinerAddress string `json:"miner_address"`
	MinerThreads int    `json:"miner_threads"`

	// Consensus configuration
	InitialDifficulty uint32 `json:"initial_difficulty"`
	BlockTime         int64  `json:"block_time"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	ClientCAs  string
	ServerName string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:        ":8333",
		BootstrapNodes:    []string{},
		MaxPeers:          125,
		DBPath:            "data/blockchain",
		MinerAddress:      "",
		MinerThreads:      4,
		InitialDifficulty: 1,
		BlockTime:         600, // 10 minutes
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Read config file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config file
			if err := SaveConfig(config, path); err != nil {
				return nil, err
			}
			return config, nil
		}
		return nil, err
	}

	// Parse config file
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "   ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// TODO: Add validation logic
	return nil
}
