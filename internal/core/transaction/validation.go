package transaction

import (
	"fmt"

	"github.com/youngchain/internal/core/common"
)

// ValidationError represents a transaction validation error
type ValidationError struct {
	Code    int
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateTransaction validates a transaction
func ValidateTransaction(tx *common.Transaction, utxoGetter UTXOGetter) error {
	// Validate basic structure
	if err := validateBasicStructure(tx); err != nil {
		return err
	}

	// Validate inputs
	if err := validateInputs(tx, utxoGetter); err != nil {
		return err
	}

	// Validate outputs
	if err := validateOutputs(tx); err != nil {
		return err
	}

	// Validate fee
	if err := validateFee(tx, utxoGetter); err != nil {
		return err
	}

	return nil
}

// UTXOGetter defines the interface for getting UTXOs
type UTXOGetter interface {
	GetUTXO(txHash []byte, index uint32) (*common.UTXO, error)
}

// validateBasicStructure validates the basic structure of a transaction
func validateBasicStructure(tx *common.Transaction) error {
	if tx == nil {
		return &ValidationError{Code: 1, Message: "transaction is nil"}
	}

	inputs := tx.Inputs()
	if len(inputs) == 0 {
		return &ValidationError{Code: 2, Message: "transaction has no inputs"}
	}

	outputs := tx.Outputs()
	if len(outputs) == 0 {
		return &ValidationError{Code: 3, Message: "transaction has no outputs"}
	}

	return nil
}

// validateInputs validates the inputs of a transaction
func validateInputs(tx *common.Transaction, utxoGetter UTXOGetter) error {
	var totalInput uint64
	spentUTXOs := make(map[string]bool)

	inputs := tx.Inputs()
	for _, input := range inputs {
		// Check for double spend within the transaction
		key := fmt.Sprintf("%x:%d", input.PreviousTxHash, input.PreviousTxIndex)
		if spentUTXOs[key] {
			return &ValidationError{Code: 4, Message: "double spend detected"}
		}
		spentUTXOs[key] = true

		// Get and validate UTXO
		utxo, err := utxoGetter.GetUTXO(input.PreviousTxHash, input.PreviousTxIndex)
		if err != nil {
			return &ValidationError{Code: 5, Message: fmt.Sprintf("failed to get UTXO: %v", err)}
		}

		if utxo == nil {
			return &ValidationError{Code: 6, Message: "UTXO not found"}
		}

		if utxo.IsSpent {
			return &ValidationError{Code: 7, Message: "UTXO is already spent"}
		}

		// Validate signature
		if !validateSignature() {
			return &ValidationError{Code: 8, Message: "invalid signature"}
		}

		totalInput += utxo.Amount
	}

	// Check if inputs are sufficient
	var totalOutput uint64
	outputs := tx.Outputs()
	for _, output := range outputs {
		totalOutput += output.Value
	}

	// Calculate fee as the difference between inputs and outputs
	fee := totalInput - totalOutput
	if fee < 0 {
		return &ValidationError{Code: 9, Message: "insufficient funds"}
	}

	return nil
}

// validateOutputs validates the outputs of a transaction
func validateOutputs(tx *common.Transaction) error {
	outputs := tx.Outputs()
	for _, output := range outputs {
		if output.Value == 0 {
			return &ValidationError{Code: 10, Message: "output value cannot be zero"}
		}

		if len(output.ScriptPubKey) == 0 {
			return &ValidationError{Code: 11, Message: "output script cannot be empty"}
		}

		if len(output.Address) == 0 {
			return &ValidationError{Code: 12, Message: "output address cannot be empty"}
		}
	}

	return nil
}

// validateFee validates the transaction fee
func validateFee(tx *common.Transaction, utxoGetter UTXOGetter) error {
	// Calculate total input and output values
	var totalInput, totalOutput uint64

	inputs := tx.Inputs()
	for _, input := range inputs {
		utxo, err := utxoGetter.GetUTXO(input.PreviousTxHash, input.PreviousTxIndex)
		if err != nil {
			return &ValidationError{Code: 13, Message: fmt.Sprintf("failed to get UTXO: %v", err)}
		}
		totalInput += utxo.Amount
	}

	outputs := tx.Outputs()
	for _, output := range outputs {
		totalOutput += output.Value
	}

	// Fee is the difference between inputs and outputs
	fee := totalInput - totalOutput
	if fee < 0 {
		return &ValidationError{Code: 14, Message: "fee cannot be negative"}
	}

	// TODO: Add more fee validation rules
	return nil
}

// validateSignature validates the signature of a transaction input
func validateSignature() bool {
	// TODO: Implement signature validation
	// When implementing, convert common.Input to types.TxInput as needed
	return true
}
