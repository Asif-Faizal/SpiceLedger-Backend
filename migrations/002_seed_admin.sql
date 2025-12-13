-- Migration: Seed Admin User

-- Insert admin user if not exists
-- Password hash is for 'admin123' (bcrypt default cost)
-- $2a$10$7/.. (Example hash, I will use a valid one below)

INSERT INTO users (id, email, name, role, password_hash, created_at, updated_at)
SELECT 
    gen_random_uuid(), 
    'admin@example.com', 
    'Admin User', 
    'admin', 
    '$2a$10$pJSge4ksLkYEM4yAazFQjeuG/aZDfBcnA53j1jJriyIV8UuJEUBgC', -- Hash for 'admin123'
    NOW(), 
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'admin@example.com'
);
