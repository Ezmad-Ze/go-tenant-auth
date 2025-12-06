package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/ezmad/auth-service/internal/security"
	"github.com/google/uuid"
)

type totpService struct {
	totpRepo       repository.TOTPRepository
	userRepo       repository.UserRepository
	totpManager    *security.TOTPManager
	passwordHasher *security.Argon2Params
}

// NewTOTPService creates a new instance of TOTPService
func NewTOTPService(
	totpRepo repository.TOTPRepository,
	userRepo repository.UserRepository,
	totpManager *security.TOTPManager,
	passwordHasher *security.Argon2Params,
) TOTPService {
	return &totpService{
		totpRepo:       totpRepo,
		userRepo:       userRepo,
		totpManager:    totpManager,
		passwordHasher: passwordHasher,
	}
}

// Setup initiates the 2FA setup process
func (s *totpService) Setup(ctx context.Context, tenantID, userID uuid.UUID) (*TOTPSetupResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if 2FA is already enabled
	existing, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		// If error is not "not found", return error
		// For now, assume error means not found or DB error.
		// Ideally repo should return specific error for not found.
		// Let's proceed assuming we can overwrite or create new.
	}
	if existing != nil && existing.IsEnabled {
		return nil, errors.New("2FA is already enabled")
	}

	// Generate secret
	secret, err := s.totpManager.GenerateSecret(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	// Generate QR code URL
	qrCodeURL, err := s.totpManager.GenerateQRCode(user.Email, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Generate backup codes
	backupCodes, err := s.totpManager.GenerateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Save to DB (disabled initially)
	totp := &domain.TOTP2FA{
		UserID:      userID,
		TenantID:    tenantID,
		Secret:      secret,
		BackupCodes: backupCodes,
		IsEnabled:   false,
	}

	if err := s.totpRepo.Create(ctx, totp); err != nil {
		return nil, fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	return &TOTPSetupResponse{
		Secret:      secret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
	}, nil
}

// Verify verifies a TOTP code without enabling it (used during setup or login)
func (s *totpService) Verify(ctx context.Context, tenantID, userID uuid.UUID, code string) error {
	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get TOTP config: %w", err)
	}
	if totp == nil {
		return errors.New("2FA not set up")
	}

	if !s.totpManager.VerifyCode(totp.Secret, code) {
		return errors.New("invalid code")
	}

	return nil
}

// Enable verifies a code and enables 2FA
func (s *totpService) Enable(ctx context.Context, tenantID, userID uuid.UUID, code string) error {
	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get TOTP config: %w", err)
	}
	if totp == nil {
		return errors.New("2FA not set up")
	}

	if !s.totpManager.VerifyCode(totp.Secret, code) {
		return errors.New("invalid code")
	}

	if err := s.totpRepo.Enable(ctx, userID); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

// Disable disables 2FA (requires password verification)
func (s *totpService) Disable(ctx context.Context, tenantID, userID uuid.UUID, password string) error {
	// Verify password
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	match, err := security.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	if !match {
		return errors.New("invalid password")
	}

	if err := s.totpRepo.Disable(ctx, userID); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}

// GenerateBackupCodes generates new backup codes
func (s *totpService) GenerateBackupCodes(ctx context.Context, tenantID, userID uuid.UUID) ([]string, error) {
	codes, err := s.totpManager.GenerateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	if err := s.totpRepo.UpdateBackupCodes(ctx, userID, codes); err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return codes, nil
}

// ValidateBackupCode validates and consumes a backup code
func (s *totpService) ValidateBackupCode(ctx context.Context, tenantID, userID uuid.UUID, code string) error {
	totp, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get TOTP config: %w", err)
	}
	if totp == nil || !totp.IsEnabled {
		return errors.New("2FA not enabled")
	}

	// Check if code exists in backup codes
	found := false
	remainingCodes := make([]string, 0, len(totp.BackupCodes))
	for _, c := range totp.BackupCodes {
		if c == code {
			found = true
			// Don't add to remaining codes (consume it)
		} else {
			remainingCodes = append(remainingCodes, c)
		}
	}

	if !found {
		return errors.New("invalid backup code")
	}

	// Update backup codes in DB
	if err := s.totpRepo.UpdateBackupCodes(ctx, userID, remainingCodes); err != nil {
		return fmt.Errorf("failed to update backup codes: %w", err)
	}

	return nil
}
