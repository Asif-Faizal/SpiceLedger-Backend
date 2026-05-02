-- Market FIFO Trading Engine Schema
-- user_id and spice_grade_id mirror the CHAR(27) IDs from the control service
-- (accounts.id → user_id, grade.id → spice_grade_id)
-- PKs on high-volume tables stay BIGINT AUTO_INCREMENT for insert performance.

SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS buy_lots;
DROP TABLE IF EXISTS sell_allocations;
DROP TABLE IF EXISTS positions;
SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE IF NOT EXISTS transactions (
  id             CHAR(27)     PRIMARY KEY,
  user_id        CHAR(27)     NOT NULL,
  spice_grade_id CHAR(27)     NOT NULL,
  type           ENUM('BUY','SELL') NOT NULL,
  quantity       DECIMAL(15,4) NOT NULL CHECK (quantity > 0),
  price          DECIMAL(15,4) NOT NULL CHECK (price > 0),
  trade_date     DATE          NOT NULL,
  created_at     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_txn_user_grade (user_id, spice_grade_id),
  INDEX idx_txn_trade_date (trade_date)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS buy_lots (
  id             CHAR(27)      PRIMARY KEY,
  transaction_id CHAR(27)      NOT NULL,
  user_id        CHAR(27)      NOT NULL,
  spice_grade_id CHAR(27)      NOT NULL,
  original_qty   DECIMAL(15,4) NOT NULL CHECK (original_qty > 0),
  remaining_qty  DECIMAL(15,4) NOT NULL CHECK (remaining_qty >= 0),
  price          DECIMAL(15,4) NOT NULL CHECK (price > 0),
  trade_date     DATE          NOT NULL,
  created_at     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

  -- FIFO query: user+grade filtered, then ordered by trade_date ASC
  INDEX idx_lots_fifo (user_id, spice_grade_id, remaining_qty, trade_date, id),

  CONSTRAINT fk_lot_transaction FOREIGN KEY (transaction_id) REFERENCES transactions(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS sell_allocations (
  id                  CHAR(27)      PRIMARY KEY,
  sell_transaction_id CHAR(27)      NOT NULL,
  buy_lot_id          CHAR(27)      NOT NULL,
  quantity            DECIMAL(15,4) NOT NULL CHECK (quantity > 0),
  buy_price           DECIMAL(15,4) NOT NULL CHECK (buy_price > 0),
  sell_price          DECIMAL(15,4) NOT NULL CHECK (sell_price > 0),
  realized_pnl        DECIMAL(15,4) NOT NULL,
  created_at          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_alloc_sell_txn (sell_transaction_id),
  INDEX idx_alloc_buy_lot  (buy_lot_id),

  CONSTRAINT fk_alloc_sell_txn FOREIGN KEY (sell_transaction_id) REFERENCES transactions(id),
  CONSTRAINT fk_alloc_buy_lot  FOREIGN KEY (buy_lot_id)          REFERENCES buy_lots(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS positions (
  user_id        CHAR(27)      NOT NULL,
  spice_grade_id CHAR(27)      NOT NULL,
  total_qty      DECIMAL(15,4) NOT NULL DEFAULT 0 CHECK (total_qty >= 0),
  total_cost     DECIMAL(15,4) NOT NULL DEFAULT 0 CHECK (total_cost >= 0),
  realized_pnl   DECIMAL(15,4) NOT NULL DEFAULT 0,
  updated_at     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (user_id, spice_grade_id)
) ENGINE=InnoDB;

-- Seed Data for Market Module
-- Market Transactions for Merchant (Last 2 days: T-2 and T-1)
INSERT INTO transactions (id, user_id, spice_grade_id, type, quantity, price, trade_date) VALUES
-- T-2: Buy 10kg Turmeric A @ 100
('txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY', 10.00, 100.00, CURRENT_DATE - INTERVAL 2 DAY),
-- T-1: Buy 5kg Turmeric A @ 110
('txn_buy_t1_tur_a_0000000002', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'BUY', 5.00, 110.00, CURRENT_DATE - INTERVAL 1 DAY),
-- T-1: Sell 3kg Turmeric A @ 130 (Different price action)
('txn_sell_t1_tur_a_000000003', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 'SELL', 3.00, 130.00, CURRENT_DATE - INTERVAL 1 DAY),
-- T-2: Buy 20kg Pepper A @ 200
('txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 'BUY', 20.00, 200.00, CURRENT_DATE - INTERVAL 2 DAY);

-- Market State: Buy Lots (Initial inventory from BUYs)
INSERT INTO buy_lots (id, transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date) VALUES
-- T-2 Buy Lot (7kg remaining after 3kg sold) - Carryover product
('lot_buy_t2_tur_a_0000000001', 'txn_buy_t2_tur_a_0000000001', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 10.00, 7.00, 100.00, CURRENT_DATE - INTERVAL 2 DAY),
-- T-1 Buy Lot (5kg remaining)
('lot_buy_t1_tur_a_0000000002', 'txn_buy_t1_tur_a_0000000002', 'acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 5.00, 5.00, 110.00, CURRENT_DATE - INTERVAL 1 DAY),
-- T-2 Buy Lot Pepper (20kg remaining)
('lot_buy_t2_pep_a_0000000001', 'txn_buy_t2_pep_a_0000000001', 'acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 20.00, 20.00, 200.00, CURRENT_DATE - INTERVAL 2 DAY);

-- Market State: Sell Allocations (Audit trail for the 3kg sold)
INSERT INTO sell_allocations (id, sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl) VALUES
('alloc_sell_t1_tur_a_0000001', 'txn_sell_t1_tur_a_000000003', 'lot_buy_t2_tur_a_0000000001', 3.00, 100.00, 130.00, 90.00);

-- Market State: Positions (Aggregated view)
INSERT INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl) VALUES
-- Turmeric A: 7kg + 5kg = 12kg. Cost: (7*100) + (5*110) = 1250. PnL: 90.
('acc_merchant_00000000000001', 'grd_turmeric_a_000000000001', 12.00, 1250.00, 90.00),
-- Pepper A: 20kg. Cost: 20*200 = 4000. PnL: 0.
('acc_merchant_00000000000001', 'grd_pepper_a_00000000000001', 20.00, 4000.00, 0.00);