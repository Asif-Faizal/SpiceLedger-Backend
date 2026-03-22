# Market — FIFO Trading Engine

The market module implements a **FIFO (First-In, First-Out) cost-basis trading engine** for spice grades. Every BUY creates inventory; every SELL consumes the oldest inventory first.

---

## Tables at a Glance

| Table | Role |
|---|---|
| `transactions` | Immutable log of every BUY / SELL event |
| `buy_lots` | Inventory units created on BUY; depleted on SELL |
| `sell_allocations` | FIFO audit trail — maps each SELL qty to a specific BuyLot |
| `positions` | Live aggregate state (qty held, cost basis, realised P&L) |
| `daily_price` *(control DB)* | Single canonical market price per grade per day — used for unrealised P&L |

---

## P&L — Two Types

### Realized P&L
Locked in permanently when a SELL happens. Stored in `sell_allocations` and accumulated in `positions`.

```
realized_pnl = (sell_price - buy_price) × quantity_sold
```

`sell_price` is the price from the SELL transaction itself (the price the user agreed to trade at on that day).

### Unrealized P&L
Computed on-the-fly; **never written to the DB**. Reflects what the position would be worth if closed today.

```
today_price   = daily_price WHERE grade_id = ? AND date = TODAY()
avg_cost      = positions.total_cost / positions.total_qty
unrealized_pnl = (today_price - avg_cost) × positions.total_qty
```

The service layer calls `GetDailyPrice(gradeID, today)` from the repository and computes this in memory before returning to the caller. `ErrNoPriceAvailable` is returned when the price for today has not been published yet.

---

## Flow: BUY

**Steps:**
1. Insert a row into `transactions` (type = `BUY`)
2. Insert a row into `buy_lots` (`original_qty = remaining_qty = purchased qty`)
3. Upsert `positions` — add qty and cost

### Example 1 — User 1 buys 10 kg of Grade `grd_01` at ₹200/kg on 2024-01-01

```sql
INSERT INTO transactions (user_id, spice_grade_id, type, quantity, price, trade_date)
VALUES (1, 5, 'BUY', 10.0000, 200.0000, '2024-01-01');
-- → id = 1

INSERT INTO buy_lots (transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date)
VALUES (1, 1, 5, 10.0000, 10.0000, 200.0000, '2024-01-01');
-- → id = 1

INSERT INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl)
VALUES (1, 5, 10.0000, 2000.0000, 0.0000)
ON DUPLICATE KEY UPDATE
  total_qty  = total_qty + 10.0000,
  total_cost = total_cost + 2000.0000;
```

**State:**

`buy_lots`
| id | original_qty | remaining_qty | price | trade_date |
|---|---|---|---|---|
| 1 | 10.0000 | 10.0000 | 200.0000 | 2024-01-01 |

`positions`
| total_qty | total_cost | realized_pnl |
|---|---|---|
| 10.0000 | 2000.0000 | 0.0000 |

---

### Example 2 — Same user buys 6 kg at ₹220/kg on 2024-01-05

```sql
INSERT INTO transactions (user_id, spice_grade_id, type, quantity, price, trade_date)
VALUES (1, 5, 'BUY', 6.0000, 220.0000, '2024-01-05');
-- → id = 2

INSERT INTO buy_lots (transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date)
VALUES (2, 1, 5, 6.0000, 6.0000, 220.0000, '2024-01-05');
-- → id = 2

INSERT INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl)
VALUES (1, 5, 6.0000, 1320.0000, 0.0000)
ON DUPLICATE KEY UPDATE
  total_qty  = total_qty + 6.0000,
  total_cost = total_cost + 1320.0000;
```

**State:**

`buy_lots`
| id | original_qty | remaining_qty | price | trade_date |
|---|---|---|---|---|
| 1 | 10.0000 | 10.0000 | 200.0000 | 2024-01-01 |
| 2 | 6.0000 | 6.0000 | 220.0000 | 2024-01-05 |

`positions`
| total_qty | total_cost | realized_pnl |
|---|---|---|
| 16.0000 | 3320.0000 | 0.0000 |

---

## Flow: SELL (FIFO matching)

A SELL consumes `buy_lots` in **trade_date ASC** order (oldest first). The entire operation runs inside a **single DB transaction**.

**Steps:**
1. Insert a row into `transactions` (type = `SELL`)
2. Lock open lots with `SELECT … FOR UPDATE`
3. Walk lots oldest→newest, consuming `remaining_qty` until sell qty is filled
4. For each lot touched: `DeductBuyLotQty` (UPDATE with guard)
5. For each lot touched: `InsertSellAllocation` with per-lot `realized_pnl = (sell_price - buy_price) × consumed_qty`
6. Upsert `positions` — decrease qty + cost, accumulate `realized_pnl`

### Example 3 — Sell 7 kg at ₹250/kg on 2024-01-10

Open lots (FIFO): Lot 1 (10 kg @ ₹200) → consumed first. 7 kg fits entirely in Lot 1.

```sql
INSERT INTO transactions (user_id, spice_grade_id, type, quantity, price, trade_date)
VALUES (1, 5, 'SELL', 7.0000, 250.0000, '2024-01-10');
-- → id = 3

-- Lock open lots (FOR UPDATE inside DB transaction)

-- Deduct 7 from Lot 1 (3 remains)
UPDATE buy_lots SET remaining_qty = remaining_qty - 7.0000
WHERE id = 1 AND remaining_qty >= 7.0000;

-- FIFO allocation: realized_pnl = (250 - 200) × 7 = 350
INSERT INTO sell_allocations (sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl)
VALUES (3, 1, 7.0000, 200.0000, 250.0000, 350.0000);

-- Position: qty 16→9, cost 3320→1920 (removed 200×7=1400)
INSERT INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl)
VALUES (1, 5, -7.0000, -1400.0000, 350.0000)
ON DUPLICATE KEY UPDATE
  total_qty    = total_qty - 7.0000,
  total_cost   = total_cost - 1400.0000,
  realized_pnl = realized_pnl + 350.0000;
```

