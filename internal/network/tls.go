package network

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
)

// TLSConfig holds TLS configuration
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	CAFile     string
	ServerName string
}

// NewTLSConfig creates a new TLS configuration
func NewTLSConfig(certFile, keyFile, caFile, serverName string) (*TLSConfig, error) {
	return &TLSConfig{
		CertFile:   certFile,
		KeyFile:    keyFile,
		CAFile:     caFile,
		ServerName: serverName,
	}, nil
}

// LoadTLSConfig loads TLS configuration from files
func (c *TLSConfig) LoadTLSConfig() (*tls.Config, error) {
	// Load certificate
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(c.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %v", err)
	}

	// Create certificate pool
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// Create TLS configuration
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   c.ServerName,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

	return config, nil
}

// WrapConn wraps a connection with TLS
func WrapConn(conn net.Conn, config *tls.Config) (*tls.Conn, error) {
	return tls.Client(conn, config), nil
}

// ListenTLS creates a TLS listener
func ListenTLS(addr string, config *tls.Config) (net.Listener, error) {
	return tls.Listen("tcp", addr, config)
}

// DialTLS creates a TLS connection
func DialTLS(addr string, config *tls.Config) (*tls.Conn, error) {
	return tls.Dial("tcp", addr, config)
}

// VerifyPeerCertificate verifies a peer's certificate
func VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(verifiedChains) == 0 {
		return fmt.Errorf("no verified certificate chains")
	}
	return nil
}
