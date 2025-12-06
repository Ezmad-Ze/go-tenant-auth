-- name: CreateTenant :one
INSERT INTO
    tenants (name, slug, domain, settings)
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1 AND is_active = true;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1 AND is_active = true;

-- name: GetTenantByDomain :one
SELECT * FROM tenants WHERE domain = $1 AND is_active = true;

-- name: UpdateTenant :one
UPDATE tenants
SET
    name = COALESCE(sqlc.narg ('name'), name),
    domain = COALESCE(sqlc.narg ('domain'), domain),
    settings = COALESCE(
        sqlc.narg ('settings'),
        settings
    )
WHERE
    id = $1
RETURNING
    *;

-- name: ListTenants :many
SELECT *
FROM tenants
WHERE
    is_active = true
ORDER BY created_at DESC;