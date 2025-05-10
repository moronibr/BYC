package testing

import (
	"testing"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/types"
)

// TestBlock creates a test block with given parameters
func TestBlock(t *testing.T, prevHash []byte, timestamp time.Time) *block.Block {
	return &block.Block{
		Header: block.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			Timestamp:     timestamp,
			Difficulty:    1,
		},
		Transactions: []*transaction.Transaction{},
	}
}

// TestTransaction creates a test transaction
func TestTransaction(t *testing.T, from, to string, amount uint64, coinType types.CoinType) *transaction.Transaction {
	return &transaction.Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		CoinType:  coinType,
		Timestamp: time.Now(),
	}
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
