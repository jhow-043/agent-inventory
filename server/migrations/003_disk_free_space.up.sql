-- Add logical disk (partition) info: drive letter, free space, and total partition size.
ALTER TABLE disks
    ADD COLUMN drive_letter VARCHAR(5) NOT NULL DEFAULT '',
    ADD COLUMN partition_size_bytes BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN free_space_bytes BIGINT NOT NULL DEFAULT 0;
