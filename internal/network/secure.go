package network

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/moroni/BYC/internal/network/common"
)

// SecureConfig holds configuration for secure networking
type SecureConfig struct {
	CertFile     string
	KeyFile      string
	CAFile       string
	VerifyPeer   bool
	MinVersion   uint16
	CipherSuites []uint16
}

// SecurePeer represents a secure peer connection
type SecurePeer struct {
	*common.Peer
	cert    *x509.Certificate
	privKey *ecdsa.PrivateKey
	conn    *tls.Conn
}

// SecureMessage represents a signed and encrypted message
type SecureMessage struct {
	*common.NetworkMessage
	Signature []byte
	Nonce     []byte
}

// NewSecureConfig creates a new secure configuration
func NewSecureConfig() *SecureConfig {
	return &SecureConfig{
		CertFile:   "cert.pem",
		KeyFile:    "key.pem",
		CAFile:     "ca.pem",
		VerifyPeer: true,
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}
}

// GenerateCertificate generates a self-signed certificate
func GenerateCertificate(host string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"BYC Network"},
			CommonName:   host,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{host},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, privKey, nil
}

// SaveCertificate saves a certificate and private key to files
func SaveCertificate(cert *x509.Certificate, privKey *ecdsa.PrivateKey, certFile, keyFile string) error {
	// Save certificate
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
		return fmt.Errorf("failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("error closing cert.pem: %v", err)
	}

	// Save private key
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to open key.pem for writing: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("error closing key.pem: %v", err)
	}

	return nil
}

// NewSecurePeer creates a new secure peer
func NewSecurePeer(peer *common.Peer) (*SecurePeer, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	return &SecurePeer{
		Peer:    peer,
		privKey: privKey,
	}, nil
}

// SignMessage signs a message using the peer's private key
func (sp *SecurePeer) SignMessage(msg *common.NetworkMessage) (*common.NetworkMessage, error) {
	// TODO: Implement message signing
	return msg, nil
}

// VerifyMessage verifies a signed message
func (sp *SecurePeer) VerifyMessage(msg *common.NetworkMessage) bool {
	// TODO: Implement message verification
	return true
}

// SecureNetworkManager extends NetworkManager with security features
type SecureNetworkManager struct {
	*NetworkManager
	config *SecureConfig
	server net.Listener
}

// NewSecureNetworkManager creates a new secure network manager
func NewSecureNetworkManager(netConfig *common.NetworkConfig, secureConfig *SecureConfig) (*SecureNetworkManager, error) {
	// Convert common.NetworkConfig to NetworkConfig
	config := &NetworkConfig{
		NodeID:         netConfig.NodeID,
		ListenPort:     netConfig.ListenPort,
		MaxPeers:       netConfig.MaxPeers,
		BootstrapPeers: netConfig.BootstrapPeers,
		PingInterval:   netConfig.PingInterval,
		DialTimeout:    netConfig.DialTimeout,
		ReadTimeout:    netConfig.ReadTimeout,
		WriteTimeout:   netConfig.WriteTimeout,
	}

	nm := NewNetworkManager(config)
	return &SecureNetworkManager{
		NetworkManager: nm,
		config:         secureConfig,
	}, nil
}

// Start starts the secure network manager
func (snm *SecureNetworkManager) Start() error {
	// Generate self-signed certificate if not exists
	if err := snm.generateCertificate(); err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}

	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(snm.config.CertFile, snm.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %v", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   snm.config.MinVersion,
		CipherSuites: snm.config.CipherSuites,
	}

	// Start TLS listener
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", snm.NetworkManager.config.ListenPort), tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to start TLS listener: %v", err)
	}

	snm.server = listener
	go snm.acceptConnections()

	return nil
}

// Stop stops the secure network manager
func (snm *SecureNetworkManager) Stop() error {
	if snm.server != nil {
		return snm.server.Close()
	}
	return nil
}

