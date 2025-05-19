package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the configuration for the blockchain node
type Config struct {
	// Environment settings
	Environment string `json:"environment" env:"BYC_ENV"` // "development", "staging", "production"

	// Network settings
	Network struct {
		ListenAddress  string   `json:"listen_address" env:"BYC_NETWORK_LISTEN_ADDRESS"`
		MaxPeers       int      `json:"max_peers" env:"BYC_NETWORK_MAX_PEERS"`
		BootstrapPeers []string `json:"bootstrap_peers" env:"BYC_NETWORK_BOOTSTRAP_PEERS"`
		PeerTimeout    int      `json:"peer_timeout" env:"BYC_NETWORK_PEER_TIMEOUT"` // in seconds
		MaxConnections int      `json:"max_connections" env:"BYC_NETWORK_MAX_CONNECTIONS"`
		EnableUPnP     bool     `json:"enable_upnp" env:"BYC_NETWORK_ENABLE_UPNP"`
	} `json:"network"`

	// Mining settings
	Mining struct {
		Enabled      bool   `json:"enabled" env:"BYC_MINING_ENABLED"`
		Threads      int    `json:"threads" env:"BYC_MINING_THREADS"`
		CoinType     string `json:"coin_type" env:"BYC_MINING_COIN_TYPE"`
		BlockType    string `json:"block_type" env:"BYC_MINING_BLOCK_TYPE"`
		MaxNonce     int64  `json:"max_nonce" env:"BYC_MINING_MAX_NONCE"`
		HashRateLog  bool   `json:"hash_rate_log" env:"BYC_MINING_HASH_RATE_LOG"`
		HashRateFile string `json:"hash_rate_file" env:"BYC_MINING_HASH_RATE_FILE"`
	} `json:"mining"`

	// Blockchain settings
	Blockchain struct {
		Difficulty    int   `json:"difficulty" env:"BYC_BLOCKCHAIN_DIFFICULTY"`
		BlockTime     int   `json:"block_time" env:"BYC_BLOCKCHAIN_BLOCK_TIME"`
		MaxBlockSize  int   `json:"max_block_size" env:"BYC_BLOCKCHAIN_MAX_BLOCK_SIZE"`
		MaxTxPerBlock int   `json:"max_tx_per_block" env:"BYC_BLOCKCHAIN_MAX_TX_PER_BLOCK"`
		GenesisTime   int64 `json:"genesis_time" env:"BYC_BLOCKCHAIN_GENESIS_TIME"`
		RewardAmount  int64 `json:"reward_amount" env:"BYC_BLOCKCHAIN_REWARD_AMOUNT"`
	} `json:"blockchain"`

	// Database settings
	Database struct {
		Type           string `json:"type" env:"BYC_DATABASE_TYPE"`
		Path           string `json:"path" env:"BYC_DATABASE_PATH"`
		MaxSize        int64  `json:"max_size" env:"BYC_DATABASE_MAX_SIZE"`
		Compression    bool   `json:"compression" env:"BYC_DATABASE_COMPRESSION"`
		CacheSize      int    `json:"cache_size" env:"BYC_DATABASE_CACHE_SIZE"`
		WriteBuffer    int    `json:"write_buffer" env:"BYC_DATABASE_WRITE_BUFFER"`
		MaxOpenFiles   int    `json:"max_open_files" env:"BYC_DATABASE_MAX_OPEN_FILES"`
		CompactOnStart bool   `json:"compact_on_start" env:"BYC_DATABASE_COMPACT_ON_START"`
	} `json:"database"`

	// Security settings
	Security struct {
		RateLimit      int      `json:"rate_limit" env:"BYC_SECURITY_RATE_LIMIT"`
		MaxConnections int      `json:"max_connections" env:"BYC_SECURITY_MAX_CONNECTIONS"`
		Whitelist      []string `json:"whitelist" env:"BYC_SECURITY_WHITELIST"`
		EnableTLS      bool     `json:"enable_tls" env:"BYC_SECURITY_ENABLE_TLS"`
		CertFile       string   `json:"cert_file" env:"BYC_SECURITY_CERT_FILE"`
		KeyFile        string   `json:"key_file" env:"BYC_SECURITY_KEY_FILE"`
		AllowedOrigins []string `json:"allowed_origins" env:"BYC_SECURITY_ALLOWED_ORIGINS"`
	} `json:"security"`

	// Logging settings
	Logging struct {
		Level      string `json:"level" env:"BYC_LOGGING_LEVEL"`
		File       string `json:"file" env:"BYC_LOGGING_FILE"`
		MaxSize    int    `json:"max_size" env:"BYC_LOGGING_MAX_SIZE"`
		MaxBackups int    `json:"max_backups" env:"BYC_LOGGING_MAX_BACKUPS"`
		MaxAge     int    `json:"max_age" env:"BYC_LOGGING_MAX_AGE"`
		Compress   bool   `json:"compress" env:"BYC_LOGGING_COMPRESS"`
		Console    bool   `json:"console" env:"BYC_LOGGING_CONSOLE"`
		Format     string `json:"format" env:"BYC_LOGGING_FORMAT"` // "json" or "text"
	} `json:"logging"`

	// API settings
	API struct {
		Enabled      bool   `json:"enabled" env:"BYC_API_ENABLED"`
		ListenAddr   string `json:"listen_addr" env:"BYC_API_LISTEN_ADDR"`
		ReadTimeout  int    `json:"read_timeout" env:"BYC_API_READ_TIMEOUT"`
		WriteTimeout int    `json:"write_timeout" env:"BYC_API_WRITE_TIMEOUT"`
		MaxBodySize  int64  `json:"max_body_size" env:"BYC_API_MAX_BODY_SIZE"`
		EnableCORS   bool   `json:"enable_cors" env:"BYC_API_ENABLE_CORS"`
	} `json:"api"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	config := &Config{}

	// Set default environment
	config.Environment = "development"

	// Set default network settings
	config.Network.ListenAddress = "localhost:3000"
	config.Network.MaxPeers = 50
	config.Network.BootstrapPeers = []string{}
	config.Network.PeerTimeout = 30
	config.Network.MaxConnections = 1000
	config.Network.EnableUPnP = false

	// Set default mining settings
	config.Mining.Enabled = false
	config.Mining.Threads = 1
	config.Mining.CoinType = "leah"
	config.Mining.BlockType = "golden"
	config.Mining.MaxNonce = 1000000
	config.Mining.HashRateLog = true
	config.Mining.HashRateFile = "logs/hashrate.log"

	// Set default blockchain settings
	config.Blockchain.Difficulty = 4
	config.Blockchain.BlockTime = 60
	config.Blockchain.MaxBlockSize = 1024 * 1024 // 1MB
	config.Blockchain.MaxTxPerBlock = 1000
	config.Blockchain.GenesisTime = time.Now().Unix()
	config.Blockchain.RewardAmount = 50

	// Set default database settings
	config.Database.Type = "leveldb"
	config.Database.Path = "data/blockchain"
	config.Database.MaxSize = 1024 * 1024 * 1024 // 1GB
	config.Database.Compression = true
	config.Database.CacheSize = 32 * 1024 * 1024  // 32MB
	config.Database.WriteBuffer = 4 * 1024 * 1024 // 4MB
	config.Database.MaxOpenFiles = 1000
	config.Database.CompactOnStart = true

	// Set default security settings
	config.Security.RateLimit = 100
	config.Security.MaxConnections = 1000
	config.Security.Whitelist = []string{}
	config.Security.EnableTLS = false
	config.Security.CertFile = "certs/server.crt"
	config.Security.KeyFile = "certs/server.key"
	config.Security.AllowedOrigins = []string{"*"}

	// Set default logging settings
	config.Logging.Level = "info"
	config.Logging.File = "logs/byc.log"
	config.Logging.MaxSize = 100 // MB
	config.Logging.MaxBackups = 3
	config.Logging.MaxAge = 28 // days
	config.Logging.Compress = true
	config.Logging.Console = true
	config.Logging.Format = "text"

	// Set default API settings
	config.API.Enabled = true
	config.API.ListenAddr = "localhost:8080"
	config.API.ReadTimeout = 30
	config.API.WriteTimeout = 30
	config.API.MaxBodySize = 1024 * 1024 // 1MB
	config.API.EnableCORS = true

	return config
}

// LoadConfig loads the configuration from a file and environment variables
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	// Read config file if it exists
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, fmt.Errorf("failed to decode config file: %v", err)
		}
	} else {
		// Create default config file
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create config file: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(config); err != nil {
			return nil, fmt.Errorf("failed to write default config: %v", err)
		}
	}

	// Override with environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load from environment: %v", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return config, nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate environment
	if config.Environment != "development" && config.Environment != "staging" && config.Environment != "production" {
		return fmt.Errorf("invalid environment: %s", config.Environment)
	}

	// Validate network settings
	if config.Network.MaxPeers < 0 {
		return fmt.Errorf("invalid max peers: %d", config.Network.MaxPeers)
	}
	if config.Network.PeerTimeout < 0 {
		return fmt.Errorf("invalid peer timeout: %d", config.Network.PeerTimeout)
	}

	// Validate mining settings
	if config.Mining.Threads < 1 {
		return fmt.Errorf("invalid mining threads: %d", config.Mining.Threads)
	}
	if config.Mining.MaxNonce < 0 {
		return fmt.Errorf("invalid max nonce: %d", config.Mining.MaxNonce)
	}

	// Validate blockchain settings
	if config.Blockchain.Difficulty < 1 {
		return fmt.Errorf("invalid difficulty: %d", config.Blockchain.Difficulty)
	}
	if config.Blockchain.BlockTime < 1 {
		return fmt.Errorf("invalid block time: %d", config.Blockchain.BlockTime)
	}
	if config.Blockchain.MaxBlockSize < 1024 {
		return fmt.Errorf("invalid max block size: %d", config.Blockchain.MaxBlockSize)
	}
	if config.Blockchain.MaxTxPerBlock < 1 {
		return fmt.Errorf("invalid max transactions per block: %d", config.Blockchain.MaxTxPerBlock)
	}

	// Validate database settings
	if config.Database.MaxSize < 1024*1024 {
		return fmt.Errorf("invalid database max size: %d", config.Database.MaxSize)
	}
	if config.Database.CacheSize < 1024*1024 {
		return fmt.Errorf("invalid database cache size: %d", config.Database.CacheSize)
	}
	if config.Database.WriteBuffer < 1024*1024 {
		return fmt.Errorf("invalid database write buffer: %d", config.Database.WriteBuffer)
	}
	if config.Database.MaxOpenFiles < 100 {
		return fmt.Errorf("invalid max open files: %d", config.Database.MaxOpenFiles)
	}

	// Validate security settings
	if config.Security.RateLimit < 0 {
		return fmt.Errorf("invalid rate limit: %d", config.Security.RateLimit)
	}
	if config.Security.MaxConnections < 0 {
		return fmt.Errorf("invalid max connections: %d", config.Security.MaxConnections)
	}

	// Validate logging settings
	if config.Logging.MaxSize < 1 {
		return fmt.Errorf("invalid log max size: %d", config.Logging.MaxSize)
	}
	if config.Logging.MaxBackups < 0 {
		return fmt.Errorf("invalid log max backups: %d", config.Logging.MaxBackups)
	}
	if config.Logging.MaxAge < 0 {
		return fmt.Errorf("invalid log max age: %d", config.Logging.MaxAge)
	}

	// Validate API settings
	if config.API.ReadTimeout < 0 {
		return fmt.Errorf("invalid API read timeout: %d", config.API.ReadTimeout)
	}
	if config.API.WriteTimeout < 0 {
		return fmt.Errorf("invalid API write timeout: %d", config.API.WriteTimeout)
	}
	if config.API.MaxBodySize < 0 {
		return fmt.Errorf("invalid API max body size: %d", config.API.MaxBodySize)
	}

	return nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}
