package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moroni/BYC/internal/crypto"
)

// CoinType represents the different types of coins in the system
type CoinType string

const (
	// Golden Block Coins
	Leah    CoinType = "LEAH"
	Shiblum CoinType = "SHIBLUM"
	Shiblon CoinType = "SHIBLON"
	Senine  CoinType = "SENINE"
	Seon    CoinType = "SEON"
	Shum    CoinType = "SHUM"
	Limnah  CoinType = "LIMNAH"
	Antion  CoinType = "ANTION"

	// Silver Block Coins
	Senum CoinType = "SENUM"
	Amnor CoinType = "AMNOR"
	Ezrom CoinType = "EZROM"
	Onti  CoinType = "ONTI"

	// Special Coins
	Ephraim  CoinType = "EPHRAIM"
	Manasseh CoinType = "MANASSEH"
	Joseph   CoinType = "JOSEPH"
)

// BlockType represents which blockchain a block belongs to
type BlockType string

const (
	GoldenBlock BlockType = "GOLDEN"
	SilverBlock BlockType = "SILVER"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp    int64
	Transactions []Transaction
	PrevHash     []byte
	Hash         []byte
	Nonce        uint64
	BlockType    BlockType
	Difficulty   int
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	ID        []byte
	Inputs    []TxInput
	Outputs   []TxOutput
	Timestamp time.Time
	BlockType BlockType
}

// TxInput represents a transaction input
type TxInput struct {
	TxID        []byte
	OutputIndex int
	Amount      float64
	Signature   []byte
	PublicKey   []byte
	Address     string
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value         float64
	CoinType      CoinType
	PublicKeyHash []byte
	Address       string
}

// Wallet represents a user's wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

// Node represents a node in the P2P network
type Node struct {
	Address    string
	Peers      []string
	BlockType  BlockType
	IsMining   bool
	MiningCoin CoinType
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount float64, coinType CoinType, inputs []TxInput, outputs []TxOutput) *Transaction {
	tx := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now(),
		BlockType: GetBlockType(coinType),
	}

	// Calculate transaction ID
	tx.ID = tx.CalculateHash()

	return tx
}

// CalculateHash calculates the hash of a transaction
func (tx *Transaction) CalculateHash() []byte {
	// Create a copy of the transaction without signatures
	txCopy := *tx
	txCopy.Inputs = make([]TxInput, len(tx.Inputs))
	copy(txCopy.Inputs, tx.Inputs)

	// Clear signatures and public keys
	for i := range txCopy.Inputs {
		txCopy.Inputs[i].Signature = nil
		txCopy.Inputs[i].PublicKey = nil
	}

	// Convert the transaction to bytes
	data, err := json.Marshal(txCopy)
	if err != nil {
		return nil
	}

	// Calculate the hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// Sign signs a transaction with the given private key
func (tx *Transaction) Sign(privateKey []byte) error {
	txCopy := tx.TrimmedCopy()

	for i, input := range txCopy.Inputs {
		// Set the public key for this input
		txCopy.Inputs[i].PublicKey = input.PublicKey

		// Calculate the hash of the transaction
		hash := txCopy.CalculateHash()

		// Sign the hash with the private key
		signature, err := crypto.Sign(hash, privateKey)
		if err != nil {
			return err
		}

		// Set the signature for this input
		tx.Inputs[i].Signature = signature
	}

	return nil
}

// IsCoinbase checks if a transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TxID) == 0 && tx.Inputs[0].OutputIndex == -1
}

// GetTotalInput returns the total input amount
func (tx *Transaction) GetTotalInput() float64 {
	var total float64
	for _, input := range tx.Inputs {
		total += input.Amount
	}
	return total
}

// GetTotalOutput returns the total output amount
func (tx *Transaction) GetTotalOutput() float64 {
	var total float64
	for _, output := range tx.Outputs {
		total += output.Value
	}
	return total
}

// GetFee returns the transaction fee
func (tx *Transaction) GetFee() float64 {
	return tx.GetTotalInput() - tx.GetTotalOutput()
}

