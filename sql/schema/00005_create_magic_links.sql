-- +goose Up
-- +goose StatementBegin

-- Create magic_links table for passwordless authentication
CREATE TABLE magic_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    is_used BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_magic_links_user_id ON magic_links (user_id);

CREATE INDEX idx_magic_links_tenant_id ON magic_links (tenant_id);

CREATE INDEX idx_magic_links_email ON magic_links (email);

CREATE INDEX idx_magic_links_token_hash ON magic_links (token_hash);

CREATE INDEX idx_magic_links_expires_at ON magic_links (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS magic_links;
-- +goose StatementEnd