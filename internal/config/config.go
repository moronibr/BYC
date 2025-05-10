package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Network configuration
	Network    string `json:"network"`
	ListenAddr string `json:"listen_addr"`
	RPCPort    int    `json:"rpc_port"`
	MaxPeers   int    `json:"max_peers"`

	// Mining configuration
	MiningEnabled bool   `json:"mining_enabled"`
	MiningThreads int    `json:"mining_threads"`
	MiningCoin    string `json:"mining_coin"`

	// Database configuration
	DataDir   string        `json:"data_dir"`
	DBPath    string        `json:"db_path"`
	MaxDBSize int64         `json:"max_db_size"`
	DBTimeout time.Duration `json:"db_timeout"`

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
		Network:           "mainnet",
		ListenAddr:        "0.0.0.0",
		RPCPort:           8333,
		MaxPeers:          125,
		MiningEnabled:     false,
		MiningThreads:     1,
		MiningCoin:        "leah",
		DataDir:           "data",
		DBPath:            "byc.db",
		MaxDBSize:         1024 * 1024 * 1024, // 1GB
		DBTimeout:         30 * time.Second,
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

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	// Validate network
	if config.Network != "mainnet" && config.Network != "testnet" {
		return errors.New("invalid network")
	}

	// Validate RPC port
	if config.RPCPort <= 0 || config.RPCPort > 65535 {
		return errors.New("invalid RPC port")
	}

	// Validate max peers
	if config.MaxPeers <= 0 {
		return errors.New("invalid max peers")
	}

	// Validate mining threads
	if config.MiningThreads <= 0 {
		return errors.New("invalid mining threads")
	}

	// Validate database settings
	if config.MaxDBSize <= 0 {
		return errors.New("invalid max DB size")
	}
	if config.DBTimeout <= 0 {
		return errors.New("invalid DB timeout")
	}

	return nil
}
