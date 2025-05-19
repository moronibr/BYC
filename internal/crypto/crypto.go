package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

// Sign signs a message with a private key
func Sign(message []byte, privateKey []byte) ([]byte, error) {
	// Create a new private key
	key := new(ecdsa.PrivateKey)
	key.Curve = elliptic.P256()

	// Set the private key value
	key.D = new(big.Int).SetBytes(privateKey)

	// Calculate the public key
	key.PublicKey.X, key.PublicKey.Y = key.Curve.ScalarBaseMult(key.D.Bytes())

	// Hash the message
	hash := sha256.Sum256(message)

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		return nil, err
	}

	// Combine r and s into a single signature
	signature := append(r.Bytes(), s.Bytes()...)

	return signature, nil
}

// Verify verifies a signature
func Verify(message []byte, signature []byte, publicKey []byte) bool {
	// Create a new public key
	key := new(ecdsa.PublicKey)
	key.Curve = elliptic.P256()

	// Extract x and y coordinates
	x, y := elliptic.Unmarshal(key.Curve, publicKey)
	if x == nil {
		return false
	}

	key.X = x
	key.Y = y

	// Hash the message
	hash := sha256.Sum256(message)

	// Split signature into r and s
	if len(signature) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	// Verify the signature
	return ecdsa.Verify(key, hash[:], r, s)
}

// HashPublicKey hashes a public key
func HashPublicKey(publicKey []byte) []byte {
	hash := sha256.Sum256(publicKey)
	return hash[:]
}

// GenerateKeyPair generates a new key pair
func GenerateKeyPair() ([]byte, []byte, error) {
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Get the private key bytes
	privateKeyBytes := privateKey.D.Bytes()

	// Get the public key bytes
	publicKeyBytes := elliptic.Marshal(privateKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	return privateKeyBytes, publicKeyBytes, nil
}
