# SpiceLedger Backend

SpiceLedger is a microservices-based backend for managing accounts, ledger operations, and market trading.

## Architecture

```
Client → Proxy (:8080) → REST (:8082) → Control gRPC (:50051) → MySQL
                       → GraphQL (:8081) → Control + Market gRPC (:50052) → MySQL
```

| Service | Role | Port |
|---------|------|------|
| **control** | Accounts, auth, products, prices | 50051 |
| **market** | FIFO trading engine | 50052 |
| **rest** | REST → gRPC gateway | 8082 |
| **graphql** | GraphQL → gRPC gateway | 8081 |
| **proxy** | Unified HTTP entrypoint | 8080 |

---

## Quick Start

### Option A — Local MySQL (recommended for active development)

```bash
brew services start mysql          # if not already running
make setup                         # creates .env + initializes DB
make run-control                   # or use VS Code "Full Stack (local)" compound
```

Run all services locally with VS Code: **Run and Debug → Full Stack (local)**

### Option B — Docker (full stack, team parity)

```bash
make install-docker                # installs Docker Desktop (macOS)
open -a Docker                     # start Docker Desktop, wait until running
make setup-docker                  # copies .env.docker.example → .env
make up-full                       # builds and starts all 6 containers
```

Docker MySQL is exposed on host port **3307** (not 3306) so it can run alongside Homebrew MySQL. Connect with: `mysql -h 127.0.0.1 -P 3307 -u root -p`

Database-only in Docker (services run natively):

```bash
make up-db                         # docker compose --profile infra up db -d
make db-init                       # apply schemas to local or Docker MySQL
```

### Useful commands

```bash
make help          # all targets
make down          # stop Docker stack
make logs          # tail container logs
make db-reset      # re-apply schemas (destructive)
make test          # go test ./...
```

---

## Environment Configuration

All services load config from `.env` via `util/config.go`. Copy the right template:

| Template | Use case |
|----------|----------|
| `.env.local.example` | Homebrew MySQL on `localhost` |
| `.env.docker.example` | Full Docker Compose stack |

Key variables:

- `APP_ENV` — `development` | `staging` | `production` (production rejects default secrets)
- `DB_*` — database connection
- `*_GRPC_URL` / `*_SERVICE_URL` — inter-service discovery (auto-defaults from ports if unset)

---

## Database Migrations

Schema changes are versioned in `migrations/` and tracked in the `schema_migrations` table (version, name, description, applied_at). Goose uses `goose_migrations` internally to apply only pending updates.

```bash
make migrate-up        # apply pending migrations (via Docker)
make migrate-status    # list applied / pending
make migrate-version   # current version number (via Docker)
make db-init           # create DB + migrate
```

Migration commands run **inside Docker** (`docker compose run migrate`) so they work with `DB_HOST=db` in `.env`. Ensure the DB container is running first (`make up-db` or `make up-full`).

**Adding a new migration:** create `migrations/00006_your_change.sql`:

```sql
-- +goose Up
ALTER TABLE accounts ADD COLUMN ...;
INSERT IGNORE INTO schema_migrations (version, name, description)
VALUES (6, 'your_change', 'What this migration does');

-- +goose Down
ALTER TABLE accounts DROP COLUMN ...;
```

**Docker persistence:** `make down` stops containers but keeps data in the `spice_ledger_mysql_data` volume. Only `make down-volumes` wipes the database. On each `make up-full`, the migrate service runs and applies any new migrations without re-seeding existing data.

---

## Local Setup (manual)

### Prerequisites
1. **MySQL** on `localhost:3306` (`brew services start mysql`)
2. **Database** initialized with both schemas:
   ```bash
   make db-init
   ```

### Run services (separate terminals or VS Code compound)

```bash
make run-control
make run-market
make run-rest
make run-graphql
make run-proxy
```

---

## 🚀 API Reference

**Base URL**: `http://localhost:8080/rest` (via Proxy)
**Default Basic Auth**: `admin:secret123` (Base64: `YWRtaW46c2VjcmV0MTIz`)

### Health Check
```bash
curl http://localhost:8080/rest/health
```

### Account Management

#### 1. Check Email Availability
```bash
# Requires Basic Auth Header
curl "http://localhost:8080/rest/accounts/check-email?email=user@example.com" \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz"
```

