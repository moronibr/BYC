package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Check network settings
	if config.Network != "mainnet" {
		t.Errorf("Expected network to be mainnet, got %s", config.Network)
	}
	if config.ListenAddr != "0.0.0.0" {
		t.Errorf("Expected ListenAddr to be 0.0.0.0, got %s", config.ListenAddr)
	}
	if config.RPCPort != 8333 {
		t.Errorf("Expected RPCPort to be 8333, got %d", config.RPCPort)
	}

	// Check mining settings
	if config.MiningEnabled {
		t.Error("Expected MiningEnabled to be false")
	}
	if config.MiningThreads != 1 {
		t.Errorf("Expected MiningThreads to be 1, got %d", config.MiningThreads)
	}
	if config.MiningCoin != "leah" {
		t.Errorf("Expected MiningCoin to be leah, got %s", config.MiningCoin)
	}

	// Check database settings
	if config.DataDir != "data" {
		t.Errorf("Expected DataDir to be data, got %s", config.DataDir)
	}
	if config.DBPath != "byc.db" {
		t.Errorf("Expected DBPath to be byc.db, got %s", config.DBPath)
	}
	if config.MaxDBSize != 1024*1024*1024 {
		t.Errorf("Expected MaxDBSize to be 1GB, got %d", config.MaxDBSize)
	}
	if config.DBTimeout != 30*time.Second {
		t.Errorf("Expected DBTimeout to be 30s, got %v", config.DBTimeout)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "byc-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file path
	configPath := filepath.Join(tmpDir, "config.json")

	// Test loading non-existent config (should create default)
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify default values were created
	if config.Network != "mainnet" {
		t.Errorf("Expected network to be mainnet, got %s", config.Network)
	}

	// Modify config
	config.Network = "testnet"
	config.RPCPort = 8334

	// Save modified config
	if err := SaveConfig(config, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load modified config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load modified config: %v", err)
	}

	// Verify modifications were saved
	if loadedConfig.Network != "testnet" {
		t.Errorf("Expected network to be testnet, got %s", loadedConfig.Network)
	}
	if loadedConfig.RPCPort != 8334 {
		t.Errorf("Expected RPCPort to be 8334, got %d", loadedConfig.RPCPort)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "byc-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file path
	configPath := filepath.Join(tmpDir, "config.json")

	// Create and save config
	config := DefaultConfig()
	if err := SaveConfig(config, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestValidateConfig(t *testing.T) {
	config := DefaultConfig()

	// Test valid config
	if err := ValidateConfig(config); err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test invalid network
	config.Network = "invalid"
	if err := ValidateConfig(config); err == nil {
		t.Error("Expected invalid network to fail validation")
	}
	config.Network = "mainnet" // Reset

	// Test invalid port
	config.RPCPort = -1
	if err := ValidateConfig(config); err == nil {
		t.Error("Expected invalid port to fail validation")
	}
	config.RPCPort = 8333 // Reset

	// Test invalid max peers
	config.MaxPeers = 0
	if err := ValidateConfig(config); err == nil {
		t.Error("Expected invalid max peers to fail validation")
	}
}
