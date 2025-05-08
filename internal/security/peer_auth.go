package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// PeerAuth represents peer authentication
type PeerAuth struct {
	mu         sync.RWMutex
	peers      map[string]*PeerInfo
	blacklist  map[string]time.Time
	reputation map[string]int
}

// PeerInfo represents peer information
type PeerInfo struct {
	PublicKey  ed25519.PublicKey
	LastSeen   time.Time
	Reputation int
	IsBanned   bool
	BanExpiry  time.Time
}

// NewPeerAuth creates a new peer authentication system
func NewPeerAuth() *PeerAuth {
	return &PeerAuth{
		peers:      make(map[string]*PeerInfo),
		blacklist:  make(map[string]time.Time),
		reputation: make(map[string]int),
	}
}

// GenerateKeyPair generates a new key pair
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// SignMessage signs a message with a private key
func SignMessage(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

// VerifySignature verifies a message signature
func VerifySignature(publicKey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// AddPeer adds a new peer
func (pa *PeerAuth) AddPeer(address string, publicKey ed25519.PublicKey) error {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	// Check if peer is blacklisted
	if banExpiry, banned := pa.blacklist[address]; banned {
		if time.Now().Before(banExpiry) {
			return fmt.Errorf("peer is blacklisted until %v", banExpiry)
		}
		delete(pa.blacklist, address)
	}

	pa.peers[address] = &PeerInfo{
		PublicKey:  publicKey,
		LastSeen:   time.Now(),
		Reputation: 0,
		IsBanned:   false,
	}
	return nil
}

// RemovePeer removes a peer
func (pa *PeerAuth) RemovePeer(address string) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	delete(pa.peers, address)
}

// UpdatePeerReputation updates a peer's reputation
func (pa *PeerAuth) UpdatePeerReputation(address string, delta int) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	if peer, exists := pa.peers[address]; exists {
		peer.Reputation += delta
		peer.LastSeen = time.Now()

		// Ban peer if reputation is too low
		if peer.Reputation < -100 {
			peer.IsBanned = true
			peer.BanExpiry = time.Now().Add(24 * time.Hour)
			pa.blacklist[address] = peer.BanExpiry
		}
	}
}

// IsPeerBanned checks if a peer is banned
func (pa *PeerAuth) IsPeerBanned(address string) bool {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if peer, exists := pa.peers[address]; exists {
		if peer.IsBanned {
			if time.Now().After(peer.BanExpiry) {
				peer.IsBanned = false
				return false
			}
			return true
		}
	}
	return false
}

// GetPeerInfo gets peer information
func (pa *PeerAuth) GetPeerInfo(address string) (*PeerInfo, error) {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if peer, exists := pa.peers[address]; exists {
		return peer, nil
	}
	return nil, fmt.Errorf("peer not found")
}

// CleanupExpiredBans removes expired bans
func (pa *PeerAuth) CleanupExpiredBans() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	now := time.Now()
	for address, expiry := range pa.blacklist {
		if now.After(expiry) {
			delete(pa.blacklist, address)
			if peer, exists := pa.peers[address]; exists {
				peer.IsBanned = false
			}
		}
	}
}

// EncodePublicKey encodes a public key to base64
func EncodePublicKey(key ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(key)
}

// DecodePublicKey decodes a base64 public key
func DecodePublicKey(key string) (ed25519.PublicKey, error) {
	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(data), nil
}
