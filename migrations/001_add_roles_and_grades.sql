-- Migration: Add Role to Users and Create Grades Table

-- 1. Add role column to users table if it doesn't exist
-- Using a DO block to safely add column only if missing (Postgres specific)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='role') THEN
        ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user';
    END IF;
END $$;

-- 2. Create grades table
CREATE TABLE IF NOT EXISTS grades (
    name VARCHAR(50) PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
