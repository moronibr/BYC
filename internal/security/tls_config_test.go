package security

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultTLSServerConfig(t *testing.T) {
	config := DefaultTLSServerConfig()

	// Test TLS version settings
	if config.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinVersion to be TLS 1.2, got %v", GetTLSVersionString(config.MinVersion))
	}
	if config.MaxVersion != tls.VersionTLS13 {
		t.Errorf("Expected MaxVersion to be TLS 1.3, got %v", GetTLSVersionString(config.MaxVersion))
	}

	// Test cipher suites
	expectedSuites := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
	if len(config.CipherSuites) != len(expectedSuites) {
		t.Errorf("Expected %d cipher suites, got %d", len(expectedSuites), len(config.CipherSuites))
	}
	for i, suite := range expectedSuites {
		if config.CipherSuites[i] != suite {
			t.Errorf("Expected cipher suite %s, got %s",
				GetCipherSuiteString(suite),
				GetCipherSuiteString(config.CipherSuites[i]))
		}
	}

	// Test certificate rotation settings
	if config.CertRotationInterval != 24*time.Hour {
		t.Errorf("Expected CertRotationInterval to be 24h, got %v", config.CertRotationInterval)
	}
	if config.CertRenewalThreshold != 12*time.Hour {
		t.Errorf("Expected CertRenewalThreshold to be 12h, got %v", config.CertRenewalThreshold)
	}
}

func TestLoadTLSServerConfig(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificates
	certFile := filepath.Join(tempDir, "server.crt")
	keyFile := filepath.Join(tempDir, "server.key")
	caFile := filepath.Join(tempDir, "ca.crt")

	// Run certificate generation script
	cmd := exec.Command("bash", "scripts/generate_certs.sh")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to generate test certificates: %v", err)
	}

	// Test loading TLS config
	config := &TLSServerConfig{
		CertFile:   certFile,
		KeyFile:    keyFile,
		ClientCAs:  caFile,
		ServerName: "localhost",
	}

	tlsConfig, err := LoadTLSServerConfig(config)
	if err != nil {
		t.Fatalf("Failed to load TLS config: %v", err)
	}

	// Verify TLS config
	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("Expected 1 certificate, got %d", len(tlsConfig.Certificates))
	}
	if !tlsConfig.PreferServerCipherSuites {
		t.Error("Expected PreferServerCipherSuites to be true")
	}
	if !tlsConfig.SessionTicketsDisabled {
		t.Error("Expected SessionTicketsDisabled to be true")
	}
	if tlsConfig.ClientSessionCache == nil {
		t.Error("Expected ClientSessionCache to be initialized")
	}
	if tlsConfig.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Error("Expected ClientAuth to be RequireAndVerifyClientCert")
	}
}

func TestValidateCertificate(t *testing.T) {
	// Create test certificate
	cert := &x509.Certificate{
		NotBefore: time.Now().Add(-24 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	// Test valid certificate
	if err := ValidateCertificate(cert); err != nil {
		t.Errorf("Expected valid certificate, got error: %v", err)
	}

	// Test expired certificate
	expiredCert := *cert
	expiredCert.NotAfter = time.Now().Add(-1 * time.Hour)
	if err := ValidateCertificate(&expiredCert); err == nil {
		t.Error("Expected error for expired certificate")
	}

	// Test certificate about to expire
	expiringCert := *cert
	expiringCert.NotAfter = time.Now().Add(12 * time.Hour)
	if err := ValidateCertificate(&expiringCert); err == nil {
		t.Error("Expected error for certificate about to expire")
	}

	// Test certificate missing key usage
	invalidCert := *cert
	invalidCert.KeyUsage = 0
	if err := ValidateCertificate(&invalidCert); err == nil {
		t.Error("Expected error for certificate missing key usage")
	}

	// Test certificate missing extended key usage
	invalidCert = *cert
	invalidCert.ExtKeyUsage = nil
	if err := ValidateCertificate(&invalidCert); err == nil {
		t.Error("Expected error for certificate missing extended key usage")
	}
}

func TestGetCertificateExpiry(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificates
	cmd := exec.Command("bash", "scripts/generate_certs.sh")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to generate test certificates: %v", err)
	}

	// Test getting certificate expiry
	certFile := filepath.Join(tempDir, "server.crt")
	expiry, err := GetCertificateExpiry(certFile)
	if err != nil {
		t.Fatalf("Failed to get certificate expiry: %v", err)
	}

	// Certificate should be valid for 365 days
	expectedExpiry := time.Now().Add(365 * 24 * time.Hour)
	if expiry.Sub(expectedExpiry) > 24*time.Hour {
		t.Errorf("Expected expiry around %v, got %v", expectedExpiry, expiry)
	}
}

func TestShouldRotateCertificate(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificates
	cmd := exec.Command("bash", "scripts/generate_certs.sh")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to generate test certificates: %v", err)
	}

	config := DefaultTLSServerConfig()
	certFile := filepath.Join(tempDir, "server.crt")

	// Test certificate that doesn't need rotation
	shouldRotate, err := ShouldRotateCertificate(certFile, config)
	if err != nil {
		t.Fatalf("Failed to check certificate rotation: %v", err)
	}
	if shouldRotate {
		t.Error("Expected certificate to not need rotation")
	}

	// Test certificate that needs rotation
	config.CertRenewalThreshold = 366 * 24 * time.Hour // Longer than certificate validity
	shouldRotate, err = ShouldRotateCertificate(certFile, config)
	if err != nil {
		t.Fatalf("Failed to check certificate rotation: %v", err)
	}
	if !shouldRotate {
		t.Error("Expected certificate to need rotation")
	}
}

func TestGetTLSVersionString(t *testing.T) {
	tests := []struct {
		version uint16
		want    string
	}{
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0, "Unknown"},
	}

	for _, tt := range tests {
		got := GetTLSVersionString(tt.version)
		if got != tt.want {
			t.Errorf("GetTLSVersionString(%v) = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestGetCipherSuiteString(t *testing.T) {
	tests := []struct {
		suite uint16
		want  string
	}{
		{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
		{tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"},
		{tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"},
		{tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"},
		{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"},
		{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		{0, "Unknown"},
	}

	for _, tt := range tests {
		got := GetCipherSuiteString(tt.suite)
		if got != tt.want {
			t.Errorf("GetCipherSuiteString(%v) = %v, want %v", tt.suite, got, tt.want)
		}
	}
}
