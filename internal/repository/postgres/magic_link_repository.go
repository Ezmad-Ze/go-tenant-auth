package postgres

import (
	"context"
	"net/netip"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type magicLinkRepository struct {
	q *db.Queries
}

func NewMagicLinkRepository(q *db.Queries) repository.MagicLinkRepository {
	return &magicLinkRepository{
		q: q,
	}
}

func (r *magicLinkRepository) Create(ctx context.Context, link *domain.MagicLink) error {
	var ipAddr *netip.Addr
	if link.IPAddress != nil {
		if addr, err := netip.ParseAddr(*link.IPAddress); err == nil {
			ipAddr = &addr
		}
	}

	// Handle optional UserID
	var userID pgtype.UUID
	if link.UserID != nil {
		userID = uuidPtrToPgUUID(link.UserID)
	} else {
		userID.Valid = false
	}

	l, err := r.q.CreateMagicLink(ctx, db.CreateMagicLinkParams{
		UserID:    userID,
		TenantID:  link.TenantID,
		Email:     link.Email,
		TokenHash: link.TokenHash,
		IpAddress: ipAddr,
		UserAgent: link.UserAgent,
		ExpiresAt: pgtype.Timestamptz{Time: link.ExpiresAt, Valid: true},
	})
	if err != nil {
		return err
	}

	link.ID = l.ID
	link.CreatedAt = l.CreatedAt.Time
	if l.IsUsed != nil {
		link.IsUsed = *l.IsUsed
	} else {
		link.IsUsed = false
	}

	return nil
}

func (r *magicLinkRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.MagicLink, error) {
	l, err := r.q.GetMagicLinkByToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return toDomainMagicLink(l), nil
}

func (r *magicLinkRepository) MarkUsed(ctx context.Context, linkID uuid.UUID) error {
	return r.q.MarkMagicLinkUsed(ctx, linkID)
}

func toDomainMagicLink(l db.MagicLink) *domain.MagicLink {
	return &domain.MagicLink{
		ID:        l.ID,
		UserID:    pgUUIDToUUIDPtr(l.UserID),
		TenantID:  l.TenantID,
		Email:     l.Email,
		TokenHash: l.TokenHash,
		IPAddress: ipAddrToStringPtr(l.IpAddress),
		UserAgent: l.UserAgent,
		ExpiresAt: l.ExpiresAt.Time,
		UsedAt:    ptrTime(l.UsedAt),
		IsUsed:    l.IsUsed != nil && *l.IsUsed,
		CreatedAt: l.CreatedAt.Time,
	}
}
