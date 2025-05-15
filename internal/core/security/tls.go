package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
)

var (
	tlsConfig     *tls.Config
	tlsConfigOnce sync.Once
)

// TLSConfig holds the TLS configuration
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	CAFile     string
	MinVersion uint16
}

// NewTLSConfig creates a new TLS configuration
func NewTLSConfig(certFile, keyFile, caFile string) (*TLSConfig, error) {
	if _, err := os.Stat(certFile); err != nil {
		return nil, fmt.Errorf("certificate file not found: %v", err)
	}
	if _, err := os.Stat(keyFile); err != nil {
		return nil, fmt.Errorf("key file not found: %v", err)
	}
	if _, err := os.Stat(caFile); err != nil {
		return nil, fmt.Errorf("CA file not found: %v", err)
	}

	return &TLSConfig{
		CertFile:   certFile,
		KeyFile:    keyFile,
		CAFile:     caFile,
		MinVersion: tls.VersionTLS12,
	}, nil
}

// GetTLSConfig returns a singleton TLS configuration
func GetTLSConfig(config *TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfigOnce.Do(func() {
		// Load certificate
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return
		}

		// Load CA certificate
		caCert, err := os.ReadFile(config.CAFile)
		if err != nil {
			return
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			err = fmt.Errorf("failed to append CA certificate")
			return
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			MinVersion:   config.MinVersion,
			// Enable modern cipher suites
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
			// Enable perfect forward secrecy
			PreferServerCipherSuites: true,
			// Enable session tickets
			SessionTicketsDisabled: false,
			// Enable client authentication
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
	})

	return tlsConfig, err
}
