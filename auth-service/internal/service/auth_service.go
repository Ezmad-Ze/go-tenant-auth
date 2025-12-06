package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/ezmad/auth-service/internal/security"
	"github.com/ezmad/auth-service/internal/util"
	"github.com/google/uuid"
)

// authService implements the AuthService interface
type authService struct {
	userRepo           repository.UserRepository
	sessionRepo        repository.SessionRepository
	deviceRepo         repository.DeviceRepository
	magicLinkRepo      repository.MagicLinkRepository
	totpRepo           repository.TOTPRepository
	cacheRepo          repository.CacheRepository
	emailService       EmailService
	failedLoginTracker FailedLoginTracker
	totpRateLimiter    *TOTPRateLimiter
	passwordHasher     *security.Argon2Params
	tokenManager       *security.PasetoToken
	totpManager        *security.TOTPManager
	config             AuthConfig
}

type AuthConfig struct {
	AccessTokenDuration      time.Duration
	RefreshTokenDuration     time.Duration
	MagicLinkDuration        time.Duration
	MaxFailedAttempts        int
	FailedLoginBlockDuration time.Duration
	SessionMaxDevices        int
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	deviceRepo repository.DeviceRepository,
	magicLinkRepo repository.MagicLinkRepository,
	totpRepo repository.TOTPRepository,
	cacheRepo repository.CacheRepository,
	emailService EmailService,
	failedLoginTracker FailedLoginTracker,
	totpRateLimiter *TOTPRateLimiter,
	passwordHasher *security.Argon2Params,
	tokenManager *security.PasetoToken,
	totpManager *security.TOTPManager,
	config AuthConfig,
) AuthService {
	return &authService{
		userRepo:           userRepo,
		sessionRepo:        sessionRepo,
		deviceRepo:         deviceRepo,
		magicLinkRepo:      magicLinkRepo,
		totpRepo:           totpRepo,
		cacheRepo:          cacheRepo,
		emailService:       emailService,
		failedLoginTracker: failedLoginTracker,
		totpRateLimiter:    totpRateLimiter,
		passwordHasher:     passwordHasher,
		tokenManager:       tokenManager,
		totpManager:        totpManager,
		config:             config,
	}
}

// Register implements user registration
func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*domain.TokenPair, error) {
	// Check if user already exists
	existing, _ := s.userRepo.GetByEmail(ctx, req.TenantID, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("user with email already exists")
	}

	// Hash password
	passwordHash, err := security.HashPassword(req.Password, s.passwordHasher)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     req.TenantID,
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Metadata:     req.Metadata,
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	return s.createTokenPair(ctx, user)
}

