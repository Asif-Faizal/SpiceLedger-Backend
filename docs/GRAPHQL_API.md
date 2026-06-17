# GraphQL API Reference

GraphQL is served at **`POST /graphql`** on the gateway: `http://localhost:8080/graphql`.

Schema: [`graphql/schema.graphql`](../graphql/schema.graphql)  
Resolvers: [`graphql/query_resolver.go`](../graphql/query_resolver.go), [`graphql/mutation_resolver.go`](../graphql/mutation_resolver.go)

## Response format

Unlike standard GraphQL `{ data, errors }`, SpiceLedger wraps responses to match REST:

**Success:**
```json
{
  "success": true,
  "message": "",
  "data": {
    "products": [ ... ]
  }
}
```

**Error:**
```json
{
  "success": false,
  "message": "token has invalid claims: token is expired",
  "data": null
}
```

HTTP status reflects the error (401 for auth, 403 for permission, 400 for validation).

**Auth:** `Authorization: Bearer <access_token>` on every request. Obtain tokens via REST `POST /rest/accounts/login`.

---

## Data flow

```
HTTP POST /graphql
    │
    ▼
authMiddleware          → stores JWT in context (AccessTokenKey)
    │
    ▼
gqlgen resolver         → maps GraphQL args to protobuf request
    │
    ▼
gRPC client interceptor → forwards "authorization: Bearer <jwt>" metadata
    │
    ├─► Control gRPC (:50051)   AuthInterceptor + SessionInterceptor
    └─► Market gRPC (:50052)    AuthInterceptor
            │
            ▼
        MySQL
    │
    ▼
restResponseEnvelope    → { success, message, data }
```

---

## Queries

### `products(date, search)`

| | |
|---|---|
| **gRPC** | `ControlService.GetProductsWithGradesAndPrices` |
| **Auth** | Authenticated (Bearer) |
| **Default date** | Today (`YYYY-MM-DD`) if omitted |

**GraphQL request:**
```graphql
query {
  products(date: "2026-06-16", search: "") {
    id
    name
    category
    description
    status
    grades {
      id
      productId
      name
      description
      status
      price
    }
  }
}
```

**gRPC request** (`GetProductsWithGradesAndPricesRequest`):
```protobuf
{ date: "2026-06-16", search: "" }
```

**gRPC response** → mapped to `[Product!]!` with nested `grades` and each grade's `price` for the given date.

---

### `getGradePosition(spiceGradeId)`

| | |
|---|---|
| **gRPC** | `MarketService.GetGradePosition` |
| **Auth** | Merchant Bearer (user from JWT context) |

**GraphQL request:**
```graphql
query {
  getGradePosition(spiceGradeId: "grd_turmeric_a_000000000001") {
    userId
    spiceGradeId
    totalQty
    totalCost
    avgCost
    todayPrice
    realizedPnL
    unrealizedPnL
    updatedAt
  }
}
```

**gRPC request** (`GetGradePositionRequest`):
```protobuf
{ user_id: "<from JWT account_id>", spice_grade_id: "grd_turmeric_a_000000000001" }
```

`user_id` is injected server-side from the JWT — not passed in GraphQL args.

