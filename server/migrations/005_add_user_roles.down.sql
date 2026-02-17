-- Remove role column
ALTER TABLE users DROP COLUMN role;

-- Drop index
DROP INDEX IF EXISTS idx_users_role;
