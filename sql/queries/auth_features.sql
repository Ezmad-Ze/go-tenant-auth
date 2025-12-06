-- name: CreateDevice :one
INSERT INTO
    devices (
        user_id,
        tenant_id,
        device_name,
        device_type,
        fingerprint,
        user_agent,
        ip_address
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: GetDeviceByID :one
SELECT * FROM devices WHERE id = $1 AND tenant_id = $2;

-- name: GetDeviceByFingerprint :one
SELECT * FROM devices WHERE user_id = $1 AND fingerprint = $2;

-- name: ListUserDevices :many
SELECT *
FROM devices
WHERE
    user_id = $1
    AND tenant_id = $2
ORDER BY last_used_at DESC;

-- name: UpdateDeviceLastUsed :exec
UPDATE devices SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: TrustDevice :exec
UPDATE devices SET is_trusted = true WHERE id = $1;

-- name: DeleteDevice :exec
DELETE FROM devices WHERE id = $1 AND user_id = $2;

-- name: CreateMagicLink :one
INSERT INTO
    magic_links (
        user_id,
        tenant_id,
        email,
        token_hash,
        ip_address,
        user_agent,
        expires_at
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: GetMagicLinkByToken :one
SELECT *
FROM magic_links
WHERE
    token_hash = $1
    AND is_used = false
    AND expires_at > CURRENT_TIMESTAMP;

-- name: MarkMagicLinkUsed :exec
UPDATE magic_links
SET
    is_used = true,
    used_at = CURRENT_TIMESTAMP
WHERE
    id = $1;

-- name: CreateTOTP :one
INSERT INTO
    totp_2fa (
        user_id,
        tenant_id,
        secret,
        backup_codes
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetTOTPByUserID :one
SELECT * FROM totp_2fa WHERE user_id = $1;

-- name: EnableTOTP :exec
UPDATE totp_2fa
SET
    is_enabled = true,
    verified_at = CURRENT_TIMESTAMP
WHERE
    user_id = $1;

-- name: DisableTOTP :exec
UPDATE totp_2fa SET is_enabled = false WHERE user_id = $1;

-- name: UpdateTOTPBackupCodes :exec
UPDATE totp_2fa SET backup_codes = $1 WHERE user_id = $2;