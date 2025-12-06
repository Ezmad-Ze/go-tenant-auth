package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ezmad/auth-service/internal/observability"
	"github.com/ezmad/auth-service/internal/service"
	"github.com/ezmad/auth-service/internal/util"
	"github.com/go-chi/chi/v5"
)

type TOTPHandler struct {
	totpService  service.TOTPService
	auditService observability.AuditService
}

func NewTOTPHandler(totpService service.TOTPService, auditService observability.AuditService) *TOTPHandler {
	return &TOTPHandler{
		totpService:  totpService,
		auditService: auditService,
	}
}

func (h *TOTPHandler) RegisterRoutes(r chi.Router) {
	r.Post("/2fa/setup", h.Setup)
	r.Post("/2fa/verify", h.Verify)
	r.Post("/2fa/enable", h.Enable)
	r.Post("/2fa/disable", h.Disable)
	r.Post("/2fa/backup-codes", h.GenerateBackupCodes)
}

// Setup godoc
// @Summary Setup TOTP 2FA
// @Description Generate TOTP secret and QR code for 2FA setup
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} TOTPSetupResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/2fa/setup [post]
func (h *TOTPHandler) Setup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := util.GetTenantIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "Tenant ID not found")
		return
	}
	userID, ok := util.GetUserIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "User ID not found")
		return
	}

	resp, err := h.totpService.Setup(ctx, tenantID, userID)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.RespondJSON(w, http.StatusOK, resp)
}

// Verify godoc
// @Summary Verify TOTP code
// @Description Verify a TOTP code
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body TOTPVerifyRequest true "TOTP code"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/2fa/verify [post]
func (h *TOTPHandler) Verify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := util.GetTenantIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "Tenant ID not found")
		return
	}
	userID, ok := util.GetUserIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "User ID not found")
		return
	}

	var req TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.totpService.Verify(ctx, tenantID, userID, req.Code); err != nil {
		util.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, ctx, tenantID, &userID, observability.ActionTOTPVerified, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Code verified successfully"})
}

// Enable godoc
// @Summary Enable TOTP 2FA
// @Description Enable TOTP 2FA after successful verification
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body TOTPVerifyRequest true "TOTP code"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/2fa/enable [post]
func (h *TOTPHandler) Enable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := util.GetTenantIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "Tenant ID not found")
		return
	}
	userID, ok := util.GetUserIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "User ID not found")
		return
	}

	var req TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.totpService.Enable(ctx, tenantID, userID, req.Code); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, ctx, tenantID, &userID, observability.ActionTOTPEnabled, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "2FA enabled successfully"})
}

// Disable godoc
// @Summary Disable TOTP 2FA
// @Description Disable TOTP 2FA (requires password)
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body TOTPDisableRequest true "Password"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/2fa/disable [post]
func (h *TOTPHandler) Disable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := util.GetTenantIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "Tenant ID not found")
		return
	}
	userID, ok := util.GetUserIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "User ID not found")
		return
	}

	var req TOTPDisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.totpService.Disable(ctx, tenantID, userID, req.Password); err != nil {
		util.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, ctx, tenantID, &userID, observability.ActionTOTPDisabled, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "2FA disabled successfully"})
}

// GenerateBackupCodes godoc
// @Summary Generate backup codes
// @Description Generate new backup codes for 2FA
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} BackupCodesResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/2fa/backup-codes [post]
func (h *TOTPHandler) GenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := util.GetTenantIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "Tenant ID not found")
		return
	}
	userID, ok := util.GetUserIDFromContext(ctx)
	if !ok {
		util.RespondError(w, http.StatusUnauthorized, "User ID not found")
		return
	}

	codes, err := h.totpService.GenerateBackupCodes(ctx, tenantID, userID)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, ctx, tenantID, &userID, observability.ActionBackupCodeGenerated, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"backup_codes": codes,
	})
}
