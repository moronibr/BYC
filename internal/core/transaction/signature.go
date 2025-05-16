package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/youngchain/internal/core/types"
)

var (
	ErrInvalidPublicKey = errors.New("invalid public key")
)

// Signature represents a transaction signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// SignTransaction signs a transaction with a private key
func SignTransaction(tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*Signature, error) {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return nil, fmt.Errorf("failed to calculate transaction hash")
	}

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, tx.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return &Signature{
		R: r,
		S: s,
	}, nil
}

// VerifySignature verifies a transaction signature
func VerifySignature(tx *types.Transaction, signature *Signature, publicKey *ecdsa.PublicKey) bool {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return false
	}

	// Verify the signature
	return ecdsa.Verify(publicKey, tx.Hash, signature.R, signature.S)
}

// MultisigSignature represents a multisignature
type MultisigSignature struct {
	Signatures []*Signature
	PublicKeys []*ecdsa.PublicKey
	Threshold  int
}

// SignMultisigTransaction signs a transaction with multiple private keys
func SignMultisigTransaction(tx *types.Transaction, privateKeys []*ecdsa.PrivateKey, publicKeys []*ecdsa.PublicKey, threshold int) (*MultisigSignature, error) {
	if len(privateKeys) < threshold {
		return nil, fmt.Errorf("insufficient private keys for threshold %d", threshold)
	}

	signatures := make([]*Signature, 0, len(privateKeys))
	for _, privateKey := range privateKeys {
		signature, err := SignTransaction(tx, privateKey)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signature)
	}

	return &MultisigSignature{
		Signatures: signatures,
		PublicKeys: publicKeys,
		Threshold:  threshold,
	}, nil
}

// VerifyMultisigSignature verifies a multisignature
func VerifyMultisigSignature(tx *types.Transaction, multisig *MultisigSignature) bool {
	if len(multisig.Signatures) < multisig.Threshold {
		return false
	}

	validSignatures := 0
	for i, signature := range multisig.Signatures {
		if i >= len(multisig.PublicKeys) {
			break
		}
		if VerifySignature(tx, signature, multisig.PublicKeys[i]) {
			validSignatures++
		}
	}

	return validSignatures >= multisig.Threshold
}

// SchnorrSignature represents a Schnorr signature
type SchnorrSignature struct {
	R *big.Int
	S *big.Int
}

// SignSchnorrTransaction signs a transaction using Schnorr signatures
func SignSchnorrTransaction(tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*SchnorrSignature, error) {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return nil, fmt.Errorf("failed to calculate transaction hash")
	}

	// Sign the hash using Schnorr
	r, s, err := schnorrSign(rand.Reader, privateKey, tx.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction with Schnorr: %v", err)
	}

	return &SchnorrSignature{
		R: r,
		S: s,
	}, nil
}

// VerifySchnorrSignature verifies a Schnorr signature
func VerifySchnorrSignature(tx *types.Transaction, signature *SchnorrSignature, publicKey *ecdsa.PublicKey) bool {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return false
	}

	// Verify the signature using Schnorr
	return schnorrVerify(publicKey, tx.Hash, signature.R, signature.S)
}

// schnorrSign signs a message using Schnorr signatures
func schnorrSign(rand io.Reader, privateKey *ecdsa.PrivateKey, message []byte) (*big.Int, *big.Int, error) {
	// Convert ECDSA private key to btcec private key
	btcecPrivKey := ConvertECDSAToBTCEc(privateKey)
	if btcecPrivKey == nil {
		return nil, nil, fmt.Errorf("failed to convert private key")
	}

	// Create a hash of the message
	hash := sha256.Sum256(message)

	// Sign the hash using Schnorr
	sig, err := schnorr.Sign(btcecPrivKey, hash[:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Schnorr signature: %v", err)
	}

	// Get the R and S components from the serialized signature
	sigBytes := sig.Serialize()
	if len(sigBytes) != 64 {
		return nil, nil, fmt.Errorf("invalid signature length")
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	return r, s, nil
}

// schnorrVerify verifies a Schnorr signature
func schnorrVerify(publicKey *ecdsa.PublicKey, message []byte, r, s *big.Int) bool {
	// Convert ECDSA public key to btcec public key
	btcecPubKey := ConvertECDSAPubToBTCEc(publicKey)
	if btcecPubKey == nil {
		return false
	}

	// Create a hash of the message
	hash := sha256.Sum256(message)

	// Create a Schnorr signature from R and S
	sigBytes := append(r.Bytes(), s.Bytes()...)
	sig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return false
	}

	// Verify the signature
	return sig.Verify(hash[:], btcecPubKey)
}

// TaprootSignature represents a Taproot signature
type TaprootSignature struct {
	R *big.Int
	S *big.Int
}

// SignTaprootTransaction signs a transaction using Taproot
func SignTaprootTransaction(tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*TaprootSignature, error) {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return nil, fmt.Errorf("failed to calculate transaction hash")
	}

	// Sign the hash using Taproot
	r, s, err := taprootSign(rand.Reader, privateKey, tx.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction with Taproot: %v", err)
	}

	return &TaprootSignature{
		R: r,
		S: s,
	}, nil
}

