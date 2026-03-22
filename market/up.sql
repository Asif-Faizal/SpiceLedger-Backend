-- Market FIFO Trading Engine Schema
-- user_id and spice_grade_id mirror the CHAR(27) IDs from the control service
-- (accounts.id → user_id, grade.id → spice_grade_id)
-- PKs on high-volume tables stay BIGINT AUTO_INCREMENT for insert performance.

CREATE TABLE IF NOT EXISTS transactions (
  id             BIGINT       PRIMARY KEY AUTO_INCREMENT,
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
  id             BIGINT        PRIMARY KEY AUTO_INCREMENT,
  transaction_id BIGINT        NOT NULL,
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
  id                  BIGINT        PRIMARY KEY AUTO_INCREMENT,
  sell_transaction_id BIGINT        NOT NULL,
  buy_lot_id          BIGINT        NOT NULL,
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