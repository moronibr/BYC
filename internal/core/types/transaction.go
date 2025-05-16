package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"
)

// Transaction represents a transaction in the blockchain
type Transaction struct {
	ID        []byte
	Vin       []TXInput
	Vout      []TXOutput
	Timestamp int64
}

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// TXOutput represents a transaction output
type TXOutput struct {
	Value      int64
	PubKeyHash []byte
}

// NewTransaction creates a new transaction
func NewTransaction(vin []TXInput, vout []TXOutput) *Transaction {
	tx := &Transaction{
		Vin:       vin,
		Vout:      vout,
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.Hash()
	return tx
}

// Hash returns the hash of the transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Serialize serializes the transaction into a byte array
func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer

	// Write timestamp
	binary.Write(&result, binary.LittleEndian, tx.Timestamp)

	// Write inputs
	binary.Write(&result, binary.LittleEndian, int64(len(tx.Vin)))
	for _, vin := range tx.Vin {
		// Write txid
		binary.Write(&result, binary.LittleEndian, int64(len(vin.Txid)))
		result.Write(vin.Txid)

		// Write vout
		binary.Write(&result, binary.LittleEndian, int64(vin.Vout))

		// Write signature
		binary.Write(&result, binary.LittleEndian, int64(len(vin.Signature)))
		result.Write(vin.Signature)

		// Write pubkey
		binary.Write(&result, binary.LittleEndian, int64(len(vin.PubKey)))
		result.Write(vin.PubKey)
	}

	// Write outputs
	binary.Write(&result, binary.LittleEndian, int64(len(tx.Vout)))
	for _, vout := range tx.Vout {
		// Write value
		binary.Write(&result, binary.LittleEndian, vout.Value)

		// Write pubkey hash
		binary.Write(&result, binary.LittleEndian, int64(len(vout.PubKeyHash)))
		result.Write(vout.PubKeyHash)
	}

	return result.Bytes()
}

// DeserializeTransaction deserializes a transaction from a byte array
func DeserializeTransaction(data []byte) (*Transaction, error) {
	tx := &Transaction{}
	reader := bytes.NewReader(data)

	// Read timestamp
	binary.Read(reader, binary.LittleEndian, &tx.Timestamp)

	// Read inputs
	var vinCount int64
	binary.Read(reader, binary.LittleEndian, &vinCount)
	tx.Vin = make([]TXInput, vinCount)
	for i := int64(0); i < vinCount; i++ {
		// Read txid
		var txidLen int64
		binary.Read(reader, binary.LittleEndian, &txidLen)
		tx.Vin[i].Txid = make([]byte, txidLen)
		reader.Read(tx.Vin[i].Txid)

		// Read vout
		var vout int64
		binary.Read(reader, binary.LittleEndian, &vout)
		tx.Vin[i].Vout = int(vout)

		// Read signature
		var sigLen int64
		binary.Read(reader, binary.LittleEndian, &sigLen)
		tx.Vin[i].Signature = make([]byte, sigLen)
		reader.Read(tx.Vin[i].Signature)

		// Read pubkey
		var pubKeyLen int64
		binary.Read(reader, binary.LittleEndian, &pubKeyLen)
		tx.Vin[i].PubKey = make([]byte, pubKeyLen)
		reader.Read(tx.Vin[i].PubKey)
	}

	// Read outputs
	var voutCount int64
	binary.Read(reader, binary.LittleEndian, &voutCount)
	tx.Vout = make([]TXOutput, voutCount)
	for i := int64(0); i < voutCount; i++ {
		// Read value
		binary.Read(reader, binary.LittleEndian, &tx.Vout[i].Value)

		// Read pubkey hash
		var pubKeyHashLen int64
		binary.Read(reader, binary.LittleEndian, &pubKeyHashLen)
		tx.Vout[i].PubKeyHash = make([]byte, pubKeyHashLen)
		reader.Read(tx.Vout[i].PubKeyHash)
	}

	return tx, nil
}

// String returns a string representation of the transaction
func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Transaction %x:", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("  Input %d:", i))
		lines = append(lines, fmt.Sprintf("    TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("    Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("    Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("    PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("  Output %d:", i))
		lines = append(lines, fmt.Sprintf("    Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("    Script: %x", output.PubKeyHash))
	}

	return fmt.Sprintf("%s\n", lines)
}

// Validate validates the transaction
func (tx *Transaction) Validate() error {
	// Check if transaction is nil
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	// Check if transaction has inputs
	if len(tx.Vin) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}

	// Check if transaction has outputs
	if len(tx.Vout) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Check if transaction ID is valid
	if tx.ID == nil {
		return fmt.Errorf("transaction has no ID")
	}

	// Check if transaction timestamp is valid
	if tx.Timestamp > time.Now().Unix() {
		return fmt.Errorf("invalid timestamp")
	}

	// Validate inputs
	for i, input := range tx.Vin {
		if input.Txid == nil {
			return fmt.Errorf("input %d has no transaction ID", i)
		}
		if input.Vout < 0 {
			return fmt.Errorf("input %d has invalid output index", i)
		}
		if input.Signature == nil {
			return fmt.Errorf("input %d has no signature", i)
		}
		if input.PubKey == nil {
			return fmt.Errorf("input %d has no public key", i)
		}
	}

	// Validate outputs
	for i, output := range tx.Vout {
		if output.Value <= 0 {
			return fmt.Errorf("output %d has invalid value", i)
		}
		if output.PubKeyHash == nil {
			return fmt.Errorf("output %d has no public key hash", i)
		}
	}

	return nil
}

// Size returns the size of the transaction in bytes
func (tx *Transaction) Size() int {
	size := 0

	// ID size
	size += len(tx.ID)

	// Timestamp size
	size += 8

	// Input count size
	size += 4

	// Input sizes
	for _, input := range tx.Vin {
		size += len(input.Txid)
		size += 4 // Vout size
		size += len(input.Signature)
		size += len(input.PubKey)
	}

	// Output count size
	size += 4

	// Output sizes
	for _, output := range tx.Vout {
		size += 8 // Value size
		size += len(output.PubKeyHash)
	}

	return size
}

// IsCoinbase checks if the transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}
