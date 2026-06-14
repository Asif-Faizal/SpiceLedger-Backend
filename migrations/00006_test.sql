-- +goose Up
ALTER TABLE accounts ADD COLUMN status VARCHAR(20) DEFAULT 'active';
INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (6, 'account_status', 'Add status column to accounts');

-- +goose Down
ALTER TABLE accounts DROP COLUMN status;