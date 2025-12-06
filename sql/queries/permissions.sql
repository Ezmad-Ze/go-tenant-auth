-- name: CreatePermission :one
INSERT INTO
    permissions (
        tenant_id,
        resource_id,
        scope_id,
        name,
        description,
        abac_rule
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetPermissionByID :one
SELECT p.*, r.name as resource_name, s.name as scope_name
FROM
    permissions p
    JOIN resources r ON p.resource_id = r.id
    JOIN scopes s ON p.scope_id = s.id
WHERE
    p.id = $1
    AND p.tenant_id = $2;

-- name: GetPermissionByName :one
SELECT p.*, r.name as resource_name, s.name as scope_name
FROM
    permissions p
    JOIN resources r ON p.resource_id = r.id
    JOIN scopes s ON p.scope_id = s.id
WHERE
    p.name = $1
    AND p.tenant_id = $2;

-- name: ListPermissions :many
SELECT p.*, r.name as resource_name, s.name as scope_name
FROM
    permissions p
    JOIN resources r ON p.resource_id = r.id
    JOIN scopes s ON p.scope_id = s.id
WHERE
    p.tenant_id = $1
ORDER BY p.name;

-- name: GetUserPermissions :many
-- Get all permissions for a user (both direct and from roles)
SELECT DISTINCT
    p.id,
    p.tenant_id,
    p.resource_id,
    p.scope_id,
    p.name,
    p.description,
    p.abac_rule,
    r.name as resource_name,
    s.name as scope_name
FROM
    permissions p
    JOIN resources r ON p.resource_id = r.id
    JOIN scopes s ON p.scope_id = s.id
    LEFT JOIN user_permissions up ON p.id = up.permission_id
    LEFT JOIN role_permissions rp ON p.id = rp.permission_id
    LEFT JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE (
        up.user_id = $1
        OR ur.user_id = $1
    )
    AND p.tenant_id = $2
    AND p.is_active = true
    AND (
        up.expires_at IS NULL
        OR up.expires_at > CURRENT_TIMESTAMP
    )
    AND (
        ur.expires_at IS NULL
        OR ur.expires_at > CURRENT_TIMESTAMP
    );

-- name: CheckUserPermission :one
-- Check if user has a specific permission
SELECT EXISTS (
        SELECT 1
        FROM
            permissions p
            LEFT JOIN user_permissions up ON p.id = up.permission_id
            LEFT JOIN role_permissions rp ON p.id = rp.permission_id
            LEFT JOIN user_roles ur ON rp.role_id = ur.role_id
        WHERE (
                up.user_id = $1
                OR ur.user_id = $1
            )
            AND p.name = $2
            AND p.tenant_id = $3
            AND p.is_active = true
            AND (
                up.expires_at IS NULL
                OR up.expires_at > CURRENT_TIMESTAMP
            )
            AND (
                ur.expires_at IS NULL
                OR ur.expires_at > CURRENT_TIMESTAMP
            )
    ) AS has_permission;

-- name: CreateResource :one
INSERT INTO
    resources (tenant_id, name, description)
VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreateScope :one
INSERT INTO
    scopes (tenant_id, name, description)
VALUES ($1, $2, $3)
RETURNING
    *;

-- name: ListResources :many
SELECT *
FROM resources
WHERE
    tenant_id = $1
    AND is_active = true
ORDER BY name;

-- name: ListScopes :many
SELECT *
FROM scopes
WHERE
    tenant_id = $1
    AND is_active = true
ORDER BY name;