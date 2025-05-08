package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/youngchain/internal/config"
	"github.com/youngchain/internal/consensus"
	"github.com/youngchain/internal/network/messages"
	"github.com/youngchain/internal/security"
)

func setupTestServer(t *testing.T) (*Server, string, func()) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "server-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Generate test certificates
	cmd := exec.Command("bash", "scripts/generate_certs.sh")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to generate test certificates: %v", err)
	}

	// Create test configuration
	cfg := config.DefaultConfig()
	cfg.ListenAddr = "127.0.0.1:0" // Use random port
	cfg.TLSEnabled = true
	cfg.TLS.CertFile = filepath.Join(tempDir, "server.crt")
	cfg.TLS.KeyFile = filepath.Join(tempDir, "server.key")
	cfg.TLS.ClientCAs = filepath.Join(tempDir, "client-ca.crt")
	cfg.TLS.ServerName = "localhost"

	// Create consensus instance
	cons := consensus.NewConsensus()

	// Create server
	server := NewServer(cfg, cons)

	// Cleanup function
	cleanup := func() {
		server.Stop()
		os.RemoveAll(tempDir)
	}

	return server, tempDir, cleanup
}

func TestServerTLSConnection(t *testing.T) {
	server, tempDir, cleanup := setupTestServer(t)
	defer cleanup()

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Get server address
	addr := server.config.ListenAddr

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(
		filepath.Join(tempDir, "client.crt"),
		filepath.Join(tempDir, "client.key"),
	)
	if err != nil {
		t.Fatalf("Failed to load client certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(filepath.Join(tempDir, "ca.crt"))
	if err != nil {
		t.Fatalf("Failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		t.Fatalf("Failed to parse CA certificate")
	}

	// Create TLS config for client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "localhost",
	}

	// Connect to server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Verify TLS connection state
	state := conn.ConnectionState()
	if state.Version != tls.VersionTLS12 && state.Version != tls.VersionTLS13 {
		t.Errorf("Expected TLS 1.2 or 1.3, got %d", state.Version)
	}

	// Verify cipher suite
	validCipher := false
	for _, suite := range server.tlsConfig.CipherSuites {
		if state.CipherSuite == suite {
			validCipher = true
			break
		}
	}
	if !validCipher {
		t.Errorf("Unexpected cipher suite: %d", state.CipherSuite)
	}
}

func TestServerTLSMessageExchange(t *testing.T) {
	server, tempDir, cleanup := setupTestServer(t)
	defer cleanup()

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Get server address
	addr := server.config.ListenAddr

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(
		filepath.Join(tempDir, "client.crt"),
		filepath.Join(tempDir, "client.key"),
	)
	if err != nil {
		t.Fatalf("Failed to load client certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(filepath.Join(tempDir, "ca.crt"))
	if err != nil {
		t.Fatalf("Failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		t.Fatalf("Failed to parse CA certificate")
	}

	// Create TLS config for client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "localhost",
	}

	// Connect to server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create test message
	msg := messages.Message{
		Type: messages.PingMsg,
		Data: []byte("ping"),
	}

	// Sign message
	signedMsg, err := server.msgSigner.SignMessage(msg)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Marshal message
	msgData, err := json.Marshal(signedMsg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// Send message length
	if err := binary.Write(conn, binary.BigEndian, uint32(len(msgData))); err != nil {
		t.Fatalf("Failed to send message length: %v", err)
	}

	// Send message
	if _, err := conn.Write(msgData); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response length
	var respLen uint32
	if err := binary.Read(conn, binary.BigEndian, &respLen); err != nil {
		t.Fatalf("Failed to read response length: %v", err)
	}

	// Read response
	respData := make([]byte, respLen)
	if _, err := io.ReadFull(conn, respData); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Unmarshal response
	var respMsg security.SignedMessage
	if err := json.Unmarshal(respData, &respMsg); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response
	valid, err := server.msgSigner.VerifyMessage(&respMsg)
	if err != nil {
		t.Fatalf("Failed to verify response: %v", err)
	}
	if !valid {
		t.Fatal("Invalid response signature")
	}
}

func TestServerTLSReconnection(t *testing.T) {
	server, tempDir, cleanup := setupTestServer(t)
	defer cleanup()

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Get server address
	addr := server.config.ListenAddr

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(
		filepath.Join(tempDir, "client.crt"),
		filepath.Join(tempDir, "client.key"),
	)
	if err != nil {
		t.Fatalf("Failed to load client certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(filepath.Join(tempDir, "ca.crt"))
	if err != nil {
		t.Fatalf("Failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		t.Fatalf("Failed to parse CA certificate")
	}

	// Create TLS config for client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "localhost",
	}

	// Test multiple connections
	for i := 0; i < 5; i++ {
		// Connect to server
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			t.Fatalf("Failed to connect to server (attempt %d): %v", i+1, err)
		}

		// Verify connection
		state := conn.ConnectionState()
		if state.Version != tls.VersionTLS12 && state.Version != tls.VersionTLS13 {
			t.Errorf("Expected TLS 1.2 or 1.3, got %d", state.Version)
		}

		// Close connection
		conn.Close()

		// Wait a bit before next connection
		time.Sleep(100 * time.Millisecond)
	}
}

func TestServerTLSInvalidCertificate(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Get server address
	addr := server.config.ListenAddr

	// Create invalid TLS config (no certificates)
	tlsConfig := &tls.Config{
		ServerName: "localhost",
	}

	// Try to connect to server
	_, err := tls.Dial("tcp", addr, tlsConfig)
	if err == nil {
		t.Fatal("Expected connection to fail with invalid certificate")
	}
}

func TestServerTLSExpiredCertificate(t *testing.T) {
	server, tempDir, cleanup := setupTestServer(t)
	defer cleanup()

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Get server address
	addr := server.config.ListenAddr

	// Load expired client certificate
	cert, err := tls.LoadX509KeyPair(
		filepath.Join(tempDir, "client.crt"),
		filepath.Join(tempDir, "client.key"),
	)
	if err != nil {
		t.Fatalf("Failed to load client certificate: %v", err)
	}

	// Modify certificate to be expired
	cert.Leaf.NotAfter = time.Now().Add(-24 * time.Hour)

	// Load CA certificate
	caCert, err := os.ReadFile(filepath.Join(tempDir, "ca.crt"))
	if err != nil {
		t.Fatalf("Failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		t.Fatalf("Failed to parse CA certificate")
	}

	// Create TLS config for client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "localhost",
	}

	// Try to connect to server
	_, err = tls.Dial("tcp", addr, tlsConfig)
	if err == nil {
		t.Fatal("Expected connection to fail with expired certificate")
	}
}
