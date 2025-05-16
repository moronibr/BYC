package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

const (
	// DefaultTransactionVersion is the default version for transactions
	DefaultTransactionVersion = 1
	// DefaultTransactionLockTime is the default lock time for transactions
	DefaultTransactionLockTime = 0
)

// BitcoinTransaction represents a Bitcoin transaction
type BitcoinTransaction struct {
	// Version is the transaction version
	Version uint32
	// Inputs are the transaction inputs
	Inputs []*TransactionInput
	// Outputs are the transaction outputs
	Outputs []*TransactionOutput
	// LockTime is the transaction lock time
	LockTime uint32
}

// TransactionInput represents a transaction input
type TransactionInput struct {
	// PreviousTxID is the ID of the previous transaction
	PreviousTxID []byte
	// PreviousTxIndex is the index of the previous transaction output
	PreviousTxIndex int
	// ScriptSig is the input script
	ScriptSig []byte
	// Sequence is the input sequence
	Sequence uint32
}

// TransactionOutput represents a transaction output
type TransactionOutput struct {
	// Value is the output value in satoshis
	Value int64
	// ScriptPubKey is the output script
	ScriptPubKey []byte
}

// NewBitcoinTransaction creates a new Bitcoin transaction
func NewBitcoinTransaction() *BitcoinTransaction {
	return &BitcoinTransaction{
		Version:  DefaultTransactionVersion,
		Inputs:   make([]*TransactionInput, 0),
		Outputs:  make([]*TransactionOutput, 0),
		LockTime: DefaultTransactionLockTime,
	}
}

// AddInput adds an input to the transaction
func (tx *BitcoinTransaction) AddInput(previousTxID []byte, previousTxIndex int, scriptSig []byte) {
	tx.Inputs = append(tx.Inputs, &TransactionInput{
		PreviousTxID:    previousTxID,
		PreviousTxIndex: previousTxIndex,
		ScriptSig:       scriptSig,
		Sequence:        0xffffffff,
	})
}

// AddOutput adds an output to the transaction
func (tx *BitcoinTransaction) AddOutput(value int64, scriptPubKey []byte) {
	tx.Outputs = append(tx.Outputs, &TransactionOutput{
		Value:        value,
		ScriptPubKey: scriptPubKey,
	})
}

// Sign signs the transaction
func (tx *BitcoinTransaction) Sign(privateKey *ecdsa.PrivateKey, utxoSet *UTXOSet) error {
	// Create a copy of the transaction for signing
	txCopy := tx.Copy()

	// Sign each input
	for i, input := range txCopy.Inputs {
		// Get the previous transaction output
		utxo, err := utxoSet.GetUTXO(input.PreviousTxID, uint32(input.PreviousTxIndex))
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}

		// Create the signature hash
		hash, err := txCopy.SignatureHash(i, utxo.ScriptPubKey)
		if err != nil {
			return fmt.Errorf("failed to create signature hash: %v", err)
		}

		// Sign the hash
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
		if err != nil {
			return fmt.Errorf("failed to sign transaction: %v", err)
		}

		// Create the signature
		signature := append(r.Bytes(), s.Bytes()...)
		signature = append(signature, byte(0x01)) // SIGHASH_ALL

		// Create the script signature
		scriptSig := append(signature, utxo.ScriptPubKey...)
		tx.Inputs[i].ScriptSig = scriptSig
	}

	return nil
}

// Verify verifies the transaction
func (tx *BitcoinTransaction) Verify(utxoSet *UTXOSet) error {
	// Create a copy of the transaction for verification
	txCopy := tx.Copy()

	// Verify each input
	for i, input := range txCopy.Inputs {
		// Get the previous transaction output
		utxo, err := utxoSet.GetUTXO(input.PreviousTxID, uint32(input.PreviousTxIndex))
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}

		// Create the signature hash
		hash, err := txCopy.SignatureHash(i, utxo.ScriptPubKey)
		if err != nil {
			return fmt.Errorf("failed to create signature hash: %v", err)
		}

		// Extract the signature and public key
		signature := input.ScriptSig[:len(input.ScriptSig)-1]
		publicKeyBytes := input.ScriptSig[len(input.ScriptSig)-1:]

		// Parse the public key
		publicKey, err := parsePublicKey(publicKeyBytes)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %v", err)
		}

		// Verify the signature
		r := new(big.Int).SetBytes(signature[:32])
		s := new(big.Int).SetBytes(signature[32:64])
		if !ecdsa.Verify(publicKey, hash, r, s) {
			return fmt.Errorf("invalid signature")
		}
	}

	return nil
}

