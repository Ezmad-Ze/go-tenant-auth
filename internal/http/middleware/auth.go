package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ezmad/auth-service/internal/security"
	"github.com/ezmad/auth-service/internal/util"
)

// AuthMiddleware handles authentication via JWT/PASETO tokens
func AuthMiddleware(tokenMaker *security.PasetoToken) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public routes
			path := r.URL.Path
			if path == "/health" || 
			   path == "/auth/login" || 
			   path == "/auth/register" || 
			   path == "/auth/refresh" || 
			   path == "/auth/magic-link/send" || 
			   path == "/auth/magic-link/verify" ||
			   path == "/auth/password/reset" ||
			   path == "/auth/password/confirm-reset" ||
			   path == "/auth/email/verify" ||
			   (path == "/tenants" && r.Method == "POST") {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				util.RespondError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			fields := strings.Fields(authHeader)
			if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
				util.RespondError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			token := fields[1]
			claims, err := tokenMaker.VerifyToken(token)
			if err != nil {
				util.RespondError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), util.UserIDKey, claims.UserID)
			// TenantID might already be set by TenantMiddleware, but token is source of truth for auth
			ctx = context.WithValue(ctx, util.TenantIDKey, claims.TenantID)
			// We could also add roles/email if needed
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
