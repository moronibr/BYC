package coin

import (
	"time"
)

// CoinType represents the type of coin in the system
type CoinType string

const (
	// Mineable coins (present in both blocks)
	Leah    CoinType = "leah"
	Shiblum CoinType = "shiblum"
	Shiblon CoinType = "shiblon"

	// Gold-based derived units
	Senine CoinType = "senine"
	Seon   CoinType = "seon"
	Shum   CoinType = "shum"
	Limnah CoinType = "limnah"
	Antion CoinType = "antion"

	// Silver-based derived units
	Senum CoinType = "senum"
	Amnor CoinType = "amnor"
	Ezrom CoinType = "ezrom"
	Onti  CoinType = "onti"

	// Special block completion coins
	Ephraim  CoinType = "ephraim"
	Manasseh CoinType = "manasseh"
	Joseph   CoinType = "joseph"
)

// Supply limits
const (
	MaxEphraimSupply  = 11_000_000
	MaxManassehSupply = 11_000_000
	MaxJosephSupply   = 11_000_000
)

// Coin represents a basic coin in the system
type Coin struct {
	Type        CoinType
	Value       uint64
	CreatedAt   time.Time
	BlockHash   []byte
	Transaction []byte
}

// SpecialCoin represents Ephraim, Manasseh, or Joseph coins
type SpecialCoin struct {
	Type         CoinType
	BlockHash    []byte
	CreatedAt    time.Time
	EphraimHash  []byte // For Joseph coins
	ManassehHash []byte // For Joseph coins
}

// SupplyTracker keeps track of coin supply
type SupplyTracker struct {
	EphraimCount  uint64
	ManassehCount uint64
	JosephCount   uint64
}

// NewSupplyTracker creates a new supply tracker
func NewSupplyTracker() *SupplyTracker {
	return &SupplyTracker{
		EphraimCount:  0,
		ManassehCount: 0,
		JosephCount:   0,
	}
}

// CanCreateEphraim checks if we can create more Ephraim coins
func (st *SupplyTracker) CanCreateEphraim() bool {
	return st.EphraimCount < MaxEphraimSupply
}

// CanCreateManasseh checks if we can create more Manasseh coins
func (st *SupplyTracker) CanCreateManasseh() bool {
	return st.ManassehCount < MaxManassehSupply
}

// CanCreateJoseph checks if we can create more Joseph coins
func (st *SupplyTracker) CanCreateJoseph() bool {
	return st.JosephCount < MaxJosephSupply
}
