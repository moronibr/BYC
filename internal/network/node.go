package network

import (
	"fmt"
)

// NetworkError represents a network-related error
type NetworkError struct {
	Operation string
	Reason    string
	Details   map[string]interface{}
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s: %s", e.Operation, e.Reason)
}

// ConnectionError represents a peer connection error
type ConnectionError struct {
	PeerAddress string
	Reason      string
	Details     map[string]interface{}
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("failed to connect to peer %s: %s", e.PeerAddress, e.Reason)
}

// MessageError represents a network message error
type MessageError struct {
	MessageType string
	Reason      string
	Details     map[string]interface{}
}

func (e *MessageError) Error() string {
	return fmt.Sprintf("message error for type %s: %s", e.MessageType, e.Reason)
}
