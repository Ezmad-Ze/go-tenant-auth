-- name: CreateUser :one
INSERT INTO
    users (
        tenant_id,
        email,
        username,
        password_hash,
        first_name,
        last_name,
        phone,
        metadata
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    )
RETURNING
    *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND tenant_id = $2;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND tenant_id = $2;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 AND tenant_id = $2;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = COALESCE(
        sqlc.narg ('first_name'),
        first_name
    ),
    last_name = COALESCE(
        sqlc.narg ('last_name'),
        last_name
    ),
    phone = COALESCE(sqlc.narg ('phone'), phone),
    metadata = COALESCE(
        sqlc.narg ('metadata'),
        metadata
    )
WHERE
    id = $1
    AND tenant_id = $2
RETURNING
    *;

-- name: UpdateUserPassword :exec
UPDATE users
SET
    password_hash = $1,
    password_changed_at = CURRENT_TIMESTAMP
WHERE
    id = $2
    AND tenant_id = $3;

-- name: VerifyUserEmail :exec
UPDATE users
SET
    is_verified = true,
    email_verified_at = CURRENT_TIMESTAMP
WHERE
    id = $1
    AND tenant_id = $2;

-- name: IncrementFailedLoginAttempts :exec
UPDATE users
SET
    failed_login_attempts = failed_login_attempts + 1
WHERE
    id = $1
    AND tenant_id = $2;

-- name: ResetFailedLoginAttempts :exec
UPDATE users
SET
    failed_login_attempts = 0,
    locked_until = NULL
WHERE
    id = $1
    AND tenant_id = $2;

-- name: LockUser :exec
UPDATE users
SET
    locked_until = $1
WHERE
    id = $2
    AND tenant_id = $3;

-- name: UpdateLastLogin :exec
UPDATE users
SET
    last_login_at = CURRENT_TIMESTAMP
WHERE
    id = $1
    AND tenant_id = $2;

-- name: DeactivateUser :exec
UPDATE users
SET
    is_active = false
WHERE
    id = $1
    AND tenant_id = $2;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1 AND tenant_id = $2;

-- name: ListUsers :many
SELECT *
FROM users
WHERE
    tenant_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3;