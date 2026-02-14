-- Remote access tools (TeamViewer, AnyDesk, RustDesk, etc.)
CREATE TABLE remote_tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    tool_name VARCHAR(100) NOT NULL DEFAULT '',
    remote_id VARCHAR(255) NOT NULL DEFAULT '',
    version VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_remote_tools_device_id ON remote_tools(device_id);