// parsePublicKey parses a public key from bytes
func parsePublicKey(publicKeyBytes []byte) (*ecdsa.PublicKey, error) {
	// TODO: Implement public key parsing
	return nil, nil
}

// Copy creates a copy of the transaction
func (tx *BitcoinTransaction) Copy() *BitcoinTransaction {
	txCopy := &BitcoinTransaction{
		Version:  tx.Version,
		LockTime: tx.LockTime,
		Inputs:   make([]*TransactionInput, len(tx.Inputs)),
		Outputs:  make([]*TransactionOutput, len(tx.Outputs)),
	}

	// Copy inputs
	for i, input := range tx.Inputs {
		txCopy.Inputs[i] = &TransactionInput{
			PreviousTxID:    make([]byte, len(input.PreviousTxID)),
			PreviousTxIndex: input.PreviousTxIndex,
			ScriptSig:       make([]byte, len(input.ScriptSig)),
			Sequence:        input.Sequence,
		}
		copy(txCopy.Inputs[i].PreviousTxID, input.PreviousTxID)
		copy(txCopy.Inputs[i].ScriptSig, input.ScriptSig)
	}

	// Copy outputs
	for i, output := range tx.Outputs {
		txCopy.Outputs[i] = &TransactionOutput{
			Value:        output.Value,
			ScriptPubKey: make([]byte, len(output.ScriptPubKey)),
		}
		copy(txCopy.Outputs[i].ScriptPubKey, output.ScriptPubKey)
	}

	return txCopy
}

// SignatureHash creates the signature hash for an input
func (tx *BitcoinTransaction) SignatureHash(inputIndex int, scriptPubKey []byte) ([]byte, error) {
	// Create a copy of the transaction
	txCopy := tx.Copy()

	// Clear all script signatures
	for i := range txCopy.Inputs {
		txCopy.Inputs[i].ScriptSig = nil
	}

	// Set the script signature for the input being signed
	txCopy.Inputs[inputIndex].ScriptSig = scriptPubKey

	// Serialize the transaction
	serialized, err := txCopy.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %v", err)
	}

	// Add the hash type
	serialized = append(serialized, 0x01, 0x00, 0x00, 0x00) // SIGHASH_ALL

	// Hash the serialized transaction
	hash := sha256.Sum256(serialized)
	hash = sha256.Sum256(hash[:])

	return hash[:], nil
}

// Serialize serializes the transaction
func (tx *BitcoinTransaction) Serialize() ([]byte, error) {
	// TODO: Implement transaction serialization
	return nil, nil
}

// Deserialize deserializes the transaction
func (tx *BitcoinTransaction) Deserialize(data []byte) error {
	// TODO: Implement transaction deserialization
	return nil
}

// GetID returns the transaction ID
func (tx *BitcoinTransaction) GetID() string {
	// Serialize the transaction
	serialized, err := tx.Serialize()
	if err != nil {
		return ""
	}

	// Hash the serialized transaction
	hash := sha256.Sum256(serialized)
	hash = sha256.Sum256(hash[:])

	// Return the hash as a hex string
	return hex.EncodeToString(hash[:])
}

// GetSize returns the size of the transaction in bytes
func (tx *BitcoinTransaction) GetSize() int {
	// Serialize the transaction
	serialized, err := tx.Serialize()
	if err != nil {
		return 0
	}

	return len(serialized)
}

// GetWeight returns the weight of the transaction
func (tx *BitcoinTransaction) GetWeight() int {
	// Serialize the transaction
	serialized, err := tx.Serialize()
	if err != nil {
		return 0
	}

	return len(serialized) * 4
}

// GetFee returns the fee of the transaction
func (tx *BitcoinTransaction) GetFee(utxoSet *UTXOSet) (int64, error) {
	// Calculate input sum
	var inputSum int64
	for _, input := range tx.Inputs {
		utxo, err := utxoSet.GetUTXO(input.PreviousTxID, uint32(input.PreviousTxIndex))
		if err != nil {
			return 0, fmt.Errorf("failed to get UTXO: %v", err)
		}
		inputSum += utxo.Value
	}

	// Calculate output sum
	var outputSum int64
	for _, output := range tx.Outputs {
		outputSum += output.Value
	}

	// Calculate fee
	return inputSum - outputSum, nil
}

