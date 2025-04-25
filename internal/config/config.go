package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Network settings
	Network        string   `json:"network"`
	ListenAddr     string   `json:"listen_addr"`
	RPCAddr        string   `json:"rpc_addr"`
	RPCPort        int      `json:"rpc_port"`
	MaxPeers       int      `json:"max_peers"`
	BootstrapNodes []string `json:"bootstrap_nodes"`

	// Mining settings
	MiningEnabled bool   `json:"mining_enabled"`
	MiningThreads int    `json:"mining_threads"`
	MiningCoin    string `json:"mining_coin"`
	MiningPool    string `json:"mining_pool"`

	// Database settings
	DataDir   string        `json:"data_dir"`
	DBPath    string        `json:"db_path"`
	MaxDBSize int64         `json:"max_db_size"`
	DBTimeout time.Duration `json:"db_timeout"`

	// Security settings
	RPCAuth     bool   `json:"rpc_auth"`
	RPCAuthFile string `json:"rpc_auth_file"`
	TLSEnabled  bool   `json:"tls_enabled"`
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`

	// Logging settings
	LogLevel      string `json:"log_level"`
	LogFile       string `json:"log_file"`
	LogMaxSize    int    `json:"log_max_size"`
	LogMaxBackups int    `json:"log_max_backups"`
	LogMaxAge     int    `json:"log_max_age"`

	// Performance settings
	MaxMempoolSize int `json:"max_mempool_size"`
	MaxBlockSize   int `json:"max_block_size"`
	MaxTxSize      int `json:"max_tx_size"`
	RateLimit      int `json:"rate_limit"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Network:    "mainnet",
		ListenAddr: "0.0.0.0",
		RPCAddr:    "127.0.0.1",
		RPCPort:    8333,
		MaxPeers:   50,
		BootstrapNodes: []string{
			"node1.byc.network:8333",
			"node2.byc.network:8333",
		},

		MiningEnabled: false,
		MiningThreads: 1,
		MiningCoin:    "leah",
		MiningPool:    "",

		DataDir:   "data",
		DBPath:    "byc.db",
		MaxDBSize: 1024 * 1024 * 1024, // 1GB
		DBTimeout: 30 * time.Second,

		RPCAuth:     false,
		RPCAuthFile: "rpc_auth.json",
		TLSEnabled:  false,
		TLSCertFile: "cert.pem",
		TLSKeyFile:  "key.pem",

		LogLevel:      "info",
		LogFile:       "byc.log",
		LogMaxSize:    100, // MB
		LogMaxBackups: 3,
		LogMaxAge:     28, // days

		MaxMempoolSize: 100 * 1024 * 1024, // 100MB
		MaxBlockSize:   2 * 1024 * 1024,   // 2MB
		MaxTxSize:      100 * 1024,        // 100KB
		RateLimit:      1000,              // requests per second
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
	// TODO: Implement validation logic
	return nil
}
