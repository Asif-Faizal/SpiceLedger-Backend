CREATE TABLE IF NOT EXISTS accounts (
    id CHAR(27) PRIMARY KEY,
    name VARCHAR(24),
    user_type ENUM('admin', 'merchant') NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id CHAR(27) PRIMARY KEY,
    account_id CHAR(27) NOT NULL,
    device_id CHAR(27) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_revoked TINYINT(1) NOT NULL DEFAULT 0,
    UNIQUE (account_id, device_id),
    FOREIGN KEY (account_id) REFERENCES accounts(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS merchants (
    id CHAR(27) PRIMARY KEY,
    account_id CHAR(27) NOT NULL,
    address VARCHAR(255) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
) ENGINE=InnoDB;
