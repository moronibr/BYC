package utils

import (
	"fmt"
	"net"
)

// FindAvailablePort finds an available port starting from the given port
func FindAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+1000; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found in range %d-%d", startPort, startPort+1000)
}

// ParseAddress extracts the host and port from an address string
func ParseAddress(address string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}

	port := 0
	fmt.Sscanf(portStr, "%d", &port)
	if port == 0 {
		return "", 0, fmt.Errorf("invalid port number: %s", portStr)
	}

	return host, port, nil
}

// FindAvailableAddress finds an available address by trying ports starting from the given address
func FindAvailableAddress(address string) (string, error) {
	host, port, err := ParseAddress(address)
	if err != nil {
		return "", err
	}

	// Determine port range based on the port number
	startPort := port
	if port >= 8000 && port < 9000 {
		// API server port range (8000-8999)
		startPort = 8000
	} else if port >= 3000 && port < 4000 {
		// P2P server port range (3000-3999)
		startPort = 3000
	}

	availablePort, err := FindAvailablePort(startPort)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", host, availablePort), nil
}
