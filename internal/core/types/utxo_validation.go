package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
)

const (
	// DefaultValidationInterval is the default interval for validation
	DefaultValidationInterval = 1000
	// DefaultValidationThreshold is the default threshold for validation
	DefaultValidationThreshold = 0.95
)

// ValidationType represents the type of validation
type ValidationType byte

const (
	// ValidationTypeNone indicates no validation
	ValidationTypeNone ValidationType = iota
	// ValidationTypeHash indicates hash-based validation
	ValidationTypeHash
	// ValidationTypeMerkle indicates Merkle tree validation
	ValidationTypeMerkle
)

// ValidationResult represents the result of a validation
type ValidationResult struct {
	// Valid indicates whether the validation was successful
	Valid bool
	// Type is the type of validation
	Type ValidationType
	// Hash is the hash of the validated data
	Hash []byte
	// Size is the size of the validated data
	Size int64
	// Error is the error that occurred during validation
	Error error
}

// UTXOValidation handles validation of the UTXO set
type UTXOValidation struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Validation state
	validationType ValidationType
	interval       int
	threshold      float64
	lastHash       []byte
	lastSize       int64
	errorCount     int
}

// NewUTXOValidation creates a new UTXO validation handler
func NewUTXOValidation(utxoSet *UTXOSet) *UTXOValidation {
	return &UTXOValidation{
		utxoSet:        utxoSet,
		validationType: ValidationTypeNone,
		interval:       DefaultValidationInterval,
		threshold:      DefaultValidationThreshold,
	}
}

// Validate validates the UTXO set
func (uv *UTXOValidation) Validate() *ValidationResult {
	uv.mu.Lock()
	defer uv.mu.Unlock()

	// Check if validation is enabled
	if uv.validationType == ValidationTypeNone {
		return &ValidationResult{
			Valid: true,
			Type:  ValidationTypeNone,
		}
	}

	// Get UTXO set size
	size := uv.utxoSet.Size()

	// Validate based on type
	switch uv.validationType {
	case ValidationTypeHash:
		// Calculate hash
		hash := uv.calculateHash()

		// Check if hash matches
		if uv.lastHash != nil && !bytes.Equal(hash, uv.lastHash) {
			uv.errorCount++
			return &ValidationResult{
				Valid: false,
				Type:  ValidationTypeHash,
				Hash:  hash,
				Size:  size,
				Error: fmt.Errorf("hash mismatch: expected %x, got %x", uv.lastHash, hash),
			}
		}

		// Update last hash
		uv.lastHash = hash
		uv.lastSize = size

		return &ValidationResult{
			Valid: true,
			Type:  ValidationTypeHash,
			Hash:  hash,
			Size:  size,
		}

	case ValidationTypeMerkle:
		// Calculate Merkle root
		root := uv.calculateMerkleRoot()

		// Check if root matches
		if uv.lastHash != nil && !bytes.Equal(root, uv.lastHash) {
			uv.errorCount++
			return &ValidationResult{
				Valid: false,
				Type:  ValidationTypeMerkle,
				Hash:  root,
				Size:  size,
				Error: fmt.Errorf("Merkle root mismatch: expected %x, got %x", uv.lastHash, root),
			}
		}

		// Update last hash
		uv.lastHash = root
		uv.lastSize = size

		return &ValidationResult{
			Valid: true,
			Type:  ValidationTypeMerkle,
			Hash:  root,
			Size:  size,
		}

	default:
		return &ValidationResult{
			Valid: false,
			Type:  uv.validationType,
			Error: fmt.Errorf("unsupported validation type: %d", uv.validationType),
		}
	}
}

// calculateHash calculates the hash of the UTXO set
func (uv *UTXOValidation) calculateHash() []byte {
	// Create hash
	hash := sha256.New()

	// Get UTXOs
	utxos := uv.utxoSet.GetUTXOs()

	// Sort UTXOs by transaction ID and output index
	sort.Slice(utxos, func(i, j int) bool {
		if bytes.Equal(utxos[i].TxID, utxos[j].TxID) {
			return utxos[i].Vout < utxos[j].Vout
		}
		return bytes.Compare(utxos[i].TxID, utxos[j].TxID) < 0
	})

	// Hash each UTXO
	for _, utxo := range utxos {
		// Hash transaction ID
		hash.Write(utxo.TxID)

		// Hash output index
		binary.Write(hash, binary.LittleEndian, uint32(utxo.Vout))

		// Hash value
		binary.Write(hash, binary.LittleEndian, utxo.Value)

		// Hash public key hash
		hash.Write(utxo.PubKeyHash)

		// Hash spent flag
		binary.Write(hash, binary.LittleEndian, utxo.Spent)
	}

	return hash.Sum(nil)
}

