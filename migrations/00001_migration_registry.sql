-- +goose Up
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     BIGINT       NOT NULL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    applied_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (1, 'migration_registry', 'Creates schema_migrations audit table for tracking DB versions');

-- +goose Down
DROP TABLE IF EXISTS schema_migrations;
