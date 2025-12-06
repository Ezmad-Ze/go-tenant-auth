package email

import (
	"context"

	"github.com/ezmad/auth-service/internal/service"
	"github.com/rs/zerolog/log"
)

type mockEmailService struct{}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() service.EmailService {
	return &mockEmailService{}
}

func (s *mockEmailService) SendMagicLink(ctx context.Context, toEmail, magicLink string) error {
	log.Info().
		Str("to_email", toEmail).
		Str("magic_link", magicLink).
		Msg("📧 [MOCK EMAIL] Sending Magic Link")
	return nil
}

func (s *mockEmailService) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	log.Info().
		Str("to_email", toEmail).
		Str("token", token).
		Msg("📧 [MOCK EMAIL] Sending Verification Email")
	return nil
}

func (s *mockEmailService) SendPasswordResetEmail(ctx context.Context, toEmail, token string) error {
	log.Info().
		Str("to_email", toEmail).
		Str("token", token).
		Msg("📧 [MOCK EMAIL] Sending Password Reset Email")
	return nil
}
