package postgres

import (
	"context"
	"net/netip"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
)

type deviceRepository struct {
	q *db.Queries
}

func NewDeviceRepository(q *db.Queries) repository.DeviceRepository {
	return &deviceRepository{
		q: q,
	}
}

func (r *deviceRepository) Create(ctx context.Context, device *domain.Device) error {
	var ipAddr *netip.Addr
	if device.IPAddress != nil {
		if addr, err := netip.ParseAddr(*device.IPAddress); err == nil {
			ipAddr = &addr
		}
	}

	d, err := r.q.CreateDevice(ctx, db.CreateDeviceParams{
		UserID:      device.UserID,
		TenantID:    device.TenantID,
		DeviceName:  device.DeviceName,
		DeviceType:  device.DeviceType,
		Fingerprint: device.Fingerprint,
		UserAgent:   device.UserAgent,
		IpAddress:   ipAddr,
	})
	if err != nil {
		return err
	}

	device.ID = d.ID
	device.CreatedAt = d.CreatedAt.Time
	device.LastUsedAt = d.LastUsedAt.Time
	if d.IsTrusted != nil {
		device.IsTrusted = *d.IsTrusted
	} else {
		device.IsTrusted = false
	}

	return nil
}

func (r *deviceRepository) GetByID(ctx context.Context, tenantID, deviceID uuid.UUID) (*domain.Device, error) {
	d, err := r.q.GetDeviceByID(ctx, db.GetDeviceByIDParams{
		ID:       deviceID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainDevice(d), nil
}

func (r *deviceRepository) GetByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error) {
	d, err := r.q.GetDeviceByFingerprint(ctx, db.GetDeviceByFingerprintParams{
		UserID:      userID,
		Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, err
	}
	return toDomainDevice(d), nil
}

func (r *deviceRepository) ListUserDevices(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Device, error) {
	devices, err := r.q.ListUserDevices(ctx, db.ListUserDevicesParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Device, len(devices))
	for i, d := range devices {
		result[i] = toDomainDevice(d)
	}
	return result, nil
}

func (r *deviceRepository) UpdateLastUsed(ctx context.Context, deviceID uuid.UUID) error {
	return r.q.UpdateDeviceLastUsed(ctx, deviceID)
}

func (r *deviceRepository) TrustDevice(ctx context.Context, deviceID uuid.UUID) error {
	return r.q.TrustDevice(ctx, deviceID)
}

func (r *deviceRepository) Delete(ctx context.Context, deviceID, userID uuid.UUID) error {
	return r.q.DeleteDevice(ctx, db.DeleteDeviceParams{
		ID:     deviceID,
		UserID: userID,
	})
}

func toDomainDevice(d db.Device) *domain.Device {
	return &domain.Device{
		ID:          d.ID,
		UserID:      d.UserID,
		TenantID:    d.TenantID,
		DeviceName:  d.DeviceName,
		DeviceType:  d.DeviceType,
		Fingerprint: d.Fingerprint,
		UserAgent:   d.UserAgent,
		IPAddress:   ipAddrToStringPtr(d.IpAddress),
		IsTrusted:   d.IsTrusted != nil && *d.IsTrusted,
		LastUsedAt:  d.LastUsedAt.Time,
		CreatedAt:   d.CreatedAt.Time,
	}
}
