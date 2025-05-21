package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byc/internal/security"
)

// SecurityMiddleware provides security features for the API
type SecurityMiddleware struct {
	rateLimiter  *security.RateLimiter
	keyManager   *security.KeyManager
	auditor      *security.SecurityAuditor
	allowedHosts []string
	allowedPaths []string
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(
	rateLimiter *security.RateLimiter,
	keyManager *security.KeyManager,
	auditor *security.SecurityAuditor,
	allowedHosts []string,
	allowedPaths []string,
) *SecurityMiddleware {
	return &SecurityMiddleware{
		rateLimiter:  rateLimiter,
		keyManager:   keyManager,
		auditor:      auditor,
		allowedHosts: allowedHosts,
		allowedPaths: allowedPaths,
	}
}

// Middleware returns a middleware function that applies security measures
func (sm *SecurityMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := getClientIP(r)

		// Check rate limit
		if !sm.rateLimiter.Allow(ip) {
			sm.auditor.LogRateLimit(ip, 100)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Check host
		if !sm.isAllowedHost(r.Host) {
			sm.auditor.LogEvent(security.SecurityEvent{
				EventType:   "unauthorized_host",
				IP:          ip,
				Description: "Unauthorized host access attempt",
				Severity:    security.SeverityHigh,
			})
			http.Error(w, "Unauthorized host", http.StatusForbidden)
			return
		}

		// Check path
		if !sm.isAllowedPath(r.URL.Path) {
			sm.auditor.LogEvent(security.SecurityEvent{
				EventType:   "unauthorized_path",
				IP:          ip,
				Description: "Unauthorized path access attempt",
				Severity:    security.SeverityHigh,
			})
			http.Error(w, "Unauthorized path", http.StatusForbidden)
			return
		}

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Add rate limit headers
		minuteTokens, hourTokens := sm.rateLimiter.GetRemainingTokens(ip)
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", minuteTokens))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		w.Header().Set("X-RateLimit-Hour-Remaining", fmt.Sprintf("%.0f", hourTokens))

		// Process request
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		// Log request
		sm.auditor.LogEvent(security.SecurityEvent{
			EventType:   "request",
			IP:          ip,
			Description: fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			Severity:    security.SeverityLow,
			Details: map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"duration":   duration.String(),
				"user_agent": r.UserAgent(),
			},
		})
	})
}

// isAllowedHost checks if the host is allowed
func (sm *SecurityMiddleware) isAllowedHost(host string) bool {
	for _, allowed := range sm.allowedHosts {
		if host == allowed {
			return true
		}
	}
	return false
}

// isAllowedPath checks if the path is allowed
func (sm *SecurityMiddleware) isAllowedPath(path string) bool {
	for _, allowed := range sm.allowedPaths {
		if strings.HasPrefix(path, allowed) {
			return true
		}
	}
	return false
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Use remote address
	return strings.Split(r.RemoteAddr, ":")[0]
}

// SecurityHeaders adds security headers to the response
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}

// RateLimit adds rate limiting to the request
func RateLimit(rateLimiter *security.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !rateLimiter.Allow(ip) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireHTTPS ensures the request is using HTTPS
func RequireHTTPS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			http.Error(w, "HTTPS required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