// VerifyTaprootSignature verifies a Taproot signature
func VerifyTaprootSignature(tx *types.Transaction, signature *TaprootSignature, publicKey *ecdsa.PublicKey) bool {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return false
	}

	// Verify the signature using Taproot
	return taprootVerify(publicKey, tx.Hash, signature.R, signature.S)
}

// taprootSign signs a message using Taproot (Schnorr) signatures
func taprootSign(rand io.Reader, privateKey *ecdsa.PrivateKey, message []byte) (*big.Int, *big.Int, error) {
	// Convert ECDSA private key to secp256k1 private key
	privKeyBytes := privateKey.D.Bytes()
	secpPrivKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	if secpPrivKey == nil {
		return nil, nil, fmt.Errorf("failed to convert private key")
	}

	// Create a hash of the message
	hash := sha256.Sum256(message)

	// For Taproot, we need to tweak the internal key
	internalKey := secpPrivKey.PubKey()
	// Create a tap tweak (this would normally be derived from the script tree)
	tweak := sha256.Sum256(internalKey.SerializeCompressed())

	// Create a new private key with the tweak added
	tweakScalar := new(secp256k1.ModNScalar)
	if !tweakScalar.SetByteSlice(tweak[:]) {
		return nil, nil, fmt.Errorf("invalid tweak")
	}
	tweakedPrivKey := new(secp256k1.PrivateKey)
	tweakedPrivKey.Key = secpPrivKey.Key
	tweakedPrivKey.Key.Add(tweakScalar)

	// Convert back to btcec for Schnorr signing
	btcecPrivKey, err := btcec.PrivKeyFromBytes(tweakedPrivKey.Serialize())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert to btcec key: %v", err)
	}

	// Sign the hash using Schnorr with the tweaked key
	sig, err := schnorr.Sign(btcecPrivKey, hash[:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Taproot signature: %v", err)
	}

	// Get the R and S components from the serialized signature
	sigBytes := sig.Serialize()
	if len(sigBytes) != 64 {
		return nil, nil, fmt.Errorf("invalid signature length")
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	return r, s, nil
}

// taprootVerify verifies a Taproot signature
func taprootVerify(publicKey *ecdsa.PublicKey, message []byte, r, s *big.Int) bool {
	// Convert ECDSA public key to secp256k1 public key
	xVal := new(secp256k1.FieldVal)
	yVal := new(secp256k1.FieldVal)
	if !xVal.SetByteSlice(publicKey.X.Bytes()) || !yVal.SetByteSlice(publicKey.Y.Bytes()) {
		return false
	}
	secpPubKey := secp256k1.NewPublicKey(xVal, yVal)
	if secpPubKey == nil {
		return false
	}

	// Create a hash of the message
	hash := sha256.Sum256(message)

	// Create the tap tweak (same as in signing)
	tweak := sha256.Sum256(secpPubKey.SerializeCompressed())

	// Create a new public key with the tweak added
	tweakScalar := new(secp256k1.ModNScalar)
	if !tweakScalar.SetByteSlice(tweak[:]) {
		return false
	}

	// Add the tweak to the public key (in Jacobian coordinates)
	var pubJac, tweakJac, outJac secp256k1.JacobianPoint
	secpPubKey.AsJacobian(&pubJac)
	secp256k1.ScalarBaseMultNonConst(tweakScalar, &tweakJac)
	secp256k1.AddNonConst(&pubJac, &tweakJac, &outJac)
	tweakedPubKey := secp256k1.NewPublicKey(&outJac.X, &outJac.Y)

	// Convert to btcec for Schnorr verification
	btcecPubKey, err := btcec.ParsePubKey(tweakedPubKey.SerializeCompressed())
	if err != nil {
		return false
	}

	// Create a Schnorr signature from R and S
	sigBytes := append(r.Bytes(), s.Bytes()...)
	sig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return false
	}

	// Verify the signature using the tweaked public key
	return sig.Verify(hash[:], btcecPubKey)
}

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// PublicKeyToAddress converts a public key to an address
func PublicKeyToAddress(publicKey *ecdsa.PublicKey) string {
	// Serialize public key
	pubKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// Hash the public key
	hash := sha256.Sum256(pubKeyBytes)

	// Take the last 20 bytes as the address
	address := hash[len(hash)-20:]

	// Convert to hex string
	return fmt.Sprintf("0x%x", address)
}

// Script represents a transaction script
type Script struct {
	// Script type (P2PKH, P2SH, etc.)
	Type string
	// Script data
	Data []byte
}

