package api

import (
	"github.com/byc/internal/blockchain"
)

// Config represents the API server configuration
type Config struct {
	// NodeAddress is the address to listen on for node connections
	NodeAddress string
	// BlockType is the type of blocks to handle
	BlockType blockchain.BlockType
	// BootstrapPeers is a list of peer addresses to connect to on startup
	BootstrapPeers []string
}

// NewConfig creates a new API server configuration
func NewConfig(nodeAddress string, blockType blockchain.BlockType, bootstrapPeers []string) *Config {
	return &Config{
		NodeAddress:    nodeAddress,
		BlockType:      blockType,
		BootstrapPeers: bootstrapPeers,
	}
}
