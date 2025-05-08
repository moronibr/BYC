package security

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// SignedMessage represents a signed message
type SignedMessage struct {
	Message   []byte `json:"message"`
	Signature []byte `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	PublicKey string `json:"public_key"`
}

// MessageSigner represents a message signer
type MessageSigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewMessageSigner creates a new message signer
func NewMessageSigner(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *MessageSigner {
	return &MessageSigner{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// SignMessage signs a message
func (ms *MessageSigner) SignMessage(message interface{}) (*SignedMessage, error) {
	// Marshal message to JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	// Create message hash
	hash := sha256.Sum256(messageBytes)

	// Sign hash
	signature := ed25519.Sign(ms.privateKey, hash[:])

	// Create signed message
	signedMessage := &SignedMessage{
		Message:   messageBytes,
		Signature: signature,
		Timestamp: time.Now().Unix(),
		PublicKey: EncodePublicKey(ms.publicKey),
	}

	return signedMessage, nil
}

// VerifyMessage verifies a signed message
func (ms *MessageSigner) VerifyMessage(signedMessage *SignedMessage) (bool, error) {
	// Verify timestamp (within 5 minutes)
	if time.Now().Unix()-signedMessage.Timestamp > 300 {
		return false, fmt.Errorf("message too old")
	}

	// Decode public key
	publicKey, err := DecodePublicKey(signedMessage.PublicKey)
	if err != nil {
		return false, fmt.Errorf("invalid public key: %v", err)
	}

	// Create message hash
	hash := sha256.Sum256(signedMessage.Message)

	// Verify signature
	return ed25519.Verify(publicKey, hash[:], signedMessage.Signature), nil
}

// VerifyMessageWithKey verifies a signed message with a specific public key
func (ms *MessageSigner) VerifyMessageWithKey(signedMessage *SignedMessage, publicKey ed25519.PublicKey) (bool, error) {
	// Verify timestamp (within 5 minutes)
	if time.Now().Unix()-signedMessage.Timestamp > 300 {
		return false, fmt.Errorf("message too old")
	}

	// Create message hash
	hash := sha256.Sum256(signedMessage.Message)

	// Verify signature
	return ed25519.Verify(publicKey, hash[:], signedMessage.Signature), nil
}

// UnmarshalMessage unmarshals a signed message
func (ms *MessageSigner) UnmarshalMessage(signedMessage *SignedMessage, target interface{}) error {
	// Verify message
	valid, err := ms.VerifyMessage(signedMessage)
	if err != nil {
		return fmt.Errorf("failed to verify message: %v", err)
	}
	if !valid {
		return fmt.Errorf("invalid message signature")
	}

	// Unmarshal message
	if err := json.Unmarshal(signedMessage.Message, target); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return nil
}

// UnmarshalMessageWithKey unmarshals a signed message with a specific public key
func (ms *MessageSigner) UnmarshalMessageWithKey(signedMessage *SignedMessage, publicKey ed25519.PublicKey, target interface{}) error {
	// Verify message
	valid, err := ms.VerifyMessageWithKey(signedMessage, publicKey)
	if err != nil {
		return fmt.Errorf("failed to verify message: %v", err)
	}
	if !valid {
		return fmt.Errorf("invalid message signature")
	}

	// Unmarshal message
	if err := json.Unmarshal(signedMessage.Message, target); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return nil
}
