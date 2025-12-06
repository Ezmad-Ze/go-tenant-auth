-- +goose Up
-- +goose StatementBegin

-- Create devices table for device tracking
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    device_name VARCHAR(255),
    device_type VARCHAR(50), -- mobile, desktop, tablet
    fingerprint VARCHAR(255) NOT NULL,
    user_agent TEXT,
    ip_address INET,
    is_trusted BOOLEAN DEFAULT false,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, fingerprint)
);

-- Create indexes
CREATE INDEX idx_devices_user_id ON devices (user_id);

CREATE INDEX idx_devices_tenant_id ON devices (tenant_id);

CREATE INDEX idx_devices_fingerprint ON devices (fingerprint);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS devices;
-- +goose StatementEnd