// MiningDifficulty returns the difficulty multiplier for a given coin type
func MiningDifficulty(coinType CoinType) int {
	switch coinType {
	case Leah:
		return 1
	case Shiblum:
		return 3
	case Shiblon:
		return 5
	default:
		return 0 // Not mineable
	}
}

// IsMineable checks if a coin type is mineable
func IsMineable(coinType CoinType) bool {
	switch coinType {
	case Leah, Shiblum, Shiblon, Senine, Seon, Shum, Limnah, Antion, Senum, Amnor, Ezrom, Onti:
		return true
	default:
		return false
	}
}

// CanTransferBetweenBlocks checks if a coin type can be transferred between blocks
func CanTransferBetweenBlocks(coinType CoinType) bool {
	switch coinType {
	case Antion, Senum, Amnor, Ezrom, Onti:
		return true
	default:
		return false
	}
}

// GetBlockType returns the block type for a coin type
func GetBlockType(coinType CoinType) BlockType {
	switch coinType {
	case Leah, Shiblum, Shiblon, Senine, Seon, Shum, Limnah, Antion:
		return GoldenBlock
	case Senum, Amnor, Ezrom, Onti:
		return SilverBlock
	default:
		return ""
	}
}

// TrimmedCopy creates a copy of the transaction without signatures
func (tx *Transaction) TrimmedCopy() *Transaction {
	txCopy := *tx
	txCopy.Inputs = make([]TxInput, len(tx.Inputs))
	copy(txCopy.Inputs, tx.Inputs)

	// Clear signatures and public keys
	for i := range txCopy.Inputs {
		txCopy.Inputs[i].Signature = nil
		txCopy.Inputs[i].PublicKey = nil
	}

	return &txCopy
}

// String returns the string representation of the coin type
func (c CoinType) String() string {
	switch c {
	case Leah:
		return "Leah"
	case Shiblum:
		return "Shiblum"
	case Shiblon:
		return "Shiblon"
	case Senine:
		return "Senine"
	case Seon:
		return "Seon"
	case Shum:
		return "Shum"
	case Limnah:
		return "Limnah"
	case Antion:
		return "Antion"
	case Senum:
		return "Senum"
	case Amnor:
		return "Amnor"
	case Ezrom:
		return "Ezrom"
	case Onti:
		return "Onti"
	case Ephraim:
		return "Ephraim"
	case Manasseh:
		return "Manasseh"
	case Joseph:
		return "Joseph"
	default:
		return "Unknown"
	}
}

// ConvertLeahToShiblum converts Leah to Shiblum (1 Shiblum = 2 Leah)
func ConvertLeahToShiblum(leah float64) float64 {
	return leah / 2
}

// ConvertShiblumToShiblon converts Shiblum to Shiblon (1 Shiblon = 2 Shiblum)
func ConvertShiblumToShiblon(shiblum float64) float64 {
	return shiblum / 2
}

// ConvertShiblonToSenum converts Shiblon to Senum (1 Senum = 2 Shiblon)
func ConvertShiblonToSenum(shiblon float64) float64 {
	return shiblon * 2
}

// ConvertLeahToSenum converts Leah directly to Senum (1 Senum = 8 Leah)
func ConvertLeahToSenum(leah float64) float64 {
	return leah / 8
}

// Gold coin conversions
// ConvertSenineToSeon converts Senine to Seon (1 Seon = 2 Senine)
func ConvertSenineToSeon(senine float64) float64 {
	return senine * 2
}

// ConvertSeonToShum converts Seon to Shum (1 Shum = 2 Seon)
func ConvertSeonToShum(seon float64) float64 {
	return seon * 2
}

// ConvertShumToLimnah converts Shum to Limnah (1 Limnah = 7 Senine)
func ConvertShumToLimnah(shum float64) float64 {
	return shum * 1.75 // Since 1 Shum = 4 Senine, and 1 Limnah = 7 Senine
}

// Silver coin conversions
// ConvertSenumToAmnor converts Senum to Amnor (1 Amnor = 2 Senum)
func ConvertSenumToAmnor(senum float64) float64 {
	return senum * 2
}

