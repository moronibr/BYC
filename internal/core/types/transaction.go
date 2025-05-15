package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/youngchain/internal/core/coin"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	Version   uint32
	Inputs    []*TxInput
	Outputs   []*TxOutput
	LockTime  uint32
	Hash      []byte
	Timestamp time.Time
	Fee       uint64
	CoinType  coin.CoinType
	Data      []byte
	Witness   [][]byte
}

// TxInput represents a transaction input
type TxInput struct {
	PreviousTxHash  []byte
	PreviousTxIndex uint32
	ScriptSig       []byte
	Sequence        uint32
	Address         string
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value        uint64
	ScriptPubKey []byte
	Address      string
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash    []byte
	TxIndex   uint32
	Value     uint64
	Address   string
	CoinType  coin.CoinType
	IsSpent   bool
	BlockHash []byte
}

// NewTransaction creates a new transaction
func NewTransaction(from, to []byte, amount uint64, data []byte) *Transaction {
	tx := &Transaction{
		Version:   1,
		Timestamp: time.Now(),
		Data:      data,
		Inputs:    make([]*TxInput, 0),
		Outputs:   make([]*TxOutput, 0),
		Witness:   make([][]byte, 0),
		LockTime:  0,
		Fee:       0,
	}

	// Add input
	if from != nil {
		tx.AddInput(&TxInput{
			PreviousTxHash:  make([]byte, 0),
			PreviousTxIndex: 0,
			ScriptSig:       make([]byte, 0),
			Sequence:        0xffffffff,
			Address:         string(from),
		})
	}

	// Add output
	if to != nil {
		tx.AddOutput(&TxOutput{
			Value:        amount,
			ScriptPubKey: make([]byte, 0),
			Address:      string(to),
		})
	}

	tx.CalculateHash()
	return tx
}

// CalculateHash calculates the transaction hash
func (t *Transaction) CalculateHash() {
	// Initialize data slice
	data := make([]byte, 0)

	// Add version
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, t.Version)
	data = append(data, versionBytes...)

	// Add timestamp
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(t.Timestamp.Unix()))
	data = append(data, timestampBytes...)

	// Add inputs
	if t.Inputs != nil {
		for _, input := range t.Inputs {
			if input == nil {
				continue
			}
			// Add previous transaction hash
			if input.PreviousTxHash != nil {
				data = append(data, input.PreviousTxHash...)
			}
			// Add previous transaction index
			indexBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(indexBytes, input.PreviousTxIndex)
			data = append(data, indexBytes...)
			// Add script signature
			if input.ScriptSig != nil {
				data = append(data, input.ScriptSig...)
			}
			// Add sequence
			seqBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(seqBytes, input.Sequence)
			data = append(data, seqBytes...)
		}
	}

	// Add outputs
	if t.Outputs != nil {
		for _, output := range t.Outputs {
			if output == nil {
				continue
			}
			// Add value
			valueBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(valueBytes, output.Value)
			data = append(data, valueBytes...)
			// Add script public key
			if output.ScriptPubKey != nil {
				data = append(data, output.ScriptPubKey...)
			}
		}
	}

	// Add lock time
	lockTimeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lockTimeBytes, t.LockTime)
	data = append(data, lockTimeBytes...)

	// Add fee
	feeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(feeBytes, t.Fee)
	data = append(data, feeBytes...)

	// Add data
	if t.Data != nil {
		data = append(data, t.Data...)
	}

	// Add witness data
	if t.Witness != nil {
		for _, witness := range t.Witness {
			if witness != nil {
				data = append(data, witness...)
			}
		}
	}

	// Calculate hash
	hash := sha256.Sum256(data)
	t.Hash = hash[:]
}

// AddInput adds an input to the transaction
func (t *Transaction) AddInput(input *TxInput) {
	t.Inputs = append(t.Inputs, input)
}

// AddOutput adds an output to the transaction
func (t *Transaction) AddOutput(output *TxOutput) {
	t.Outputs = append(t.Outputs, output)
}

// Encode encodes the transaction to JSON
func (tx *Transaction) Encode() ([]byte, error) {
	return json.Marshal(tx)
}

// Decode decodes the transaction from JSON
func (tx *Transaction) Decode(data []byte) error {
	return json.Unmarshal(data, tx)
}

