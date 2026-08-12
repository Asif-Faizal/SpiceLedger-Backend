-- +goose Up
-- Rich merchant portfolio seed (~7 days) with profits AND losses for trends charts.
-- Merchant: acc_merchant_00000000000001

INSERT IGNORE INTO transactions (id, user_id, spice_grade_id, type, quantity, price, trade_date) VALUES
-- Day -7: buy turmeric A
('txn_buy_t7_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY',  20.00, 100.00, CURRENT_DATE - INTERVAL 7 DAY),
-- Day -6: buy turmeric B + pepper
('txn_buy_t6_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'BUY',  30.00,  80.00, CURRENT_DATE - INTERVAL 6 DAY),
('txn_buy_t6_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY',  25.00, 210.00, CURRENT_DATE - INTERVAL 6 DAY),
-- Day -5: LOSS sell turmeric A below cost + buy more pepper
('txn_sel_t5_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL',  8.00,  90.00, CURRENT_DATE - INTERVAL 5 DAY),
('txn_buy_t5_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY',  10.00, 200.00, CURRENT_DATE - INTERVAL 5 DAY),
-- Day -4: profit sell turmeric B + buy turmeric A
('txn_sel_t4_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'SELL', 12.00,  95.00, CURRENT_DATE - INTERVAL 4 DAY),
('txn_buy_t4_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY',  10.00, 105.00, CURRENT_DATE - INTERVAL 4 DAY),
-- Day -3: LOSS sell pepper + profit sell turmeric A
('txn_sel_t3_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'SELL', 10.00, 190.00, CURRENT_DATE - INTERVAL 3 DAY),
('txn_sel_t3_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL',  5.00, 120.00, CURRENT_DATE - INTERVAL 3 DAY),
('txn_buy_t3_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY',  15.00, 205.00, CURRENT_DATE - INTERVAL 3 DAY),
-- Day -2: buys only
('txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY',   8.00, 110.00, CURRENT_DATE - INTERVAL 2 DAY),
('txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY',  12.00, 215.00, CURRENT_DATE - INTERVAL 2 DAY),
-- Day -1: mixed — profit turmeric B, loss turmeric A
('txn_sel_t1_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'SELL',  8.00,  92.00, CURRENT_DATE - INTERVAL 1 DAY),
('txn_sel_t1_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL',  4.00,  95.00, CURRENT_DATE - INTERVAL 1 DAY),
('txn_buy_t1_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'BUY',  10.00,  85.00, CURRENT_DATE - INTERVAL 1 DAY),
-- Today: profit pepper + profit turmeric A
('txn_sel_t0_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'SELL',  6.00, 230.00, CURRENT_DATE),
('txn_sel_t0_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL',  3.00, 125.00, CURRENT_DATE),
('txn_buy_t0_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 'BUY',   5.00,  88.00, CURRENT_DATE);

INSERT IGNORE INTO buy_lots (id, transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date) VALUES
-- Turmeric A lots (FIFO)
('lot_buy_t7_tur_a_0000000001', 'txn_buy_t7_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 20.00,  0.00, 100.00, CURRENT_DATE - INTERVAL 7 DAY),
('lot_buy_t4_tur_a_0000000001', 'txn_buy_t4_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 10.00, 10.00, 105.00, CURRENT_DATE - INTERVAL 4 DAY),
('lot_buy_t2_tur_a_0000000001', 'txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001',  8.00,  8.00, 110.00, CURRENT_DATE - INTERVAL 2 DAY),
-- Turmeric B lots
('lot_buy_t6_tur_b_0000000001', 'txn_buy_t6_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 30.00, 10.00,  80.00, CURRENT_DATE - INTERVAL 6 DAY),
('lot_buy_t1_tur_b_0000000001', 'txn_buy_t1_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 10.00, 10.00,  85.00, CURRENT_DATE - INTERVAL 1 DAY),
('lot_buy_t0_tur_b_0000000001', 'txn_buy_t0_tur_b_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_b_000000000001',  5.00,  5.00,  88.00, CURRENT_DATE),
-- Pepper A lots
('lot_buy_t6_pep_a_0000000001', 'txn_buy_t6_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 25.00,  9.00, 210.00, CURRENT_DATE - INTERVAL 6 DAY),
('lot_buy_t5_pep_a_0000000001', 'txn_buy_t5_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 10.00, 10.00, 200.00, CURRENT_DATE - INTERVAL 5 DAY),
('lot_buy_t3_pep_a_0000000001', 'txn_buy_t3_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 15.00, 15.00, 205.00, CURRENT_DATE - INTERVAL 3 DAY),
('lot_buy_t2_pep_a_0000000001', 'txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 12.00, 12.00, 215.00, CURRENT_DATE - INTERVAL 2 DAY);

INSERT IGNORE INTO sell_allocations (id, sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl) VALUES
-- Day -5: LOSS turmeric A (90 - 100) * 8 = -80
('alloc_sel_t5_tur_a_0000001', 'txn_sel_t5_tur_a_0000000001', 'lot_buy_t7_tur_a_0000000001', 8.00, 100.00,  90.00, -80.00),
-- Day -4: PROFIT turmeric B (95 - 80) * 12 = 180
('alloc_sel_t4_tur_b_0000001', 'txn_sel_t4_tur_b_0000000001', 'lot_buy_t6_tur_b_0000000001', 12.00,  80.00,  95.00, 180.00),
-- Day -3: LOSS pepper (190 - 210) * 10 = -200
('alloc_sel_t3_pep_a_0000001', 'txn_sel_t3_pep_a_0000000001', 'lot_buy_t6_pep_a_0000000001', 10.00, 210.00, 190.00, -200.00),
-- Day -3: PROFIT turmeric A (120 - 100) * 5 = 100
('alloc_sel_t3_tur_a_0000001', 'txn_sel_t3_tur_a_0000000001', 'lot_buy_t7_tur_a_0000000001', 5.00, 100.00, 120.00, 100.00),
-- Day -1: PROFIT turmeric B (92 - 80) * 8 = 96
('alloc_sel_t1_tur_b_0000001', 'txn_sel_t1_tur_b_0000000001', 'lot_buy_t6_tur_b_0000000001', 8.00,  80.00,  92.00,  96.00),
-- Day -1: LOSS turmeric A (95 - 100) * 4 = -20  (uses remaining of t7 lot: 20-8-5=7, sell 4)
('alloc_sel_t1_tur_a_0000001', 'txn_sel_t1_tur_a_0000000001', 'lot_buy_t7_tur_a_0000000001', 4.00, 100.00,  95.00, -20.00),
-- Today: PROFIT pepper (230 - 210) * 6 = 120
('alloc_sel_t0_pep_a_0000001', 'txn_sel_t0_pep_a_0000000001', 'lot_buy_t6_pep_a_0000000001', 6.00, 210.00, 230.00, 120.00),
-- Today: PROFIT turmeric A (125 - 100) * 3 = 75  (lot t7 remaining after sells: 20-8-5-4=3)
('alloc_sel_t0_tur_a_0000001', 'txn_sel_t0_tur_a_0000000001', 'lot_buy_t7_tur_a_0000000001', 3.00, 100.00, 125.00,  75.00);

-- Positions from remaining lots:
-- Turmeric A: lot t4 10@105 + lot t2 8@110 = 18 kg, cost 1050+880=1930; realized -80+100-20+75 = 75
-- Turmeric B: lot t6 10@80 + lot t1 10@85 + lot t0 5@88 = 25 kg, cost 800+850+440=2090; realized 180+96 = 276
-- Pepper A: lot t6 9@210 + lot t5 10@200 + lot t3 15@205 + lot t2 12@215 = 46 kg;
--   cost 1890+2000+3075+2580=9545; realized -200+120 = -80
INSERT IGNORE INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl) VALUES
('acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 18.00, 1930.00, 75.00),
('acc_merchant_00000000000001', 'grd_turmeric_b_000000000001', 25.00, 2090.00, 276.00),
('acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 46.00, 9545.00, -80.00);

INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (5, 'market_seed', 'Seed data for market transactions, lots, allocations, and positions');

-- +goose Down
DELETE FROM sell_allocations WHERE id LIKE 'alloc_%';
DELETE FROM buy_lots WHERE id LIKE 'lot_%';
DELETE FROM positions WHERE user_id = 'acc_merchant_00000000000001';
DELETE FROM transactions WHERE user_id = 'acc_merchant_00000000000001';
