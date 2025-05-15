package coin

import (
	"encoding/json"
	"fmt"
)

// Type represents the type of coin in the system
type Type string

const (
	// Golden represents the primary coin type
	Golden Type = "golden"
	// Silver represents the secondary coin type
	Silver Type = "silver"
	// Shiblum represents the tertiary coin type
	Shiblum Type = "shiblum"
	// Shiblon represents the quaternary coin type
	Shiblon Type = "shiblon"
)

// Coin represents a coin with its type and amount
type Coin struct {
	Type   Type  `json:"type"`
	Amount int64 `json:"amount"`
}

// New creates a new coin of the specified type and amount
func New(coinType Type, amount int64) *Coin {
	return &Coin{
		Type:   coinType,
		Amount: amount,
	}
}

// String returns the string representation of the coin
func (c *Coin) String() string {
	return fmt.Sprintf("%d %s", c.Amount, c.Type)
}

// MarshalJSON implements the json.Marshaler interface
func (c *Coin) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type   Type  `json:"type"`
		Amount int64 `json:"amount"`
	}{
		Type:   c.Type,
		Amount: c.Amount,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *Coin) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type   Type  `json:"type"`
		Amount int64 `json:"amount"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Type = aux.Type
	c.Amount = aux.Amount
	return nil
}

// Validate checks if the coin is valid
func (c *Coin) Validate() error {
	if c.Amount <= 0 {
		return fmt.Errorf("invalid coin amount: %d", c.Amount)
	}
	switch c.Type {
	case Golden, Silver, Shiblum, Shiblon:
		return nil
	default:
		return fmt.Errorf("invalid coin type: %s", c.Type)
	}
}

// Conversion rates between coin types
const (
	GoldenToSilverRate   = 2.0 // 1 Golden = 2 Silver
	SilverToShiblumRate  = 2.0 // 1 Silver = 2 Shiblum
	ShiblumToShiblonRate = 2.0 // 1 Shiblum = 2 Shiblon
)

// Convert converts the coin to another type
func (c *Coin) Convert(targetType Type) (*Coin, error) {
	if c.Type == targetType {
		return c, nil
	}

	var rate float64
	switch {
	case c.Type == Golden && targetType == Silver:
		rate = GoldenToSilverRate
	case c.Type == Silver && targetType == Golden:
		rate = 1.0 / GoldenToSilverRate
	case c.Type == Silver && targetType == Shiblum:
		rate = SilverToShiblumRate
	case c.Type == Shiblum && targetType == Silver:
		rate = 1.0 / SilverToShiblumRate
	case c.Type == Shiblum && targetType == Shiblon:
		rate = ShiblumToShiblonRate
	case c.Type == Shiblon && targetType == Shiblum:
		rate = 1.0 / ShiblumToShiblonRate
	default:
		return nil, fmt.Errorf("unsupported conversion from %s to %s", c.Type, targetType)
	}

	convertedAmount := int64(float64(c.Amount) * rate)
	return New(targetType, convertedAmount), nil
}
