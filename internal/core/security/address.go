package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Address represents a secure address
type Address struct {
	PublicKey  *ecdsa.PublicKey
	AddressStr string
}

// GenerateAddress generates a new secure address
func GenerateAddress() (*Address, error) {
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Get the public key
	publicKey := &privateKey.PublicKey

	// Generate the address string (e.g., using the public key hash)
	addressStr := generateAddressString(publicKey)

	return &Address{
		PublicKey:  publicKey,
		AddressStr: addressStr,
	}, nil
}

// generateAddressString generates a string representation of the address
func generateAddressString(publicKey *ecdsa.PublicKey) string {
	// Concatenate the x and y coordinates of the public key
	keyBytes := append(publicKey.X.Bytes(), publicKey.Y.Bytes()...)

	// Hash the key bytes
	hash := sha256.Sum256(keyBytes)

	// Convert the hash to a hex string
	return hex.EncodeToString(hash[:])
}

// ValidateAddress validates an address string
func ValidateAddress(addressStr string) bool {
	// Decode the address string
	addressBytes, err := hex.DecodeString(addressStr)
	if err != nil {
		return false
	}

	// Check the length of the address bytes
	if len(addressBytes) != 32 {
		return false
	}

	return true
}