// Copy creates a deep copy of the transaction
func (t *Transaction) Copy() *Transaction {
	tx := &Transaction{
		Version:   t.Version,
		LockTime:  t.LockTime,
		Hash:      make([]byte, len(t.Hash)),
		Timestamp: t.Timestamp,
		Fee:       t.Fee,
		CoinType:  t.CoinType,
		Data:      make([]byte, len(t.Data)),
		Witness:   make([][]byte, len(t.Witness)),
	}

	copy(tx.Hash, t.Hash)
	copy(tx.Data, t.Data)

	for i, w := range t.Witness {
		tx.Witness[i] = make([]byte, len(w))
		copy(tx.Witness[i], w)
	}

	tx.Inputs = make([]*TxInput, len(t.Inputs))
	for i, input := range t.Inputs {
		tx.Inputs[i] = &TxInput{
			PreviousTxHash:  make([]byte, len(input.PreviousTxHash)),
			PreviousTxIndex: input.PreviousTxIndex,
			ScriptSig:       make([]byte, len(input.ScriptSig)),
			Sequence:        input.Sequence,
			Address:         input.Address,
		}
		copy(tx.Inputs[i].PreviousTxHash, input.PreviousTxHash)
		copy(tx.Inputs[i].ScriptSig, input.ScriptSig)
	}

	tx.Outputs = make([]*TxOutput, len(t.Outputs))
	for i, output := range t.Outputs {
		tx.Outputs[i] = &TxOutput{
			Value:        output.Value,
			ScriptPubKey: make([]byte, len(output.ScriptPubKey)),
			Address:      output.Address,
		}
		copy(tx.Outputs[i].ScriptPubKey, output.ScriptPubKey)
	}

	return tx
}

// Size returns the size of the transaction in bytes
func (t *Transaction) Size() int {
	size := 4 // Version
	size += 8 // Timestamp
	size += 1 // Input count
	for _, input := range t.Inputs {
		size += 32 // PreviousTxHash
		size += 4  // PreviousTxIndex
		size += len(input.ScriptSig)
		size += 4 // Sequence
	}
	size += 1 // Output count
	for _, output := range t.Outputs {
		size += 8 // Value
		size += len(output.ScriptPubKey)
	}
	size += 4 // LockTime
	size += 8 // Fee
	return size
}

// Validate validates the transaction
func (t *Transaction) Validate() error {
	// TODO: Implement transaction validation
	return nil
}

// Serialize serializes the transaction to bytes
func (tx *Transaction) Serialize() []byte {
	buf := new(bytes.Buffer)

	// Write version
	binary.Write(buf, binary.BigEndian, tx.Version)

	// Write timestamp
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(tx.Timestamp.Unix()))
	buf.Write(timestampBytes)

	// Write number of inputs
	binary.Write(buf, binary.BigEndian, uint32(len(tx.Inputs)))

	// Write inputs
	for _, input := range tx.Inputs {
		// Write previous tx hash
		binary.Write(buf, binary.BigEndian, input.PreviousTxHash)

		// Write previous tx index
		binary.Write(buf, binary.BigEndian, input.PreviousTxIndex)

		// Write script sig length and script sig
		binary.Write(buf, binary.BigEndian, uint32(len(input.ScriptSig)))
		buf.Write(input.ScriptSig)

		// Write sequence
		binary.Write(buf, binary.BigEndian, input.Sequence)
	}

	// Write number of outputs
	binary.Write(buf, binary.BigEndian, uint32(len(tx.Outputs)))

	// Write outputs
	for _, output := range tx.Outputs {
		// Write value
		binary.Write(buf, binary.BigEndian, output.Value)

		// Write script pub key length and script pub key
		binary.Write(buf, binary.BigEndian, uint32(len(output.ScriptPubKey)))
		buf.Write(output.ScriptPubKey)
	}

	// Write lock time
	binary.Write(buf, binary.BigEndian, tx.LockTime)

	// Write fee
	binary.Write(buf, binary.BigEndian, tx.Fee)

	// Write coin type length and coin type
	binary.Write(buf, binary.BigEndian, uint32(len(tx.CoinType)))
	buf.WriteString(string(tx.CoinType))

	// Write data length and data
	binary.Write(buf, binary.BigEndian, uint32(len(tx.Data)))
	buf.Write(tx.Data)

	return buf.Bytes()
}

