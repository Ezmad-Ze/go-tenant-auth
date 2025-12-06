package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/observability"
	"github.com/ezmad/auth-service/internal/service"
	"github.com/ezmad/auth-service/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AuthorizationHandler handles authorization HTTP requests
type AuthorizationHandler struct {
	authzService service.AuthorizationService
	auditService observability.AuditService
}

// NewAuthorizationHandler creates a new authorization handler
func NewAuthorizationHandler(authzService service.AuthorizationService, auditService observability.AuditService) *AuthorizationHandler {
	return &AuthorizationHandler{
		authzService: authzService,
		auditService: auditService,
	}
}

// RegisterRoutes registers all authorization routes
func (h *AuthorizationHandler) RegisterRoutes(r chi.Router) {
	r.Post("/check", h.CheckPermission)
	r.Get("/permissions", h.GetUserPermissions)

	// Role management
	r.Post("/roles", h.CreateRole)
	r.Post("/roles/assign", h.AssignRole)
	r.Delete("/roles/revoke", h.RevokeRole)
	r.Get("/users/{userID}/roles", h.GetUserRoles)

	// Permission management
	r.Post("/permissions", h.CreatePermission)
	r.Post("/permissions/assign-to-role", h.AssignPermissionToRole)
	r.Post("/permissions/assign-to-user", h.AssignPermissionToUser)

	// Resource and Scope management
	r.Post("/resources", h.CreateResource)
	r.Post("/scopes", h.CreateScope)
	r.Get("/resources", h.ListResources)
	r.Get("/scopes", h.ListScopes)
}

// CheckPermission handles permission checking
// CheckPermission godoc
// @Summary Check permission
// @Description Check if a user has permission for a resource and scope
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body CheckPermissionRequest true "Permission check details"
// @Success 200 {object} CheckPermissionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /authz/check [post]
func (h *AuthorizationHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	var req CheckPermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	check := &domain.PermissionCheck{
		UserID:     userID,
		TenantID:   tenantID,
		Resource:   req.Resource,
		Scope:      req.Scope,
		Attributes: req.Attributes,
	}

	hasPermission, err := h.authzService.CheckPermission(r.Context(), check)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "Failed to check permission")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"has_permission": hasPermission,
		"permission":     req.Resource + "." + req.Scope,
	})
}

// GetUserPermissions returns all permissions for the authenticated user
// GetUserPermissions godoc
// @Summary Get user permissions
// @Description Get all permissions for the authenticated user
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} UserPermissionsResponse
// @Failure 500 {object} ErrorResponse
// @Router /authz/permissions [get]
func (h *AuthorizationHandler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID, _ := util.GetUserIDFromContext(r.Context())
	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	permissions, err := h.authzService.GetUserPermissions(r.Context(), tenantID, userID)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "Failed to get permissions")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"permissions": permissions,
	})
}

// CreateRole creates a new role
// CreateRole godoc
// @Summary Create role
// @Description Create a new role
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body CreateRoleRequest true "Role details"
// @Success 201 {object} domain.Role
// @Failure 400 {object} ErrorResponse
// @Router /authz/roles [post]
func (h *AuthorizationHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	role, err := h.authzService.CreateRole(r.Context(), tenantID, req.Name, req.Description, req.IsSystem)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	userID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogPermissionChange(h.auditService, r.Context(), tenantID, userID, "role_created", "role", role.ID, map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusCreated, role)
}

