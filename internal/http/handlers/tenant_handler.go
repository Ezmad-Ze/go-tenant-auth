package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ezmad/auth-service/internal/observability"
	"github.com/ezmad/auth-service/internal/service"
	"github.com/ezmad/auth-service/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TenantHandler struct {
	tenantService service.TenantService
	auditService  observability.AuditService
}

func NewTenantHandler(tenantService service.TenantService, auditService observability.AuditService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		auditService:  auditService,
	}
}

func (h *TenantHandler) RegisterRoutes(r chi.Router) {
	r.Post("/tenants", h.CreateTenant)
	r.Get("/tenants", h.ListTenants)
	r.Get("/tenants/{id}", h.GetTenant)
	r.Put("/tenants/{id}", h.UpdateTenant)
}

// CreateTenant godoc
// @Summary Create tenant
// @Description Create a new tenant with an admin user
// @Tags Tenants
// @Accept json
// @Produce json
// @Param request body CreateTenantRequest true "Tenant details"
// @Success 201 {object} TenantResponse
// @Failure 400 {object} ErrorResponse
// @Router /tenants [post]
func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" {
		util.RespondError(w, http.StatusBadRequest, "Name and slug are required")
		return
	}

	if req.AdminEmail == "" || req.AdminPassword == "" {
		util.RespondError(w, http.StatusBadRequest, "Admin email and password are required")
		return
	}

	result, err := h.tenantService.CreateTenant(r.Context(), req.Name, req.Slug, req.AdminEmail, req.AdminPassword, req.Settings)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	// UserID might be nil if public signup
	userID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	
	// Log Tenant Created
	observability.LogAdminAction(h.auditService, r.Context(), result.Tenant.ID, userID, observability.ActionTenantCreated, "tenant", &result.Tenant.ID, map[string]interface{}{
		"name": result.Tenant.Name,
		"slug": result.Tenant.Slug,
	}, &ipAddr, &userAgent)

	// Log Admin User Created
	observability.LogAdminAction(h.auditService, r.Context(), result.Tenant.ID, userID, observability.ActionUserCreated, "user", &result.AdminUser.ID, map[string]interface{}{
		"email": result.AdminUser.Email,
		"role":  "admin",
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusCreated, result)
}

// GetTenant godoc
// @Summary Get tenant
// @Description Get tenant by ID
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Tenant ID"
// @Success 200 {object} TenantResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /tenants/{id} [get]
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	tenant, err := h.tenantService.GetTenant(r.Context(), tenantID)
	if err != nil {
		util.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	util.RespondJSON(w, http.StatusOK, tenant)
}

// UpdateTenant godoc
// @Summary Update tenant
// @Description Update tenant settings
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Tenant ID"
// @Param request body service.TenantUpdateRequest true "Update details"
// @Success 200 {object} TenantResponse
// @Failure 400 {object} ErrorResponse
// @Router /tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req service.TenantUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tenant, err := h.tenantService.UpdateTenant(r.Context(), tenantID, &req)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	userID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAdminAction(h.auditService, r.Context(), tenantID, userID, observability.ActionTenantUpdated, "tenant", &tenantID, map[string]interface{}{
		"updates": req,
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusOK, tenant)
}

// ListTenants godoc
// @Summary List tenants
// @Description List all tenants
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} TenantListResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants [get]
func (h *TenantHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.tenantService.ListTenants(r.Context())
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"tenants": tenants,
	})
}
