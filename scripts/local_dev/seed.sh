#!/bin/bash

# Seed script for development database
# This script seeds the database with sample data for testing

set -e

# Database connection
DB_URL="${DATABASE_URL:-postgres://authuser:authpass@localhost:5432/authdb?sslmode=disable}"

echo "Seeding database with sample data..."

# Generate UUIDs (or use fixed ones for testing)
TENANT_ID="00000000-0000-0000-0000-000000000001"
ADMIN_USER_ID="00000000-0000-0000-0000-000000000010"

# SQL to insert sample data
psql "$DB_URL" <<-EOSQL
    -- Insert default tenant
    INSERT INTO tenants (id, name, slug, is_active) 
    VALUES ('$TENANT_ID', 'Default Tenant', 'default', true)
    ON CONFLICT (slug) DO NOTHING;

    -- Insert default resources
    INSERT INTO resources (tenant_id, name, description) VALUES
        ('$TENANT_ID', 'user', 'User management'),
        ('$TENANT_ID', 'project', 'Project management'),
        ('$TENANT_ID', 'invoice', 'Invoice management'),
        ('$TENANT_ID', 'company', 'Company management')
    ON CONFLICT (tenant_id, name) DO NOTHING;

    -- Insert default scopes
    INSERT INTO scopes (tenant_id, name, description) VALUES
        ('$TENANT_ID', 'create', 'Create permission'),
        ('$TENANT_ID', 'read', 'Read permission'),
        ('$TENANT_ID', 'update', 'Update permission'),
        ('$TENANT_ID', 'delete', 'Delete permission'),
        ('$TENANT_ID', 'manage', 'Full management permission')
    ON CONFLICT (tenant_id, name) DO NOTHING;

    -- Insert sample permissions
    INSERT INTO permissions (tenant_id, resource_id, scope_id, name)
    SELECT 
        '$TENANT_ID',
        r.id,
        s.id,
        r.name || '.' || s.name
    FROM resources r
    CROSS JOIN scopes s
    WHERE r.tenant_id = '$TENANT_ID' AND s.tenant_id = '$TENANT_ID'
    ON CONFLICT (tenant_id, name) DO NOTHING;

    -- Insert default roles
    INSERT INTO roles (tenant_id, name, description, is_system) VALUES
        ('$TENANT_ID', 'admin', 'System administrator with full access', true),
        ('$TENANT_ID', 'manager', 'Manager with read/write access', false),
        ('$TENANT_ID', 'user', 'Standard user with limited access', false)
    ON CONFLICT (tenant_id, name) DO NOTHING;

    -- Assign all permissions to admin role
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT r.id, p.id
    FROM roles r
    CROSS JOIN permissions p
    WHERE r.name = 'admin' AND r.tenant_id = '$TENANT_ID' AND p.tenant_id = '$TENANT_ID'
    ON CONFLICT (role_id, permission_id) DO NOTHING;

    EOSQL

echo "✓ Database seeded successfully!"
echo "  Tenant ID: $TENANT_ID"
echo "  Default roles created: admin, manager, user"
