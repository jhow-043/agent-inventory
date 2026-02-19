-- Revert structured change tracking columns from hardware_history.

DROP INDEX IF EXISTS idx_hardware_history_component;

ALTER TABLE hardware_history
    DROP COLUMN IF EXISTS component,
    DROP COLUMN IF EXISTS change_type,
    DROP COLUMN IF EXISTS field,
    DROP COLUMN IF EXISTS old_value,
    DROP COLUMN IF EXISTS new_value;
