package security

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// TLSServerConfig represents TLS configuration for the server
type TLSServerConfig struct {
	// Certificate configuration
	CertFile   string
	KeyFile    string
	ClientCAs  string
	ServerName string

	// Security settings
	MinVersion   uint16
	MaxVersion   uint16
	CipherSuites []uint16

	// Certificate rotation
	CertRotationInterval time.Duration
	CertRenewalThreshold time.Duration
}

// DefaultTLSServerConfig returns a secure default TLS configuration
func DefaultTLSServerConfig() *TLSServerConfig {
	return &TLSServerConfig{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		CertRotationInterval: 24 * time.Hour,
		CertRenewalThreshold: 12 * time.Hour,
	}
}

// LoadTLSServerConfig loads TLS configuration
func LoadTLSServerConfig(config *TLSServerConfig) (*tls.Config, error) {
	// Load server certificate and private key
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   config.MinVersion,
		MaxVersion:   config.MaxVersion,
		CipherSuites: config.CipherSuites,
		ServerName:   config.ServerName,
	}

	// Set secure defaults
	tlsConfig.PreferServerCipherSuites = true
	tlsConfig.SessionTicketsDisabled = true
	tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(128)

	// Load client CAs if specified
	if config.ClientCAs != "" {
		clientCAs, err := os.ReadFile(config.ClientCAs)
		if err != nil {
			return nil, fmt.Errorf("failed to read client CAs: %v", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(clientCAs) {
			return nil, fmt.Errorf("failed to parse client CAs")
		}

		tlsConfig.ClientCAs = certPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

// ValidateCertificate validates a certificate
func ValidateCertificate(cert *x509.Certificate) error {
	// Check expiration
	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("certificate expired at %v", cert.NotAfter)
	}

	// Check if certificate is about to expire
	if time.Until(cert.NotAfter) < 24*time.Hour {
		return fmt.Errorf("certificate expires soon at %v", cert.NotAfter)
	}

	// Check key usage
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return fmt.Errorf("certificate missing digital signature key usage")
	}

	// Check extended key usage
	hasServerAuth := false
	hasClientAuth := false
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			hasServerAuth = true
		}
		if usage == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
		}
	}
	if !hasServerAuth || !hasClientAuth {
		return fmt.Errorf("certificate missing required extended key usage")
	}

	return nil
}

// GetCertificateExpiry returns the expiry time of a certificate
func GetCertificateExpiry(certFile string) (time.Time, error) {
	cert, err := os.ReadFile(certFile)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read certificate: %v", err)
	}

	block, _ := pem.Decode(cert)
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to decode certificate")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return x509Cert.NotAfter, nil
}

// ShouldRotateCertificate checks if a certificate should be rotated
func ShouldRotateCertificate(certFile string, config *TLSServerConfig) (bool, error) {
	expiry, err := GetCertificateExpiry(certFile)
	if err != nil {
		return false, err
	}

	// Rotate if certificate is about to expire
	return time.Until(expiry) < config.CertRenewalThreshold, nil
}

// GetTLSVersionString returns a string representation of a TLS version
func GetTLSVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

// GetCipherSuiteString returns a string representation of a cipher suite
func GetCipherSuiteString(suite uint16) string {
	switch suite {
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:
		return "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"
	case tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:
		return "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	default:
		return "Unknown"
	}
}
