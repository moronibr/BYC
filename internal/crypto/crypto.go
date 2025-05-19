package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"errors"
	"math/big"
)

// PrivateKeyToBytes converts an ECDSA private key to bytes
func PrivateKeyToBytes(privateKey *ecdsa.PrivateKey) []byte {
	return privateKey.D.Bytes()
}

// PublicKeyToBytes converts an ECDSA public key to bytes
func PublicKeyToBytes(publicKey *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
}

// BytesToPrivateKey converts bytes to an ECDSA private key
func BytesToPrivateKey(privateKeyBytes []byte) (*ecdsa.PrivateKey, error) {
	curve := elliptic.P256()
	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = curve
	privateKey.D = new(big.Int).SetBytes(privateKeyBytes)

	// Calculate public key
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKey.D.Bytes())

	return privateKey, nil
}

// BytesToPublicKey converts bytes to an ECDSA public key
func BytesToPublicKey(publicKeyBytes []byte) (*ecdsa.PublicKey, error) {
	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, publicKeyBytes)
	if x == nil {
		return nil, errors.New("invalid public key bytes")
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// Sign signs a message using the private key
func Sign(message []byte, privateKeyBytes []byte) ([]byte, error) {
	curve := elliptic.P256()
	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = curve
	privateKey.D = new(big.Int).SetBytes(privateKeyBytes)

	// Calculate public key
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKey.D.Bytes())

	// Sign the message
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, message)
	if err != nil {
		return nil, err
	}

	// Encode the signature
	signature, err := asn1.Marshal(struct {
		R, S *big.Int
	}{r, s})
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// Verify verifies a message signature using the public key
func Verify(message []byte, signature []byte, publicKeyBytes []byte) bool {
	// Parse the signature
	var sig struct {
		R, S *big.Int
	}
	_, err := asn1.Unmarshal(signature, &sig)
	if err != nil {
		return false
	}

	// Parse the public key
	publicKey, err := BytesToPublicKey(publicKeyBytes)
	if err != nil {
		return false
	}

	// Verify the signature
	return ecdsa.Verify(publicKey, message, sig.R, sig.S)
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
