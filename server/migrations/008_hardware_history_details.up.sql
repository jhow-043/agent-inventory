-- Add structured change tracking columns to hardware_history.
-- Previously only a JSONB snapshot was stored; now each change gets
-- component, field, change_type, old_value, and new_value for easy querying.

ALTER TABLE hardware_history
    ADD COLUMN component   VARCHAR(50),   -- cpu, ram, motherboard, bios, disk, network
    ADD COLUMN change_type VARCHAR(50),   -- changed, added, removed
    ADD COLUMN field       VARCHAR(100),  -- e.g. total_bytes, model, serial
    ADD COLUMN old_value   TEXT,
    ADD COLUMN new_value   TEXT;

CREATE INDEX idx_hardware_history_component
    ON hardware_history (device_id, component, changed_at DESC);