// Login implements user login
func (s *authService) Login(ctx context.Context, req *LoginRequest) (*domain.TokenPair, error) {
	// Get user
	user, err := s.userRepo.GetByEmail(ctx, req.TenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked using failed login tracker
	if s.failedLoginTracker != nil {
		locked, lockUntil, err := s.failedLoginTracker.IsAccountLocked(ctx, req.TenantID, user.ID)
		if err == nil && locked {
			return nil, fmt.Errorf("account is locked until %v", lockUntil)
		}
	}

	// Verify password
	valid, err := security.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		// Record failed login attempt using tracker
		if s.failedLoginTracker != nil {
			_ = s.failedLoginTracker.RecordFailedAttempt(ctx, req.TenantID, user.ID)
		}

		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if 2FA is enabled
	totp, _ := s.totpRepo.GetByUserID(ctx, user.ID)
	if totp != nil && totp.IsEnabled {
		// Check if account is locked for TOTP attempts
		if s.totpRateLimiter != nil {
			locked, lockUntil, err := s.totpRateLimiter.IsAccountLockedForTOTP(ctx, req.TenantID, user.ID)
			if err == nil && locked {
				return nil, fmt.Errorf("account locked for 2FA attempts until %v", lockUntil)
			}
		}

		// Check if either TOTP or backup code is provided
		if (req.TOTP == nil || *req.TOTP == "") && (req.BackupCode == nil || *req.BackupCode == "") {
			return nil, fmt.Errorf("2FA_REQUIRED")
		}
		
		// Try TOTP first if provided
		if req.TOTP != nil && *req.TOTP != "" {
			if !s.totpManager.VerifyCode(totp.Secret, *req.TOTP) {
				// Record failed TOTP attempt
				if s.totpRateLimiter != nil {
					_ = s.totpRateLimiter.RecordFailedTOTPAttempt(ctx, req.TenantID, user.ID)
				}
				return nil, fmt.Errorf("invalid 2FA code")
			}
			// TOTP verified successfully, reset attempts
			if s.totpRateLimiter != nil {
				_ = s.totpRateLimiter.ResetTOTPAttempts(ctx, req.TenantID, user.ID)
			}
		} else if req.BackupCode != nil && *req.BackupCode != "" {
			// Try backup code
			// Validate backup code by checking if it exists in the list
			found := false
			remainingCodes := make([]string, 0, len(totp.BackupCodes))
			for _, code := range totp.BackupCodes {
				if code == *req.BackupCode {
					found = true
					// Don't add to remaining codes (consume it)
				} else {
					remainingCodes = append(remainingCodes, code)
				}
			}

			if !found {
				// Record failed attempt
				if s.totpRateLimiter != nil {
					_ = s.totpRateLimiter.RecordFailedTOTPAttempt(ctx, req.TenantID, user.ID)
				}
				return nil, fmt.Errorf("invalid backup code")
			}

			// Backup code is valid, update the list
			if err := s.totpRepo.UpdateBackupCodes(ctx, user.ID, remainingCodes); err != nil {
				return nil, fmt.Errorf("failed to update backup codes: %w", err)
			}

			// Reset TOTP attempts on successful backup code use
			if s.totpRateLimiter != nil {
				_ = s.totpRateLimiter.ResetTOTPAttempts(ctx, req.TenantID, user.ID)
			}
		}
	}

	// Reset failed login attempts using tracker
	if s.failedLoginTracker != nil {
		_ = s.failedLoginTracker.ResetFailedAttempts(ctx, req.TenantID, user.ID)
	}
	
	// Update last login in database
	_ = s.userRepo.UpdateLastLogin(ctx, req.TenantID, user.ID)

	// Register/update device if provided
	// ... device handling logic

	// Create tokens
	return s.createTokenPair(ctx, user)
}

// createTokenPair is a helper to create access and refresh tokens
func (s *authService) createTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	// Create access token
	accessClaims := &security.TokenClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		TokenType: "access",
	}
	accessToken, err := s.tokenManager.CreateToken(accessClaims, s.config.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshClaims := &security.TokenClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		TokenType: "refresh",
	}
	refreshToken, err := s.tokenManager.CreateToken(refreshClaims, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	// Store session
	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		TenantID:         user.TenantID,
		TokenHash:        security.HashToken(accessToken),
		RefreshTokenHash: util.StringPtr(security.HashToken(refreshToken)),
		ExpiresAt:        time.Now().Add(s.config.AccessTokenDuration),
		IsActive:         true,
		CreatedAt:        time.Now(),
	}
	_ = s.sessionRepo.Create(ctx, session)

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.config.AccessTokenDuration.Seconds()),
		UserID:       user.ID,
	}, nil
}

// Logout invalidates a session
func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.Invalidate(ctx, sessionID)
}

// LogoutAll invalidates all sessions for a user
func (s *authService) LogoutAll(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.sessionRepo.InvalidateAllUserSessions(ctx, tenantID, userID)
}

// RefreshToken implements token refresh
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// Verify refresh token
	claims, err := s.tokenManager.VerifyToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("not a refresh token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Create new token pair (with token rotation)
	return s.createTokenPair(ctx, user)
}

// VerifyAccessToken verifies an access token and returns the user
func (s *authService) VerifyAccessToken(ctx context.Context, accessToken string) (*domain.User, error) {
	claims, err := s.tokenManager.VerifyToken(accessToken)
	if err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, claims.TenantID, claims.UserID)
}

// Stub implementations for other methods
func (s *authService) SendMagicLink(ctx context.Context, tenantID uuid.UUID, email string) error {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		// Return nil to avoid user enumeration
		return nil
	}

	// Generate token
	token, err := security.GenerateRandomToken(32)
	if err != nil {
		return err
	}

	// Hash token
	tokenHash := security.HashToken(token)

	// Create magic link record
	magicLink := &domain.MagicLink{
		ID:        uuid.New(),
		UserID:    &user.ID,
		TenantID:  tenantID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.config.MagicLinkDuration),
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.magicLinkRepo.Create(ctx, magicLink); err != nil {
		return err
	}

	// Send email
	// In a real app, we would construct the full URL here
	magicLinkURL := fmt.Sprintf("http://localhost:3000/auth/magic-link?token=%s", token)
	return s.emailService.SendMagicLink(ctx, email, magicLinkURL)
}

