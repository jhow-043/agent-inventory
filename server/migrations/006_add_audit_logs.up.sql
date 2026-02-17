-- Create audit_logs table for tracking all important actions
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username TEXT NOT NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_composite ON audit_logs(user_id, created_at DESC);

-- Add comments for documentation
COMMENT ON TABLE audit_logs IS 'Audit trail of all important system actions';
COMMENT ON COLUMN audit_logs.action IS 'Action performed (e.g., auth.login, device.status.update)';
COMMENT ON COLUMN audit_logs.resource_type IS 'Type of resource affected (e.g., device, user, department)';
COMMENT ON COLUMN audit_logs.resource_id IS 'UUID of the affected resource (nullable for non-entity actions)';
COMMENT ON COLUMN audit_logs.details IS 'JSON object with action-specific details (old/new values, etc.)';
