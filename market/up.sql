CREATE TABLE transactions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  spice_grade_id BIGINT NOT NULL,
  type ENUM('BUY','SELL') NOT NULL,
  quantity DECIMAL(15,4) NOT NULL,
  price DECIMAL(15,4) NOT NULL,
  trade_date DATE NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

  INDEX(user_id, spice_grade_id),
  INDEX(trade_date)
)ENGINE=InnoDB;

CREATE TABLE buy_lots (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  transaction_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  spice_grade_id BIGINT NOT NULL,

  original_qty DECIMAL(15,4) NOT NULL,
  remaining_qty DECIMAL(15,4) NOT NULL,

  price DECIMAL(15,4) NOT NULL,
  trade_date DATE NOT NULL,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

  INDEX(user_id, spice_grade_id, trade_date)
)ENGINE=InnoDB;

CREATE TABLE sell_allocations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,

  sell_transaction_id BIGINT NOT NULL,
  buy_lot_id BIGINT NOT NULL,

  quantity DECIMAL(15,4) NOT NULL,
  buy_price DECIMAL(15,4) NOT NULL,
  sell_price DECIMAL(15,4) NOT NULL,

  realized_pnl DECIMAL(15,4) NOT NULL,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

  INDEX(sell_transaction_id),
  INDEX(buy_lot_id)
)ENGINE=InnoDB;

CREATE TABLE positions (
  user_id BIGINT NOT NULL,
  spice_grade_id BIGINT NOT NULL,

  total_qty DECIMAL(15,4) NOT NULL,
  total_cost DECIMAL(15,4) NOT NULL,
  realized_pnl DECIMAL(15,4) NOT NULL,

  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY(user_id, spice_grade_id)
)ENGINE=InnoDB;