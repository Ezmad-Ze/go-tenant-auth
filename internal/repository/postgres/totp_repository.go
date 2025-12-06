package postgres

import (
	"context"
	"encoding/json"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
)

type totpRepository struct {
	q *db.Queries
}

func NewTOTPRepository(q *db.Queries) repository.TOTPRepository {
	return &totpRepository{
		q: q,
	}
}

func (r *totpRepository) Create(ctx context.Context, totp *domain.TOTP2FA) error {
	backupCodesJSON, err := json.Marshal(totp.BackupCodes)
	if err != nil {
		return err
	}

	t, err := r.q.CreateTOTP(ctx, db.CreateTOTPParams{
		UserID:      totp.UserID,
		TenantID:    totp.TenantID,
		Secret:      totp.Secret,
		BackupCodes: backupCodesJSON,
	})
	if err != nil {
		return err
	}

	totp.ID = t.ID
	totp.CreatedAt = t.CreatedAt.Time
	if t.IsEnabled != nil {
		totp.IsEnabled = *t.IsEnabled
	} else {
		totp.IsEnabled = false
	}

	return nil
}

func (r *totpRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TOTP2FA, error) {
	t, err := r.q.GetTOTPByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toDomainTOTP(t), nil
}

func (r *totpRepository) Enable(ctx context.Context, userID uuid.UUID) error {
	return r.q.EnableTOTP(ctx, userID)
}

func (r *totpRepository) Disable(ctx context.Context, userID uuid.UUID) error {
	return r.q.DisableTOTP(ctx, userID)
}

func (r *totpRepository) UpdateBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error {
	backupCodesJSON, err := json.Marshal(codes)
	if err != nil {
		return err
	}

	return r.q.UpdateTOTPBackupCodes(ctx, db.UpdateTOTPBackupCodesParams{
		BackupCodes: backupCodesJSON,
		UserID:      userID,
	})
}

func toDomainTOTP(t db.Totp2fa) *domain.TOTP2FA {
	totp := &domain.TOTP2FA{
		ID:          t.ID,
		UserID:      t.UserID,
		TenantID:    t.TenantID,
		Secret:      t.Secret,
		IsEnabled:   t.IsEnabled != nil && *t.IsEnabled,
		VerifiedAt:  ptrTime(t.VerifiedAt),
		CreatedAt:   t.CreatedAt.Time,
		UpdatedAt:   t.UpdatedAt.Time,
	}

	if len(t.BackupCodes) > 0 {
		var codes []string
		if err := json.Unmarshal(t.BackupCodes, &codes); err == nil {
			totp.BackupCodes = codes
		}
	}

	return totp
}
