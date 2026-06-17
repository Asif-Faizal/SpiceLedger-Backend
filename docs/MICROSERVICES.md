# Microservices Architecture

SpiceLedger Backend is a Go microservices system with two gRPC domain services, a unified HTTP gateway, and a shared MySQL database.

## Service overview

```
                         ┌─────────────────────────────────────────┐
                         │         gateway (:8080)                 │
                         │  /health  /rest/*  /graphql  /playground│
                         │  (in-process REST + GraphQL handlers)   │
                         └───────────┬─────────────┬─────────────────┘
                                     │             │
                                     ▼             ▼
                         ┌───────────────┐  ┌──────────────────┐
                         │  rest package │  │ graphql package  │
                         │  HTTP → gRPC  │  │ gqlgen → gRPC    │
                         └───────┬───────┘  └────────┬─────────┘
                                 │                   │
                                 │         ┌─────────┴─────────┐
                                 │         │                   │
                                 ▼         ▼                   ▼
                         ┌───────────────┐           ┌─────────────────┐
                         │ control       │           │ market          │
                         │ gRPC (:50051) │           │ gRPC (:50052)   │
                         │ accounts      │           │ FIFO trading    │
                         │ auth          │           │ positions/PnL   │
                         │ products      │           │ transactions    │
                         │ daily prices  │           │                 │
                         └───────┬───────┘           └────────┬────────┘
                                 │                            │
                                 └────────────┬───────────────┘
                                              ▼
                                    ┌─────────────────┐
                                    │  MySQL (:3306)  │
                                    │  spice_ledger   │
                                    └─────────────────┘
```

| Service | Type | Port (default) | Responsibility |
|---------|------|----------------|----------------|
| **gateway** | HTTP edge | 8080 | Single public entrypoint; mounts REST + GraphQL |
| **control** | gRPC | 50051 | Identity, sessions, catalog (products/grades), daily prices |
| **market** | gRPC | 50052 | Buy/sell, FIFO inventory, positions, P&L, transaction history |
| **db** | MySQL 8 | 3306 (3306 on host) | Persistent storage |
| **migrate** | One-shot job | — | Applies goose migrations on startup |

`rest/` and `graphql/` are Go packages used by gateway — not separate containers in Docker Compose.

---

## Domain split

### Control service

**Package:** [`control/`](../control/)  
**Proto:** [`control/control.proto`](../control/control.proto)  
**Tables:** `accounts`, `sessions`, `merchant_details`, `products`, `grade`, `daily_price`

Handles:

- Account CRUD, email check, merchant profile
- Login / logout / refresh (JWT + session rows)
- Product and grade catalog
- Daily price create/list/today queries
- `GetProductsWithGradesAndPrices` (used by GraphQL `products` query)
- `GetSystemMetrics` (admin dashboard user/product counts)

### Market service

**Package:** [`market/`](../market/)  
**Proto:** [`market/market.proto`](../market/market.proto)  
**Tables:** `transactions`, `buy_lots`, `sell_allocations`, `positions`

Handles:

- **Buy** — creates transaction + buy lot, updates position
- **Sell** — FIFO allocation against buy lots, realizes P&L
- **Positions** — quantity, average cost, unrealized P&L (uses today's `daily_price`)
- **Transaction history** — per user or per grade
- **Market metrics** — volume, top products (admin dashboard)

Market reads `daily_price` from the same MySQL database for mark-to-market pricing.

---

## Integration patterns

### 1. REST → Control only

[`rest/`](../rest/) is a thin HTTP layer. Each handler:

1. Parses HTTP request (query params, JSON body)
2. Forwards `Authorization` header to gRPC metadata (or uses internal Basic auth)
3. Calls `ControlClient` method
4. Maps protobuf response → JSON via `util.WriteJSONResponse`

REST does **not** talk to market. Trading is GraphQL-only.

### 2. GraphQL → Control + Market

[`graphql/`](../graphql/) uses gqlgen resolvers that call:

- `controlClient` — catalog mutations, products query, system metrics
- `marketClient` — buy/sell, positions, transactions, market metrics

JWT from the HTTP `Authorization` header is stored in context and re-attached to every outbound gRPC call via a client interceptor ([`graphql/graph.go`](../graphql/graph.go)).

Market handlers read `user_id` from gRPC context (`AccountIDKey`) when the GraphQL resolver does not pass it explicitly.

### 3. Gateway (unified HTTP edge)

[`gateway/`](../gateway/) mounts REST and GraphQL handlers in one process ([`gateway/server.go`](../gateway/server.go)):

| Path | Handler |
|------|---------|
| `/rest/*` | `rest.NewHandler` with `/rest` prefix stripped |
| `/graphql` | gqlgen executable schema |
| `/playground` | GraphQL playground UI |
| `/health` | Gateway liveness |
| `/ready` | Readiness stub (extend with upstream checks) |

No reverse-proxy hop — outbound gRPC connections are owned by the gateway process. See [ENGINEERING.md](./ENGINEERING.md) for ADRs.

---

## Docker Compose wiring

From [`docker-compose.yaml`](../docker-compose.yaml):

```
db (healthy)
  └─ migrate (completed)
       ├─ control (started)
       └─ market (started, after control)
            └─ gateway (needs control + market)
```

Environment overrides inside Docker:

- `DB_HOST=db`
- `ACCOUNT_GRPC_URL=control:50051`
- `MARKET_GRPC_URL=market:50052`
- `PROXY_PORT=8080`

Host port mapping (defaults): gateway **8080**, control **50051**, market **50052**, MySQL **3306**→3306 when `DB_HOST_PORT=3306`.

---

## Authentication model

Two layers work together:

1. **Basic auth** (`admin:secret123`) — bootstrap and internal calls; satisfies `is_authenticated` on gRPC for public-ish endpoints
2. **Bearer JWT** — user sessions after login; injects `account_id`, `user_type`, admin/merchant flags

| Endpoint type | Typical auth |
|---------------|--------------|
| Login, refresh, check-email, list products | Basic |
| Admin operations (list accounts, create product) | Bearer (admin JWT) |
| Merchant operations (buy/sell, positions) | Bearer (merchant JWT) |
| GraphQL queries/mutations | Bearer JWT |

Control additionally validates that Bearer tokens match a non-revoked row in `sessions`.

---

## Response format (all HTTP APIs)

Both REST and GraphQL HTTP gateways return:

```json
{ "success": bool, "message": string, "data": object | null }
```

See [MIDDLEWARE_AND_UTIL.md](./MIDDLEWARE_AND_UTIL.md) for error mapping details.

---

## API documentation

| API | Documentation |
|-----|---------------|
| REST | [README.md](../README.md#rest-api-quick-reference), Bruno collection in `SpiceLedger-API/` |
| GraphQL | [GRAPHQL_API.md](./GRAPHQL_API.md) |
| gRPC (internal) | `control/control.proto`, `market/market.proto` |

---

## Local vs Docker

| Mode | MySQL | Services | Entry URL |
|------|-------|----------|-----------|
| Full local | Homebrew `:3306` | `make run-gateway` (+ control/market) | `http://localhost:8080` |
| Full Docker | Container `:3306` host | `make up-full-build` | `http://localhost:8080` |
| Hybrid | Docker DB only | `make up-db` + `make run-*` | Same |

Use `make up-full-build` after code changes — `make up-full` alone does not rebuild images.
