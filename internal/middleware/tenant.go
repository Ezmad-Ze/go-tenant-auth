package middleware

import (
	"net/http"

	"github.com/ezmad/auth-service/internal/repository"
	"github.com/ezmad/auth-service/internal/util"
	"github.com/google/uuid"
)

// TenantMiddleware extracts tenant ID from request
type TenantMiddleware struct {
	tenantRepo      repository.TenantRepository
	headerName      string
	defaultTenantID string
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantRepo repository.TenantRepository, headerName, defaultTenantID string) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo:      tenantRepo,
		headerName:      headerName,
		defaultTenantID: defaultTenantID,
	}
}

// ExtractTenant middleware
func (m *TenantMiddleware) ExtractTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tenantID uuid.UUID
		var err error

		// Try to get tenant ID from header
		tenantIDStr := r.Header.Get(m.headerName)
		if tenantIDStr != "" {
			tenantID, err = uuid.Parse(tenantIDStr)
			if err != nil {
				util.RespondError(w, http.StatusBadRequest, "Invalid tenant ID")
				return
			}

			// Verify tenant exists and is active
			tenant, err := m.tenantRepo.GetByID(r.Context(), tenantID)
			if err != nil || tenant == nil || !tenant.IsActive {
				util.RespondError(w, http.StatusBadRequest, "Invalid or inactive tenant")
				return
			}
		} else {
			// Try to extract from subdomain
			// Implementation would parse Host header

			// Fall back to default tenant
			if m.defaultTenantID != "" {
				tenantID, err = uuid.Parse(m.defaultTenantID)
				if err != nil {
					util.RespondError(w, http.StatusInternalServerError, "Invalid default tenant configuration")
					return
				}
			}
		}

		// Set tenant ID in context
		ctx := util.SetTenantIDInContext(r.Context(), tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
