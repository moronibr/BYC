package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// TransactionValidator manages transaction validation
type TransactionValidator struct {
	// Configuration
	maxTxSize     int
	maxInputs     int
	maxOutputs    int
	maxScriptSize int
	maxValue      uint64
	minFee        uint64
	maxLockTime   uint32
	utxoSet       UTXOSet
}

// UTXOSet represents the set of unspent transaction outputs
type UTXOSet interface {
	GetUTXO(txHash [32]byte, outputIndex uint32) (*UTXO, error)
	SpendUTXO(txHash [32]byte, outputIndex uint32) error
	AddUTXO(txHash [32]byte, outputIndex uint32, utxo *UTXO) error
}

// UTXO represents an unspent transaction output
type UTXO struct {
	Value     uint64
	Script    []byte
	BlockTime int64
}

// NewTransactionValidator creates a new transaction validator
func NewTransactionValidator(maxTxSize, maxInputs, maxOutputs, maxScriptSize int, maxValue, minFee uint64, maxLockTime uint32, utxoSet UTXOSet) *TransactionValidator {
	return &TransactionValidator{
		maxTxSize:     maxTxSize,
		maxInputs:     maxInputs,
		maxOutputs:    maxOutputs,
		maxScriptSize: maxScriptSize,
		maxValue:      maxValue,
		minFee:        minFee,
		maxLockTime:   maxLockTime,
		utxoSet:       utxoSet,
	}
}

// ValidateTransaction validates a transaction
func (tv *TransactionValidator) ValidateTransaction(tx *Transaction, blockTime int64) error {
	// Validate transaction size
	if err := tv.validateTxSize(tx); err != nil {
		return fmt.Errorf("invalid transaction size: %v", err)
	}

	// Validate inputs and outputs
	if err := tv.validateInputsOutputs(tx); err != nil {
		return fmt.Errorf("invalid inputs/outputs: %v", err)
	}

	// Validate scripts
	if err := tv.validateScripts(tx); err != nil {
		return fmt.Errorf("invalid scripts: %v", err)
	}

	// Validate values and fees
	if err := tv.validateValuesAndFees(tx, blockTime); err != nil {
		return fmt.Errorf("invalid values/fees: %v", err)
	}

	// Validate lock time
	if err := tv.validateLockTime(tx, blockTime); err != nil {
		return fmt.Errorf("invalid lock time: %v", err)
	}

	return nil
}

// validateTxSize validates the transaction size
func (tv *TransactionValidator) validateTxSize(tx *Transaction) error {
	size := tx.Size()
	if size > tv.maxTxSize {
		return fmt.Errorf("transaction size %d exceeds maximum %d", size, tv.maxTxSize)
	}
	return nil
}

// validateInputsOutputs validates the transaction inputs and outputs
func (tv *TransactionValidator) validateInputsOutputs(tx *Transaction) error {
	// Skip validation for coinbase transactions
	if tx.IsCoinbase() {
		return nil
	}

	// Validate input count
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction must have at least one input")
	}
	if len(tx.Inputs) > tv.maxInputs {
		return fmt.Errorf("transaction has too many inputs")
	}

	// Validate output count
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction must have at least one output")
	}
	if len(tx.Outputs) > tv.maxOutputs {
		return fmt.Errorf("transaction has too many outputs")
	}

	// Check for duplicate inputs
	seen := make(map[string]bool)
	for _, input := range tx.Inputs {
		key := fmt.Sprintf("%x:%d", input.PrevHash, input.PrevIndex)
		if seen[key] {
			return fmt.Errorf("duplicate input")
		}
		seen[key] = true
	}

	return nil
}

// validateScripts validates the transaction scripts
func (tv *TransactionValidator) validateScripts(tx *Transaction) error {
	// Skip validation for coinbase transactions
	if tx.IsCoinbase() {
		return nil
	}

	// Validate input scripts
	for i, input := range tx.Inputs {
		if len(input.Script) > tv.maxScriptSize {
			return fmt.Errorf("input script at index %d exceeds maximum size", i)
		}
	}

	// Validate output scripts
	for i, output := range tx.Outputs {
		if len(output.Script) > tv.maxScriptSize {
			return fmt.Errorf("output script at index %d exceeds maximum size", i)
		}
	}

	return nil
}

// validateValuesAndFees validates the transaction values and fees
func (tv *TransactionValidator) validateValuesAndFees(tx *Transaction, blockTime int64) error {
	// Skip validation for coinbase transactions
	if tx.IsCoinbase() {
		return nil
	}

	var inputSum uint64
	var outputSum uint64

	// Sum input values
	for _, input := range tx.Inputs {
		utxo, err := tv.utxoSet.GetUTXO(input.PrevHash, input.PrevIndex)
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}
		inputSum += utxo.Value
	}

	// Sum output values
	for _, output := range tx.Outputs {
		if output.Value > tv.maxValue {
			return fmt.Errorf("output value exceeds maximum")
		}
		outputSum += output.Value
	}

	// Calculate fee
	fee := inputSum - outputSum
	if fee < tv.minFee {
		return fmt.Errorf("fee %d is below minimum %d", fee, tv.minFee)
	}

	return nil
}

// validateLockTime validates the transaction lock time
func (tv *TransactionValidator) validateLockTime(tx *Transaction, blockTime int64) error {
	if tx.LockTime > tv.maxLockTime {
		return fmt.Errorf("lock time exceeds maximum")
	}

	// If lock time is a timestamp, it must be in the future
	if tx.LockTime >= 500000000 {
		if tx.LockTime <= uint32(blockTime) {
			return fmt.Errorf("lock time must be in the future")
		}
	}

	return nil
}

// CalculateTransactionHash calculates the transaction hash
func (tv *TransactionValidator) CalculateTransactionHash(tx *Transaction) [32]byte {
	data := make([]byte, 0, tx.Size())
	data = binary.BigEndian.AppendUint32(data, tx.Version)

	// Add inputs
	data = append(data, byte(len(tx.Inputs)))
	for _, input := range tx.Inputs {
		data = append(data, input.PrevHash[:]...)
		data = binary.BigEndian.AppendUint32(data, input.PrevIndex)
		data = append(data, byte(len(input.Script)))
		data = append(data, input.Script...)
		data = binary.BigEndian.AppendUint32(data, input.Sequence)
	}

	// Add outputs
	data = append(data, byte(len(tx.Outputs)))
	for _, output := range tx.Outputs {
		data = binary.BigEndian.AppendUint64(data, output.Value)
		data = append(data, byte(len(output.Script)))
		data = append(data, output.Script...)
	}

	data = binary.BigEndian.AppendUint32(data, tx.LockTime)
	return sha256.Sum256(data)
}
