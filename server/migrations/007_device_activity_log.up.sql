-- Create device_activity_log table for tracking device changes
-- (user logins, software installs/uninstalls, OS changes, etc.)
CREATE TABLE device_activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    activity_type TEXT NOT NULL,      -- 'user_login', 'software_installed', 'software_removed', 'os_updated', 'boot'
    description TEXT NOT NULL,        -- Human-readable description
    old_value TEXT,                   -- Previous value (e.g., old user, old OS version)
    new_value TEXT,                   -- New value
    metadata JSONB,                   -- Extra structured data (e.g., software details)
    detected_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_device_activity_device_id ON device_activity_log(device_id);
CREATE INDEX idx_device_activity_type ON device_activity_log(activity_type);
CREATE INDEX idx_device_activity_detected_at ON device_activity_log(detected_at DESC);
CREATE INDEX idx_device_activity_device_time ON device_activity_log(device_id, detected_at DESC);

COMMENT ON TABLE device_activity_log IS 'Tracks changes detected during inventory submissions';
COMMENT ON COLUMN device_activity_log.activity_type IS 'Type of change: user_login, software_installed, software_removed, os_updated, boot';