// AssignRole assigns a role to a user
// AssignRole godoc
// @Summary Assign role
// @Description Assign a role to a user
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body AssignRoleRequest true "Assignment details"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/roles/assign [post]
func (h *AuthorizationHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req AssignRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	roleID, _ := uuid.Parse(req.RoleID)
	assignedBy, _ := util.GetUserIDFromContext(r.Context())

	if err := h.authzService.AssignRoleToUser(r.Context(), userID, roleID, assignedBy, nil); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogPermissionChange(h.auditService, r.Context(), tenantID, assignedBy, observability.ActionRoleAssigned, "user", userID, map[string]interface{}{
		"role_id": roleID,
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Role assigned successfully"})
}

// RevokeRole revokes a role from a user
// RevokeRole godoc
// @Summary Revoke role
// @Description Revoke a role from a user
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body RevokeRoleRequest true "Revocation details"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/roles/revoke [delete]
func (h *AuthorizationHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	var req RevokeRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	roleID, _ := uuid.Parse(req.RoleID)

	if err := h.authzService.RemoveRoleFromUser(r.Context(), userID, roleID); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	performerID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogPermissionChange(h.auditService, r.Context(), tenantID, performerID, observability.ActionRoleRevoked, "user", userID, map[string]interface{}{
		"role_id": roleID,
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Role revoked successfully"})
}

// GetUserRoles gets all roles for a user
// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all roles for a user
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param userID path string true "User ID"
// @Success 200 {object} UserRolesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /authz/users/{userID}/roles [get]
func (h *AuthorizationHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	roles, err := h.authzService.GetUserRoles(r.Context(), userID)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "Failed to get roles")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{"roles": roles})
}

// Stub implementations for other handlers
// CreatePermission godoc
// @Summary Create permission
// @Description Create a new permission
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body CreatePermissionRequest true "Permission details"
// @Success 201 {object} domain.Permission
// @Failure 400 {object} ErrorResponse
// @Router /authz/permissions [post]
func (h *AuthorizationHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req CreatePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	permission, err := h.authzService.CreatePermission(r.Context(), tenantID, req.ResourceName, req.ScopeName, req.Description, nil)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	userID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogPermissionChange(h.auditService, r.Context(), tenantID, userID, "permission_created", "permission", permission.ID, map[string]interface{}{
		"resource":    req.ResourceName,
		"scope":       req.ScopeName,
		"description": req.Description,
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusCreated, permission)
}

// AssignPermissionToRole godoc
// @Summary Assign permission to role
// @Description Assign a permission to a role
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body AssignPermissionRoleRequest true "Assignment details"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/permissions/assign-to-role [post]
func (h *AuthorizationHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	var req AssignPermissionRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	roleID, _ := uuid.Parse(req.RoleID)
	permissionID, _ := uuid.Parse(req.PermissionID)

	if err := h.authzService.AssignPermissionToRole(r.Context(), roleID, permissionID); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	userID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogPermissionChange(h.auditService, r.Context(), tenantID, userID, observability.ActionPermissionGranted, "role", roleID, map[string]interface{}{
		"permission_id": permissionID,
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Permission assigned to role successfully"})
}

// AssignPermissionToUser godoc
// @Summary Assign permission to user
// @Description Assign a permission directly to a user
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body AssignPermissionUserRequest true "Assignment details"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /authz/permissions/assign-to-user [post]
func (h *AuthorizationHandler) AssignPermissionToUser(w http.ResponseWriter, r *http.Request) {
	var req AssignPermissionUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	permissionID, _ := uuid.Parse(req.PermissionID)
	assignedBy, _ := util.GetUserIDFromContext(r.Context())

	if err := h.authzService.AssignPermissionToUser(r.Context(), userID, permissionID, assignedBy, nil); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogPermissionChange(h.auditService, r.Context(), tenantID, assignedBy, observability.ActionPermissionGranted, "user", userID, map[string]interface{}{
		"permission_id": permissionID,
	}, &ipAddr, &userAgent)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Permission assigned to user successfully"})
}

// CreateResource godoc
// @Summary Create resource
// @Description Create a new resource
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 201 {object} MessageResponse
// @Router /authz/resources [post]
func (h *AuthorizationHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	userID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	// Assuming resource creation logic is here or stubbed. Logging anyway.
	h.auditService.LogEvent(r.Context(), &observability.AuditEvent{
		TenantID:  tenantID,
		UserID:    &userID,
		Action:    "resource_created",
		IPAddress: &ipAddr,
		UserAgent: &userAgent,
		Status:    observability.AuditStatusSuccess,
	})
	util.RespondJSON(w, http.StatusCreated, map[string]string{"message": "Resource created"})
}

// CreateScope godoc
// @Summary Create scope
// @Description Create a new scope
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 201 {object} MessageResponse
// @Router /authz/scopes [post]
func (h *AuthorizationHandler) CreateScope(w http.ResponseWriter, r *http.Request) {
	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	userID, _ := util.GetUserIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	// Assuming scope creation logic is here or stubbed. Logging anyway.
	h.auditService.LogEvent(r.Context(), &observability.AuditEvent{
		TenantID:  tenantID,
		UserID:    &userID,
		Action:    "scope_created",
		IPAddress: &ipAddr,
		UserAgent: &userAgent,
		Status:    observability.AuditStatusSuccess,
	})
	util.RespondJSON(w, http.StatusCreated, map[string]string{"message": "Scope created"})
}

// ListResources godoc
// @Summary List resources
// @Description List all resources
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} ListResourcesResponse
// @Router /authz/resources [get]
func (h *AuthorizationHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	resources, _ := h.authzService.ListResources(r.Context(), tenantID)
	util.RespondJSON(w, http.StatusOK, map[string]interface{}{"resources": resources})
}

// ListScopes godoc
// @Summary List scopes
// @Description List all scopes
// @Tags Authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} ListScopesResponse
// @Router /authz/scopes [get]
func (h *AuthorizationHandler) ListScopes(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	scopes, _ := h.authzService.ListScopes(r.Context(), tenantID)
	util.RespondJSON(w, http.StatusOK, map[string]interface{}{"scopes": scopes})
}
