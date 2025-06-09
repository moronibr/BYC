package main

import (
	"fmt"
	"sync"

	"byc/internal/blockchain"
	"byc/internal/network"
)

var (
	currentNode *network.Node
	nodeMutex   sync.RWMutex
)

// getNode returns the current node instance
func getNode() (*network.Node, error) {
	nodeMutex.RLock()
	defer nodeMutex.RUnlock()

	if currentNode == nil {
		return nil, fmt.Errorf("node not initialized. Please start the node first")
	}
	return currentNode, nil
}

// setNode sets the current node instance
func setNode(node *network.Node) {
	nodeMutex.Lock()
	defer nodeMutex.Unlock()
	currentNode = node
}

// findAvailablePort finds an available port for the node
func findAvailablePort() (string, error) {
	// Try ports starting from 3000
	for port := 3000; port < 4000; port++ {
		addr := fmt.Sprintf("localhost:%d", port)
		config := &network.Config{
			Address:        addr,
			BlockType:      blockchain.GoldenBlock,
			BootstrapPeers: []string{},
		}

		node, err := network.NewNode(config)
		if err == nil {
			// Port is available
			node.Stop() // Close the test node
			return addr, nil
		}
	}
	return "", fmt.Errorf("no available ports found")
}

// initializeNode initializes a new node with an available port
func initializeNode() (*network.Node, error) {
	addr, err := findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %v", err)
	}

	config := &network.Config{
		Address:        addr,
		BlockType:      blockchain.GoldenBlock,
		BootstrapPeers: []string{},
	}

	node, err := network.NewNode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %v", err)
	}

	setNode(node)
	return node, nil
}

// ensureNode ensures a node is initialized
func ensureNode() (*network.Node, error) {
	node, err := getNode()
	if err == nil {
		return node, nil
	}

	// Node not initialized, create a new one
	return initializeNode()
}
