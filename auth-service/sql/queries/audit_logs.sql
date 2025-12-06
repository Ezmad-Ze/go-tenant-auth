-- name: CreateAuditLog :one
INSERT INTO
    audit_logs (
        tenant_id,
        user_id,
        action,
        resource_type,
        resource_id,
        metadata,
        ip_address,
        user_agent,
        status,
        error_message
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
        $9,
        $10
    )
RETURNING
    *;

-- name: GetAuditLogsByTenant :many
SELECT *
FROM audit_logs
WHERE
    tenant_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3;

-- name: GetAuditLogsByUser :many
SELECT *
FROM audit_logs
WHERE
    tenant_id = $1
    AND user_id = $2
ORDER BY created_at DESC
LIMIT $3
OFFSET
    $4;

-- name: GetAuditLogsByAction :many
SELECT *
FROM audit_logs
WHERE
    tenant_id = $1
    AND action = $2
ORDER BY created_at DESC
LIMIT $3
OFFSET
    $4;

-- name: GetAuditLogsByResource :many
SELECT *
FROM audit_logs
WHERE
    tenant_id = $1
    AND resource_type = $2
    AND resource_id = $3
ORDER BY created_at DESC;

-- name: GetAuditLogsByDateRange :many
SELECT *
FROM audit_logs
WHERE
    tenant_id = $1
    AND created_at >= $2
    AND created_at <= $3
ORDER BY created_at DESC
LIMIT $4
OFFSET
    $5;

-- name: CountAuditLogsByTenant :one
SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1;

-- name: CountAuditLogsByUser :one
SELECT COUNT(*)
FROM audit_logs
WHERE
    tenant_id = $1
    AND user_id = $2;