-- name: CreateSession :one
INSERT INTO
    sessions (
        user_id,
        tenant_id,
        device_id,
        token_hash,
        refresh_token_hash,
        ip_address,
        user_agent,
        expires_at,
        refresh_expires_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
    )
RETURNING
    *;

-- name: GetSessionByTokenHash :one
SELECT *
FROM sessions
WHERE
    token_hash = $1
    AND is_active = true
    AND expires_at > CURRENT_TIMESTAMP;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1 AND tenant_id = $2;

-- name: ListUserSessions :many
SELECT *
FROM sessions
WHERE
    user_id = $1
    AND tenant_id = $2
    AND is_active = true
ORDER BY last_activity_at DESC;

-- name: UpdateSessionActivity :exec
UPDATE sessions
SET
    last_activity_at = CURRENT_TIMESTAMP
WHERE
    id = $1;

-- name: InvalidateSession :exec
UPDATE sessions SET is_active = false WHERE id = $1;

-- name: InvalidateAllUserSessions :exec
UPDATE sessions
SET
    is_active = false
WHERE
    user_id = $1
    AND tenant_id = $2;

-- name: InvalidateExpiredSessions :exec
UPDATE sessions
SET
    is_active = false
WHERE
    expires_at < CURRENT_TIMESTAMP
    AND is_active = true;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: CountUserActiveSessions :one
SELECT COUNT(*)
FROM sessions
WHERE
    user_id = $1
    AND tenant_id = $2
    AND is_active = true
    AND expires_at > CURRENT_TIMESTAMP;