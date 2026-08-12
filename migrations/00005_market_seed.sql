-- +goose Up
-- Rich merchant portfolio seed (~7 days) for dashboard, trends, filters, and FIFO demos.
-- Merchant: acc_merchant_00000000000001

INSERT IGNORE INTO transactions (id, user_id, spice_grade_id, type, quantity, price, trade_date) VALUES
-- Turmeric A
('txn_buy_t7_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY',  15.00,  95.00, CURRENT_DATE - INTERVAL 7 DAY),
('txn_sel_t4_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL',  5.00, 108.00, CURRENT_DATE - INTERVAL 4 DAY),
('txn_buy_t4_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY',   8.00, 105.00, CURRENT_DATE - INTERVAL 4 DAY),
('txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY',  10.00, 100.00, CURRENT_DATE - INTERVAL 2 DAY),
('txn_buy_t1_tur_a_0000000002', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY',   5.00, 110.00, CURRENT_DATE - INTERVAL 1 DAY),
('txn_sell_t1_tur_a_000000003', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL',  3.00, 130.00, CURRENT_DATE - INTERVAL 1 DAY),
('txn_sel_t0_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL',  2.00, 125.00, CURRENT_DATE),
-- Turmeric B
('txn_buy_t6_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'BUY',  25.00,  78.00, CURRENT_DATE - INTERVAL 6 DAY),
('txn_sel_t1_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'SELL', 10.00,  88.00, CURRENT_DATE - INTERVAL 1 DAY),
('txn_buy_t0_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'BUY',  12.00,  88.00, CURRENT_DATE),
-- Pepper A
('txn_buy_t5_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY',  30.00, 195.00, CURRENT_DATE - INTERVAL 5 DAY),
('txn_sel_t3_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'SELL', 10.00, 215.00, CURRENT_DATE - INTERVAL 3 DAY),
('txn_buy_t3_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY',  15.00, 205.00, CURRENT_DATE - INTERVAL 3 DAY),
('txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY',  20.00, 200.00, CURRENT_DATE - INTERVAL 2 DAY),
('txn_sel_t0_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'SELL',  5.00, 225.00, CURRENT_DATE);

INSERT IGNORE INTO buy_lots (id, transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date) VALUES
-- Turmeric A lots (FIFO order)
('lot_buy_t7_tur_a_0000000001', 'txn_buy_t7_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 15.00,  5.00,  95.00, CURRENT_DATE - INTERVAL 7 DAY),
('lot_buy_t4_tur_a_0000000001', 'txn_buy_t4_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001',  8.00,  8.00, 105.00, CURRENT_DATE - INTERVAL 4 DAY),
('lot_buy_t2_tur_a_0000000001', 'txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 10.00, 10.00, 100.00, CURRENT_DATE - INTERVAL 2 DAY),
('lot_buy_t1_tur_a_0000000002', 'txn_buy_t1_tur_a_0000000002', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001',  5.00,  5.00, 110.00, CURRENT_DATE - INTERVAL 1 DAY),
-- Turmeric B lots
('lot_buy_t6_tur_b_0000000001', 'txn_buy_t6_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 25.00, 15.00,  78.00, CURRENT_DATE - INTERVAL 6 DAY),
('lot_buy_t0_tur_b_0000000001', 'txn_buy_t0_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 12.00, 12.00,  88.00, CURRENT_DATE),
-- Pepper A lots
('lot_buy_t5_pep_a_0000000001', 'txn_buy_t5_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 30.00, 15.00, 195.00, CURRENT_DATE - INTERVAL 5 DAY),
('lot_buy_t3_pep_a_0000000001', 'txn_buy_t3_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 15.00, 15.00, 205.00, CURRENT_DATE - INTERVAL 3 DAY),
('lot_buy_t2_pep_a_0000000001', 'txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 20.00, 20.00, 200.00, CURRENT_DATE - INTERVAL 2 DAY);

INSERT IGNORE INTO sell_allocations (id, sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl) VALUES
-- Turmeric A sells (oldest lots first)
('alloc_sel_t4_tur_a_0000001', 'txn_sel_t4_tur_a_0000000001', 'lot_buy_t7_tur_a_0000000001', 5.00,  95.00, 108.00,  65.00),
('alloc_sell_t1_tur_a_0000001', 'txn_sell_t1_tur_a_000000003', 'lot_buy_t7_tur_a_0000000001', 3.00,  95.00, 130.00, 105.00),
('alloc_sel_t0_tur_a_0000001', 'txn_sel_t0_tur_a_0000000001', 'lot_buy_t7_tur_a_0000000001', 2.00,  95.00, 125.00,  60.00),
-- Turmeric B sell
('alloc_sel_t1_tur_b_0000001', 'txn_sel_t1_tur_b_0000000001', 'lot_buy_t6_tur_b_0000000001', 10.00, 78.00,  88.00, 100.00),
-- Pepper A sells
('alloc_sel_t3_pep_a_0000001', 'txn_sel_t3_pep_a_0000000001', 'lot_buy_t5_pep_a_0000000001', 10.00, 195.00, 215.00, 200.00),
('alloc_sel_t0_pep_a_0000001', 'txn_sel_t0_pep_a_0000000001', 'lot_buy_t5_pep_a_0000000001',  5.00, 195.00, 225.00, 150.00);

INSERT IGNORE INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl) VALUES
('acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 28.00, 2865.00, 230.00),
('acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 27.00, 2226.00, 100.00),
('acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 50.00, 10000.00, 350.00);

INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (5, 'market_seed', 'Seed data for market transactions, lots, allocations, and positions');

-- +goose Down
DELETE FROM sell_allocations WHERE id LIKE 'alloc_%';
DELETE FROM buy_lots WHERE id LIKE 'lot_%';
DELETE FROM positions WHERE user_id = 'acc_merchant_00000000000001';
DELETE FROM transactions WHERE user_id = 'acc_merchant_00000000000001';
