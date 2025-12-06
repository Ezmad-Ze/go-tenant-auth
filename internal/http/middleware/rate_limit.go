package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ezmad/auth-service/internal/service"
	"github.com/ezmad/auth-service/internal/util"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool
	RequestsPerMin int
	Window         time.Duration
	KeyPrefix      string
}

// RateLimitMiddleware handles rate limiting for HTTP requests
type RateLimitMiddleware struct {
	limiter service.RateLimiter
	config  RateLimitConfig
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(limiter service.RateLimiter, config RateLimitConfig) *RateLimitMiddleware {
	// Default window to 1 minute if not specified
	if config.Window == 0 {
		config.Window = time.Minute
	}
	
	// Default key prefix
	if config.KeyPrefix == "" {
		config.KeyPrefix = "ratelimit"
	}

	return &RateLimitMiddleware{
		limiter: limiter,
		config:  config,
	}
}

// Limit is the middleware handler that enforces rate limiting
func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if rate limiting is disabled
		if !m.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract client identifier (IP address or user ID from context)
		identifier := m.getClientIdentifier(r)
		
		// Create rate limit key
		key := fmt.Sprintf("%s:%s:%s", m.config.KeyPrefix, r.URL.Path, identifier)

		// Check rate limit
		result, err := m.limiter.Allow(r.Context(), key, m.config.RequestsPerMin, m.config.Window)
		if err != nil {
			// On error, allow request (fail-open)
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		// Check if request is allowed
		if !result.Allowed {
			// Set Retry-After header
			retryAfterSeconds := int(result.RetryAfter.Seconds())
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))

			// Return 429 Too Many Requests
			util.RespondError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			return
		}

		// Request is allowed, continue
		next.ServeHTTP(w, r)
	})
}

// getClientIdentifier extracts a unique identifier for the client
// Tries user ID from context first, falls back to IP address
func (m *RateLimitMiddleware) getClientIdentifier(r *http.Request) string {
	// Try to get user ID from context (for authenticated requests)
	if userID, ok := util.GetUserIDFromContext(r.Context()); ok {
		return userID.String()
	}

	// Fall back to IP address
	return util.GetIPAddress(r)
}
