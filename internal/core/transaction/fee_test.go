package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/youngchain/internal/core/block"
)

func TestFeeCalculator(t *testing.T) {
	// Create a new fee calculator with test values
	calculator := NewFeeCalculator(1000, 10, 500, 10000)

	t.Run("Basic Fee Calculation", func(t *testing.T) {
		// Create a test transaction
		tx := &block.Transaction{
			Version: 1,
			Inputs:  []block.TxInput{{}},
			Outputs: []block.TxOutput{{}},
		}

		// Calculate fee
		fee := calculator.CalculateFee(tx)

		// Verify fee is within bounds
		assert.GreaterOrEqual(t, fee, calculator.minFee)
		assert.LessOrEqual(t, fee, calculator.maxFee)
	})

	t.Run("Fee Estimation", func(t *testing.T) {
		// Test fee estimation with different input/output counts
		fee1 := calculator.EstimateFee(1, 1)
		fee2 := calculator.EstimateFee(2, 2)

		// Verify that more inputs/outputs result in higher fees
		assert.Greater(t, fee2, fee1)
	})

	t.Run("Priority Multiplier", func(t *testing.T) {
		// Set priority multiplier
		calculator.SetPriorityMultiplier(1.5)

		// Create a test transaction
		tx := &block.Transaction{
			Version: 1,
			Inputs:  []block.TxInput{{}},
			Outputs: []block.TxOutput{{}},
		}

		// Calculate fee
		fee := calculator.CalculateFee(tx)

		// Verify fee is within bounds
		assert.GreaterOrEqual(t, fee, calculator.minFee)
		assert.LessOrEqual(t, fee, calculator.maxFee)
	})

	t.Run("Fee Parameter Updates", func(t *testing.T) {
		// Test base fee update
		newBaseFee := int64(2000)
		calculator.SetBaseFee(newBaseFee)
		assert.Equal(t, newBaseFee, calculator.GetBaseFee())

		// Test fee per byte update
		newFeePerByte := int64(20)
		calculator.SetFeePerByte(newFeePerByte)
		assert.Equal(t, newFeePerByte, calculator.GetFeePerByte())

		// Test invalid updates
		calculator.SetBaseFee(100) // Should not update as it's below minFee
		assert.Equal(t, newBaseFee, calculator.GetBaseFee())

		calculator.SetFeePerByte(-10) // Should not update as it's negative
		assert.Equal(t, newFeePerByte, calculator.GetFeePerByte())
	})
}
