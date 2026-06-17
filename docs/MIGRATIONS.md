# Database Migrations

SpiceLedger uses [goose](https://github.com/pressly/goose) for versioned SQL migrations. Files live in [`migrations/`](../migrations/) and are applied by [`cmd/migrate/main.go`](../cmd/migrate/main.go).

## Two tracking tables

| Table | Purpose |
|-------|---------|
| `goose_migrations` | Managed by goose — tracks which migration **files** have been applied |
| `schema_migrations` | Application audit log — human-readable version, name, description |

Each migration SQL file should insert a row into `schema_migrations` in its `-- +goose Up` section so you can query migration history from the app layer.

---

## Current migration files

| Version | File | Description |
|---------|------|-------------|
| 1 | `00001_migration_registry.sql` | Creates `schema_migrations` audit table |
| 2 | `00002_control_schema.sql` | Control tables: accounts, sessions, merchant_details, products, grade, daily_price |
| 3 | `00003_control_seed.sql` | Seed admin/merchant users, products, grades, sample daily prices |
| 4 | `00004_market_schema.sql` | Market tables: transactions, buy_lots, sell_allocations, positions |
| 5 | `00005_market_seed.sql` | Sample buy/sell transactions, FIFO lots, positions |
| 6 | `00006_test.sql` | Adds `status` column to `accounts` |

Deprecated (do not use): `control/up.sql`, `market/up.sql` — replaced by numbered migrations.

---

## Running migrations

### Via Makefile (Docker)

Requires the `db` container to be running:

```bash
make up-db              # start MySQL only
make migrate-up         # apply pending
make migrate-status     # list state
make migrate-version    # print current version
make migrate-down       # rollback last migration
```

These run `docker compose --profile infra run --rm migrate <command>`.

### Via init script (local or Docker-aware)

```bash
make db-init            # scripts/init-db.sh local
```

`init-db.sh` behavior:

- Waits for MySQL to accept connections
- Creates `spice_ledger` database if missing
- If `.env` has `DB_HOST=db`, connects via `127.0.0.1` and `DB_HOST_PORT` (default **3307**) from the host
- Runs migrations through Docker when targeting Docker MySQL; otherwise `go run ./cmd/migrate/main.go up`

### Direct CLI (host, local MySQL only)

```bash
# .env must have DB_HOST=localhost
go run ./cmd/migrate/main.go up
go run ./cmd/migrate/main.go status
go run ./cmd/migrate/main.go version
go run ./cmd/migrate/main.go down
```

---

## Docker persistence

- `make down` stops containers but **keeps data** in the named volume `spice_ledger_mysql_data`.
- `make down-volumes` deletes that volume and all data.
- On `make up-full`, the `migrate` service runs once and applies only **pending** migrations — existing data is not re-seeded.

---

## Adding a new migration

1. Create the next numbered file, e.g. `migrations/00007_add_foo.sql`:

```sql
-- +goose Up
ALTER TABLE accounts ADD COLUMN foo VARCHAR(50) DEFAULT '';
INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (7, 'add_foo', 'Adds foo column to accounts');

-- +goose Down
ALTER TABLE accounts DROP COLUMN foo;
```

2. Apply:

```bash
make migrate-up
```

3. Verify:

```bash
make migrate-status
make migrate-version
```

### Rules

- Use `-- +goose Up` and `-- +goose Down` sections in every file.
- File names must sort lexicographically (`00007_...` after `00006_...`).
- Prefer `INSERT IGNORE` for `schema_migrations` rows so re-runs are safe.
- Never edit a migration that has already been applied in shared environments — add a new file instead.

---

## Seed credentials (from migration 3)

| Role | Email | Password |
|------|-------|----------|
| Admin | `admin@spice.com` | `secret123` |
| Merchant | `merchant@spice.com` | `secret123` |

Seed IDs (used in Bruno `local` environment):

- Admin account: `acc_admin_00000000000000001`
- Merchant account: `acc_merchant_00000000000001`
- Turmeric product: `prd_turmeric_00000000000001`
- Turmeric grade A: `grd_turmeric_a_000000000001`
