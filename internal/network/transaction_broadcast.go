package network

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/youngchain/internal/core/types"
)

const (
	transactionPort = 8333
)

// TransactionBroadcaster handles broadcasting transactions to the network
type TransactionBroadcaster struct {
	peers     map[string]net.Conn
	peersLock sync.RWMutex
}

// NewTransactionBroadcaster creates a new transaction broadcaster
func NewTransactionBroadcaster() *TransactionBroadcaster {
	return &TransactionBroadcaster{
		peers: make(map[string]net.Conn),
	}
}

// AddPeer adds a new peer to the network
func (tb *TransactionBroadcaster) AddPeer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	tb.peersLock.Lock()
	tb.peers[addr] = conn
	tb.peersLock.Unlock()

	return nil
}

// BroadcastTransaction broadcasts a transaction to all peers
func (tb *TransactionBroadcaster) BroadcastTransaction(tx *types.Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}

	tb.peersLock.RLock()
	defer tb.peersLock.RUnlock()

	for addr, conn := range tb.peers {
		_, err := conn.Write(data)
		if err != nil {
			fmt.Printf("Failed to broadcast to peer %s: %v\n", addr, err)
			continue
		}
	}

	return nil
}

// StartServer starts the transaction broadcast server
func (tb *TransactionBroadcaster) StartServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", transactionPort))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Failed to accept connection: %v\n", err)
				continue
			}

			go tb.handleConnection(conn)
		}
	}()

	return nil
}

// handleConnection handles incoming peer connections
func (tb *TransactionBroadcaster) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		var tx types.Transaction
		if err := json.Unmarshal(buf[:n], &tx); err != nil {
			fmt.Printf("Failed to unmarshal transaction: %v\n", err)
			continue
		}

		// Process received transaction
		// TODO: Add transaction validation and processing
	}
}
