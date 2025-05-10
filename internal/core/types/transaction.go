package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"

	"github.com/youngchain/internal/core/coin"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	// Transaction hash
	Hash []byte

	// Transaction version
	Version uint32

	// Transaction timestamp
	Timestamp time.Time

	// Transaction inputs
	Inputs []*Input

	// Transaction outputs
	Outputs []*Output

	// Transaction lock time
	LockTime uint32

	// Transaction fee
	Fee uint64

	// Transaction coin type
	CoinType coin.CoinType

	// Transaction data
	Data []byte
}

// Input represents a transaction input
type Input struct {
	PreviousTxHash  []byte
	PreviousTxIndex uint32
	ScriptSig       []byte
	Sequence        uint32
	Address         string
}

// Output represents a transaction output
type Output struct {
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
		Inputs:    make([]*Input, 0),
		Outputs:   make([]*Output, 0),
		LockTime:  0,
		Fee:       0,
		CoinType:  coin.Leah,
		Data:      data,
	}

	// Add input
	tx.Inputs = append(tx.Inputs, &Input{
		Address:         string(from),
		ScriptSig:       nil,
		Sequence:        0xffffffff,
		PreviousTxHash:  nil,
		PreviousTxIndex: 0,
	})

	// Add output
	tx.Outputs = append(tx.Outputs, &Output{
		Address:      string(to),
		Value:        amount,
		ScriptPubKey: nil,
	})

	// Calculate transaction hash
	tx.Hash = tx.CalculateHash()

	return tx
}

// CalculateHash calculates the transaction hash
func (tx *Transaction) CalculateHash() []byte {
	// Create a buffer to store transaction data
	buf := make([]byte, 0)

	// Add version
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, tx.Version)
	buf = append(buf, versionBytes...)

	// Add inputs
	for _, input := range tx.Inputs {
		buf = append(buf, input.PreviousTxHash...)
		indexBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(indexBytes, input.PreviousTxIndex)
		buf = append(buf, indexBytes...)
		buf = append(buf, input.ScriptSig...)
		seqBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(seqBytes, input.Sequence)
		buf = append(buf, seqBytes...)
	}

	// Add outputs
	for _, output := range tx.Outputs {
		valueBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(valueBytes, output.Value)
		buf = append(buf, valueBytes...)
		buf = append(buf, output.ScriptPubKey...)
	}

	// Add lock time and fee
	lockTimeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lockTimeBytes, tx.LockTime)
	buf = append(buf, lockTimeBytes...)
	feeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(feeBytes, tx.Fee)
	buf = append(buf, feeBytes...)

	// Add coin type
	buf = append(buf, []byte(tx.CoinType)...)

	// Add data
	buf = append(buf, tx.Data...)

	// Calculate hash
	hash := sha256.Sum256(buf)
	return hash[:]
}

// Copy creates a deep copy of the transaction
func (tx *Transaction) Copy() *Transaction {
	txCopy := &Transaction{
		Version:   tx.Version,
		Timestamp: tx.Timestamp,
		Inputs:    make([]*Input, len(tx.Inputs)),
		Outputs:   make([]*Output, len(tx.Outputs)),
		LockTime:  tx.LockTime,
		Fee:       tx.Fee,
		CoinType:  tx.CoinType,
		Data:      make([]byte, len(tx.Data)),
		Hash:      make([]byte, len(tx.Hash)),
	}

	// Copy inputs
	for i, input := range tx.Inputs {
		txCopy.Inputs[i] = &Input{
			PreviousTxHash:  make([]byte, len(input.PreviousTxHash)),
			PreviousTxIndex: input.PreviousTxIndex,
			ScriptSig:       make([]byte, len(input.ScriptSig)),
			Sequence:        input.Sequence,
			Address:         input.Address,
		}
		copy(txCopy.Inputs[i].PreviousTxHash, input.PreviousTxHash)
		copy(txCopy.Inputs[i].ScriptSig, input.ScriptSig)
	}

	// Copy outputs
	for i, output := range tx.Outputs {
		txCopy.Outputs[i] = &Output{
			Value:        output.Value,
			ScriptPubKey: make([]byte, len(output.ScriptPubKey)),
			Address:      output.Address,
		}
		copy(txCopy.Outputs[i].ScriptPubKey, output.ScriptPubKey)
	}

	copy(txCopy.Data, tx.Data)
	copy(txCopy.Hash, tx.Hash)

	return txCopy
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 0

	size += 4 // Version
	size += 8 // Timestamp
	size += 4 // LockTime
	size += 8 // Fee
	size += len(tx.Data)

	// Add size of inputs
	for _, input := range tx.Inputs {
		size += len(input.PreviousTxHash)
		size += 4 // PreviousTxIndex
		size += len(input.ScriptSig)
		size += 4 // Sequence
		size += len(input.Address)
	}

	// Add size of outputs
	for _, output := range tx.Outputs {
		size += 8 // Value
		size += len(output.ScriptPubKey)
		size += len(output.Address)
	}

	// Add size of coin type
	size += len(tx.CoinType)

	return size
}

