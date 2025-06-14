package main

import (
	"fmt"
	"net"
	"sync"

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
		return nil, fmt.Errorf("no node is running")
	}
	return currentNode, nil
}

// setNode sets the current node instance
func setNode(node *network.Node) {
	nodeMutex.Lock()
	defer nodeMutex.Unlock()
	currentNode = node
}

// findAvailablePort searches for an available port starting from 3000
func findAvailablePort() (string, error) {
	for port := 3000; port < 4000; port++ {
		addr := fmt.Sprintf("localhost:%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return addr, nil
		}
	}
	return "", fmt.Errorf("no available ports found")
}

// initializeNode initializes a new node with an available port
func initializeNode() (*network.Node, error) {
	addr, err := findAvailablePort()
	if err != nil {
		return nil, err
	}

	config := &network.Config{
		Address:   addr,
		BlockType: "golden",
	}

	node, err := network.NewNode(config)
	if err != nil {
		return nil, err
	}

	setNode(node)
	return node, nil
}

// ensureNode ensures that a node is initialized
func ensureNode() (*network.Node, error) {
	node, err := getNode()
	if err == nil {
		return node, nil
	}
	return initializeNode()
}
