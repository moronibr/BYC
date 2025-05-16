package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config represents the application configuration
type Config struct {
	Environment string           `json:"environment"`
	Database    DatabaseConfig   `json:"database"`
	Network     NetworkConfig    `json:"network"`
	Security    SecurityConfig   `json:"security"`
	Monitoring  MonitoringConfig `json:"monitoring"`
	Backup      BackupConfig     `json:"backup"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Port int `json:"port"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	KeyPath string `json:"keyPath"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Interval time.Duration `json:"interval"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	Path string `json:"path"`
}

// ConfigManager manages the application configuration
type ConfigManager struct {
	mu     sync.RWMutex
	config *Config
	path   string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(path string) *ConfigManager {
	return &ConfigManager{
		path: path,
	}
}

// LoadConfig loads the configuration from a file
func (cm *ConfigManager) LoadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	file, err := os.Open(cm.path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	cm.config = &Config{}
	if err := json.NewDecoder(file).Decode(cm.config); err != nil {
		return fmt.Errorf("failed to decode config: %v", err)
	}

	return cm.validateConfig()
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// UpdateConfig updates the configuration dynamically
func (cm *ConfigManager) UpdateConfig(newConfig *Config) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := cm.validateConfig(); err != nil {
		return err
	}

	cm.config = newConfig
	return cm.saveConfig()
}

// saveConfig saves the configuration to a file
func (cm *ConfigManager) saveConfig() error {
	file, err := os.Create(cm.path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(cm.config)
}

// validateConfig validates the configuration
func (cm *ConfigManager) validateConfig() error {
	if cm.config == nil {
		return errors.New("config is nil")
	}

	if cm.config.Environment == "" {
		return errors.New("environment is required")
	}

	if cm.config.Database.Host == "" {
		return errors.New("database host is required")
	}

	if cm.config.Database.Port <= 0 {
		return errors.New("database port must be positive")
	}

	if cm.config.Network.Port <= 0 {
		return errors.New("network port must be positive")
	}

	if cm.config.Security.KeyPath == "" {
		return errors.New("key path is required")
	}

	if cm.config.Monitoring.Interval <= 0 {
		return errors.New("monitoring interval must be positive")
	}

	if cm.config.Backup.Path == "" {
		return errors.New("backup path is required")
	}

	return nil
}

// LoadEnvironmentConfig loads environment-specific configuration
func (cm *ConfigManager) LoadEnvironmentConfig(env string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	envPath := filepath.Join(filepath.Dir(cm.path), fmt.Sprintf("config.%s.json", env))
	file, err := os.Open(envPath)
	if err != nil {
		return fmt.Errorf("failed to open environment config file: %v", err)
	}
	defer file.Close()

	cm.config = &Config{}
	if err := json.NewDecoder(file).Decode(cm.config); err != nil {
		return fmt.Errorf("failed to decode environment config: %v", err)
	}

	return cm.validateConfig()
}
