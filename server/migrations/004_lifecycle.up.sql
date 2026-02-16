-- Phase 3: Device lifecycle, departments, hardware history
-- Add status column and department support to devices.
-- Create departments and hardware_history tables.

CREATE TABLE departments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

ALTER TABLE devices
    ADD COLUMN status        VARCHAR(20) NOT NULL DEFAULT 'active',
    ADD COLUMN department_id UUID REFERENCES departments(id) ON DELETE SET NULL;

CREATE TABLE hardware_history (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id  UUID         NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    snapshot   JSONB        NOT NULL,
    changed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_devices_status     ON devices(status);
CREATE INDEX idx_devices_department ON devices(department_id);
CREATE INDEX idx_hw_history_device  ON hardware_history(device_id);
