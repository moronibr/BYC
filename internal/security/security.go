package security

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter for API endpoints
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

// GetLimiter gets or creates a limiter for the given key
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// RateLimitMiddleware is a middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get limiter for IP
		limiter := rl.GetLimiter(r.RemoteAddr)

		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	ClientCAs  string
	ServerName string
}

// LoadTLSConfig loads TLS configuration
func LoadTLSConfig(config TLSConfig) (*tls.Config, error) {
	// Load server certificate and private key
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

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

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// InputValidator validates input data
type InputValidator struct {
	maxSize int64
}

// NewInputValidator creates a new input validator
func NewInputValidator(maxSize int64) *InputValidator {
	return &InputValidator{
		maxSize: maxSize,
	}
}

// ValidateJSON validates JSON input
func (v *InputValidator) ValidateJSON(data []byte) error {
	// Check size
	if int64(len(data)) > v.maxSize {
		return fmt.Errorf("input size exceeds maximum allowed size")
	}

	// Validate JSON structure
	var js json.RawMessage
	if err := json.Unmarshal(data, &js); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	return nil
}

// ValidateRequest validates an HTTP request
func (v *InputValidator) ValidateRequest(r *http.Request) error {
	// Check content length
	if r.ContentLength > v.maxSize {
		return fmt.Errorf("request size exceeds maximum allowed size")
	}

	// Check content type
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("unsupported content type")
	}

	return nil
}
