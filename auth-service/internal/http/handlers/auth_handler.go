package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ezmad/auth-service/internal/observability"
	"github.com/ezmad/auth-service/internal/service"
	"github.com/ezmad/auth-service/internal/util"
	"github.com/go-chi/chi/v5"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService  service.AuthService
	auditService observability.AuditService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService, auditService observability.AuditService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		auditService: auditService,
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/magic-link/send", h.SendMagicLink)
	r.Post("/magic-link/verify", h.VerifyMagicLink)
	r.Post("/password/change", h.ChangePassword)
	r.Post("/password/reset", h.ResetPassword)
	r.Post("/email/send-verification", h.SendVerificationEmail)
	r.Post("/email/verify", h.VerifyEmail)
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user in the tenant
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get tenant ID from context (set by middleware)
	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	// Call service
	tokens, err := h.authService.Register(r.Context(), &service.RegisterRequest{
		TenantID:  tenantID,
		Email:     req.Email,
		Password:  req.Password,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Metadata:  req.Metadata,
	})

	if err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, r.Context(), tenantID, &tokens.UserID, observability.ActionRegister, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"token_type":    "Bearer",
		"user_id":       tokens.UserID,
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return access/refresh tokens. Supports 2FA and backup codes.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param X-Device-Fingerprint header string false "Device Fingerprint"
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	tokens, err := h.authService.Login(r.Context(), &service.LoginRequest{
		TenantID:   tenantID,
		Email:      req.Email,
		Password:   req.Password,
		TOTP:       req.TOTP,
		BackupCode: req.BackupCode,
		DeviceInfo: &service.DeviceInfo{
			Fingerprint: r.Header.Get("X-Device-Fingerprint"),
			UserAgent:   util.StringPtr(r.UserAgent()),
			IPAddress:   util.StringPtr(util.GetIPAddress(r)),
		},
	})

	if err != nil {
		// Audit Log Failure
		ipAddr := util.GetIPAddress(r)
		userAgent := r.UserAgent()
		observability.LogAuthEvent(h.auditService, r.Context(), tenantID, nil, observability.ActionLoginFailed, &ipAddr, &userAgent, err)

		util.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Audit Log Success
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, r.Context(), tenantID, &tokens.UserID, observability.ActionLogin, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"token_type":    "Bearer",
		"user_id":       tokens.UserID,
	})
}

// Logout godoc
// @Summary User logout
// @Description Invalidate the current session
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session ID from context (set by auth middleware)
	sessionID, _ := util.GetSessionIDFromContext(r.Context())
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	userID, _ := util.GetUserIDFromContext(r.Context())

	if err := h.authService.Logout(r.Context(), sessionID); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, r.Context(), tenantID, &userID, observability.ActionLogout, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		util.RespondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"token_type":    "Bearer",
		"user_id":       tokens.UserID,
	})
}

// Stub implementations for other handlers
// SendMagicLink godoc
// @Summary Send magic link
// @Description Send a magic link to the user's email for passwordless login
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body EmailRequest true "Email address"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/magic-link/send [post]
func (h *AuthHandler) SendMagicLink(w http.ResponseWriter, r *http.Request) {
	var req EmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	if err := h.authService.SendMagicLink(r.Context(), tenantID, req.Email); err != nil {
		// We generally don't want to expose if the email failed (user enumeration),
		// but for internal errors we might log it. For now, just return OK.
		// If it's a validation error, we might return 400.
		// For security, we usually return 200 OK even if user not found.
		// However, if it's a system error, we might want to log it.
		// Let's assume SendMagicLink returns error only on system failure or validation.
		// If user not found, it returns nil (as implemented in service).
		util.RespondError(w, http.StatusInternalServerError, "Failed to send magic link")
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	// Use LogEvent directly to include email in metadata
	h.auditService.LogEvent(r.Context(), &observability.AuditEvent{
		TenantID:  tenantID,
		Action:    observability.ActionMagicLinkSent,
		Metadata:  map[string]interface{}{"email": req.Email},
		IPAddress: &ipAddr,
		UserAgent: &userAgent,
		Status:    observability.AuditStatusSuccess,
	})

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "If the email exists, a magic link has been sent"})
}

// VerifyMagicLink godoc
// @Summary Verify magic link
// @Description Verify a magic link token and return access/refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body TokenRequest true "Magic link token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/magic-link/verify [post]
func (h *AuthHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokens, err := h.authService.VerifyMagicLink(r.Context(), req.Token)
	if err != nil {
		util.RespondError(w, http.StatusUnauthorized, "Invalid or expired magic link")
		return
	}

	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, r.Context(), tenantID, &tokens.UserID, observability.ActionMagicLinkUsed, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"token_type":    "Bearer",
	})
}

// ChangePassword godoc
// @Summary Change password
// @Description Change the authenticated user's password
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body ChangePasswordRequest true "Password change details"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/password/change [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get tenant and user ID from context (set by auth middleware)
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	userID, _ := util.GetUserIDFromContext(r.Context())

	if err := h.authService.ChangePassword(r.Context(), tenantID, userID, req.OldPassword, req.NewPassword); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, r.Context(), tenantID, &userID, observability.ActionPasswordChange, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// ResetPassword godoc
// @Summary Request password reset
// @Description Send a password reset link to the user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body EmailRequest true "Email address"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/password/reset [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req EmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get tenant ID from context
	tenantID, _ := util.GetTenantIDFromContext(r.Context())

	if err := h.authService.ResetPassword(r.Context(), tenantID, req.Email); err != nil {
		// Don't reveal errors to avoid user enumeration
		// Just return success
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	// Use LogEvent directly to include email in metadata
	h.auditService.LogEvent(r.Context(), &observability.AuditEvent{
		TenantID:  tenantID,
		Action:    observability.ActionPasswordReset,
		Metadata:  map[string]interface{}{"email": req.Email},
		IPAddress: &ipAddr,
		UserAgent: &userAgent,
		Status:    observability.AuditStatusSuccess,
	})

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "If the email exists, a password reset email has been sent"})
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user's email address using a token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param request body TokenRequest true "Verification token"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/email/verify [post]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		util.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Audit Log
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, r.Context(), tenantID, nil, observability.ActionEmailVerified, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

// SendVerificationEmail godoc
// @Summary Send verification email
// @Description Send a new verification email to the authenticated user
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/email/send-verification [post]
func (h *AuthHandler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	// Get tenant and user ID from context (set by auth middleware)
	tenantID, _ := util.GetTenantIDFromContext(r.Context())
	userID, _ := util.GetUserIDFromContext(r.Context())

	if err := h.authService.SendVerificationEmail(r.Context(), tenantID, userID); err != nil {
		util.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit Log
	ipAddr := util.GetIPAddress(r)
	userAgent := r.UserAgent()
	observability.LogAuthEvent(h.auditService, r.Context(), tenantID, &userID, observability.ActionVerificationSent, &ipAddr, &userAgent, nil)

	util.RespondJSON(w, http.StatusOK, map[string]string{"message": "Verification email sent"})
}