#### 2. Create or Update Account
```bash
curl -X POST http://localhost:8080/rest/accounts \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "acc_123",
    "name": "John Doe",
    "user_type": "merchant",
    "email": "john@example.com",
    "password": "securepassword"
  }'
```

#### 3. List All Accounts
```bash
curl http://localhost:8080/rest/accounts \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### 4. Get Account by ID
```bash
curl http://localhost:8080/rest/accounts/acc_123 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Merchant Details
These endpoints require a valid **Merchant** Bearer token.

#### 1. Create or Update Merchant Details
```bash
curl -X POST http://localhost:8080/rest/accounts/merchant-details \
  -H "Authorization: Bearer YOUR_MERCHANT_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "acc_123",
    "phone_number": "+1234567890",
    "address": "123 Spice Street",
    "city": "Spice City",
    "state": "Spice State",
    "pincode": "123456"
  }'
```

#### 2. Get Merchant Details
```bash
curl http://localhost:8080/rest/accounts/merchant-details/acc_123 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Authentication & Sessions

#### 1. Login
```bash
curl -X POST http://localhost:8080/rest/accounts/login \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword",
    "device_id": "dev_987"
  }'
```

#### 2. Refresh Token
```bash
curl -X POST http://localhost:8080/rest/accounts/refresh \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN",
    "device_id": "dev_987"
  }'
```

#### 3. Logout
Requires a valid Bearer token in the `Authorization` header.
```bash
curl -X POST http://localhost:8080/rest/accounts/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"device_id": "dev_987"}'
```

### Products

#### 1. Create or Update Product
```bash
curl -X POST http://localhost:8080/rest/products \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "prod_001",
    "name": "Black Pepper",
    "category": "Spices",
    "description": "High quality black pepper",
    "status": "active"
  }'
```

#### 2. List All Products
```bash
curl "http://localhost:8080/rest/products/" \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz"
```

### Grades

#### 1. Create or Update Grade
```bash
curl -X POST http://localhost:8080/rest/grades \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "grade_001",
    "product_id": "prod_001",
    "name": "Grade A",
    "description": "Premium Grade",
    "status": "active"
  }'
```

#### 2. List Grades by Product ID
```bash
curl "http://localhost:8080/rest/grades/?product_id=prod_001" \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz"
```

### Daily Prices

#### 1. Create or Update Daily Price
```bash
curl -X POST http://localhost:8080/rest/daily-prices \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "price_001",
    "product_id": "prod_001",
    "grade_id": "grade_001",
    "price": 450.50,
    "date": "2026-03-08",
    "time": "10:00:00"
  }'
```

#### 2. List Daily Prices (Historical)
```bash
curl "http://localhost:8080/rest/daily-prices/?grade_id=grade_001&duration=7&date=2026-03-08" \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz"
```

#### 3. Get Today's Price by Grade ID
```bash
curl "http://localhost:8080/rest/daily-prices/grade/today/?grade_id=grade_001" \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz"
```

#### 4. Get Today's Price by Product ID
```bash
curl "http://localhost:8080/rest/daily-prices/product/today/?product_id=prod_001&date=2026-03-08" \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz"
```

---

## 🔐 Authentication Details

The system supports two layers of authentication:
1. **Internal Basic Auth**: Required by the gRPC service. The REST gateway defaults to `admin:secret123` for internal calls IF no header is passed, but for public API testing via Proxy, you should include the header: `-H "Authorization: Basic YWRtaW46c2VjcmV0MTIz"`.
2. **Session-based JWT**: Used for authenticated client requests (Logout).

## Run Services

### 1. Control Service (gRPC)
```bash
DB_HOST=localhost go run control/cmd/control/main.go
```

### 2. Market Service (gRPC)
```bash
DB_HOST=localhost go run market/cmd/market/main.go
```

### 3. Rest Service
```bash
ACCOUNT_GRPC_URL=localhost:50051 PORT=8082 go run rest/cmd/rest/main.go
```

### 4. GraphQL Service
```bash
ACCOUNT_GRPC_URL=localhost:50051 MARKET_GRPC_URL=localhost:50052 go run graphql/cmd/graph/main.go
```

### 5. Proxy Service
```bash
ACCOUNT_SERVICE_URL=http://localhost:8082 GRAPHQL_GATEWAY_URL=http://localhost:8081 PROXY_PORT=8080 go run proxy/main.go
```