package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/security"
	"github.com/ezmad/auth-service/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		request       *RegisterRequest
		setupMocks    func(*testutil.MockUserRepository, *testutil.MockSessionRepository)
		expectedError string
	}{
		{
			name: "successful registration",
			request: &RegisterRequest{
				TenantID: uuid.New(),
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(userRepo *testutil.MockUserRepository, sessionRepo *testutil.MockSessionRepository) {
				userRepo.GetByEmailFunc = func(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
					return nil, errors.New("not found")
				}
				userRepo.CreateFunc = func(ctx context.Context, user *domain.User) error {
					user.ID = uuid.New()
					user.CreatedAt = time.Now()
					user.UpdatedAt = time.Now()
					return nil
				}
				sessionRepo.CreateFunc = func(ctx context.Context, session *domain.Session) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name: "duplicate email",
			request: &RegisterRequest{
				TenantID: uuid.New(),
				Email:    "existing@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(userRepo *testutil.MockUserRepository, sessionRepo *testutil.MockSessionRepository) {
				userRepo.GetByEmailFunc = func(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
					return &domain.User{ID: uuid.New(), Email: email}, nil
				}
			},
			expectedError: "user with email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &testutil.MockUserRepository{}
			sessionRepo := &testutil.MockSessionRepository{}
			totpRepo := &testutil.MockTOTPRepository{}
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, sessionRepo)
			}

			// Create service
			tokenManager, err := security.NewPasetoToken("12345678901234567890123456789012", "test-issuer")
			require.NoError(t, err, "Failed to create token manager")
			totpManager := security.NewTOTPManager("TestIssuer", 30, 6)
			
			authService := NewAuthService(
				userRepo,
				sessionRepo,
				nil, // deviceRepo
				nil, // magicLinkRepo
				totpRepo,
				nil, // cacheRepo
				nil, // emailService
				nil, // failedLoginTracker
				nil, // totpRateLimiter
				&security.Argon2Params{
					Memory:      64 * 1024,
					Iterations:  3,
					Parallelism: 2,
					SaltLength:  16,
					KeyLength:   32,
				},
				tokenManager,
				totpManager,
				AuthConfig{
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 168 * time.Hour,
				},
			)

			// Execute
			tokens, err := authService.Register(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, tokens)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tokens)
				assert.NotEmpty(t, tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
				assert.Greater(t, tokens.ExpiresIn, 0)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	email := "test@example.com"
	password := "SecurePass123!"
	
	// Pre-hash password for test user
	argon2Params := &security.Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
	passwordHash, _ := security.HashPassword(password, argon2Params)

	tests := []struct {
		name          string
		request       *LoginRequest
		setupMocks    func(*testutil.MockUserRepository, *testutil.MockTOTPRepository)
		expectedError string
	}{
		{
			name: "successful login without 2FA",
			request: &LoginRequest{
				TenantID: tenantID,
				Email:    email,
				Password: password,
			},
			setupMocks: func(userRepo *testutil.MockUserRepository, totpRepo *testutil.MockTOTPRepository) {
				userRepo.GetByEmailFunc = func(ctx context.Context, tid uuid.UUID, e string) (*domain.User, error) {
					return &domain.User{
						ID:           userID,
						TenantID:     tenantID,
						Email:        email,
						PasswordHash: passwordHash,
						IsActive:     true,
					}, nil
				}
				userRepo.UpdateLastLoginFunc = func(ctx context.Context, tid, uid uuid.UUID) error {
					return nil
				}
				totpRepo.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) (*domain.TOTP2FA, error) {
					return nil, nil // No 2FA enabled
				}
			},
			expectedError: "",
		},
		{
			name: "invalid credentials",
			request: &LoginRequest{
				TenantID: tenantID,
				Email:    email,
				Password: "WrongPassword",
			},
			setupMocks: func(userRepo *testutil.MockUserRepository, totpRepo *testutil.MockTOTPRepository) {
				userRepo.GetByEmailFunc = func(ctx context.Context, tid uuid.UUID, e string) (*domain.User, error) {
					return &domain.User{
						ID:           userID,
						TenantID:     tenantID,
						Email:        email,
						PasswordHash: passwordHash,
						IsActive:     true,
					}, nil
				}
			},
			expectedError: "invalid credentials",
		},
		{
			name: "2FA required",
			request: &LoginRequest{
				TenantID: tenantID,
				Email:    email,
				Password: password,
			},
			setupMocks: func(userRepo *testutil.MockUserRepository, totpRepo *testutil.MockTOTPRepository) {
				userRepo.GetByEmailFunc = func(ctx context.Context, tid uuid.UUID, e string) (*domain.User, error) {
					return &domain.User{
						ID:           userID,
						TenantID:     tenantID,
						Email:        email,
						PasswordHash: passwordHash,
						IsActive:     true,
					}, nil
				}
				totpRepo.GetByUserIDFunc = func(ctx context.Context, uid uuid.UUID) (*domain.TOTP2FA, error) {
					return &domain.TOTP2FA{
						UserID:    userID,
						IsEnabled: true,
						Secret:    "TESTSECRET",
					}, nil
				}
			},
			expectedError: "2FA_REQUIRED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &testutil.MockUserRepository{}
			sessionRepo := &testutil.MockSessionRepository{
				CreateFunc: func(ctx context.Context, session *domain.Session) error {
					return nil
				},
			}
			totpRepo := &testutil.MockTOTPRepository{}
			
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, totpRepo)
			}

			// Create service
			tokenManager, err := security.NewPasetoToken("12345678901234567890123456789012", "test-issuer")
			require.NoError(t, err, "Failed to create token manager")
			totpManager := security.NewTOTPManager("TestIssuer", 30, 6)
			
			authService := NewAuthService(
				userRepo,
				sessionRepo,
				nil,
				nil,
				totpRepo,
				nil,
				nil,
				nil,
				nil,
				argon2Params,
				tokenManager,
				totpManager,
				AuthConfig{
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 168 * time.Hour,
				},
			)

			// Execute
			tokens, err := authService.Login(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, tokens)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tokens)
				assert.NotEmpty(t, tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
			}
		})
	}
}
