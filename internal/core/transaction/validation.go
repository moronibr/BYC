package transaction

import (
	"errors"
	"fmt"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/storage"
)

var (
	ErrInvalidSignature    = errors.New("invalid transaction signature")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidOutput       = errors.New("invalid output")
	ErrInvalidCoinType     = errors.New("invalid coin type")
	ErrInvalidFee          = errors.New("invalid fee")
	ErrInvalidScript       = errors.New("invalid script")
	ErrDoubleSpend         = errors.New("double spend detected")
	ErrTransactionTooLarge = errors.New("transaction too large")
	ErrInvalidCrossChain   = errors.New("invalid cross-chain transaction")
	ErrInvalidMaturity     = errors.New("transaction not mature")
	ErrInvalidLockTime     = errors.New("invalid lock time")
)

// Validator handles transaction validation
type Validator struct {
	db *storage.DB
}

// NewValidator creates a new transaction validator
func NewValidator(db *storage.DB) *Validator {
	return &Validator{
		db: db,
	}
}

// ValidateTransaction validates a transaction
func (v *Validator) ValidateTransaction(tx *block.Transaction, mempool *Mempool) error {
	// Check transaction size
	if err := v.validateSize(tx); err != nil {
		return err
	}

	// Check for double spends
	if err := v.checkDoubleSpend(tx, mempool); err != nil {
		return err
	}

	// Validate inputs
	if err := v.validateInputs(tx); err != nil {
		return err
	}

	// Validate outputs
	if err := v.validateOutputs(tx); err != nil {
		return err
	}

	// Validate signatures
	if err := v.validateSignatures(tx); err != nil {
		return err
	}

	// Validate coin types
	if err := v.validateCoinTypes(tx); err != nil {
		return err
	}

	// Validate fee
	if err := v.validateFee(tx); err != nil {
		return err
	}

	// Validate scripts
	if err := v.validateScripts(tx); err != nil {
		return err
	}

	// Validate cross-chain rules
	if err := v.validateCrossChain(tx); err != nil {
		return err
	}

	// Validate maturity
	if err := v.validateMaturity(tx); err != nil {
		return err
	}

	// Validate lock time
	if err := v.validateLockTime(tx); err != nil {
		return err
	}

	return nil
}

// validateSize checks if the transaction size is within limits
func (v *Validator) validateSize(tx *block.Transaction) error {
	// TODO: Implement size validation based on network rules
	// For now, use a simple size check
	if tx.Size() > 100*1024 { // 100KB limit
		return ErrTransactionTooLarge
	}
	return nil
}

// checkDoubleSpend checks if any input is already spent
func (v *Validator) checkDoubleSpend(tx *block.Transaction, mempool *Mempool) error {
	// Check against mempool
	if mempool != nil {
		for _, input := range tx.Inputs {
			if mempool.IsInputSpent(&input) {
				return ErrDoubleSpend
			}
		}
	}

	// Check against blockchain
	// TODO: Implement blockchain double spend check
	return nil
}

// validateInputs validates transaction inputs
func (v *Validator) validateInputs(tx *block.Transaction) error {
	if len(tx.Inputs) == 0 {
		return ErrInvalidInput
	}

	for _, input := range tx.Inputs {
		// Check if input exists in UTXO set
		utxo, err := v.db.GetUTXO(input.PreviousTx, input.OutputIndex)
		if err != nil {
			return fmt.Errorf("input not found: %v", err)
		}

		// Check if input is spent
		if utxo.IsSpent {
			return ErrDoubleSpend
		}

		// Check if input is mature
		if !utxo.IsMature() {
			return ErrInvalidMaturity
		}
	}

	return nil
}

// validateOutputs validates transaction outputs
func (v *Validator) validateOutputs(tx *block.Transaction) error {
	if len(tx.Outputs) == 0 {
		return ErrInvalidOutput
	}

	for _, output := range tx.Outputs {
		// Check if output amount is valid
		if output.Value <= 0 {
			return ErrInvalidOutput
		}

		// Check if output script is valid
		if len(output.Script) == 0 {
			return ErrInvalidScript
		}
	}

	return nil
}

// validateSignatures validates transaction signatures
func (v *Validator) validateSignatures(tx *block.Transaction) error {
	// TODO: Implement signature validation
	// For now, just check if signature exists
	if len(tx.Signature) == 0 {
		return ErrInvalidSignature
	}
	return nil
}

// validateCoinTypes validates coin types in the transaction
func (v *Validator) validateCoinTypes(tx *block.Transaction) error {
	// Check if all inputs have the same coin type
	inputCoinType := tx.CoinType
	for _, input := range tx.Inputs {
		// Get the UTXO for this input to check its coin type
		utxo, err := v.db.GetUTXO(input.PreviousTx, input.OutputIndex)
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}
		if utxo.CoinType != inputCoinType {
			return ErrInvalidCoinType
		}
	}

	// Check if all outputs have the same coin type
	outputCoinType := tx.Outputs[0].CoinType
	for _, output := range tx.Outputs {
		if output.CoinType != outputCoinType {
			return ErrInvalidCoinType
		}
	}

	// Check if cross-chain rules are followed
	if inputCoinType != outputCoinType {
		// Only Antion can cross chains
		if inputCoinType != coin.Antion && outputCoinType != coin.Antion {
			return ErrInvalidCrossChain
		}
	}

	return nil
}

