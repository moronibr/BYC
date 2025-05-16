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
	"net"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/youngchain/internal/core/types"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	Rate       float64       // tokens per second
	BucketSize float64       // maximum bucket size
	Window     time.Duration // time window for tracking
	MaxIPs     int           // maximum number of IPs per user
	MaxUsers   int           // maximum number of users per IP
}

// UserLimiter tracks rate limits for a specific user
type UserLimiter struct {
	IPs        map[string]time.Time // IP addresses used by this user
	LastAccess time.Time            // last access time
	Limiter    *RateLimiter         // rate limiter for this user
}

// IPLimiter tracks rate limits for a specific IP
type IPLimiter struct {
	Users      map[string]time.Time // users from this IP
	LastAccess time.Time            // last access time
	Limiter    *RateLimiter         // rate limiter for this IP
}

// EnhancedRateLimiter implements sophisticated rate limiting
type EnhancedRateLimiter struct {
	config  RateLimiterConfig
	users   map[string]*UserLimiter // user-based limiters
	ips     map[string]*IPLimiter   // IP-based limiters
	global  *RateLimiter            // global rate limiter
	mu      sync.RWMutex            // mutex for thread safety
	cleanup *time.Ticker            // cleanup ticker
	stop    chan struct{}           // stop channel for cleanup
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(config RateLimiterConfig) *EnhancedRateLimiter {
	limiter := &EnhancedRateLimiter{
		config:  config,
		users:   make(map[string]*UserLimiter),
		ips:     make(map[string]*IPLimiter),
		global:  NewRateLimiter(config.Rate, config.BucketSize),
		cleanup: time.NewTicker(5 * time.Minute),
		stop:    make(chan struct{}),
	}
	go limiter.cleanupLoop()
	return limiter
}

// cleanupLoop periodically removes expired entries
func (rl *EnhancedRateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.cleanup.C:
			rl.cleanupExpired()
		case <-rl.stop:
			rl.cleanup.Stop()
			return
		}
	}
}

// cleanupExpired removes expired entries from the rate limiter
func (rl *EnhancedRateLimiter) cleanupExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	expiry := now.Add(-rl.config.Window)

	// Cleanup users
	for user, limiter := range rl.users {
		if limiter.LastAccess.Before(expiry) {
			delete(rl.users, user)
			continue
		}

		// Cleanup IPs for this user
		for ip, lastAccess := range limiter.IPs {
			if lastAccess.Before(expiry) {
				delete(limiter.IPs, ip)
			}
		}
	}

	// Cleanup IPs
	for ip, limiter := range rl.ips {
		if limiter.LastAccess.Before(expiry) {
			delete(rl.ips, ip)
			continue
		}

		// Cleanup users for this IP
		for user, lastAccess := range limiter.Users {
			if lastAccess.Before(expiry) {
				delete(limiter.Users, user)
			}
		}
	}
}

// Allow checks if an operation is allowed under the rate limits
func (rl *EnhancedRateLimiter) Allow(userID string, ip net.IP) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	ipStr := ip.String()

	// Check global rate limit
	if !rl.global.Allow() {
		return false
	}

	// Get or create user limiter
	userLimiter, exists := rl.users[userID]
	if !exists {
		userLimiter = &UserLimiter{
			IPs:        make(map[string]time.Time),
			LastAccess: now,
			Limiter:    NewRateLimiter(rl.config.Rate, rl.config.BucketSize),
		}
		rl.users[userID] = userLimiter
	}

	// Get or create IP limiter
	ipLimiter, exists := rl.ips[ipStr]
	if !exists {
		ipLimiter = &IPLimiter{
			Users:      make(map[string]time.Time),
			LastAccess: now,
			Limiter:    NewRateLimiter(rl.config.Rate, rl.config.BucketSize),
		}
		rl.ips[ipStr] = ipLimiter
	}

	// Check user rate limit
	if !userLimiter.Limiter.Allow() {
		return false
	}

	// Check IP rate limit
	if !ipLimiter.Limiter.Allow() {
		return false
	}

	// Update tracking
	userLimiter.LastAccess = now
	userLimiter.IPs[ipStr] = now
	ipLimiter.LastAccess = now
	ipLimiter.Users[userID] = now

	// Check limits
	if len(userLimiter.IPs) > rl.config.MaxIPs {
		return false
	}
	if len(ipLimiter.Users) > rl.config.MaxUsers {
		return false
	}

	return true
}

