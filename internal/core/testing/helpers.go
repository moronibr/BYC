package testing

import (
	"testing"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/common"
)

// TestBlock creates a test block with given parameters
func TestBlock(t *testing.T, prevHash []byte, timestamp time.Time) *block.Block {
	return &block.Block{
		Header: &common.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			Timestamp:     timestamp,
			Difficulty:    1,
		},
		Transactions: []*common.Transaction{},
	}
}

// TestTransaction creates a test transaction
func TestTransaction(t *testing.T, from, to string, amount uint64, coinType coin.CoinType) *common.Transaction {
	tx := &common.Transaction{
		Version:   1,
		Timestamp: time.Now(),
		From:      []byte(from),
		To:        []byte(to),
		Amount:    amount,
		Inputs:    make([]common.Input, 0),
		Outputs:   make([]common.Output, 0),
		LockTime:  0,
		Data:      nil,
	}

	// Add input
	tx.Inputs = append(tx.Inputs, common.Input{
		PreviousTxHash:  nil,
		PreviousTxIndex: 0,
		ScriptSig:       nil,
		Sequence:        0xffffffff,
	})

	// Add output
	tx.Outputs = append(tx.Outputs, common.Output{
		Value:        amount,
		ScriptPubKey: nil,
		Address:      to,
	})

	return tx
}

// AssertNoError is a helper to check for errors
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError is a helper to check for expected errors
func AssertError(t *testing.T, err error, expected string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if err.Error() != expected {
		t.Fatalf("expected error %q but got %q", expected, err.Error())
	}
}
