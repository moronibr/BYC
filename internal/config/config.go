package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
)

// NetworkType represents different blockchain networks
type NetworkType string

const (
	Mainnet NetworkType = "mainnet"
	Testnet NetworkType = "testnet"
	Devnet  NetworkType = "devnet"
)

// NetworkConfig holds network-specific configuration
type NetworkConfig struct {
	Type           NetworkType   `json:"type"`
	RPCURL         string        `json:"rpc_url"`
	P2PPort        int           `json:"p2p_port"`
	BootstrapNodes []string      `json:"bootstrap_nodes"`
	BlockTime      time.Duration `json:"block_time"`
	Difficulty     int           `json:"difficulty"`
	MaxBlockSize   int           `json:"max_block_size"`
	MaxConnections int           `json:"max_connections"`
	SyncTimeout    time.Duration `json:"sync_timeout"`
	ReconnectDelay time.Duration `json:"reconnect_delay"`
}

// FeeConfig holds fee estimation parameters
type FeeConfig struct {
	BaseFee            float64       `json:"base_fee"`
	SizeMultiplier     float64       `json:"size_multiplier"`
	PriorityMultiplier float64       `json:"priority_multiplier"`
	MinFee             float64       `json:"min_fee"`
	MaxFee             float64       `json:"max_fee"`
	FeeUpdateInterval  time.Duration `json:"fee_update_interval"`
}

// SecurityConfig holds security-related parameters
type SecurityConfig struct {
	EncryptionAlgorithm string        `json:"encryption_algorithm"`
	KeyDerivationCost   int           `json:"key_derivation_cost"`
	KeyRotationInterval time.Duration `json:"key_rotation_interval"`
	MaxLoginAttempts    int           `json:"max_login_attempts"`
	SessionTimeout      time.Duration `json:"session_timeout"`
	PasswordMinLength   int           `json:"password_min_length"`
	RequireSpecialChars bool          `json:"require_special_chars"`
	RequireNumbers      bool          `json:"require_numbers"`
	RequireUppercase    bool          `json:"require_uppercase"`
}

// EnvironmentConfig holds environment-specific settings
type EnvironmentConfig struct {
	LogLevel        string `json:"log_level"`
	DataDir         string `json:"data_dir"`
	BackupDir       string `json:"backup_dir"`
	TempDir         string `json:"temp_dir"`
	MaxLogSize      int    `json:"max_log_size"`
	MaxLogFiles     int    `json:"max_log_files"`
	DebugMode       bool   `json:"debug_mode"`
	MetricsPort     int    `json:"metrics_port"`
	EnableProfiling bool   `json:"enable_profiling"`
}

