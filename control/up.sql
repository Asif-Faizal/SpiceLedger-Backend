SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS merchant_details;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS grade;
DROP TABLE IF EXISTS daily_price;
SET FOREIGN_KEY_CHECKS = 1;

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

-- Seed Data
-- Passwords are 'secret123'
INSERT INTO accounts (id, name, user_type, email, password) VALUES
('acc_admin_00000000000000001', 'Admin User', 'admin', 'admin@spice.com', '$2a$10$DsDmsn.LYDL5UWOF33SVtOMVIo8vd08v8zQuDyKDCmm0Z9DtfYofW'),
('acc_merchant_00000000000001', 'Merchant User', 'merchant', 'merchant@spice.com', '$2a$10$DsDmsn.LYDL5UWOF33SVtOMVIo8vd08v8zQuDyKDCmm0Z9DtfYofW');

INSERT INTO merchant_details (id, account_id, phone_number, address, city, state, pincode) VALUES
('det_merchant_00000000000001', 'acc_merchant_00000000000001', '1234567890', '123 Spice Market', 'Cochin', 'Kerala', '682001');

INSERT INTO products (id, name, category, description, status) VALUES
('prd_turmeric_00000000000001', 'Turmeric', 'spice', 'Premium quality turmeric finger', 'active'),
('prd_pepper_000000000000001', 'Black Pepper', 'spice', 'Malabar bold black pepper', 'active');

INSERT INTO grade (id, product_id, name, description, status) VALUES
('grd_turmeric_a_000000000001', 'prd_turmeric_00000000000001', 'Grade A', 'Highest curcumin content', 'active'),
('grd_turmeric_b_000000000001', 'prd_turmeric_00000000000001', 'Grade B', 'Standard quality', 'active'),
('grd_pepper_a_00000000000001', 'prd_pepper_000000000000001', 'Grade A', 'Extra bold', 'active');

-- Daily Prices for past 3 days (T-2, T-1, Today)
INSERT INTO daily_price (id, product_id, grade_id, price, date, time) VALUES
-- T-2
('prc_t2_turmeric_a_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_a_000000000001', 100.00, CURRENT_DATE - INTERVAL 2 DAY, '10:00:00'),
('prc_t2_turmeric_b_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_b_000000000001', 80.00, CURRENT_DATE - INTERVAL 2 DAY, '10:00:00'),
('prc_t2_pepper_a_00000000001', 'prd_pepper_000000000000001', 'grd_pepper_a_00000000000001', 200.00, CURRENT_DATE - INTERVAL 2 DAY, '10:00:00'),
-- T-1
('prc_t1_turmeric_a_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_a_000000000001', 110.00, CURRENT_DATE - INTERVAL 1 DAY, '10:00:00'),
('prc_t1_turmeric_b_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_b_000000000001', 85.00, CURRENT_DATE - INTERVAL 1 DAY, '10:00:00'),
('prc_t1_pepper_a_00000000001', 'prd_pepper_000000000000001', 'grd_pepper_a_00000000000001', 210.00, CURRENT_DATE - INTERVAL 1 DAY, '10:00:00'),
-- T (Today)
('prc_t0_turmeric_a_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_a_000000000001', 120.00, CURRENT_DATE, '10:00:00'),
('prc_t0_turmeric_b_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_b_000000000001', 90.00, CURRENT_DATE, '10:00:00'),
('prc_t0_pepper_a_00000000001', 'prd_pepper_000000000000001', 'grd_pepper_a_00000000000001', 220.00, CURRENT_DATE, '10:00:00');
