-- Add role column to users table
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'viewer';

-- Update existing users to admin
UPDATE users SET role = 'admin';

-- Add check constraint for valid roles
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'viewer'));

-- Create index on role for filtering
CREATE INDEX idx_users_role ON users(role);