**Response fields** come from `PositionView` protobuf (FIFO position + today's daily price for unrealized P&L).

---

### `getPositions`

| | |
|---|---|
| **gRPC** | `MarketService.GetPositions` |
| **Auth** | Merchant Bearer |

**GraphQL request:**
```graphql
query {
  getPositions {
    userId
    spiceGradeId
    totalQty
    avgCost
    todayPrice
    realizedPnL
    unrealizedPnL
  }
}
```

**gRPC:** `GetPositionsRequest { user_id: "<from JWT>" }` → `repeated PositionView`.

---

### `listGradeTransactions(spiceGradeId, skip, take)`

| | |
|---|---|
| **gRPC** | `MarketService.ListGradeTransactions` |
| **Auth** | Merchant Bearer |

**GraphQL request:**
```graphql
query {
  listGradeTransactions(spiceGradeId: "grd_turmeric_a_000000000001", skip: 0, take: 10) {
    id
    userId
    spiceGradeId
    type
    quantity
    price
    tradeDate
    createdAt
  }
}
```

**gRPC:** `ListGradeTransactionsRequest { user_id, spice_grade_id, skip, take }`.

---

### `listTransactions(skip, take)`

| | |
|---|---|
| **gRPC** | `MarketService.ListTransactions` |
| **Auth** | Merchant Bearer (own transactions) / used by admin dashboard internally |

**GraphQL request:**
```graphql
query {
  listTransactions(skip: 0, take: 20) {
    id
    type
    quantity
    price
    tradeDate
  }
}
```

**gRPC:** `ListTransactionsRequest { user_id: "<from JWT>", skip, take }`.

---

### `adminDashboard`

| | |
|---|---|
| **gRPC** | `ControlService.GetSystemMetrics` + `MarketService.GetMarketMetrics` + `MarketService.ListTransactions` |
| **Auth** | Admin Bearer |

**GraphQL request:**
```graphql
query {
  adminDashboard {
    totalUsers
    totalProducts
    totalTransactions
    totalVolume
    recentTransactions { id type quantity price }
    topProducts { name volume }
  }
}
```

**Resolver orchestration** ([`query_resolver.go`](../graphql/query_resolver.go)):

1. `GetSystemMetrics` → `totalUsers`, `totalProducts`
2. `GetMarketMetrics` → `totalTransactions`, `totalVolume`, `topProducts`
3. `ListTransactions { take: 10 }` → `recentTransactions` (all users, admin context)

---

## Mutations

### `createProduct(input)`

| | |
|---|---|
| **gRPC** | `ControlService.CreateOrUpdateProduct` |
| **Auth** | Admin Bearer |

**GraphQL:**
```graphql
mutation {
  createProduct(input: {
    id: ""
    name: "Cardamom"
    category: "spice"
    description: "Green pods"
    status: "active"
  }) {
    id name category status
  }
}
```

**gRPC:** `CreateOrUpdateProductRequest` — empty `id` generates a new KSUID server-side.

---

### `createGrade(input)`

| | |
|---|---|
| **gRPC** | `ControlService.CreateOrUpdateGrade` |
| **Auth** | Admin Bearer |

**GraphQL:**
```graphql
mutation {
  createGrade(input: {
    id: ""
    productId: "prd_turmeric_00000000000001"
    name: "Grade A"
    status: "active"
  }) {
    id productId name
  }
}
```

---

### `createDailyPrice(input)`

| | |
|---|---|
| **gRPC** | `ControlService.CreateOrUpdateDailyPrice` |
| **Auth** | Admin Bearer |

**GraphQL:**
```graphql
mutation {
  createDailyPrice(input: {
    id: ""
    productId: "prd_turmeric_00000000000001"
    gradeId: "grd_turmeric_a_000000000001"
    price: 125.5
    date: "2026-06-16"
    time: "10:00:00"
  }) {
    id price date time
  }
}
```

**gRPC:** `CreateOrUpdateDailyPriceRequest` — upserts on `(product_id, grade_id, date)`.

---

### `buy(spiceGradeId, quantity, price, tradeDate)`

| | |
|---|---|
| **gRPC** | `MarketService.Buy` |
| **Auth** | Merchant Bearer |

**GraphQL:**
```graphql
mutation {
  buy(
    spiceGradeId: "grd_turmeric_a_000000000001"
    quantity: 10
    price: 120.0
    tradeDate: "2026-06-16"
  ) {
    id type quantity price tradeDate
  }
}
```

**gRPC request** (`BuyRequest`):
```protobuf
{
  user_id: "<from JWT>",
  spice_grade_id: "grd_turmeric_a_000000000001",
  quantity: 10,
  price: 120.0,
  trade_date: "2026-06-16"   // defaults to today if empty/invalid
}
```

**gRPC response** (`BuyResponse`):
```protobuf
{ transaction: { id, user_id, spice_grade_id, type: "BUY", quantity, price, trade_date, created_at } }
```

Side effects: inserts `transactions` + `buy_lots`, updates `positions`.

---

### `sell(spiceGradeId, quantity, price, tradeDate)`

| | |
|---|---|
| **gRPC** | `MarketService.Sell` |
| **Auth** | Merchant Bearer |

Same shape as `buy`. **gRPC:** `SellRequest` / `SellResponse`.

FIFO sell allocates against oldest `buy_lots`, creates `sell_allocations`, updates `realized_pnl` on positions. Fails with error if sell quantity exceeds available inventory.

---

## gRPC method map (quick reference)

| GraphQL field | gRPC service | RPC |
|---------------|--------------|-----|
| `products` | Control | `GetProductsWithGradesAndPrices` |
| `getGradePosition` | Market | `GetGradePosition` |
| `getPositions` | Market | `GetPositions` |
| `listGradeTransactions` | Market | `ListGradeTransactions` |
| `listTransactions` | Market | `ListTransactions` |
| `adminDashboard` | Control + Market | `GetSystemMetrics`, `GetMarketMetrics`, `ListTransactions` |
| `createProduct` | Control | `CreateOrUpdateProduct` |
| `createGrade` | Control | `CreateOrUpdateGrade` |
| `createDailyPrice` | Control | `CreateOrUpdateDailyPrice` |
| `buy` | Market | `Buy` |
| `sell` | Market | `Sell` |

---

## Auth requirements summary

| Field | Admin JWT | Merchant JWT |
|-------|-----------|--------------|
| `products` | ✓ | ✓ |
| `adminDashboard` | ✓ | ✗ |
| `createProduct`, `createGrade`, `createDailyPrice` | ✓ | ✗ |
| `getGradePosition`, `getPositions`, `list*`, `buy`, `sell` | ✗ | ✓ |

Admin/merchant checks happen in gRPC handlers via context flags set by `AuthInterceptor`.

---

## Example: full HTTP call

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <merchant_access_token>" \
  -d '{
    "query": "mutation { buy(spiceGradeId: \"grd_turmeric_a_000000000001\", quantity: 5, price: 120) { id type quantity } }"
  }'
```

**Success response:**
```json
{
  "success": true,
  "message": "",
  "data": {
    "buy": {
      "id": "txn_...",
      "type": "BUY",
      "quantity": 5
    }
  }
}
```

**Expired token response:**
```json
{
  "success": false,
  "message": "token has invalid claims: token is expired",
  "data": null
}
```

---

## Playground

`http://localhost:8080/playground` serves the gqlgen playground UI. Because responses use the REST envelope, the playground may not display results correctly — use [Bruno `SpiceLedger-API`](../../SpiceLedger-API/) or curl for testing.

## Related docs

- [MICROSERVICES.md](./MICROSERVICES.md) — how services connect
- [MIDDLEWARE_AND_UTIL.md](./MIDDLEWARE_AND_UTIL.md) — auth and response envelope
- Proto definitions: [`control/control.proto`](../control/control.proto), [`market/market.proto`](../market/market.proto)