func (s *authService) VerifyMagicLink(ctx context.Context, token string) (*domain.TokenPair, error) {
	// Hash token
	tokenHash := security.HashToken(token)

	// Get magic link
	magicLink, err := s.magicLinkRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired magic link")
	}

	// Validate
	if magicLink.IsUsed {
		return nil, fmt.Errorf("magic link already used")
	}

	if time.Now().After(magicLink.ExpiresAt) {
		return nil, fmt.Errorf("magic link expired")
	}

	// Mark as used
	if err := s.magicLinkRepo.MarkUsed(ctx, magicLink.ID); err != nil {
		return nil, err
	}

	if magicLink.UserID == nil {
		return nil, fmt.Errorf("invalid magic link: no user associated")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, magicLink.TenantID, *magicLink.UserID)
	if err != nil {
		return nil, err
	}

	// Create tokens
	return s.createTokenPair(ctx, user)
}

func (s *authService) ChangePassword(ctx context.Context, tenantID, userID uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	valid, err := security.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil || !valid {
		return fmt.Errorf("invalid current password")
	}

	// Hash new password
	newPasswordHash, err := security.HashPassword(newPassword, s.passwordHasher)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, tenantID, userID, user.Email, newPasswordHash); err != nil {
		return err
	}

	// Invalidate all sessions
	_ = s.sessionRepo.InvalidateAllUserSessions(ctx, tenantID, userID)

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, tenantID uuid.UUID, email string) error {
	// Get user (don't reveal if user doesn't exist for security)
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		// Return nil to avoid user enumeration
		return nil
	}

	// Generate reset token
	token, err := security.GenerateRandomToken(32)
	if err != nil {
		return err
	}

	// Hash token
	tokenHash := security.HashToken(token)

	// Create reset token record (reusing MagicLink table)
	resetLink := &domain.MagicLink{
		ID:        uuid.New(),
		UserID:    &user.ID,
		TenantID:  tenantID,
		Email:     user.Email,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiry for password reset
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.magicLinkRepo.Create(ctx, resetLink); err != nil {
		return err
	}

	// Send reset email
	resetURL := fmt.Sprintf("http://localhost:3000/auth/reset-password?token=%s", token)
	return s.emailService.SendPasswordResetEmail(ctx, user.Email, resetURL)
}

func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// Hash token
	tokenHash := security.HashToken(token)

	// Get reset token
	resetLink, err := s.magicLinkRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Validate
	if resetLink.IsUsed {
		return fmt.Errorf("reset token already used")
	}

	if time.Now().After(resetLink.ExpiresAt) {
		return fmt.Errorf("reset token expired")
	}

	if resetLink.UserID == nil {
		return fmt.Errorf("invalid reset token: no user associated")
	}

	// Mark token as used
	if err := s.magicLinkRepo.MarkUsed(ctx, resetLink.ID); err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := security.HashPassword(newPassword, s.passwordHasher)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, resetLink.TenantID, *resetLink.UserID, resetLink.Email, newPasswordHash); err != nil {
		return err
	}

	// Invalidate all sessions
	_ = s.sessionRepo.InvalidateAllUserSessions(ctx, resetLink.TenantID, *resetLink.UserID)

	return nil
}

func (s *authService) SendVerificationEmail(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Don't send if already verified
	if user.IsVerified {
		return fmt.Errorf("email already verified")
	}

	// Generate token
	token, err := security.GenerateRandomToken(32)
	if err != nil {
		return err
	}

	// Hash token
	tokenHash := security.HashToken(token)

	// Create verification token record (reusing MagicLink table)
	verificationLink := &domain.MagicLink{
		ID:        uuid.New(),
		UserID:    &user.ID,
		TenantID:  tenantID,
		Email:     user.Email,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour expiry for email verification
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.magicLinkRepo.Create(ctx, verificationLink); err != nil {
		return err
	}

	// Send verification email
	// In a real app, we would construct the full URL here
	verificationURL := fmt.Sprintf("http://localhost:3000/auth/verify-email?token=%s", token)
	return s.emailService.SendVerificationEmail(ctx, user.Email, verificationURL)
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// Hash token
	tokenHash := security.HashToken(token)

	// Get verification token
	verificationLink, err := s.magicLinkRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Validate
	if verificationLink.IsUsed {
		return fmt.Errorf("verification token already used")
	}

	if time.Now().After(verificationLink.ExpiresAt) {
		return fmt.Errorf("verification token expired")
	}

	if verificationLink.UserID == nil {
		return fmt.Errorf("invalid verification token: no user associated")
	}

	// Mark token as used
	if err := s.magicLinkRepo.MarkUsed(ctx, verificationLink.ID); err != nil {
		return err
	}

	// Verify email in user table
	if err := s.userRepo.VerifyEmail(ctx, verificationLink.TenantID, *verificationLink.UserID); err != nil {
		return err
	}

	return nil
}
