-- +goose Up
INSERT IGNORE INTO transactions (id, user_id, spice_grade_id, type, quantity, price, trade_date) VALUES
('txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY', 10.00, 100.00, CURRENT_DATE - INTERVAL 2 DAY),
('txn_buy_t1_tur_a_0000000002', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY', 5.00, 110.00, CURRENT_DATE - INTERVAL 1 DAY),
('txn_sell_t1_tur_a_000000003', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL', 3.00, 130.00, CURRENT_DATE - INTERVAL 1 DAY),
('txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY', 20.00, 200.00, CURRENT_DATE - INTERVAL 2 DAY);

INSERT IGNORE INTO buy_lots (id, transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date) VALUES
('lot_buy_t2_tur_a_0000000001', 'txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 10.00, 7.00, 100.00, CURRENT_DATE - INTERVAL 2 DAY),
('lot_buy_t1_tur_a_0000000002', 'txn_buy_t1_tur_a_0000000002', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 5.00, 5.00, 110.00, CURRENT_DATE - INTERVAL 1 DAY),
('lot_buy_t2_pep_a_0000000001', 'txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 20.00, 20.00, 200.00, CURRENT_DATE - INTERVAL 2 DAY);

INSERT IGNORE INTO sell_allocations (id, sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl) VALUES
('alloc_sell_t1_tur_a_0000001', 'txn_sell_t1_tur_a_000000003', 'lot_buy_t2_tur_a_0000000001', 3.00, 100.00, 130.00, 90.00);

INSERT IGNORE INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl) VALUES
('acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 12.00, 1250.00, 90.00),
('acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 20.00, 4000.00, 0.00);

INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (5, 'market_seed', 'Seed data for market transactions, lots, allocations, and positions');

-- +goose Down
DELETE FROM sell_allocations WHERE id LIKE 'alloc_%';
DELETE FROM buy_lots WHERE id LIKE 'lot_%';
DELETE FROM positions WHERE user_id = 'acc_merchant_00000000000001';
DELETE FROM transactions WHERE id LIKE 'txn_%';
