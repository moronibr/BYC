package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moroni/BYC/internal/blockchain"
)

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