**State:**

`buy_lots`
| id | original_qty | remaining_qty | price | trade_date |
|---|---|---|---|---|
| 1 | 10.0000 | **3.0000** | 200.0000 | 2024-01-01 |
| 2 | 6.0000 | 6.0000 | 220.0000 | 2024-01-05 |

`sell_allocations`
| id | sell_txn_id | buy_lot_id | qty | buy_price | sell_price | realized_pnl |
|---|---|---|---|---|---|---|
| 1 | 3 | 1 | 7.0000 | 200.0000 | 250.0000 | 350.0000 |

`positions`
| total_qty | total_cost | realized_pnl |
|---|---|---|
| 9.0000 | 1920.0000 | 350.0000 |

---

### Example 4 — Sell 5 kg at ₹260/kg on 2024-01-15 (spans two lots)

Open lots (FIFO): Lot 1 has **3 kg** left → exhausted. Then 2 kg taken from Lot 2.

```sql
INSERT INTO transactions (user_id, spice_grade_id, type, quantity, price, trade_date)
VALUES (1, 5, 'SELL', 5.0000, 260.0000, '2024-01-15');
-- → id = 4

-- Lot 1: consume all 3 remaining kg
-- realized_pnl = (260 - 200) × 3 = 180
UPDATE buy_lots SET remaining_qty = remaining_qty - 3.0000
WHERE id = 1 AND remaining_qty >= 3.0000;

INSERT INTO sell_allocations (sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl)
VALUES (4, 1, 3.0000, 200.0000, 260.0000, 180.0000);

-- Lot 2: consume 2 kg (5 needed − 3 done = 2 left)
-- realized_pnl = (260 - 220) × 2 = 80
UPDATE buy_lots SET remaining_qty = remaining_qty - 2.0000
WHERE id = 2 AND remaining_qty >= 2.0000;

INSERT INTO sell_allocations (sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl)
VALUES (4, 2, 2.0000, 220.0000, 260.0000, 80.0000);

-- Position: qty 9→4, cost 1920 − (200×3 + 220×2) = 1920 − 1040 = 880
INSERT INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl)
VALUES (1, 5, -5.0000, -1040.0000, 260.0000)
ON DUPLICATE KEY UPDATE
  total_qty    = total_qty - 5.0000,
  total_cost   = total_cost - 1040.0000,
  realized_pnl = realized_pnl + 260.0000;
```

**State:**

`buy_lots`
| id | original_qty | remaining_qty | price | trade_date |
|---|---|---|---|---|
| 1 | 10.0000 | **0.0000** | 200.0000 | 2024-01-01 |
| 2 | 6.0000 | **4.0000** | 220.0000 | 2024-01-05 |

`sell_allocations`
| id | sell_txn_id | buy_lot_id | qty | buy_price | sell_price | realized_pnl |
|---|---|---|---|---|---|---|
| 1 | 3 | 1 | 7.0000 | 200.0000 | 250.0000 | 350.0000 |
| 2 | 4 | 1 | 3.0000 | 200.0000 | 260.0000 | 180.0000 |
| 3 | 4 | 2 | 2.0000 | 220.0000 | 260.0000 | 80.0000 |

`positions`
| total_qty | total_cost | realized_pnl |
|---|---|---|
| 4.0000 | 880.0000 | **610.0000** |

---

## Unrealized P&L — Calculated at Read Time

After all trades above, the user still holds **4 kg** with an average cost of ₹220/kg (`880 / 4`).

The service computes unrealized P&L **at read time** (not stored in DB):

```sql
-- Fetch today's canonical price from daily_price (one row per grade per date)
SELECT price FROM daily_price
WHERE grade_id = 'grd_01' AND date = '2024-01-20'
LIMIT 1;
-- → 240.00
```

```
avg_cost       = total_cost / total_qty  =  880 / 4  =  220.00
unrealized_pnl = (today_price - avg_cost) × total_qty
               = (240.00 - 220.00) × 4
               = 80.00
```

**Combined portfolio view returned to the caller:**

| Field | Value |
|---|---|
| `total_qty` | 4.0000 kg |
| `avg_cost` | ₹220.00/kg |
| `today_price` | ₹240.00/kg |
| `unrealized_pnl` | **₹80.00** |
| `realized_pnl` | **₹610.00** |
| `total_pnl` | **₹690.00** |

> If `GetDailyPrice` returns `ErrNoPriceAvailable` (price not yet published for today), the service returns the position without unrealized P&L rather than failing the request.

---

## Oversell Prevention

| Layer | Mechanism |
|---|---|
| **Service** | Sums `remaining_qty` from `GetOpenBuyLots` before executing; errors if total < sell qty |
| **Database** | `UPDATE … WHERE remaining_qty >= ?` — returns `ErrInsufficientLotQty` if `RowsAffected = 0` |

---

## Key Invariants

| Invariant | Enforced by |
|---|---|
| `remaining_qty` never goes negative | `WHERE remaining_qty >= ?` in UPDATE |
| FIFO ordering | `ORDER BY trade_date ASC, id ASC FOR UPDATE` |
| Atomicity across all tables | Caller wraps in `BeginTx` / `Commit` / `Rollback` |
| `transactions` are immutable | INSERT only — never updated |
| `sell_allocations` are immutable | INSERT only — never updated |
| `realized_pnl` accumulates | `realized_pnl = realized_pnl + VALUES(realized_pnl)` |
| `unrealized_pnl` is never persisted | Computed in service layer at read time from `daily_price` |