// generateCertificate generates a self-signed certificate
func (snm *SecureNetworkManager) generateCertificate() error {
	// Generate private key
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Write certificate to file
	certOut, err := os.Create(snm.config.CertFile)
	if err != nil {
		return fmt.Errorf("failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("error closing cert.pem: %v", err)
	}

	// Write private key to file
	keyOut, err := os.Create(snm.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to open key.pem for writing: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("error closing key.pem: %v", err)
	}

	return nil
}

// acceptConnections accepts incoming TLS connections
func (snm *SecureNetworkManager) acceptConnections() {
	for {
		conn, err := snm.server.Accept()
		if err != nil {
			// TODO: Handle error
			continue
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			conn.Close()
			continue
		}

		// Handle connection
		go snm.handleConnection(tlsConn)
	}
}

// handleConnection handles a TLS connection
func (snm *SecureNetworkManager) handleConnection(conn *tls.Conn) {
	defer conn.Close()

	// Get peer certificate
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return
	}

	// Create peer
	peer := common.NewPeer(state.PeerCertificates[0].Subject.CommonName, conn.RemoteAddr().String(), 0)
	peer.SetConnection(conn)

	// Add peer
	snm.mu.Lock()
	snm.peers[peer.Address] = &Peer{
		ID:         peer.ID,
		Address:    peer.Address,
		LastSeen:   peer.LastSeen,
		Connection: peer.GetConnection(),
	}
	snm.mu.Unlock()

	// Handle messages
	for {
		var msg common.NetworkMessage
		if err := json.NewDecoder(conn).Decode(&msg); err != nil {
			break
		}

		// Convert common.NetworkMessage to NetworkMessage
		networkMsg := &NetworkMessage{
			Type:      MessageType(msg.Type),
			From:      msg.From,
			To:        msg.To,
			Payload:   msg.Payload,
			Timestamp: msg.Timestamp,
		}

		if err := snm.handleMessage(networkMsg); err != nil {
			// TODO: Handle error
			continue
		}
	}
}

// SendSecureMessage sends a signed and encrypted message
func (snm *SecureNetworkManager) SendSecureMessage(msg *common.NetworkMessage) error {
	peer, ok := snm.peers[msg.To]
	if !ok {
		return fmt.Errorf("peer %s not found", msg.To)
	}

	// Convert Peer to common.Peer
	commonPeer := common.NewPeer(peer.ID, peer.Address, 0)
	commonPeer.SetConnection(peer.Connection)

	securePeer, err := NewSecurePeer(commonPeer)
	if err != nil {
		return fmt.Errorf("failed to create secure peer: %v", err)
	}

	signedMsg, err := securePeer.SignMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to sign message: %v", err)
	}

	// Convert common.NetworkMessage to NetworkMessage
	networkMsg := &NetworkMessage{
		Type:      MessageType(signedMsg.Type),
		From:      signedMsg.From,
		To:        signedMsg.To,
		Payload:   signedMsg.Payload,
		Timestamp: signedMsg.Timestamp,
	}

	return snm.SendMessage(networkMsg)
}

// HandleSecureMessage handles a received secure message
func (snm *SecureNetworkManager) HandleSecureMessage(data []byte) error {
	var smsg SecureMessage
	if err := json.Unmarshal(data, &smsg); err != nil {
		return fmt.Errorf("failed to unmarshal secure message: %v", err)
	}

	peer, ok := snm.peers[smsg.From]
	if !ok {
		return fmt.Errorf("peer %s not found", smsg.From)
	}

	// Convert Peer to common.Peer
	commonPeer := common.NewPeer(peer.ID, peer.Address, 0)
	commonPeer.SetConnection(peer.Connection)

	securePeer, err := NewSecurePeer(commonPeer)
	if err != nil {
		return fmt.Errorf("failed to create secure peer: %v", err)
	}

	// Convert NetworkMessage to common.NetworkMessage
	commonMsg := &common.NetworkMessage{
		Type:      common.MessageType(smsg.Type),
		From:      smsg.From,
		To:        smsg.To,
		Payload:   smsg.Payload,
		Timestamp: smsg.Timestamp,
	}

	if !securePeer.VerifyMessage(commonMsg) {
		return fmt.Errorf("invalid message signature")
	}

	// Convert common.NetworkMessage to NetworkMessage
	networkMsg := &NetworkMessage{
		Type:      MessageType(smsg.Type),
		From:      smsg.From,
		To:        smsg.To,
		Payload:   smsg.Payload,
		Timestamp: smsg.Timestamp,
	}

	return snm.handleMessage(networkMsg)
}
