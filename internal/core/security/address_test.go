package security

import (
	"testing"
)

func TestGenerateAddress(t *testing.T) {
	address, err := GenerateAddress()
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	if address.PublicKey == nil {
		t.Error("Expected non-nil public key")
	}

	if address.AddressStr == "" {
		t.Error("Expected non-empty address string")
	}
}

func TestValidateAddress(t *testing.T) {
	address, err := GenerateAddress()
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	if !ValidateAddress(address.AddressStr) {
		t.Error("Expected valid address")
	}

	if ValidateAddress("invalid") {
		t.Error("Expected invalid address")
	}
}
