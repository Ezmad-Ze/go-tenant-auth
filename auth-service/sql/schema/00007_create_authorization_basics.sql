-- +goose Up
-- +goose StatementBegin

-- Create resources table (entities like project, invoice, user, etc.)
CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- e.g., "project", "invoice", "user"
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, name)
);

-- Create scopes table (actions like create, read, update, delete)
CREATE TABLE scopes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- e.g., "create", "read", "update", "delete"
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, name)
);

-- Create permissions table (combines resource + scope)
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    resource_id UUID NOT NULL REFERENCES resources (id) ON DELETE CASCADE,
    scope_id UUID NOT NULL REFERENCES scopes (id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- e.g., "project.read", "invoice.update"
    description TEXT,
    abac_rule JSONB, -- Optional ABAC rules for contextual checks
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (
        tenant_id,
        resource_id,
        scope_id
    ),
    UNIQUE (tenant_id, name)
);

-- Create indexes
CREATE INDEX idx_resources_tenant_id ON resources (tenant_id);

CREATE INDEX idx_scopes_tenant_id ON scopes (tenant_id);

CREATE INDEX idx_permissions_tenant_id ON permissions (tenant_id);

CREATE INDEX idx_permissions_resource_id ON permissions (resource_id);

CREATE INDEX idx_permissions_scope_id ON permissions (scope_id);

CREATE INDEX idx_permissions_name ON permissions (name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS permissions;

DROP TABLE IF EXISTS scopes;

DROP TABLE IF EXISTS resources;
-- +goose StatementEnd