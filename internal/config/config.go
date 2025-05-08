package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the server configuration
type Config struct {
	ListenAddr string
	TLSEnabled bool
	TLS        TLSConfig
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
		ListenAddr: "127.0.0.1:8080",
		TLSEnabled: true,
		TLS: TLSConfig{
			CertFile:   "certs/server.crt",
			KeyFile:    "certs/server.key",
			ClientCAs:  "certs/client-ca.crt",
			ServerName: "localhost",
		},
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
