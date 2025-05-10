package transaction

import (
	"github.com/youngchain/internal/core/common"
)

// CalculateWeight calculates the weight of a transaction according to the SegWit weight formula:
// weight = base_size * 3 + total_size
// where base_size is the size of the transaction excluding witness data
// and total_size is the size of the transaction including witness data
func CalculateWeight(tx *common.Transaction) int {
	// Calculate base size (excluding witness data)
	baseSize := tx.Size()

	// Calculate witness size
	witnessSize := 0
	for _, witness := range tx.Witness() {
		witnessSize += len(witness)
	}

	// Calculate total weight
	weight := baseSize*3 + witnessSize
	return weight
}