// Validate validates the transaction
func (tx *Transaction) Validate() error {
	// Validate version
	if tx.Version == 0 {
		return errors.New("invalid version")
	}

	// Validate inputs
	if len(tx.Inputs) == 0 {
		return errors.New("invalid inputs")
	}

	// Validate outputs
	if len(tx.Outputs) == 0 {
		return errors.New("invalid outputs")
	}

	// Validate timestamp
	if tx.Timestamp.IsZero() {
		return errors.New("invalid timestamp")
	}

	// Validate hash
	if len(tx.Hash) == 0 {
		return errors.New("invalid hash")
	}

	// Validate data
	if len(tx.Data) == 0 {
		return errors.New("invalid data")
	}

	// Validate coin type
	if tx.CoinType == "" {
		return errors.New("invalid coin type")
	}

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
		// Write previous tx hash length and hash
		binary.Write(buf, binary.BigEndian, uint32(len(input.PreviousTxHash)))
		buf.Write(input.PreviousTxHash)

		// Write previous tx index
		binary.Write(buf, binary.BigEndian, input.PreviousTxIndex)

		// Write script sig length and script sig
		binary.Write(buf, binary.BigEndian, uint32(len(input.ScriptSig)))
		buf.Write(input.ScriptSig)

		// Write sequence
		binary.Write(buf, binary.BigEndian, input.Sequence)

		// Write address length and address
		binary.Write(buf, binary.BigEndian, uint32(len(input.Address)))
		buf.WriteString(input.Address)
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

		// Write address length and address
		binary.Write(buf, binary.BigEndian, uint32(len(output.Address)))
		buf.WriteString(output.Address)
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

// Deserialize deserializes a transaction from bytes
func (tx *Transaction) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	// Read version
	if err := binary.Read(buf, binary.BigEndian, &tx.Version); err != nil {
		return err
	}

	// Read timestamp
	var timestamp uint64
	if err := binary.Read(buf, binary.BigEndian, &timestamp); err != nil {
		return err
	}
	tx.Timestamp = time.Unix(int64(timestamp), 0)

	// Read number of inputs
	var numInputs uint32
	if err := binary.Read(buf, binary.BigEndian, &numInputs); err != nil {
		return err
	}

	// Read inputs
	tx.Inputs = make([]*Input, numInputs)
	for i := uint32(0); i < numInputs; i++ {
		input := &Input{}

		// Read previous tx hash length and hash
		var hashLen uint32
		if err := binary.Read(buf, binary.BigEndian, &hashLen); err != nil {
			return err
		}
		input.PreviousTxHash = make([]byte, hashLen)
		if _, err := buf.Read(input.PreviousTxHash); err != nil {
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

		// Read address length and address
		var addrLen uint32
		if err := binary.Read(buf, binary.BigEndian, &addrLen); err != nil {
			return err
		}
		addrBytes := make([]byte, addrLen)
		if _, err := buf.Read(addrBytes); err != nil {
			return err
		}
		input.Address = string(addrBytes)

		tx.Inputs[i] = input
	}

	// Read number of outputs
	var numOutputs uint32
	if err := binary.Read(buf, binary.BigEndian, &numOutputs); err != nil {
		return err
	}

	// Read outputs
	tx.Outputs = make([]*Output, numOutputs)
	for i := uint32(0); i < numOutputs; i++ {
		output := &Output{}

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

		// Read address length and address
		var addrLen uint32
		if err := binary.Read(buf, binary.BigEndian, &addrLen); err != nil {
			return err
		}
		addrBytes := make([]byte, addrLen)
		if _, err := buf.Read(addrBytes); err != nil {
			return err
		}
		output.Address = string(addrBytes)

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
	tx.Hash = tx.CalculateHash()

	return nil
}