// ConvertAmnorToEzrom converts Amnor to Ezrom (1 Ezrom = 4 Senum)
func ConvertAmnorToEzrom(amnor float64) float64 {
	return amnor * 2
}

// ConvertEzromToOnti converts Ezrom to Onti (1 Onti = 7 Senum)
func ConvertEzromToOnti(ezrom float64) float64 {
	return ezrom * 1.75 // Since 1 Ezrom = 4 Senum, and 1 Onti = 7 Senum
}

// Lesser number conversions (already implemented)
// ConvertLeahToShiblum (1 Shiblum = 2 Leah)
// ConvertShiblumToShiblon (1 Shiblon = 2 Shiblum)
// ConvertShiblonToSenum (1 Senum = 2 Shiblon)

// Direct conversions from Leah to higher denominations
// ConvertLeahToShiblon converts Leah to Shiblon (1 Shiblon = 4 Leah)
func ConvertLeahToShiblon(leah float64) float64 {
	return leah / 4
}

// ConvertLeahToSenine converts Leah to Senine (1 Senine = 8 Leah)
func ConvertLeahToSenine(leah float64) float64 {
	return leah / 8
}

// ConvertLeahToSeon converts Leah to Seon (1 Seon = 16 Leah)
func ConvertLeahToSeon(leah float64) float64 {
	return leah / 16
}

// ConvertLeahToShum converts Leah to Shum (1 Shum = 32 Leah)
func ConvertLeahToShum(leah float64) float64 {
	return leah / 32
}

// ConvertLeahToLimnah converts Leah to Limnah (1 Limnah = 56 Leah)
func ConvertLeahToLimnah(leah float64) float64 {
	return leah / 56
}

// ConvertLeahToAntion converts Leah to Antion (1 Antion = 24 Leah)
func ConvertLeahToAntion(leah float64) float64 {
	return leah / 24
}

// Special coin creation requirements using Fibonacci sequence
const (
	// Golden Block requirements (for Ephraim)
	// Fibonacci sequence: 1, 1, 2, 3, 5, 8, 13, 21
	RequiredLeah    float64 = 1  // F(1)
	RequiredShiblum float64 = 1  // F(2)
	RequiredShiblon float64 = 2  // F(3)
	RequiredSenine  float64 = 3  // F(4)
	RequiredSeon    float64 = 5  // F(5)
	RequiredShum    float64 = 8  // F(6)
	RequiredLimnah  float64 = 13 // F(7)
	RequiredAntion  float64 = 21 // F(8)

	// Silver Block requirements (for Manasseh)
	// Fibonacci sequence: 1, 1, 2, 3
	RequiredSenum float64 = 1 // F(1)
	RequiredAmnor float64 = 1 // F(2)
	RequiredEzrom float64 = 2 // F(3)
	RequiredOnti  float64 = 3 // F(4)
)

// CanCreateEphraim checks if the user has enough of each Golden Block coin to create an Ephraim
func CanCreateEphraim(balances map[CoinType]float64) bool {
	return balances[Leah] >= RequiredLeah &&
		balances[Shiblum] >= RequiredShiblum &&
		balances[Shiblon] >= RequiredShiblon &&
		balances[Senine] >= RequiredSenine &&
		balances[Seon] >= RequiredSeon &&
		balances[Shum] >= RequiredShum &&
		balances[Limnah] >= RequiredLimnah &&
		balances[Antion] >= RequiredAntion
}

// CanCreateManasseh checks if the user has enough of each Silver Block coin to create a Manasseh
func CanCreateManasseh(balances map[CoinType]float64) bool {
	return balances[Senum] >= RequiredSenum &&
		balances[Amnor] >= RequiredAmnor &&
		balances[Ezrom] >= RequiredEzrom &&
		balances[Onti] >= RequiredOnti &&
		balances[Antion] >= RequiredAntion
}

// SpecialCoinStats tracks the creation of special coins
type SpecialCoinStats struct {
	EphraimCreated  int64
	ManassehCreated int64
	LastCreated     time.Time
}

