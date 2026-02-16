ALTER TABLE disks
    DROP COLUMN IF EXISTS drive_letter,
    DROP COLUMN IF EXISTS partition_size_bytes,
    DROP COLUMN IF EXISTS free_space_bytes;