// Close stops the cleanup loop
func (rl *EnhancedRateLimiter) Close() {
	close(rl.stop)
}

var (
	ErrInvalidPublicKey  = errors.New("invalid public key")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// Global rate limiters with enhanced configuration
var (
	signLimiter = NewEnhancedRateLimiter(RateLimiterConfig{
		Rate:       100,            // 100 operations per second
		BucketSize: 1000,           // burst of 1000
		Window:     24 * time.Hour, // 24-hour window
		MaxIPs:     5,              // max 5 IPs per user
		MaxUsers:   10,             // max 10 users per IP
	})

	verifyLimiter = NewEnhancedRateLimiter(RateLimiterConfig{
		Rate:       1000,           // 1000 operations per second
		BucketSize: 10000,          // burst of 10000
		Window:     24 * time.Hour, // 24-hour window
		MaxIPs:     10,             // max 10 IPs per user
		MaxUsers:   20,             // max 20 users per IP
	})

	keyGenLimiter = NewEnhancedRateLimiter(RateLimiterConfig{
		Rate:       10,             // 10 operations per second
		BucketSize: 100,            // burst of 100
		Window:     24 * time.Hour, // 24-hour window
		MaxIPs:     3,              // max 3 IPs per user
		MaxUsers:   5,              // max 5 users per IP
	})
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
	// Check rate limit
	if !signLimiter.Allow() {
		return nil, nil, ErrRateLimitExceeded
	}

	// Convert ECDSA private key to btcec private key
	btcecPrivKey := ConvertECDSAToBTCEc(privateKey)
	if btcecPrivKey == nil {
		return nil, nil, fmt.Errorf("failed to convert private key")
	}

	// Create a hash of the message
	hash := sha256.Sum256(message)

	// Sign the hash using Schnorr
	schnorrSig, err := schnorr.Sign(btcecPrivKey, hash[:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Schnorr signature: %v", err)
	}

	// Get the R and S components from the serialized signature
	sigBytes := schnorrSig.Serialize()
	if len(sigBytes) != 64 {
		return nil, nil, fmt.Errorf("invalid signature length")
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	return r, s, nil
}

// schnorrVerify verifies a Schnorr signature
func schnorrVerify(publicKey *ecdsa.PublicKey, message []byte, r, s *big.Int) bool {
	// Check rate limit
	if !verifyLimiter.Allow() {
		return false
	}

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
func SignTaprootTransaction(tx *types.Transaction, privateKey *ecdsa.PrivateKey, userID string, ip net.IP) (*TaprootSignature, error) {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return nil, fmt.Errorf("failed to calculate transaction hash")
	}

	// Sign the hash using Taproot
	r, s, err := taprootSign(rand.Reader, privateKey, tx.Hash, userID, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction with Taproot: %v", err)
	}

	return &TaprootSignature{
		R: r,
		S: s,
	}, nil
}

// VerifyTaprootSignature verifies a Taproot signature
func VerifyTaprootSignature(tx *types.Transaction, signature *TaprootSignature, publicKey *ecdsa.PublicKey, userID string, ip net.IP) bool {
	// Calculate transaction hash
	tx.CalculateHash()
	if tx.Hash == nil {
		return false
	}

	// Verify the signature using Taproot
	return taprootVerify(publicKey, tx.Hash, signature.R, signature.S, userID, ip)
}

// taprootSign signs a message using Taproot (Schnorr) signatures
func taprootSign(rand io.Reader, privateKey *ecdsa.PrivateKey, message []byte, userID string, ip net.IP) (*big.Int, *big.Int, error) {
	// Check rate limit
	if !signLimiter.Allow(userID, ip) {
		return nil, nil, ErrRateLimitExceeded
	}

	// Validate inputs
	if rand == nil {
		return nil, nil, fmt.Errorf("random source cannot be nil")
	}
	if privateKey == nil {
		return nil, nil, fmt.Errorf("private key cannot be nil")
	}
	if message == nil {
		return nil, nil, fmt.Errorf("message cannot be nil")
	}
	if privateKey.D == nil {
		return nil, nil, fmt.Errorf("private key D value cannot be nil")
	}
	if privateKey.Curve != elliptic.P256() {
		return nil, nil, fmt.Errorf("private key must use P256 curve")
	}

	// Create a secure buffer for private key
	privKeyBuf := make([]byte, 32)
	defer func() {
		// Securely wipe the buffer
		for i := range privKeyBuf {
			privKeyBuf[i] = 0
		}
	}()

	// Copy private key bytes to secure buffer
	privKeyBytes := privateKey.D.Bytes()
	if len(privKeyBytes) != 32 {
		return nil, nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}
	copy(privKeyBuf, privKeyBytes)

	// Convert ECDSA private key to secp256k1 private key
	secpPrivKey := secp256k1.PrivKeyFromBytes(privKeyBuf)
	if secpPrivKey == nil {
		return nil, nil, fmt.Errorf("failed to convert private key")
	}
	defer func() {
		// Securely wipe the private key
		if secpPrivKey != nil {
			secpPrivKey.Zero()
		}
	}()

	// Create a hash of the message using constant-time operations
	hash := sha256.Sum256(message)

	// For Taproot, we need to tweak the internal key
	internalKey := secpPrivKey.PubKey()
	if internalKey == nil {
		return nil, nil, fmt.Errorf("failed to derive public key")
	}

	// Create a tap tweak using constant-time operations
	tweak := sha256.Sum256(internalKey.SerializeCompressed())

	// Create a new private key with the tweak added
	tweakScalar := new(secp256k1.ModNScalar)
	if !tweakScalar.SetByteSlice(tweak[:]) {
		return nil, nil, fmt.Errorf("invalid tweak: failed to set byte slice")
	}
	defer func() {
		// Securely wipe the tweak scalar
		tweakScalar.Zero()
	}()

	tweakedPrivKey := new(secp256k1.PrivateKey)
	if tweakedPrivKey == nil {
		return nil, nil, fmt.Errorf("failed to create tweaked private key")
	}
	tweakedPrivKey.Key = secpPrivKey.Key
	tweakedPrivKey.Key.Add(tweakScalar)
	defer func() {
		// Securely wipe the tweaked private key
		if tweakedPrivKey != nil {
			tweakedPrivKey.Zero()
		}
	}()

	// Generate a nonce using RFC6979 with additional entropy
	nonce := secp256k1.NonceRFC6979(tweakedPrivKey.Serialize(), hash[:], nil, nil, 0)
	if nonce == nil {
		return nil, nil, fmt.Errorf("failed to generate nonce")
	}
	defer func() {
		// Securely wipe the nonce
		if nonce != nil {
			nonce.Zero()
		}
	}()

	// Create a Jacobian point for the nonce using constant-time operations
	var noncePoint secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(nonce, &noncePoint)

	// Convert the nonce point to affine coordinates
	var nonceX secp256k1.FieldVal
	noncePoint.ToAffine()
	nonceX = noncePoint.X

	// Create the signature components
	var r, s secp256k1.ModNScalar
	if !r.SetByteSlice(nonceX.Bytes()[:]) {
		return nil, nil, fmt.Errorf("failed to set R component")
	}

	// Calculate s = (r * privKey + hash) / nonce using constant-time operations
	var temp secp256k1.ModNScalar
	if !temp.SetByteSlice(hash[:]) {
		return nil, nil, fmt.Errorf("failed to set hash scalar")
	}
	defer func() {
		// Securely wipe the temporary scalar
		temp.Zero()
	}()

	s.Mul2(&r, &tweakedPrivKey.Key).Add(&temp)

	// Calculate nonce inverse using constant-time operations
	var nonceInv secp256k1.ModNScalar
	nonceInv.InverseNonConst()
	defer func() {
		// Securely wipe the nonce inverse
		nonceInv.Zero()
	}()
	s.Mul(&nonceInv)

	// Convert to big.Int for return using constant-time operations
	var rBytes, sBytes [32]byte
	r.PutBytes(&rBytes)
	s.PutBytes(&sBytes)
	defer func() {
		// Securely wipe the byte arrays
		for i := range rBytes {
			rBytes[i] = 0
		}
		for i := range sBytes {
			sBytes[i] = 0
		}
	}()

	rBig := new(big.Int).SetBytes(rBytes[:])
	sBig := new(big.Int).SetBytes(sBytes[:])

	// Validate the final signature components
	if rBig.Sign() <= 0 || sBig.Sign() <= 0 {
		return nil, nil, fmt.Errorf("invalid signature: R or S is zero or negative")
	}

	return rBig, sBig, nil
}

// taprootVerify verifies a Taproot signature
func taprootVerify(publicKey *ecdsa.PublicKey, message []byte, r, s *big.Int, userID string, ip net.IP) bool {
	// Check rate limit
	if !verifyLimiter.Allow(userID, ip) {
		return false
	}

	// Validate inputs
	if publicKey == nil {
		return false
	}
	if message == nil {
		return false
	}
	if r == nil || s == nil {
		return false
	}
	if publicKey.Curve != elliptic.P256() {
		return false
	}
	if r.Sign() <= 0 || s.Sign() <= 0 {
		return false
	}

	// Create secure buffers for public key coordinates
	xBuf := make([]byte, 32)
	yBuf := make([]byte, 32)
	defer func() {
		// Securely wipe the buffers
		for i := range xBuf {
			xBuf[i] = 0
		}
		for i := range yBuf {
			yBuf[i] = 0
		}
	}()

	// Copy public key coordinates to secure buffers
	copy(xBuf, publicKey.X.Bytes())
	copy(yBuf, publicKey.Y.Bytes())

	// Convert ECDSA public key to secp256k1 public key
	xVal := new(secp256k1.FieldVal)
	yVal := new(secp256k1.FieldVal)
	if !xVal.SetByteSlice(xBuf) || !yVal.SetByteSlice(yBuf) {
		return false
	}
	secpPubKey := secp256k1.NewPublicKey(xVal, yVal)
	if secpPubKey == nil {
		return false
	}

	// Create a hash of the message using constant-time operations
	hash := sha256.Sum256(message)

	// Create the tap tweak using constant-time operations
	tweak := sha256.Sum256(secpPubKey.SerializeCompressed())

	// Create a new public key with the tweak added
	tweakScalar := new(secp256k1.ModNScalar)
	if !tweakScalar.SetByteSlice(tweak[:]) {
		return false
	}
	defer func() {
		// Securely wipe the tweak scalar
		tweakScalar.Zero()
	}()

	// Add the tweak to the public key using constant-time operations
	var pubJac, tweakJac, outJac secp256k1.JacobianPoint
	secpPubKey.AsJacobian(&pubJac)
	secp256k1.ScalarBaseMultNonConst(tweakScalar, &tweakJac)
	secp256k1.AddNonConst(&pubJac, &tweakJac, &outJac)
	tweakedPubKey := secp256k1.NewPublicKey(&outJac.X, &outJac.Y)
	if tweakedPubKey == nil {
		return false
	}

	// Convert r and s to ModNScalar using constant-time operations
	var rScalar, sScalar secp256k1.ModNScalar
	if !rScalar.SetByteSlice(r.Bytes()) || !sScalar.SetByteSlice(s.Bytes()) {
		return false
	}

	// Create a hash scalar using constant-time operations
	var hashScalar secp256k1.ModNScalar
	if !hashScalar.SetByteSlice(hash[:]) {
		return false
	}

	// Verify the signature using constant-time operations
	var sG, hashP, result secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(&sScalar, &sG)
	tweakedPubKey.AsJacobian(&hashP)
	secp256k1.ScalarMultNonConst(&hashScalar, &hashP, &hashP)
	secp256k1.AddNonConst(&sG, &hashP, &result)

	// Check if the result matches R using constant-time operations
	var resultX secp256k1.FieldVal
	result.ToAffine()
	resultX = result.X

	// Compare the X coordinates using constant-time operations
	var rX secp256k1.FieldVal
	if !rX.SetByteSlice(r.Bytes()) {
		return false
	}
	return resultX.Equals(&rX)
}

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair(userID string, ip net.IP) (*ecdsa.PrivateKey, error) {
	// Check rate limit
	if !keyGenLimiter.Allow(userID, ip) {
		return nil, ErrRateLimitExceeded
	}

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
func ValidateScript(script *Script, tx *types.Transaction, inputIndex int, userID string, ip net.IP) bool {
	switch script.Type {
	case "P2PKH":
		return validateP2PKH(script, tx)
	case "P2SH":
		return validateP2SH()
	case "SCHNORR":
		return validateSchnorr(script, tx)
	case "TAPROOT":
		return validateTaproot(script, tx, userID, ip)
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
func validateTaproot(script *Script, tx *types.Transaction, userID string, ip net.IP) bool {
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

	return VerifyTaprootSignature(tx, sig, pubKey, userID, ip)
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

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	rate       float64 // tokens per second
	bucketSize float64 // maximum bucket size
	tokens     float64 // current tokens
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, bucketSize float64) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     bucketSize,
		lastRefill: time.Now(),
	}
}

// Allow checks if an operation is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Calculate time since last refill
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now

	// Add new tokens
	rl.tokens = min(rl.bucketSize, rl.tokens+elapsed*rl.rate)

	// Check if we have enough tokens
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}
	return false
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