// Deserialize deserializes a transaction from a buffer
func (tx *Transaction) Deserialize(buf *bytes.Buffer) error {
	// Read version
	if err := binary.Read(buf, binary.BigEndian, &tx.Version); err != nil {
		return err
	}

	// Read number of inputs
	var numInputs uint32
	if err := binary.Read(buf, binary.BigEndian, &numInputs); err != nil {
		return err
	}

	// Read inputs
	tx.Inputs = make([]*TxInput, numInputs)
	for i := uint32(0); i < numInputs; i++ {
		input := &TxInput{}

		// Read previous tx hash
		if err := binary.Read(buf, binary.BigEndian, &input.PreviousTxHash); err != nil {
			return err
		}

		// Read previous tx index
		if err := binary.Read(buf, binary.BigEndian, &input.PreviousTxIndex); err != nil {
			return err
		}

		// Read script sig length and script sig
		var scriptLen uint32
		if err := binary.Read(buf, binary.BigEndian, &scriptLen); err != nil {
			return err
		}
		input.ScriptSig = make([]byte, scriptLen)
		if _, err := buf.Read(input.ScriptSig); err != nil {
			return err
		}

		// Read sequence
		if err := binary.Read(buf, binary.BigEndian, &input.Sequence); err != nil {
			return err
		}

		tx.Inputs[i] = input
	}

	// Read number of outputs
	var numOutputs uint32
	if err := binary.Read(buf, binary.BigEndian, &numOutputs); err != nil {
		return err
	}

	// Read outputs
	tx.Outputs = make([]*TxOutput, numOutputs)
	for i := uint32(0); i < numOutputs; i++ {
		output := &TxOutput{}

		// Read value
		if err := binary.Read(buf, binary.BigEndian, &output.Value); err != nil {
			return err
		}

		// Read script pub key length and script pub key
		var scriptLen uint32
		if err := binary.Read(buf, binary.BigEndian, &scriptLen); err != nil {
			return err
		}
		output.ScriptPubKey = make([]byte, scriptLen)
		if _, err := buf.Read(output.ScriptPubKey); err != nil {
			return err
		}

		tx.Outputs[i] = output
	}

	// Read lock time
	if err := binary.Read(buf, binary.BigEndian, &tx.LockTime); err != nil {
		return err
	}

	// Read fee
	if err := binary.Read(buf, binary.BigEndian, &tx.Fee); err != nil {
		return err
	}

	// Read coin type length and coin type
	var coinTypeLen uint32
	if err := binary.Read(buf, binary.BigEndian, &coinTypeLen); err != nil {
		return err
	}
	coinTypeBytes := make([]byte, coinTypeLen)
	if _, err := buf.Read(coinTypeBytes); err != nil {
		return err
	}
	tx.CoinType = coin.CoinType(coinTypeBytes)

	// Read data length and data
	var dataLen uint32
	if err := binary.Read(buf, binary.BigEndian, &dataLen); err != nil {
		return err
	}
	tx.Data = make([]byte, dataLen)
	if _, err := buf.Read(tx.Data); err != nil {
		return err
	}

	// Calculate hash
	tx.CalculateHash()

	return nil
}

// IsCoinbase returns whether the transaction is a coinbase transaction
func (t *Transaction) IsCoinbase() bool {
	return len(t.Inputs) == 1 && len(t.Inputs[0].PreviousTxHash) == 0 && t.Inputs[0].PreviousTxIndex == 0xFFFFFFFF
}

// MarshalJSON implements json.Marshaler
func (t *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		Hash string `json:"hash"`
		*Alias
	}{
		Hash:  hex.EncodeToString(t.Hash),
		Alias: (*Alias)(t),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		Hash string `json:"hash"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Hash != "" {
		hash, err := hex.DecodeString(aux.Hash)
		if err != nil {
			return err
		}
		t.Hash = hash
	}
	return nil
}

// Weight returns the weight of the transaction
func (tx *Transaction) Weight() int {
	// Base size
	baseSize := tx.Size()

	// Total size
	totalSize := baseSize

	// Witness size (if any)
	witnessSize := 0
	for _, input := range tx.Inputs {
		witnessSize += len(input.ScriptSig)
	}

	// Weight = base size * 3 + total size
	return baseSize*3 + totalSize + witnessSize
}
