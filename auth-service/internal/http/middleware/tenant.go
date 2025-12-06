package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ezmad/auth-service/internal/util"
	"github.com/google/uuid"
)

// TenantMiddleware extracts the tenant ID from the request header
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip tenant middleware for certain routes
		path := r.URL.Path
		if path == "/health" || 
		   strings.HasPrefix(path, "/auth/refresh") ||
		   (path == "/tenants" && r.Method == "POST") {
			next.ServeHTTP(w, r)
			return
		}

		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			util.RespondError(w, http.StatusBadRequest, "X-Tenant-ID header is required")
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			util.RespondError(w, http.StatusBadRequest, "Invalid X-Tenant-ID header")
			return
		}

		ctx := context.WithValue(r.Context(), util.TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
