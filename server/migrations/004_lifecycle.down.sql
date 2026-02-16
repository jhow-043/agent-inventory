DROP INDEX IF EXISTS idx_hw_history_device;
DROP INDEX IF EXISTS idx_devices_department;
DROP INDEX IF EXISTS idx_devices_status;

DROP TABLE IF EXISTS hardware_history;

ALTER TABLE devices
    DROP COLUMN IF EXISTS department_id,
    DROP COLUMN IF EXISTS status;

DROP TABLE IF EXISTS departments;
