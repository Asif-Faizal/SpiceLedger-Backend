# SpiceLedger Backend

SpiceLedger is a microservices-based backend for managing accounts and ledger operations.

---

## 🛠 Local Setup (Independent Services)

To run the services independently in separate terminals, follow these steps:

### Prerequisites
1. **MySQL**: Ensure MySQL is running on `localhost:3306`.
2. **Database**: Create and initialize the database:
   ```bash
   mysql -u root -p1234 -e "CREATE DATABASE IF NOT EXISTS spice_ledger;"
   mysql -u root -p1234 spice_ledger < control/up.sql
   ```

### 1️⃣ Terminal 1: Control Service (gRPC)
The core service handling business logic and database persistence.
```bash
DB_HOST=localhost go run control/cmd/control/main.go
```

### 2️⃣ Terminal 2: REST Gateway
The entry point that translates REST requests to gRPC.
```bash
ACCOUNT_GRPC_URL=localhost:50051 PORT=8082 go run rest/cmd/rest/main.go
```

### 3️⃣ Terminal 3: Proxy Service
The main router that directs traffic to different gateways (REST, GraphQL).
```bash
ACCOUNT_SERVICE_URL=http://localhost:8082 GRAPHQL_GATEWAY_URL=http://localhost:4000 PROXY_PORT=8080 go run proxy/main.go
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

### 1. Rest Service
ACCOUNT_GRPC_URL=localhost:50051 PORT=8082 go run rest/cmd/rest/main.go

### 2. Control Service (gRPC)
DB_HOST=localhost go run control/cmd/control/main.go

### 3. GraphQL Service
go run graphql/cmd/graph/main.go

### 4. Proxy Service
ACCOUNT_SERVICE_URL=http://localhost:8082 GRAPHQL_GATEWAY_URL=http://localhost:8081 PROXY_PORT=8080 go run proxy/main.go