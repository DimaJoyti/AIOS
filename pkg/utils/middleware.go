package utils

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request details
			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      wrapped.statusCode,
				"duration":    duration,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}).Info("HTTP request processed")
		})
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(requestsPerMinute int, burst int) mux.MiddlewareFunc {
	limiter := rate.NewLimiter(rate.Limit(requestsPerMinute)/60, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware handles authentication
func AuthMiddleware(jwtSecret string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health checks and public endpoints
			if isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Validate token (simplified for now)
			token := extractToken(authHeader)
			if !validateToken(token, jwtSecret) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithFields(logrus.Fields{
						"error":  err,
						"method": r.Method,
						"path":   r.URL.Path,
					}).Error("Panic recovered")

					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// isPublicEndpoint checks if an endpoint is public
func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/ready",
		"/version",
		"/metrics",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}

// extractToken extracts the token from Authorization header
func extractToken(authHeader string) string {
	// Expected format: "Bearer <token>"
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// validateToken validates a JWT token (simplified implementation)
func validateToken(token, secret string) bool {
	// TODO: Implement proper JWT validation
	// For now, just check if token is not empty
	return token != ""
}
