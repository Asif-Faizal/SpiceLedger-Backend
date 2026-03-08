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

CREATE TABLE IF NOT EXISTS merchant_details (
    id CHAR(27) PRIMARY KEY,
    account_id CHAR(27) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    address VARCHAR(255) NOT NULL,
    city VARCHAR(50) NOT NULL,
    state VARCHAR(50) NOT NULL,
    pincode CHAR(6) NOT NULL,
    UNIQUE (account_id),
    FOREIGN KEY (account_id) REFERENCES accounts(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS products (
    id CHAR(27) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category ENUM('spice', 'others') NOT NULL,
    description TEXT,
    status ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
    UNIQUE (id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS grade (
    id CHAR(27) PRIMARY KEY,
    product_id CHAR(27) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
    UNIQUE (id),
    FOREIGN KEY (product_id) REFERENCES products(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS daily_price (
    id CHAR(27) PRIMARY KEY,
    product_id CHAR(27) NOT NULL,
    grade_id CHAR(27) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    date DATE NOT NULL,
    time TIME NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (grade_id) REFERENCES grade(id),
    UNIQUE (product_id, grade_id, date),
    INDEX idx_price_lookup (product_id, grade_id, date, time)
) ENGINE=InnoDB;