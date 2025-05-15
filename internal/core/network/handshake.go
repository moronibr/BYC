package network

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"
)

// HandshakeState represents the state of a handshake
type HandshakeState int

const (
	// HandshakeStateInit represents the initial state
	HandshakeStateInit HandshakeState = iota
	// HandshakeStateChallenge represents the challenge state
	HandshakeStateChallenge
	// HandshakeStateResponse represents the response state
	HandshakeStateResponse
	// HandshakeStateComplete represents the complete state
	HandshakeStateComplete
)

// HandshakeProtocol manages the handshake protocol
type HandshakeProtocol struct {
	version     uint32
	userAgent   string
	startHeight uint64
	challenges  map[string]*HandshakeChallenge
}

// HandshakeChallenge represents a handshake challenge
type HandshakeChallenge struct {
	Nonce       [32]byte
	Timestamp   int64
	State       HandshakeState
	PeerID      string
	Version     uint32
	UserAgent   string
	StartHeight uint64
}

// NewHandshakeProtocol creates a new handshake protocol
func NewHandshakeProtocol(version uint32, userAgent string, startHeight uint64) *HandshakeProtocol {
	return &HandshakeProtocol{
		version:     version,
		userAgent:   userAgent,
		startHeight: startHeight,
		challenges:  make(map[string]*HandshakeChallenge),
	}
}

// InitiateHandshake initiates a handshake with a peer
func (hp *HandshakeProtocol) InitiateHandshake(peerID string) (*HandshakeChallenge, error) {
	// Generate random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Create challenge
	challenge := &HandshakeChallenge{
		Nonce:       [32]byte(nonce),
		Timestamp:   time.Now().Unix(),
		State:       HandshakeStateInit,
		PeerID:      peerID,
		Version:     hp.version,
		UserAgent:   hp.userAgent,
		StartHeight: hp.startHeight,
	}

	// Store challenge
	hp.challenges[peerID] = challenge

	return challenge, nil
}

// HandleChallenge handles a handshake challenge
func (hp *HandshakeProtocol) HandleChallenge(challenge *HandshakeChallenge) (*HandshakeChallenge, error) {
	// Validate challenge
	if err := hp.validateChallenge(challenge); err != nil {
		return nil, fmt.Errorf("invalid challenge: %v", err)
	}

	// Generate response nonce
	responseNonce := make([]byte, 32)
	if _, err := rand.Read(responseNonce); err != nil {
		return nil, fmt.Errorf("failed to generate response nonce: %v", err)
	}

	// Create response
	response := &HandshakeChallenge{
		Nonce:       [32]byte(responseNonce),
		Timestamp:   time.Now().Unix(),
		State:       HandshakeStateResponse,
		PeerID:      challenge.PeerID,
		Version:     hp.version,
		UserAgent:   hp.userAgent,
		StartHeight: hp.startHeight,
	}

	// Store challenge
	hp.challenges[challenge.PeerID] = challenge

	return response, nil
}

// ValidateResponse validates a handshake response
func (hp *HandshakeProtocol) ValidateResponse(response *HandshakeChallenge) error {
	// Get original challenge
	challenge, exists := hp.challenges[response.PeerID]
	if !exists {
		return fmt.Errorf("no challenge found for peer")
	}

	// Validate response
	if response.State != HandshakeStateResponse {
		return fmt.Errorf("invalid response state")
	}

	if response.Timestamp <= challenge.Timestamp {
		return fmt.Errorf("invalid response timestamp")
	}

	if response.Version != challenge.Version {
		return fmt.Errorf("version mismatch")
	}

	// Verify nonce
	expectedNonce := hp.calculateResponseNonce(challenge.Nonce)
	if response.Nonce != expectedNonce {
		return fmt.Errorf("invalid response nonce")
	}

	// Update challenge state
	challenge.State = HandshakeStateComplete

	return nil
}

// validateChallenge validates a handshake challenge
func (hp *HandshakeProtocol) validateChallenge(challenge *HandshakeChallenge) error {
	if challenge.State != HandshakeStateInit {
		return fmt.Errorf("invalid challenge state")
	}

	if challenge.Timestamp > time.Now().Unix()+30 {
		return fmt.Errorf("challenge timestamp too far in future")
	}

	if challenge.Timestamp < time.Now().Unix()-30 {
		return fmt.Errorf("challenge timestamp too old")
	}

	return nil
}

// calculateResponseNonce calculates the expected response nonce
func (hp *HandshakeProtocol) calculateResponseNonce(challengeNonce [32]byte) [32]byte {
	data := make([]byte, 0, 40)
	data = append(data, challengeNonce[:]...)
	data = binary.BigEndian.AppendUint32(data, hp.version)
	return sha256.Sum256(data)
}

// GetChallenge returns the challenge for a peer
func (hp *HandshakeProtocol) GetChallenge(peerID string) *HandshakeChallenge {
	return hp.challenges[peerID]
}

// RemoveChallenge removes a challenge
func (hp *HandshakeProtocol) RemoveChallenge(peerID string) {
	delete(hp.challenges, peerID)
}

// IsHandshakeComplete checks if a handshake is complete
func (hp *HandshakeProtocol) IsHandshakeComplete(peerID string) bool {
	challenge, exists := hp.challenges[peerID]
	return exists && challenge.State == HandshakeStateComplete
}