// CalculateTotalValueInLeah calculates the total value of all coins in terms of Leah
func CalculateTotalValueInLeah(balances map[CoinType]float64) float64 {
	var total float64

	// Golden Block coins
	total += balances[Leah] * 1    // 1 Leah = 1 Leah
	total += balances[Shiblum] * 2 // 1 Shiblum = 2 Leah
	total += balances[Shiblon] * 4 // 1 Shiblon = 4 Leah
	total += balances[Senine] * 8  // 1 Senine = 8 Leah
	total += balances[Seon] * 16   // 1 Seon = 16 Leah
	total += balances[Shum] * 32   // 1 Shum = 32 Leah
	total += balances[Limnah] * 56 // 1 Limnah = 56 Leah
	total += balances[Antion] * 24 // 1 Antion = 24 Leah

	// Silver Block coins (converted to Leah)
	total += balances[Senum] * 8  // 1 Senum = 8 Leah
	total += balances[Amnor] * 16 // 1 Amnor = 16 Leah
	total += balances[Ezrom] * 32 // 1 Ezrom = 32 Leah
	total += balances[Onti] * 56  // 1 Onti = 56 Leah

	// Special coins (using their creation cost in Leah)
	total += balances[Ephraim] * 54  // Sum of all Golden Block requirements
	total += balances[Manasseh] * 28 // Sum of all Silver Block requirements

	return total
}

// CreateEphraim creates an Ephraim coin by consuming Fibonacci amounts of each Golden Block coin
func CreateEphraim(balances map[CoinType]float64, stats *SpecialCoinStats) (bool, map[CoinType]float64) {
	if !CanCreateEphraim(balances) {
		return false, balances
	}

	// Create a copy of the balances to modify
	newBalances := make(map[CoinType]float64)
	for k, v := range balances {
		newBalances[k] = v
	}

	// Deduct Fibonacci amounts of each required coin
	newBalances[Leah] -= RequiredLeah
	newBalances[Shiblum] -= RequiredShiblum
	newBalances[Shiblon] -= RequiredShiblon
	newBalances[Senine] -= RequiredSenine
	newBalances[Seon] -= RequiredSeon
	newBalances[Shum] -= RequiredShum
	newBalances[Limnah] -= RequiredLimnah
	newBalances[Antion] -= RequiredAntion

	// Add one Ephraim
	newBalances[Ephraim]++

	// Update stats
	stats.EphraimCreated++
	stats.LastCreated = time.Now()

	return true, newBalances
}

// CreateManasseh creates a Manasseh coin by consuming Fibonacci amounts of each Silver Block coin
func CreateManasseh(balances map[CoinType]float64, stats *SpecialCoinStats) (bool, map[CoinType]float64) {
	if !CanCreateManasseh(balances) {
		return false, balances
	}

	// Create a copy of the balances to modify
	newBalances := make(map[CoinType]float64)
	for k, v := range balances {
		newBalances[k] = v
	}

	// Deduct Fibonacci amounts of each required coin
	newBalances[Senum] -= RequiredSenum
	newBalances[Amnor] -= RequiredAmnor
	newBalances[Ezrom] -= RequiredEzrom
	newBalances[Onti] -= RequiredOnti
	newBalances[Antion] -= RequiredAntion

	// Add one Manasseh
	newBalances[Manasseh]++

	// Update stats
	stats.ManassehCreated++
	stats.LastCreated = time.Now()

	return true, newBalances
}

// GetSpecialCoinStats returns the current statistics of special coin creation
func GetSpecialCoinStats(stats *SpecialCoinStats) string {
	return fmt.Sprintf("Ephraim created: %d\nManasseh created: %d\nLast created: %s",
		stats.EphraimCreated,
		stats.ManassehCreated,
		stats.LastCreated.Format(time.RFC3339))
}

// Maximum supply constants
const (
	MaxEphraimSupply  = 15_000_000
	MaxManassehSupply = 15_000_000
	MaxJosephSupply   = 3_000_000
)

