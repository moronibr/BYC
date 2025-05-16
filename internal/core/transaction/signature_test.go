package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/youngchain/internal/core/types"
)

func TestSignTransaction(t *testing.T) {
	// Generate a key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a transaction
	tx := &types.Transaction{
		// Set transaction fields as needed
	}

	// Sign the transaction
	signature, err := SignTransaction(tx, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Verify the signature
	if !VerifySignature(tx, signature, &privateKey.PublicKey) {
		t.Error("Failed to verify signature")
	}
}

func TestSignMultisigTransaction(t *testing.T) {
	// Generate key pairs
	privateKeys := make([]*ecdsa.PrivateKey, 3)
	publicKeys := make([]*ecdsa.PublicKey, 3)
	for i := 0; i < 3; i++ {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}
		privateKeys[i] = privateKey
		publicKeys[i] = &privateKey.PublicKey
	}

	// Create a transaction
	tx := &types.Transaction{
		// Set transaction fields as needed
	}

	// Sign the transaction with multiple keys
	multisig, err := SignMultisigTransaction(tx, privateKeys, publicKeys, 2)
	if err != nil {
		t.Fatalf("Failed to sign multisig transaction: %v", err)
	}

	// Verify the multisignature
	if !VerifyMultisigSignature(tx, multisig) {
		t.Error("Failed to verify multisignature")
	}
}

func TestSignSchnorrTransaction(t *testing.T) {
	// Generate a key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a transaction
	tx := &types.Transaction{
		// Set transaction fields as needed
	}

	// Sign the transaction using Schnorr
	signature, err := SignSchnorrTransaction(tx, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign transaction with Schnorr: %v", err)
	}

	// Verify the Schnorr signature
	if !VerifySchnorrSignature(tx, signature, &privateKey.PublicKey) {
		t.Error("Failed to verify Schnorr signature")
	}
}

func TestSignTaprootTransaction(t *testing.T) {
	// Generate a key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a transaction
	tx := &types.Transaction{
		// Set transaction fields as needed
	}

	// Sign the transaction using Taproot
	signature, err := SignTaprootTransaction(tx, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign transaction with Taproot: %v", err)
	}

	// Verify the Taproot signature
	if !VerifyTaprootSignature(tx, signature, &privateKey.PublicKey) {
		t.Error("Failed to verify Taproot signature")
	}
}
