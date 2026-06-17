# SpiceLedger Backend

Microservices backend for spice trading: accounts, catalog, daily prices, and FIFO market operations.

## Architecture

```
Client → gateway (:8080) → REST handler → control gRPC (:50051) → MySQL
                        → GraphQL      → control + market gRPC (:50052) → MySQL
```

| Service | Role | Port |
|---------|------|------|
| **gateway** | Unified HTTP edge (REST + GraphQL) | 8080 |
| **control** | Accounts, auth, products, prices | 50051 |
| **market** | FIFO trading engine | 50052 |
| **db** | MySQL (host **3306** when using Docker alongside Homebrew MySQL) | 3306 |

`rest/` and `graphql/` are **libraries** mounted by `gateway` — not separate processes.

**Detailed docs:**

| Topic | Document |
|-------|----------|
| Engineering & ADRs | [docs/ENGINEERING.md](docs/ENGINEERING.md) |
| Makefile commands | [docs/MAKEFILE.md](docs/MAKEFILE.md) |
| Database migrations | [docs/MIGRATIONS.md](docs/MIGRATIONS.md) |
| Middleware & `util` package | [docs/MIDDLEWARE_AND_UTIL.md](docs/MIDDLEWARE_AND_UTIL.md) |
| Microservices integration | [docs/MICROSERVICES.md](docs/MICROSERVICES.md) |
| GraphQL API & gRPC mapping | [docs/GRAPHQL_API.md](docs/GRAPHQL_API.md) |
| Bruno API collection | [../SpiceLedger-API](../SpiceLedger-API/) |

---

## Quick start

### Option A — Local MySQL (active development)

```bash
brew services start mysql
make setup                         # .env + migrations
make run-control                   # or VS Code "Full Stack (local)"
make run-market
make run-gateway                   # REST + GraphQL on :8080
```

Entry URL: `http://localhost:8080` (`/rest/*`, `/graphql`, `/playground`)

### Option B — Full Docker stack

```bash
make install-docker                # macOS: installs Docker Desktop
open -a Docker                     # wait until Docker is running
make setup-docker                  # copies .env.docker.example → .env
make up-full-build                 # rebuild images + start all services
```

> **Important:** `make up-full` reuses existing images and does **not** rebuild after code changes. Use `make up-full-build` when Go code or Dockerfiles change.

Docker MySQL on host port **3306** (configurable via `DB_HOST_PORT`) to avoid conflicting with Homebrew MySQL on 3306:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p
```

### Option C — Docker DB only, services native

```bash
make up-db
make db-init
# set .env: DB_HOST=127.0.0.1, DB_HOST_PORT=3306 if using Docker MySQL from host
make run-control
# ... other run-* targets
```

### Useful commands

```bash
make help              # all Makefile targets
make down              # stop containers (data persists)
make down-volumes      # stop + wipe MySQL volume
make logs              # tail container logs
make migrate-up        # apply pending migrations (Docker)
make test              # go test ./...
```

---

## Environment

Copy the right template:

| Template | Use case |
|----------|----------|
| `.env.local.example` | Homebrew MySQL on `localhost:3306` |
| `.env.docker.example` | Full Docker Compose stack |

All services load config from `.env` via [`util/config.go`](util/config.go). Key variables: `APP_ENV`, `DB_*`, `JWT_SECRET`, `BASIC_AUTH_*`, `*_GRPC_URL`, `*_PORT`.

Production mode (`APP_ENV=production`) rejects default secrets.

---

## Response format

All HTTP APIs (REST and GraphQL) return the same JSON envelope:

```json
{
  "success": true,
  "message": "optional message",
  "data": { }
}
```

Errors:

```json
{
  "success": false,
  "message": "human-readable error",
  "data": null
}
```

gRPC errors are mapped to appropriate HTTP status codes (401, 403, 400, etc.). See [docs/MIDDLEWARE_AND_UTIL.md](docs/MIDDLEWARE_AND_UTIL.md).

---

## Authentication

| Layer | Usage |
|-------|-------|
| **Basic** `admin:secret123` | Login, refresh, public list endpoints, internal gRPC |
| **Bearer JWT** | Authenticated user operations after login |

**Seed users** (from migrations):

| Role | Email | Password |
|------|-------|----------|
| Admin | `admin@spice.com` | `secret123` |
| Merchant | `merchant@spice.com` | `secret123` |

```bash
# Login
curl -X POST http://localhost:8080/rest/accounts/login \
  -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@spice.com","password":"secret123","device_id":"dev-001"}'
