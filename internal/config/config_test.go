package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test network config
	assert.Equal(t, Mainnet, config.Network.Type)
	assert.Equal(t, "http://localhost:8545", config.Network.RPCURL)
	assert.Equal(t, 30303, config.Network.P2PPort)
	assert.Empty(t, config.Network.BootstrapNodes)
	assert.Equal(t, 15*time.Second, config.Network.BlockTime)
	assert.Equal(t, 2, config.Network.Difficulty)
	assert.Equal(t, 1024*1024, config.Network.MaxBlockSize)
	assert.Equal(t, 50, config.Network.MaxConnections)
	assert.Equal(t, 5*time.Minute, config.Network.SyncTimeout)
	assert.Equal(t, 30*time.Second, config.Network.ReconnectDelay)

	// Test fee config
	assert.Equal(t, 0.001, config.Fee.BaseFee)
	assert.Equal(t, 0.0001, config.Fee.SizeMultiplier)
	assert.Equal(t, 0.01, config.Fee.PriorityMultiplier)
	assert.Equal(t, 0.0001, config.Fee.MinFee)
	assert.Equal(t, 1.0, config.Fee.MaxFee)
	assert.Equal(t, 1*time.Hour, config.Fee.FeeUpdateInterval)

	// Test security config
	assert.Equal(t, "aes-256-gcm", config.Security.EncryptionAlgorithm)
	assert.Equal(t, 32768, config.Security.KeyDerivationCost)
	assert.Equal(t, 30*24*time.Hour, config.Security.KeyRotationInterval)
	assert.Equal(t, 5, config.Security.MaxLoginAttempts)
	assert.Equal(t, 30*time.Minute, config.Security.SessionTimeout)
	assert.Equal(t, 12, config.Security.PasswordMinLength)
	assert.True(t, config.Security.RequireSpecialChars)
	assert.True(t, config.Security.RequireNumbers)
	assert.True(t, config.Security.RequireUppercase)

	// Test environment config
	assert.Equal(t, "info", config.Environment.LogLevel)
	assert.Equal(t, "data", config.Environment.DataDir)
	assert.Equal(t, "backups", config.Environment.BackupDir)
	assert.Equal(t, "temp", config.Environment.TempDir)
	assert.Equal(t, 100*1024*1024, config.Environment.MaxLogSize)
	assert.Equal(t, 5, config.Environment.MaxLogFiles)
	assert.False(t, config.Environment.DebugMode)
	assert.Equal(t, 9090, config.Environment.MetricsPort)
	assert.False(t, config.Environment.EnableProfiling)
}

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create test config
	config := DefaultConfig()
	config.Network.Type = Testnet
	config.Network.RPCURL = "http://testnet.example.com:8545"
	config.Fee.BaseFee = 0.002
	config.Security.PasswordMinLength = 16

	// Save config
	err := config.SaveConfig(configPath)
	require.NoError(t, err)

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	require.NoError(t, err)

	// Verify loaded config
	assert.Equal(t, Testnet, loadedConfig.Network.Type)
	assert.Equal(t, "http://testnet.example.com:8545", loadedConfig.Network.RPCURL)
	assert.Equal(t, 0.002, loadedConfig.Fee.BaseFee)
	assert.Equal(t, 16, loadedConfig.Security.PasswordMinLength)
}

func TestConfigValidation(t *testing.T) {
	config := DefaultConfig()

	// Test valid config
	err := config.Validate()
	require.NoError(t, err)

	// Test invalid P2P port
	config.Network.P2PPort = 0
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid P2P port")

	// Test invalid block time
	config = DefaultConfig()
	config.Network.BlockTime = 0
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid block time")

	// Test invalid difficulty
	config = DefaultConfig()
	config.Network.Difficulty = 0
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid difficulty")

	// Test invalid fee config
	config = DefaultConfig()
	config.Fee.BaseFee = -1
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base fee")

	// Test invalid security config
	config = DefaultConfig()
	config.Security.KeyDerivationCost = 1000
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key derivation cost must be at least 32768")

	// Test invalid environment config
	config = DefaultConfig()
	config.Environment.MaxLogSize = 0
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid maximum log size")
}

func TestNetworkTypeHelpers(t *testing.T) {
	config := DefaultConfig()

	// Test mainnet
	config.Network.Type = Mainnet
	assert.True(t, config.IsMainnet())
	assert.False(t, config.IsTestnet())
	assert.False(t, config.IsDevnet())

	// Test testnet
	config.Network.Type = Testnet
	assert.False(t, config.IsMainnet())
	assert.True(t, config.IsTestnet())
	assert.False(t, config.IsDevnet())

	// Test devnet
	config.Network.Type = Devnet
	assert.False(t, config.IsMainnet())
	assert.False(t, config.IsTestnet())
	assert.True(t, config.IsDevnet())
}

func TestFeeEstimation(t *testing.T) {
	config := DefaultConfig()

	// Test basic fee estimation
	fee := config.EstimateFee(1000, 1.0)
	assert.Greater(t, fee, config.Fee.MinFee)
	assert.Less(t, fee, config.Fee.MaxFee)

	// Test minimum fee
	fee = config.EstimateFee(1, 0.0)
	assert.Equal(t, config.Fee.MinFee, fee)

	// Test maximum fee
	config.Fee.MaxFee = 0.001
	fee = config.EstimateFee(1000000, 100.0)
	assert.Equal(t, config.Fee.MaxFee, fee)
}

func TestEnvironmentHelpers(t *testing.T) {
	config := DefaultConfig()

	// Test directory paths
	assert.Equal(t, "data", config.GetDataDir())
	assert.Equal(t, "backups", config.GetBackupDir())
	assert.Equal(t, "temp", config.GetTempDir())

	// Test debug mode
	assert.False(t, config.IsDebugMode())
	config.Environment.DebugMode = true
	assert.True(t, config.IsDebugMode())

	// Test profiling
	assert.False(t, config.IsProfilingEnabled())
	config.Environment.EnableProfiling = true
	assert.True(t, config.IsProfilingEnabled())
}

func TestConfigPersistence(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create and save config
	config := DefaultConfig()
	err := config.SaveConfig(configPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	require.NoError(t, err)

	// Verify loaded config matches original
	assert.Equal(t, config.Network.Type, loadedConfig.Network.Type)
	assert.Equal(t, config.Network.RPCURL, loadedConfig.Network.RPCURL)
	assert.Equal(t, config.Fee.BaseFee, loadedConfig.Fee.BaseFee)
	assert.Equal(t, config.Security.PasswordMinLength, loadedConfig.Security.PasswordMinLength)
}