// CreateJoseph creates a Joseph coin by combining 1 Ephraim and 1 Manasseh
func CreateJoseph(balances map[CoinType]float64) (bool, error) {
	// Check if we have enough Ephraim and Manasseh
	if balances[Ephraim] < 1 || balances[Manasseh] < 1 {
		return false, fmt.Errorf("need 1 Ephraim and 1 Manasseh to create a Joseph coin")
	}

	// Check if we've reached the maximum Joseph supply
	if balances[Joseph] >= MaxJosephSupply {
		return false, fmt.Errorf("maximum Joseph supply reached")
	}

	// Check if we've reached the maximum Ephraim or Manasseh supply
	if balances[Ephraim] >= MaxEphraimSupply || balances[Manasseh] >= MaxManassehSupply {
		return false, fmt.Errorf("maximum Ephraim or Manasseh supply reached")
	}

	// Create Joseph coin by consuming 1 Ephraim and 1 Manasseh
	balances[Ephraim]--
	balances[Manasseh]--
	balances[Joseph]++

	return true, nil
}

// GetRemainingSupply returns the remaining supply for each special coin
func GetRemainingSupply(balances map[CoinType]float64) map[CoinType]float64 {
	return map[CoinType]float64{
		Ephraim:  MaxEphraimSupply - balances[Ephraim],
		Manasseh: MaxManassehSupply - balances[Manasseh],
		Joseph:   MaxJosephSupply - balances[Joseph],
	}
}

const (
	// MaxBlockSize is the maximum size of a block in bytes
	MaxBlockSize = 1024 * 1024 // 1MB
)

// HasUTXO checks if a UTXO exists in the set
func (utxoSet *UTXOSet) HasUTXO(txID string, outputIndex int) bool {
	utxoSet.mu.RLock()
	defer utxoSet.mu.RUnlock()

	key := fmt.Sprintf("%s:%d", txID, outputIndex)
	_, exists := utxoSet.utxos[key]
	return exists
}

// ProgressTracker tracks progress towards special coin creation
type ProgressTracker struct {
	Required map[CoinType]float64
	Current  map[CoinType]float64
	Progress map[CoinType]float64 // Percentage complete for each coin
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(coinType CoinType) *ProgressTracker {
	pt := &ProgressTracker{
		Required: make(map[CoinType]float64),
		Current:  make(map[CoinType]float64),
		Progress: make(map[CoinType]float64),
	}

	if coinType == Ephraim {
		pt.Required = map[CoinType]float64{
			Leah:    RequiredLeah,
			Shiblum: RequiredShiblum,
			Shiblon: RequiredShiblon,
			Senine:  RequiredSenine,
			Seon:    RequiredSeon,
			Shum:    RequiredShum,
			Limnah:  RequiredLimnah,
			Antion:  RequiredAntion,
		}
	} else if coinType == Manasseh {
		pt.Required = map[CoinType]float64{
			Senum:  RequiredSenum,
			Amnor:  RequiredAmnor,
			Ezrom:  RequiredEzrom,
			Onti:   RequiredOnti,
			Antion: 1, // Special requirement for Manasseh
		}
	}

	return pt
}

// UpdateProgress updates the progress based on current balances
func (pt *ProgressTracker) UpdateProgress(balances map[CoinType]float64) {
	for coinType, required := range pt.Required {
		current := balances[coinType]
		pt.Current[coinType] = current
		pt.Progress[coinType] = (current / required) * 100
	}
}

// GetOverallProgress returns the overall progress percentage
func (pt *ProgressTracker) GetOverallProgress() float64 {
	var totalProgress float64
	var count float64

	for _, progress := range pt.Progress {
		totalProgress += progress
		count++
	}

	if count == 0 {
		return 0
	}

	return totalProgress / count
}

// GetMissingCoins returns a list of coins that are still needed
func (pt *ProgressTracker) GetMissingCoins() map[CoinType]float64 {
	missing := make(map[CoinType]float64)
	for coinType, required := range pt.Required {
		if pt.Current[coinType] < required {
			missing[coinType] = required - pt.Current[coinType]
		}
	}
	return missing
}
