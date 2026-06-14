-- +goose Up
-- Passwords are 'secret123'
INSERT IGNORE INTO accounts (id, name, user_type, email, password) VALUES
('acc_admin_00000000000000001', 'Admin User', 'admin', 'admin@spice.com', '$2a$10$DsDmsn.LYDL5UWOF33SVtOMVIo8vd08v8zQuDyKDCmm0Z9DtfYofW'),
('acc_merchant_00000000000001', 'Merchant User', 'merchant', 'merchant@spice.com', '$2a$10$DsDmsn.LYDL5UWOF33SVtOMVIo8vd08v8zQuDyKDCmm0Z9DtfYofW');

INSERT IGNORE INTO merchant_details (id, account_id, phone_number, address, city, state, pincode) VALUES
('det_merchant_00000000000001', 'acc_merchant_00000000000001', '1234567890', '123 Spice Market', 'Cochin', 'Kerala', '682001');

INSERT IGNORE INTO products (id, name, category, description, status) VALUES
('prd_turmeric_00000000000001', 'Turmeric', 'spice', 'Premium quality turmeric finger', 'active'),
('prd_pepper_000000000000001', 'Black Pepper', 'spice', 'Malabar bold black pepper', 'active');

INSERT IGNORE INTO grade (id, product_id, name, description, status) VALUES
('grd_turmeric_a_000000000001', 'prd_turmeric_00000000000001', 'Grade A', 'Highest curcumin content', 'active'),
('grd_turmeric_b_000000000001', 'prd_turmeric_00000000000001', 'Grade B', 'Standard quality', 'active'),
('grd_pepper_a_00000000000001', 'prd_pepper_000000000000001', 'Grade A', 'Extra bold', 'active');

INSERT IGNORE INTO daily_price (id, product_id, grade_id, price, date, time) VALUES
('prc_t2_turmeric_a_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_a_000000000001', 100.00, CURRENT_DATE - INTERVAL 2 DAY, '10:00:00'),
('prc_t2_turmeric_b_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_b_000000000001', 80.00, CURRENT_DATE - INTERVAL 2 DAY, '10:00:00'),
('prc_t2_pepper_a_00000000001', 'prd_pepper_000000000000001', 'grd_pepper_a_00000000000001', 200.00, CURRENT_DATE - INTERVAL 2 DAY, '10:00:00'),
('prc_t1_turmeric_a_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_a_000000000001', 110.00, CURRENT_DATE - INTERVAL 1 DAY, '10:00:00'),
('prc_t1_turmeric_b_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_b_000000000001', 85.00, CURRENT_DATE - INTERVAL 1 DAY, '10:00:00'),
('prc_t1_pepper_a_00000000001', 'prd_pepper_000000000000001', 'grd_pepper_a_00000000000001', 210.00, CURRENT_DATE - INTERVAL 1 DAY, '10:00:00'),
('prc_t0_turmeric_a_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_a_000000000001', 120.00, CURRENT_DATE, '10:00:00'),
('prc_t0_turmeric_b_000000001', 'prd_turmeric_00000000000001', 'grd_turmeric_b_000000000001', 90.00, CURRENT_DATE, '10:00:00'),
('prc_t0_pepper_a_00000000001', 'prd_pepper_000000000000001', 'grd_pepper_a_00000000000001', 220.00, CURRENT_DATE, '10:00:00');

INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (3, 'control_seed', 'Seed data for accounts, products, grades, and daily prices');

-- +goose Down
DELETE FROM daily_price WHERE id LIKE 'prc_%';
DELETE FROM grade WHERE id LIKE 'grd_%';
DELETE FROM products WHERE id LIKE 'prd_%';
DELETE FROM merchant_details WHERE id LIKE 'det_%';
DELETE FROM accounts WHERE id IN ('acc_admin_00000000000000001', 'acc_merchant_00000000000001');
