package block

import (
	"time"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxOutput
	IsSpent     bool      `json:"is_spent"`
	IsConfirmed bool      `json:"is_confirmed"`
	CreatedAt   time.Time `json:"created_at"`
	SpentAt     time.Time `json:"spent_at,omitempty"`
}

// NewUTXO creates a new UTXO
func NewUTXO(output TxOutput) *UTXO {
	return &UTXO{
		TxOutput:    output,
		IsSpent:     false,
		IsConfirmed: false,
		CreatedAt:   time.Now(),
	}
}

// IsMature checks if the UTXO is mature
func (u *UTXO) IsMature() bool {
	// A UTXO is mature if it's at least 1 hour old
	return time.Since(u.CreatedAt) >= time.Hour
}

// Spend marks the UTXO as spent
func (u *UTXO) Spend() {
	u.IsSpent = true
	u.SpentAt = time.Now()
}

// Confirm marks the UTXO as confirmed
func (u *UTXO) Confirm() {
	u.IsConfirmed = true
}
