package security

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var (
	// addressRegex is a regex for validating cryptocurrency addresses
	addressRegex = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	// hexRegex is a regex for validating hex strings
	hexRegex = regexp.MustCompile(`^[0-9a-fA-F]+$`)
)

// Validator provides input validation and sanitization
type Validator struct {
	maxInputLength int
}

// NewValidator creates a new validator
func NewValidator(maxInputLength int) *Validator {
	return &Validator{
		maxInputLength: maxInputLength,
	}
}

// ValidateAddress validates a cryptocurrency address
func (v *Validator) ValidateAddress(address string) error {
	if !addressRegex.MatchString(address) {
		return fmt.Errorf("invalid address format")
	}
	return nil
}

// ValidateHex validates a hex string
func (v *Validator) ValidateHex(hexStr string) error {
	if !hexRegex.MatchString(hexStr) {
		return fmt.Errorf("invalid hex format")
	}
	return nil
}

// ValidateHash validates a transaction or block hash
func (v *Validator) ValidateHash(hash []byte) error {
	if len(hash) != 32 {
		return fmt.Errorf("invalid hash length")
	}
	return nil
}

// SanitizeString sanitizes a string input
func (v *Validator) SanitizeString(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Limit length
	if len(input) > v.maxInputLength {
		input = input[:v.maxInputLength]
	}

	// Remove control characters
	input = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)

	return input
}

// ValidateAmount validates a transaction amount
func (v *Validator) ValidateAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

// ValidateFee validates a transaction fee
func (v *Validator) ValidateFee(fee int64) error {
	if fee < 0 {
		return fmt.Errorf("fee cannot be negative")
	}
	return nil
}

// ValidateScript validates a script
func (v *Validator) ValidateScript(script []byte) error {
	if len(script) == 0 {
		return fmt.Errorf("script cannot be empty")
	}
	if len(script) > v.maxInputLength {
		return fmt.Errorf("script too long")
	}
	return nil
}

// ValidateSignature validates a signature
func (v *Validator) ValidateSignature(signature []byte) error {
	if len(signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}
	if len(signature) > v.maxInputLength {
		return fmt.Errorf("signature too long")
	}
	return nil
}

// ValidateHexString validates and decodes a hex string
func (v *Validator) ValidateHexString(hexStr string) ([]byte, error) {
	if err := v.ValidateHex(hexStr); err != nil {
		return nil, err
	}
	return hex.DecodeString(hexStr)
}

// ValidateAndSanitizeInput validates and sanitizes user input
func (v *Validator) ValidateAndSanitizeInput(input string) (string, error) {
	sanitized := v.SanitizeString(input)
	if len(sanitized) == 0 {
		return "", fmt.Errorf("input cannot be empty after sanitization")
	}
	return sanitized, nil
}