// ValidateScript validates a transaction script
func ValidateScript(script *Script, tx *types.Transaction, inputIndex int) bool {
	switch script.Type {
	case "P2PKH":
		return validateP2PKH(script, tx)
	case "P2SH":
		return validateP2SH()
	case "SCHNORR":
		return validateSchnorr(script, tx)
	case "TAPROOT":
		return validateTaproot(script, tx)
	default:
		return false
	}
}

// validateP2PKH validates a Pay-to-Public-Key-Hash script
func validateP2PKH(script *Script, tx *types.Transaction) bool {
	// Extract public key and signature from script
	if len(script.Data) < 2 {
		return false
	}

	// Extract signature and public key
	sigBytes := script.Data[:len(script.Data)-1]
	pubKeyBytes := script.Data[len(script.Data)-1:]

	// Parse public key
	pubKey, err := parsePublicKey(pubKeyBytes)
	if err != nil {
		return false
	}

	// Parse signature
	sig, err := parseSignature(sigBytes)
	if err != nil {
		return false
	}

	// Verify signature
	return VerifySignature(tx, sig, pubKey)
}

// validateP2SH validates a Pay-to-Script-Hash script
func validateP2SH() bool {
	// TODO: Implement P2SH validation
	return true
}

// validateSchnorr validates a Schnorr script
func validateSchnorr(script *Script, tx *types.Transaction) bool {
	if len(script.Data) < 2 {
		return false
	}

	sigBytes := script.Data[:len(script.Data)-1]
	pubKeyBytes := script.Data[len(script.Data)-1:]

	pubKey, err := parsePublicKey(pubKeyBytes)
	if err != nil {
		return false
	}

	sig, err := parseSchnorrSignature(sigBytes)
	if err != nil {
		return false
	}

	return VerifySchnorrSignature(tx, sig, pubKey)
}

// validateTaproot validates a Taproot script
func validateTaproot(script *Script, tx *types.Transaction) bool {
	if len(script.Data) < 2 {
		return false
	}

	sigBytes := script.Data[:len(script.Data)-1]
	pubKeyBytes := script.Data[len(script.Data)-1:]

	pubKey, err := parsePublicKey(pubKeyBytes)
	if err != nil {
		return false
	}

	sig, err := parseTaprootSignature(sigBytes)
	if err != nil {
		return false
	}

	return VerifyTaprootSignature(tx, sig, pubKey)
}

// parsePublicKey parses a public key from bytes
func parsePublicKey(data []byte) (*ecdsa.PublicKey, error) {
	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, data)
	if x == nil {
		return nil, fmt.Errorf("invalid public key")
	}
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// parseSignature parses a signature from bytes
func parseSignature(data []byte) (*Signature, error) {
	if len(data) != 64 {
		return nil, fmt.Errorf("invalid signature length")
	}
	return &Signature{
		R: new(big.Int).SetBytes(data[:32]),
		S: new(big.Int).SetBytes(data[32:]),
	}, nil
}

// parseSchnorrSignature parses a Schnorr signature from bytes
func parseSchnorrSignature(data []byte) (*SchnorrSignature, error) {
	if len(data) != 64 {
		return nil, fmt.Errorf("invalid Schnorr signature length")
	}
	return &SchnorrSignature{
		R: new(big.Int).SetBytes(data[:32]),
		S: new(big.Int).SetBytes(data[32:]),
	}, nil
}

// parseTaprootSignature parses a Taproot signature from bytes
func parseTaprootSignature(data []byte) (*TaprootSignature, error) {
	if len(data) != 64 {
		return nil, fmt.Errorf("invalid Taproot signature length")
	}
	return &TaprootSignature{
		R: new(big.Int).SetBytes(data[:32]),
		S: new(big.Int).SetBytes(data[32:]),
	}, nil
}

// ConvertECDSAToBTCEc converts an ecdsa.PrivateKey to a btcec.PrivateKey
func ConvertECDSAToBTCEc(privKey *ecdsa.PrivateKey) *btcec.PrivateKey {
	privKeyBytes := privKey.D.Bytes()
	btcecPrivKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)
	return btcecPrivKey
}

// ConvertECDSAPubToBTCEc converts an ecdsa.PublicKey to a btcec.PublicKey
func ConvertECDSAPubToBTCEc(pubKey *ecdsa.PublicKey) *btcec.PublicKey {
	pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	btcecPubKey, _ := btcec.ParsePubKey(pubKeyBytes)
	return btcecPubKey
}

// ConvertBTCEcToECDSA converts a btcec.PrivateKey to an ecdsa.PrivateKey
func ConvertBTCEcToECDSA(privKey *btcec.PrivateKey) *ecdsa.PrivateKey {
	pubKey := privKey.PubKey()
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     pubKey.X(),
			Y:     pubKey.Y(),
		},
		D: new(big.Int).SetBytes(privKey.Serialize()),
	}
}

// ConvertBTCEcPubToECDSA converts a btcec.PublicKey to an ecdsa.PublicKey
func ConvertBTCEcPubToECDSA(pubKey *btcec.PublicKey) *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}
}
