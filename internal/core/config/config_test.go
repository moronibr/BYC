package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigManager_LoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	config := &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "test",
		},
		Network: NetworkConfig{
			Port: 8080,
		},
		Security: SecurityConfig{
			KeyPath: "/path/to/key",
		},
		Monitoring: MonitoringConfig{
			Interval: time.Second * 30,
		},
		Backup: BackupConfig{
			Path: "/path/to/backup",
		},
	}

	file, err := os.Create(configPath)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		t.Fatalf("Failed to encode config: %v", err)
	}

	// Test loading the config
	cm := NewConfigManager(configPath)
	if err := cm.LoadConfig(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	loadedConfig := cm.GetConfig()
	if loadedConfig.Environment != config.Environment {
		t.Errorf("Expected environment %s, got %s", config.Environment, loadedConfig.Environment)
	}
}

func TestConfigManager_UpdateConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	config := &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "test",
		},
		Network: NetworkConfig{
			Port: 8080,
		},
		Security: SecurityConfig{
			KeyPath: "/path/to/key",
		},
		Monitoring: MonitoringConfig{
			Interval: time.Second * 30,
		},
		Backup: BackupConfig{
			Path: "/path/to/backup",
		},
	}

	file, err := os.Create(configPath)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		t.Fatalf("Failed to encode config: %v", err)
	}

	// Test updating the config
	cm := NewConfigManager(configPath)
	if err := cm.LoadConfig(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	newConfig := &Config{
		Environment: "prod",
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "prod",
			Password: "prod",
			DBName:   "prod",
		},
		Network: NetworkConfig{
			Port: 8080,
		},
		Security: SecurityConfig{
			KeyPath: "/path/to/key",
		},
		Monitoring: MonitoringConfig{
			Interval: time.Second * 30,
		},
		Backup: BackupConfig{
			Path: "/path/to/backup",
		},
	}

	if err := cm.UpdateConfig(newConfig); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	loadedConfig := cm.GetConfig()
	if loadedConfig.Environment != newConfig.Environment {
		t.Errorf("Expected environment %s, got %s", newConfig.Environment, loadedConfig.Environment)
	}
}

func TestConfigManager_ValidateConfig(t *testing.T) {
	// Test validation with nil config
	cm := NewConfigManager("")
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for nil config")
	}

	// Test validation with missing environment
	cm.config = &Config{}
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for missing environment")
	}

	// Test validation with missing database host
	cm.config = &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Port: 5432,
		},
	}
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for missing database host")
	}

	// Test validation with invalid database port
	cm.config = &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host: "localhost",
			Port: -1,
		},
	}
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for invalid database port")
	}

	// Test validation with invalid network port
	cm.config = &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Network: NetworkConfig{
			Port: -1,
		},
	}
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for invalid network port")
	}

	// Test validation with missing key path
	cm.config = &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Network: NetworkConfig{
			Port: 8080,
		},
	}
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for missing key path")
	}

	// Test validation with invalid monitoring interval
	cm.config = &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Network: NetworkConfig{
			Port: 8080,
		},
		Security: SecurityConfig{
			KeyPath: "/path/to/key",
		},
		Monitoring: MonitoringConfig{
			Interval: -1,
		},
	}
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for invalid monitoring interval")
	}

	// Test validation with missing backup path
	cm.config = &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Network: NetworkConfig{
			Port: 8080,
		},
		Security: SecurityConfig{
			KeyPath: "/path/to/key",
		},
		Monitoring: MonitoringConfig{
			Interval: time.Second * 30,
		},
	}
	if err := cm.validateConfig(); err == nil {
		t.Error("Expected error for missing backup path")
	}
}

func TestConfigManager_LoadEnvironmentConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	config := &Config{
		Environment: "test",
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "test",
		},
		Network: NetworkConfig{
			Port: 8080,
		},
		Security: SecurityConfig{
			KeyPath: "/path/to/key",
		},
		Monitoring: MonitoringConfig{
			Interval: time.Second * 30,
		},
		Backup: BackupConfig{
			Path: "/path/to/backup",
		},
	}

	file, err := os.Create(configPath)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		t.Fatalf("Failed to encode config: %v", err)
	}

	// Test loading environment-specific config
	cm := NewConfigManager(configPath)
	if err := cm.LoadEnvironmentConfig("test"); err != nil {
		t.Fatalf("Failed to load environment config: %v", err)
	}

	loadedConfig := cm.GetConfig()
	if loadedConfig.Environment != config.Environment {
		t.Errorf("Expected environment %s, got %s", config.Environment, loadedConfig.Environment)
	}
}