// Config represents the complete configuration
type Config struct {
	API struct {
		Address string `json:"address"`
		CORS    struct {
			AllowedOrigins []string `json:"allowed_origins"`
		} `json:"cors"`
		RateLimit struct {
			RequestsPerSecond int `json:"requests_per_second"`
			Burst             int `json:"burst"`
		} `json:"rate_limit"`
		TLS struct {
			Enabled  bool   `json:"enabled"`
			CertFile string `json:"cert_file"`
			KeyFile  string `json:"key_file"`
		} `json:"tls"`
	} `json:"api"`

	P2P struct {
		Address        string        `json:"address"`
		BootstrapPeers []string      `json:"bootstrap_peers"`
		MaxPeers       int           `json:"max_peers"`
		PingInterval   time.Duration `json:"ping_interval"`
		PingTimeout    time.Duration `json:"ping_timeout"`
	} `json:"p2p"`

	Logging struct {
		Level  string `json:"level"`
		Format string `json:"format"`
		Output string `json:"output"`
	} `json:"logging"`

	Blockchain struct {
		BlockType    blockchain.BlockType `json:"block_type"`
		Difficulty   int                  `json:"difficulty"`
		MaxBlockSize int64                `json:"max_block_size"`
		MiningReward float64              `json:"mining_reward"`
	} `json:"blockchain"`

	Mining struct {
		Enabled               bool   `json:"enabled"`
		CoinType              string `json:"coin_type"`
		AutoStart             bool   `json:"auto_start"`
		MaxThreads            int    `json:"max_threads"`
		TargetBlocksPerMinute int    `json:"target_blocks_per_minute"`
	} `json:"mining"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		API: struct {
			Address string `json:"address"`
			CORS    struct {
				AllowedOrigins []string `json:"allowed_origins"`
			} `json:"cors"`
			RateLimit struct {
				RequestsPerSecond int `json:"requests_per_second"`
				Burst             int `json:"burst"`
			} `json:"rate_limit"`
			TLS struct {
				Enabled  bool   `json:"enabled"`
				CertFile string `json:"cert_file"`
				KeyFile  string `json:"key_file"`
			} `json:"tls"`
		}{
			Address: "localhost:3000",
			CORS: struct {
				AllowedOrigins []string `json:"allowed_origins"`
			}{
				AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
			},
			RateLimit: struct {
				RequestsPerSecond int `json:"requests_per_second"`
				Burst             int `json:"burst"`
			}{
				RequestsPerSecond: 100,
				Burst:             1000,
			},
			TLS: struct {
				Enabled  bool   `json:"enabled"`
				CertFile string `json:"cert_file"`
				KeyFile  string `json:"key_file"`
			}{
				Enabled:  true,
				CertFile: "cert.pem",
				KeyFile:  "key.pem",
			},
		},
		P2P: struct {
			Address        string        `json:"address"`
			BootstrapPeers []string      `json:"bootstrap_peers"`
			MaxPeers       int           `json:"max_peers"`
			PingInterval   time.Duration `json:"ping_interval"`
			PingTimeout    time.Duration `json:"ping_timeout"`
		}{
			Address:        "localhost:3001",
			BootstrapPeers: []string{},
			MaxPeers:       100,
			PingInterval:   30 * time.Second,
			PingTimeout:    10 * time.Second,
		},
		Logging: struct {
			Level  string `json:"level"`
			Format string `json:"format"`
			Output string `json:"output"`
		}{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Blockchain: struct {
			BlockType    blockchain.BlockType `json:"block_type"`
			Difficulty   int                  `json:"difficulty"`
			MaxBlockSize int64                `json:"max_block_size"`
			MiningReward float64              `json:"mining_reward"`
		}{
			BlockType:    blockchain.GoldenBlock,
			Difficulty:   4,
			MaxBlockSize: 1048576, // 1MB
			MiningReward: 50,
		},
		Mining: struct {
			Enabled               bool   `json:"enabled"`
			CoinType              string `json:"coin_type"`
			AutoStart             bool   `json:"auto_start"`
			MaxThreads            int    `json:"max_threads"`
			TargetBlocksPerMinute int    `json:"target_blocks_per_minute"`
		}{
			Enabled:               true,
			CoinType:              "BTC",
			AutoStart:             true,
			MaxThreads:            4,
			TargetBlocksPerMinute: 6,
		},
	}
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Marshal the config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write the config file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate API config
	if c.API.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("invalid requests per second: %d", c.API.RateLimit.RequestsPerSecond)
	}

	if c.API.RateLimit.Burst <= 0 {
		return fmt.Errorf("invalid burst: %d", c.API.RateLimit.Burst)
	}

	// Validate P2P config
	if c.P2P.MaxPeers <= 0 || c.P2P.MaxPeers > 1000 {
		return fmt.Errorf("invalid max peers: %d", c.P2P.MaxPeers)
	}

	if c.P2P.PingInterval <= 0 {
		return fmt.Errorf("invalid ping interval: %v", c.P2P.PingInterval)
	}

	if c.P2P.PingTimeout <= 0 {
		return fmt.Errorf("invalid ping timeout: %v", c.P2P.PingTimeout)
	}

	// Validate Blockchain config
	if c.Blockchain.Difficulty <= 0 {
		return fmt.Errorf("invalid difficulty: %d", c.Blockchain.Difficulty)
	}

	if c.Blockchain.MaxBlockSize <= 0 {
		return fmt.Errorf("invalid max block size: %d", c.Blockchain.MaxBlockSize)
	}

	if c.Blockchain.MiningReward <= 0 {
		return fmt.Errorf("invalid mining reward: %f", c.Blockchain.MiningReward)
	}

	// Validate Mining config
	if c.Mining.Enabled {
		if c.Mining.MaxThreads <= 0 {
			return fmt.Errorf("invalid max threads: %d", c.Mining.MaxThreads)
		}
		if c.Mining.TargetBlocksPerMinute <= 0 {
			return fmt.Errorf("invalid target blocks per minute: %d", c.Mining.TargetBlocksPerMinute)
		}
	}

	return nil
}

// GetNetworkType returns the network type
func (c *Config) GetNetworkType() NetworkType {
	return c.Network.Type
}

// IsMainnet returns true if the network is mainnet
func (c *Config) IsMainnet() bool {
	return c.Network.Type == Mainnet
}

// IsTestnet returns true if the network is testnet
func (c *Config) IsTestnet() bool {
	return c.Network.Type == Testnet
}

// IsDevnet returns true if the network is devnet
func (c *Config) IsDevnet() bool {
	return c.Network.Type == Devnet
}

// EstimateFee estimates the transaction fee
func (c *Config) EstimateFee(txSize int, priority float64) float64 {
	fee := c.Fee.BaseFee
	fee += float64(txSize) * c.Fee.SizeMultiplier
	fee += priority * c.Fee.PriorityMultiplier

	// Ensure fee is within bounds
	if fee < c.Fee.MinFee {
		fee = c.Fee.MinFee
	}
	if fee > c.Fee.MaxFee {
		fee = c.Fee.MaxFee
	}

	return fee
}

// GetDataDir returns the data directory path
func (c *Config) GetDataDir() string {
	return c.Environment.DataDir
}

// GetBackupDir returns the backup directory path
func (c *Config) GetBackupDir() string {
	return c.Environment.BackupDir
}

// GetTempDir returns the temporary directory path
func (c *Config) GetTempDir() string {
	return c.Environment.TempDir
}

// IsDebugMode returns true if debug mode is enabled
func (c *Config) IsDebugMode() bool {
	return c.Environment.DebugMode
}

// IsProfilingEnabled returns true if profiling is enabled
func (c *Config) IsProfilingEnabled() bool {
	return c.Environment.EnableProfiling
}
