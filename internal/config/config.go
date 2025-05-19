package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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

// Config represents the complete wallet configuration
type Config struct {
	Network     NetworkConfig     `json:"network"`
	Fee         FeeConfig         `json:"fee"`
	Security    SecurityConfig    `json:"security"`
	Environment EnvironmentConfig `json:"environment"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Network: NetworkConfig{
			Type:           Mainnet,
			RPCURL:         "http://localhost:8545",
			P2PPort:        30303,
			BootstrapNodes: []string{},
			BlockTime:      15 * time.Second,
			Difficulty:     2,
			MaxBlockSize:   1024 * 1024, // 1MB
			MaxConnections: 50,
			SyncTimeout:    5 * time.Minute,
			ReconnectDelay: 30 * time.Second,
		},
		Fee: FeeConfig{
			BaseFee:            0.001,
			SizeMultiplier:     0.0001,
			PriorityMultiplier: 0.01,
			MinFee:             0.0001,
			MaxFee:             1.0,
			FeeUpdateInterval:  1 * time.Hour,
		},
		Security: SecurityConfig{
			EncryptionAlgorithm: "aes-256-gcm",
			KeyDerivationCost:   32768,
			KeyRotationInterval: 30 * 24 * time.Hour, // 30 days
			MaxLoginAttempts:    5,
			SessionTimeout:      30 * time.Minute,
			PasswordMinLength:   12,
			RequireSpecialChars: true,
			RequireNumbers:      true,
			RequireUppercase:    true,
		},
		Environment: EnvironmentConfig{
			LogLevel:        "info",
			DataDir:         "data",
			BackupDir:       "backups",
			TempDir:         "temp",
			MaxLogSize:      100 * 1024 * 1024, // 100MB
			MaxLogFiles:     5,
			DebugMode:       false,
			MetricsPort:     9090,
			EnableProfiling: false,
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Create default config
	config := DefaultConfig()

	// Read config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse config file
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Marshal config
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write config file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate network config
	if c.Network.P2PPort <= 0 || c.Network.P2PPort > 65535 {
		return fmt.Errorf("invalid P2P port: %d", c.Network.P2PPort)
	}

	if c.Network.BlockTime <= 0 {
		return fmt.Errorf("invalid block time: %v", c.Network.BlockTime)
	}

	if c.Network.Difficulty <= 0 {
		return fmt.Errorf("invalid difficulty: %d", c.Network.Difficulty)
	}

	// Validate fee config
	if c.Fee.BaseFee < 0 {
		return fmt.Errorf("invalid base fee: %f", c.Fee.BaseFee)
	}

	if c.Fee.MinFee < 0 {
		return fmt.Errorf("invalid minimum fee: %f", c.Fee.MinFee)
	}

	if c.Fee.MaxFee < c.Fee.MinFee {
		return fmt.Errorf("maximum fee must be greater than minimum fee")
	}

	// Validate security config
	if c.Security.KeyDerivationCost < 32768 {
		return fmt.Errorf("key derivation cost must be at least 32768")
	}

	if c.Security.MaxLoginAttempts <= 0 {
		return fmt.Errorf("invalid maximum login attempts: %d", c.Security.MaxLoginAttempts)
	}

	if c.Security.PasswordMinLength < 8 {
		return fmt.Errorf("password minimum length must be at least 8")
	}

	// Validate environment config
	if c.Environment.MaxLogSize <= 0 {
		return fmt.Errorf("invalid maximum log size: %d", c.Environment.MaxLogSize)
	}

	if c.Environment.MaxLogFiles <= 0 {
		return fmt.Errorf("invalid maximum log files: %d", c.Environment.MaxLogFiles)
	}

	if c.Environment.MetricsPort <= 0 || c.Environment.MetricsPort > 65535 {
		return fmt.Errorf("invalid metrics port: %d", c.Environment.MetricsPort)
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