// calculateMerkleRoot calculates the Merkle root of the UTXO set
func (uv *UTXOValidation) calculateMerkleRoot() []byte {
	// Get UTXOs
	utxos := uv.utxoSet.GetUTXOs()

	// Create leaves
	leaves := make([][]byte, len(utxos))
	for i, utxo := range utxos {
		// Create leaf
		leaf := make([]byte, 0, 32+4+8+32+1)

		// Add transaction ID
		leaf = append(leaf, utxo.TxID...)

		// Add output index
		index := make([]byte, 4)
		binary.LittleEndian.PutUint32(index, uint32(utxo.Vout))
		leaf = append(leaf, index...)

		// Add value
		value := make([]byte, 8)
		binary.LittleEndian.PutUint64(value, uint64(utxo.Value))
		leaf = append(leaf, value...)

		// Add public key hash
		leaf = append(leaf, utxo.PubKeyHash...)

		// Add spent flag
		if utxo.Spent {
			leaf = append(leaf, 1)
		} else {
			leaf = append(leaf, 0)
		}

		// Hash leaf
		hash := sha256.Sum256(leaf)
		leaves[i] = hash[:]
	}

	// Build Merkle tree
	for len(leaves) > 1 {
		// Create new level
		level := make([][]byte, 0, (len(leaves)+1)/2)

		// Hash pairs
		for i := 0; i < len(leaves); i += 2 {
			// Get left hash
			left := leaves[i]

			// Get right hash
			var right []byte
			if i+1 < len(leaves) {
				right = leaves[i+1]
			} else {
				right = left
			}

			// Hash pair
			hash := sha256.Sum256(append(left, right...))
			level = append(level, hash[:])
		}

		// Update leaves
		leaves = level
	}

	// Return root
	if len(leaves) == 0 {
		return nil
	}
	return leaves[0]
}

// GetValidationStats returns statistics about the validation
func (uv *UTXOValidation) GetValidationStats() *ValidationStats {
	uv.mu.RLock()
	defer uv.mu.RUnlock()

	stats := &ValidationStats{
		ValidationType: uv.validationType,
		Interval:       uv.interval,
		Threshold:      uv.threshold,
		LastHash:       uv.lastHash,
		LastSize:       uv.lastSize,
		ErrorCount:     uv.errorCount,
	}

	// Calculate error rate
	if uv.errorCount > 0 {
		stats.ErrorRate = float64(uv.errorCount) / float64(uv.interval)
	}

	return stats
}

// SetValidationType sets the type of validation
func (uv *UTXOValidation) SetValidationType(validationType ValidationType) {
	uv.mu.Lock()
	uv.validationType = validationType
	uv.mu.Unlock()
}

// SetInterval sets the interval for validation
func (uv *UTXOValidation) SetInterval(interval int) {
	uv.mu.Lock()
	uv.interval = interval
	uv.mu.Unlock()
}

// SetThreshold sets the threshold for validation
func (uv *UTXOValidation) SetThreshold(threshold float64) {
	uv.mu.Lock()
	uv.threshold = threshold
	uv.mu.Unlock()
}

// ValidationStats holds statistics about the validation
type ValidationStats struct {
	// ValidationType is the type of validation
	ValidationType ValidationType
	// Interval is the interval for validation
	Interval int
	// Threshold is the threshold for validation
	Threshold float64
	// LastHash is the last hash of the validated data
	LastHash []byte
	// LastSize is the last size of the validated data
	LastSize int64
	// ErrorCount is the number of validation errors
	ErrorCount int
	// ErrorRate is the rate of validation errors
	ErrorRate float64
}

// String returns a string representation of the validation statistics
func (vs *ValidationStats) String() string {
	return fmt.Sprintf(
		"Validation Type: %d\n"+
			"Interval: %d, Threshold: %.2f\n"+
			"Last Hash: %x\n"+
			"Last Size: %d\n"+
			"Error Count: %d, Error Rate: %.2f",
		vs.ValidationType,
		vs.Interval, vs.Threshold,
		vs.LastHash,
		vs.LastSize,
		vs.ErrorCount, vs.ErrorRate,
	)
}
