package security

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPManager handles Time-based One-Time Password operations
type TOTPManager struct {
	issuer string
	period uint
	digits int
}

// NewTOTPManager creates a new TOTP manager
func NewTOTPManager(issuer string, period uint, digits int) *TOTPManager {
	return &TOTPManager{
		issuer: issuer,
		period: period,
		digits: digits,
	}
}

// GenerateSecret generates a new TOTP secret
func (t *TOTPManager) GenerateSecret(accountName string) (string, error) {
	// Generate a random 20-byte secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}

	// Encode to base32
	return base32.StdEncoding.EncodeToString(secret), nil
}

// GenerateQRCode generates a QR code URL for TOTP setup
func (t *TOTPManager) GenerateQRCode(accountName, secret string) (string, error) {
	// Map digits to proper Digits type from otp package
	var digits otp.Digits
	switch t.digits {
	case 6:
		digits = otp.DigitsSix
	case 8:
		digits = otp.DigitsEight
	default:
		digits = otp.DigitsSix
	}

	opts := totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: accountName,
		Period:      t.period,
		Digits:      digits,
	}

	key, err := totp.Generate(opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return key.URL(), nil
}

// VerifyCode verifies a TOTP code
func (t *TOTPManager) VerifyCode(secret, code string) bool {
	// Remove spaces and convert to uppercase
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	code = strings.TrimSpace(code)

	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates backup codes for 2FA
func (t *TOTPManager) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := GenerateRandomToken(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code
	}
	return codes, nil
}
