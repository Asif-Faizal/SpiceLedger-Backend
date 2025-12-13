-- Migration: Fix Admin Password
UPDATE users SET password_hash = '$2a$10$pJSge4ksLkYEM4yAazFQjeuG/aZDfBcnA53j1jJriyIV8UuJEUBgC' WHERE email = 'admin@example.com';
