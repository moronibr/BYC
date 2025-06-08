package main

import (
	"fmt"
	"sync"

	"github.com/moroni/BYC/internal/blockchain"
	"github.com/moroni/BYC/internal/network"
)

var (
	nodeInstance *network.Node
	nodeMutex    sync.RWMutex
)

// getNode returns the singleton node instance, creating it if necessary
func getNode() (*network.Node, error) {
	nodeMutex.RLock()
	if nodeInstance != nil {
		nodeMutex.RUnlock()
		return nodeInstance, nil
	}
	nodeMutex.RUnlock()

	nodeMutex.Lock()
	defer nodeMutex.Unlock()

	// Double-check after acquiring write lock
	if nodeInstance != nil {
		return nodeInstance, nil
	}

	// Create new node instance
	node, err := network.NewNode(&network.Config{
		Address:        "localhost:3000",
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %v", err)
	}

	nodeInstance = node
	return node, nil
}

// stopNode stops the node instance
func stopNode() error {
	nodeMutex.Lock()
	defer nodeMutex.Unlock()

	if nodeInstance != nil {
		if err := nodeInstance.Stop(); err != nil {
			return fmt.Errorf("failed to stop node: %v", err)
		}
		nodeInstance = nil
	}
	return nil
}
