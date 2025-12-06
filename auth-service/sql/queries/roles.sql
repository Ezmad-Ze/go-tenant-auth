-- name: CreateRole :one
INSERT INTO
    roles (
        tenant_id,
        name,
        description,
        is_system
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1 AND tenant_id = $2;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1 AND tenant_id = $2;

-- name: ListRoles :many
SELECT *
FROM roles
WHERE
    tenant_id = $1
    AND is_active = true
ORDER BY name;

-- name: AssignRoleToUser :one
INSERT INTO
    user_roles (
        user_id,
        role_id,
        assigned_by,
        expires_at
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2;

-- name: GetUserRoles :many
SELECT r.*
FROM roles r
    JOIN user_roles ur ON r.id = ur.role_id
WHERE
    ur.user_id = $1
    AND r.is_active = true
    AND (
        ur.expires_at IS NULL
        OR ur.expires_at > CURRENT_TIMESTAMP
    );

-- name: AssignPermissionToRole :one
INSERT INTO
    role_permissions (role_id, permission_id)
VALUES ($1, $2)
RETURNING
    *;

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE
    role_id = $1
    AND permission_id = $2;

-- name: GetRolePermissions :many
SELECT p.*
FROM
    permissions p
    JOIN role_permissions rp ON p.id = rp.permission_id
WHERE
    rp.role_id = $1
    AND p.is_active = true;

-- name: AssignPermissionToUser :one
INSERT INTO
    user_permissions (
        user_id,
        permission_id,
        assigned_by,
        expires_at
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: RemovePermissionFromUser :exec
DELETE FROM user_permissions
WHERE
    user_id = $1
    AND permission_id = $2;