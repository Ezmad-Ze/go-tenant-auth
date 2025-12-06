package middleware

import (
	"net/http"
	"strings"

	"github.com/ezmad/auth-service/internal/service"
	"github.com/ezmad/auth-service/internal/util"
)

// AuthMiddleware verifies access tokens and sets user context
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate middleware for protected routes
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			util.RespondError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			util.RespondError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]

		// Verify token and get user
		user, err := m.authService.VerifyAccessToken(r.Context(), token)
		if err != nil {
			util.RespondError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Set user and tenant in context
		ctx := util.SetUserIDInContext(r.Context(), user.ID)
		ctx = util.SetTenantIDInContext(ctx, user.TenantID)

		// Continue with authenticated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware for routes that work with or without authentication
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				if user, err := m.authService.VerifyAccessToken(r.Context(), parts[1]); err == nil {
					ctx := util.SetUserIDInContext(r.Context(), user.ID)
					ctx = util.SetTenantIDInContext(ctx, user.TenantID)
					r = r.WithContext(ctx)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