// IsCoinbase returns whether the transaction is a coinbase transaction
func (tx *BitcoinTransaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].PreviousTxID) == 0 && tx.Inputs[0].PreviousTxIndex == -1
}

// IsValid returns whether the transaction is valid
func (tx *BitcoinTransaction) IsValid(utxoSet *UTXOSet) error {
	// Check if the transaction is a coinbase transaction
	if tx.IsCoinbase() {
		return nil
	}

	// Check if the transaction has inputs
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}

	// Check if the transaction has outputs
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Check if the transaction is valid
	if err := tx.Verify(utxoSet); err != nil {
		return fmt.Errorf("transaction verification failed: %v", err)
	}

	// Check if the transaction fee is valid
	fee, err := tx.GetFee(utxoSet)
	if err != nil {
		return fmt.Errorf("failed to get transaction fee: %v", err)
	}

	if fee < 0 {
		return fmt.Errorf("transaction fee is negative")
	}

	return nil
}

// String returns a string representation of the transaction
func (tx *BitcoinTransaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Transaction %s:", tx.GetID()))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("  Input %d:", i))
		lines = append(lines, fmt.Sprintf("    TXID:      %x", input.PreviousTxID))
		lines = append(lines, fmt.Sprintf("    Out:       %d", input.PreviousTxIndex))
		lines = append(lines, fmt.Sprintf("    Script:    %x", input.ScriptSig))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("  Output %d:", i))
		lines = append(lines, fmt.Sprintf("    Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("    Script: %x", output.ScriptPubKey))
	}

	return fmt.Sprintf("%s\n", lines)
}

// Validate validates the transaction
func (tx *BitcoinTransaction) Validate() error {
	// Check if transaction has inputs
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}

	// Check if transaction has outputs
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Check if transaction ID is valid
	if tx.GetID() == "" {
		return fmt.Errorf("transaction has no ID")
	}

	// Check if transaction lock time is valid
	if tx.LockTime > uint32(time.Now().Unix()) {
		return fmt.Errorf("invalid lock time")
	}

	// Validate inputs
	for i, input := range tx.Inputs {
		if input.PreviousTxID == nil {
			return fmt.Errorf("input %d has no transaction ID", i)
		}
		if input.PreviousTxIndex < 0 {
			return fmt.Errorf("input %d has invalid output index", i)
		}
		if input.ScriptSig == nil {
			return fmt.Errorf("input %d has no script signature", i)
		}
	}

	// Validate outputs
	for i, output := range tx.Outputs {
		if output.Value <= 0 {
			return fmt.Errorf("output %d has invalid value", i)
		}
		if output.ScriptPubKey == nil {
			return fmt.Errorf("output %d has no script public key", i)
		}
	}

	return nil
}

// validateBalance validates the input/output balance using UTXO set
func (tx *BitcoinTransaction) validateBalance(utxoSet UTXOSetInterface) error {
	var inputSum, outputSum int64

	// Calculate input sum from UTXOs
	for _, input := range tx.Inputs {
		utxo, err := utxoSet.GetUTXO(input.PreviousTxID, uint32(input.PreviousTxIndex))
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}
		inputSum += utxo.Value
	}

	// Calculate output sum
	for _, output := range tx.Outputs {
		outputSum += output.Value
	}

	// Check if input sum is greater than or equal to output sum
	if inputSum < outputSum {
		return fmt.Errorf("input sum (%d) is less than output sum (%d)", inputSum, outputSum)
	}

	return nil
}

// validateFees validates transaction fees using UTXO set
func (tx *BitcoinTransaction) validateFees(utxoSet UTXOSetInterface) error {
	var inputSum, outputSum int64

	// Calculate input sum from UTXOs
	for _, input := range tx.Inputs {
		utxo, err := utxoSet.GetUTXO(input.PreviousTxID, uint32(input.PreviousTxIndex))
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}
		inputSum += utxo.Value
	}

	// Calculate output sum
	for _, output := range tx.Outputs {
		outputSum += output.Value
	}

	// Calculate fee
	fee := inputSum - outputSum

	// Minimum fee requirement (0.0001 coins)
	minFee := int64(0.0001 * 1e8) // Assuming 8 decimal places
	if fee < minFee {
		return fmt.Errorf("transaction fee (%d) is below minimum (%d)", fee, minFee)
	}

	return nil
}
