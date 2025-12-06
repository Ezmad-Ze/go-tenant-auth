-- +goose Up
-- Create audit_logs table for compliance tracking
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    metadata JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'success',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs (tenant_id);

CREATE INDEX idx_audit_logs_user_id ON audit_logs (user_id);

CREATE INDEX idx_audit_logs_action ON audit_logs (action);

CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);

CREATE INDEX idx_audit_logs_resource ON audit_logs (resource_type, resource_id);

-- Composite index for common queries
CREATE INDEX idx_audit_logs_tenant_user_created ON audit_logs (
    tenant_id,
    user_id,
    created_at DESC
);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;