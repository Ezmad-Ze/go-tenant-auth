-- +goose Up
-- +goose StatementBegin

-- Create totp_2fa table for two-factor authentication
CREATE TABLE totp_2fa (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    secret VARCHAR(255) NOT NULL,
    backup_codes JSONB, -- Array of hashed backup codes
    is_enabled BOOLEAN DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id)
);

-- Create indexes
CREATE INDEX idx_totp_2fa_user_id ON totp_2fa (user_id);

CREATE INDEX idx_totp_2fa_tenant_id ON totp_2fa (tenant_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS totp_2fa;
-- +goose StatementEnd