```

Use `access_token` from the response as `Authorization: Bearer <token>`.

---

## REST API quick reference

**Base URL:** `http://localhost:8080/rest`  
**Health:** `GET /rest/health`

| Area | Endpoints |
|------|-----------|
| **Accounts** | `GET /accounts/check-email?email=`, `POST /accounts`, `GET /accounts`, `GET /accounts/{id}`, `GET /accounts/info` |
| **Auth** | `POST /accounts/login`, `POST /accounts/refresh`, `POST /accounts/logout` |
| **Merchant** | `POST /accounts/merchant-details`, `GET /accounts/merchant-info`, `POST /accounts/merchant-info` |
| **Products** | `POST /products`, `GET /products/?` |
| **Grades** | `POST /grades`, `GET /grades/?product_id=` |
| **Daily prices** | `POST /daily-prices`, `GET /daily-prices/?grade_id=&duration=&date=`, `GET /daily-prices/grade/today/?grade_id=`, `GET /daily-prices/product/today/?product_id=` |

### Notes

- **Check email** requires `email` as a **query parameter**, not JSON body
- **List daily prices** filters `date` backward by `duration` days; omitting `date` defaults to today (server-side)
- List endpoints need trailing slashes: `/products/`, `/grades/`, `/daily-prices/`
- Use `GET /accounts/merchant-info` for merchant profile (not `/accounts/merchant-details/{id}`)

Full curl examples and Bruno requests: [SpiceLedger-API](../SpiceLedger-API/).

---

## GraphQL API

**Endpoint:** `POST http://localhost:8080/graphql`  
**Playground:** `http://localhost:8080/playground` (may not render REST-envelope responses — prefer Bruno)

Requires `Authorization: Bearer <token>`.

| Queries | Mutations |
|---------|-----------|
| `products`, `getGradePosition`, `getPositions` | `createProduct`, `createGrade`, `createDailyPrice` |
| `listGradeTransactions`, `listTransactions` | `buy`, `sell` |
| `adminDashboard` | |

Complete field-level documentation with gRPC request/response mapping: [docs/GRAPHQL_API.md](docs/GRAPHQL_API.md).

---

## Database migrations

Versioned SQL in [`migrations/`](migrations/). Tracked by goose (`goose_migrations`) and an audit table (`schema_migrations`).

```bash
make migrate-up        # apply pending (via Docker)
make migrate-status    # list applied / pending
make migrate-version   # current version number
make db-init           # create DB + migrate (local script)
```

Full guide: [docs/MIGRATIONS.md](docs/MIGRATIONS.md).

---

## Project layout

```
SpiceLedger-Backend/
├── control/          # gRPC: accounts, auth, catalog, prices
├── market/           # gRPC: FIFO trading, positions
├── gateway/          # Unified HTTP edge (REST + GraphQL)
├── rest/             # REST handlers (library; mounted by gateway)
├── graphql/          # GraphQL resolvers (library; mounted by gateway)
├── internal/platform/# Shared HTTP/gRPC lifecycle
├── util/             # Config, auth, logging, responses
├── migrations/       # Versioned SQL (goose)
├── cmd/migrate/      # Migration CLI
├── scripts/          # init-db.sh
├── docs/             # Detailed documentation
├── docker-compose.yaml
└── Makefile
```

---

## Development

```bash
make test              # run tests
make lint              # compile all packages
make clean             # remove build artifacts
```

After changing Go code used in Docker: `make up-full-build`.

Individual services: `make run-control`, `make run-market`, `make run-gateway` — see [docs/MAKEFILE.md](docs/MAKEFILE.md) and [docs/ENGINEERING.md](docs/ENGINEERING.md).
