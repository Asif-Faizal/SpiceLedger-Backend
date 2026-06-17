# Middleware & `util` Package

Shared infrastructure lives in [`util/`](../util/). HTTP gateways and gRPC servers import this package for configuration, auth, logging, and response formatting.

## Package map

| File | Responsibility |
|------|----------------|
| `config.go` | Load `.env` via envconfig, DSN builder, service URL resolution, production secret validation |
| `constants.go` | Context keys and user-type constants |
| `jwt.go` | JWT claim struct, token generation and validation |
| `crypto.go` | bcrypt password hashing and verification |
| `auth_interceptor.go` | gRPC unary interceptor — parses Bearer JWT or Basic auth from metadata |
| `logger.go` | Zerolog setup, gRPC `UnaryServerInterceptor` for request/response logging |
| `rest_middleware.go` | HTTP logging middleware for REST gateway |
| `response.go` | Standard JSON envelope `{ success, message, data }`, gRPC → HTTP error mapping |

GraphQL-specific HTTP middleware lives in [`graphql/handler.go`](../graphql/handler.go) and [`graphql/response_envelope.go`](../graphql/response_envelope.go).

---

## Configuration (`config.go`)

All services call `util.LoadConfig()` at startup. Key fields:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | `production` enforces non-default secrets |
| `DB_HOST` | `db` | MySQL host |
| `DB_PORT` | `3306` | MySQL port |
| `DB_PASSWORD` | `1234` | MySQL password |
| `DB_NAME` | `spice_ledger` | Database name |
| `JWT_SECRET` | (dev default) | HS256 signing key |
| `ACCESS_TOKEN_DURATION` | `30m` | Access token TTL |
| `REFRESH_TOKEN_DURATION` | `168h` | Refresh token TTL |
| `BASIC_AUTH_USER` / `BASIC_AUTH_PASS` | `admin` / `secret123` | Internal service auth |
| `ACCOUNT_GRPC_URL` | `localhost:50051` | Control service address |
| `MARKET_GRPC_URL` | `localhost:50052` | Market service address |

Helper methods: `DSN()`, `ResolveAccountGrpcURL()`, `ResolveMarketGrpcURL()`.

---

## Standard response envelope (`response.go`)

All REST handlers and GraphQL HTTP responses use the same JSON shape:

```json
{
  "success": true,
  "message": "optional human message",
  "data": { }
}
```

On error:

```json
{
  "success": false,
  "message": "token is expired",
  "data": null
}
```

| Function | Usage |
|----------|-------|
| `WriteJSONResponse(w, code, success, message, data)` | Write success or generic error |
| `WriteGRPCErrorResponse(w, err)` | Map gRPC `status` to HTTP status + message |
| `HTTPStatusFromGRPCCode(code)` | gRPC code → HTTP status |
| `CleanErrorMessage(msg)` | Strip `rpc error: code = X desc = ` prefix |

gRPC → HTTP status mapping:

| gRPC code | HTTP status |
|-----------|-------------|
| `Unauthenticated` | 401 |
| `PermissionDenied` | 403 |
| `InvalidArgument` | 400 |
| `NotFound` | 404 |
| `AlreadyExists` | 409 |
| `DeadlineExceeded` | 504 |
| `Unimplemented` | 501 |
| other | 500 |

---

## gRPC auth interceptor (`auth_interceptor.go`)

Registered on both **control** and **market** gRPC servers (chained with logging and, on control, session validation).

```
Incoming metadata "authorization"
    │
    ├─ Bearer <jwt>  → ValidateToken → inject account_id, user_type, email,
    │                    is_authenticated, is_admin/is_merchant into context
    │
    └─ Basic <b64>   → if matches BASIC_AUTH_* → is_authenticated = true
```

Invalid Bearer tokens return `codes.Unauthenticated` immediately (request does not reach handler).

Context keys (from `constants.go`):

- `AccountIDKey`, `UserTypeKey`, `EmailKey`
- `IsAuthenticatedKey`, `IsAdminKey`, `IsMerchantKey`
- `AccessTokenKey` — raw JWT string (used by session interceptor)

---

## Control-only: session interceptor (`control/server.go`)

After auth interceptor, control chains `SessionInterceptor`:

- If a Bearer token is present, loads the session from DB
- Rejects revoked or missing sessions with `Unauthenticated`
- Ensures logout actually invalidates tokens

---

## gRPC logging (`logger.go` → `UnaryServerInterceptor`)

Logs every gRPC call with method name, duration, status code, and JSON-serialized request/response bodies.

---

## REST HTTP middleware (`rest_middleware.go`)

`LoggingMiddleware` wraps the REST mux in [`rest/routes.go`](../rest/routes.go):

- Reads and logs request body
- Wraps `ResponseWriter` to capture response body and status
- Emits structured transport-layer log lines

REST also uses `withAuth()` in handlers to forward the client's `Authorization` header to control gRPC, or fall back to internal Basic auth when no header is sent.

---

## GraphQL HTTP middleware

### Auth middleware (`graphql/handler.go`)

Extracts `Authorization: Bearer <token>` and stores the token in `util.AccessTokenKey` on the request context. The gRPC client interceptor in [`graphql/graph.go`](../graphql/graph.go) forwards it as `authorization` metadata to control and market.

### Response envelope (`graphql/response_envelope.go`)

Wraps gqlgen output into the same `{ success, message, data }` format as REST:

1. Runs gqlgen into an internal recorder
2. On success → `{ success: true, data: <graphql data object> }`
3. On error → `{ success: false, message: <clean gRPC message>, data: null }` with appropriate HTTP status

### Error presenter (`graphql/handler.go`)

Converts gRPC errors to clean messages (no `rpc error:` prefix) and adds `grpc_code` to extensions for status mapping.

---

## JWT & passwords

**JWT** (`jwt.go`): HS256 tokens with claims `account_id`, `user_type`, `email`, plus standard `exp`/`iat`/`nbf`.

**Passwords** (`crypto.go`): bcrypt via `HashPassword` / `CheckPasswordHash` in the control service.

---

## Request flow (auth propagation)

```
Client
  │  Authorization: Bearer <jwt>
  ▼
Gateway (:8080)
  │
  ├─ REST (/rest/*) ── withAuth() ── gRPC metadata ──► Control (:50051)
  │                                                    AuthInterceptor
  │                                                    SessionInterceptor
  │
  └─ GraphQL (/graphql) ── authMiddleware (context)
                           gRPC client interceptor ──► Control / Market
                                                      AuthInterceptor
```

Bearer tokens carry user identity end-to-end. Basic auth is used for bootstrap operations (login, public list endpoints) and internal service-to-service calls when no Bearer token is provided.
