package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/youngchain/internal/core/types"
)

var (
	ErrInvalidPublicKey = errors.New("invalid public key")
)

// Signature represents a transaction signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// SignTransaction signs a transaction with a private key
func SignTransaction(tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*Signature, error) {
	// Calculate transaction hash
	hash := tx.CalculateHash()

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return &Signature{
		R: r,
		S: s,
	}, nil
}

// VerifySignature verifies a transaction signature
func VerifySignature(tx *types.Transaction, signature *Signature, publicKey *ecdsa.PublicKey) bool {
	// Calculate transaction hash
	hash := tx.CalculateHash()

	// Verify the signature
	return ecdsa.Verify(publicKey, hash, signature.R, signature.S)
}

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// PublicKeyToAddress converts a public key to an address
func PublicKeyToAddress(publicKey *ecdsa.PublicKey) string {
	// Serialize public key
	pubKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// Hash the public key
	hash := sha256.Sum256(pubKeyBytes)

	// Take the last 20 bytes as the address
	address := hash[len(hash)-20:]

	// Convert to hex string
	return fmt.Sprintf("0x%x", address)
}

// Script represents a transaction script
type Script struct {
	// Script type (P2PKH, P2SH, etc.)
	Type string
	// Script data
	Data []byte
}

// ValidateScript validates a transaction script
func ValidateScript(script *Script, tx *types.Transaction, inputIndex int) bool {
	switch script.Type {
	case "P2PKH":
		return validateP2PKH(script, tx, inputIndex)
	case "P2SH":
		return validateP2SH(script, tx, inputIndex)
	default:
		return false
	}
}

// validateP2PKH validates a Pay-to-Public-Key-Hash script
func validateP2PKH(script *Script, tx *types.Transaction, inputIndex int) bool {
	// Extract public key and signature from script
	if len(script.Data) < 2 {
		return false
	}

	// Extract signature and public key
	sigBytes := script.Data[:len(script.Data)-1]
	pubKeyBytes := script.Data[len(script.Data)-1:]

	// Parse public key
	pubKey, err := parsePublicKey(pubKeyBytes)
	if err != nil {
		return false
	}

	// Parse signature
	sig, err := parseSignature(sigBytes)
	if err != nil {
		return false
	}

	// Verify signature
	return VerifySignature(tx, sig, pubKey)
}

// validateP2SH validates a Pay-to-Script-Hash script
func validateP2SH(script *Script, tx *types.Transaction, inputIndex int) bool {
	// TODO: Implement P2SH validation
	return true
}

// parsePublicKey parses a public key from bytes
func parsePublicKey(data []byte) (*ecdsa.PublicKey, error) {
	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, data)
	if x == nil {
		return nil, fmt.Errorf("invalid public key")
	}
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// parseSignature parses a signature from bytes
func parseSignature(data []byte) (*Signature, error) {
	if len(data) != 64 {
		return nil, fmt.Errorf("invalid signature length")
	}
	return &Signature{
		R: new(big.Int).SetBytes(data[:32]),
		S: new(big.Int).SetBytes(data[32:]),
	}, nil
}