// validateFee validates transaction fee
func (v *Validator) validateFee(tx *block.Transaction) error {
	// Calculate input amount
	var inputAmount uint64
	for _, input := range tx.Inputs {
		utxo, err := v.db.GetUTXO(input.PreviousTx, input.OutputIndex)
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}
		inputAmount += utxo.Value
	}

	// Calculate output amount
	var outputAmount uint64
	for _, output := range tx.Outputs {
		outputAmount += output.Value
	}

	// Calculate fee
	fee := inputAmount - outputAmount
	if fee < 0 {
		return ErrInvalidFee
	}

	// Check if fee meets minimum requirement
	minFee := v.calculateMinFee(tx)
	if fee < minFee {
		return ErrInvalidFee
	}

	return nil
}

// calculateMinFee calculates the minimum fee for a transaction
func (v *Validator) calculateMinFee(tx *block.Transaction) uint64 {
	// Base fee
	baseFee := uint64(1000) // 0.00001 BYC

	// Size fee
	sizeFee := uint64(tx.Size()) * 10 // 0.0000001 BYC per byte

	// Priority fee (for transactions with unconfirmed inputs)
	priorityFee := uint64(0)
	for _, input := range tx.Inputs {
		utxo, err := v.db.GetUTXO(input.PreviousTx, input.OutputIndex)
		if err == nil && !utxo.IsSpent && !utxo.IsConfirmed {
			priorityFee += 1000 // 0.00001 BYC per unconfirmed input
		}
	}

	return baseFee + sizeFee + priorityFee
}

// validateScripts validates transaction scripts
func (v *Validator) validateScripts(tx *block.Transaction) error {
	// TODO: Implement script validation
	// For now, just check if scripts exist
	for _, input := range tx.Inputs {
		if len(input.Script) == 0 {
			return ErrInvalidScript
		}
	}

	for _, output := range tx.Outputs {
		if len(output.Script) == 0 {
			return ErrInvalidScript
		}
	}

	return nil
}

// validateCrossChain validates cross-chain transaction rules
func (v *Validator) validateCrossChain(tx *block.Transaction) error {
	// Check if all inputs have the same coin type
	inputCoinType := tx.CoinType
	for _, input := range tx.Inputs {
		// Get the UTXO for this input to check its coin type
		utxo, err := v.db.GetUTXO(input.PreviousTx, input.OutputIndex)
		if err != nil {
			return fmt.Errorf("failed to get UTXO: %v", err)
		}
		if utxo.CoinType != inputCoinType {
			return ErrInvalidCrossChain
		}
	}

	// Check if all outputs have the same coin type
	outputCoinType := tx.Outputs[0].CoinType
	for _, output := range tx.Outputs {
		if output.CoinType != outputCoinType {
			return ErrInvalidCrossChain
		}
	}

	// If input and output coin types are different, this is a cross-chain transaction
	if inputCoinType != outputCoinType {
		// Only Antion can cross chains
		if inputCoinType != coin.Antion && outputCoinType != coin.Antion {
			return ErrInvalidCrossChain
		}

		// Check if cross-chain fee is paid
		if !v.hasCrossChainFee(tx) {
			return ErrInvalidFee
		}
	}

	return nil
}

// hasCrossChainFee checks if a cross-chain transaction has the required fee
func (v *Validator) hasCrossChainFee(tx *block.Transaction) bool {
	// Calculate input amount
	var inputAmount uint64
	for _, input := range tx.Inputs {
		utxo, err := v.db.GetUTXO(input.PreviousTx, input.OutputIndex)
		if err != nil {
			return false
		}
		inputAmount += utxo.Value
	}

	// Calculate output amount
	var outputAmount uint64
	for _, output := range tx.Outputs {
		outputAmount += output.Value
	}

	// Calculate fee
	fee := inputAmount - outputAmount

	// Cross-chain fee is 0.1% of the transaction amount
	minCrossChainFee := outputAmount / 1000
	if outputAmount%1000 != 0 {
		minCrossChainFee++
	}

	return fee >= minCrossChainFee
}

// validateMaturity validates transaction maturity
func (v *Validator) validateMaturity(tx *block.Transaction) error {
	// Check if transaction is mature
	if !tx.IsMature() {
		return ErrInvalidMaturity
	}

	return nil
}

// validateLockTime validates transaction lock time
func (v *Validator) validateLockTime(tx *block.Transaction) error {
	// Check if lock time is valid
	if !tx.IsLockTimeValid() {
		return ErrInvalidLockTime
	}

	return nil
}
