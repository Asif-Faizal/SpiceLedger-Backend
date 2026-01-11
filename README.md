# Spice Backend API Documentation

## Getting Started

### Prerequisites
- Go 1.21+
- Docker (for PostgreSQL & Redis)

### 1. Setup Environment
Ensure your `.env` file is configured (default provided in repo):
```ini
DB_DSN="host=localhost user=postgres password=postgres dbname=spice_db port=5432 sslmode=disable"
REDIS_ADDR="localhost:6379"
JWT_SECRET="secret_key"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="admin123"
```

### 2. Start Infrastructure
Start the database and cache using Docker:
```bash
docker-compose up -d
```

### 3. Run Migrations
Initialize the schema and seed data:
```bash
go run cmd/migrate/main.go
```

### 4. Start the Server
Run the application:
```bash
go run cmd/server/main.go
```
The server will start on `http://localhost:8080`.

---

## API Reference

Base URL: `http://localhost:8080`

## Authentication

### Register User
*Role: Any (New User)*
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "password123"
  }'
```

### Login
*Role: Any*
Returns a JWT token pair and an admin flag. Use the `token` in the `Authorization` header for subsequent requests.
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "password123"
  }'
```
**Response:**
```json
{
  "token": "eyJhbGciOiJIUz...",
  "refresh_token": "eyJhbGciOiJIUz...",
  "is_admin": false
}
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<REFRESH_TOKEN>"
  }'
```
**Response:**
```json
{
  "token": "<NEW_ACCESS_TOKEN>",
  "refresh_token": "<NEW_REFRESH_TOKEN>"
}
```

---

## Admin Endpoints
**Requires:** `Authorization: Bearer <ADMIN_TOKEN>`
**Note:** The default admin is `admin@example.com` / `admin123`.

### 1. User Statistics
Get the total number of registered users.
```bash
curl -X GET http://localhost:8080/api/admin/stats \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

### 2. Create Product
Create a product (e.g., Cardamom).
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cardamom",
    "description": "Default product"
  }'
```

### 3. List Products
```bash
curl -X GET http://localhost:8080/api/products \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

### 4. Create Grade (for a Product)
Add a new grade under a specific product.
```bash
PRODUCT_ID="<UUID-FROM-LIST-PRODUCTS>"

curl -X POST http://localhost:8080/api/grades \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "'$PRODUCT_ID'",
    "name": "Cardamom Extra Bold",
    "description": "8mm+ pods, green color"
  }'
```

### 5. Set Daily Price (Product + Grade)
Set the market price for a specific product grade on a specific date.
```bash
DATE="2025-12-14"
PRODUCT_ID="<UUID-FROM-LIST-PRODUCTS>"
GRADE_ID="<UUID-FROM-LIST-GRADES>"

curl -X POST http://localhost:8080/api/prices \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "'$DATE'",
    "product_id": "'$PRODUCT_ID'",
    "grade_id": "'$GRADE_ID'",
    "price_per_kg": 1500.50
  }'
```

---

## User / Inventory Endpoints
**Requires:** `Authorization: Bearer <USER_TOKEN>` (Admins can also access these)

### GraphQL: Consumer Login and Single-Call Data Fetch
Use GraphQL to fetch products, grades, and inventory for a specific date in a single call.

```bash
# Login as consumer
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "consumer@example.com",
    "password": "admin123"
  }'
```

Example response:
```json
{
  "is_admin": false,
  "refresh_token": "<REFRESH_TOKEN>",
  "token": "<ACCESS_TOKEN>"
}
```

```bash
# GraphQL combined query
curl -s -X POST http://localhost:8080/api/graphql \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query($date:String!){ products{ID Name Description CreatedAt} grades{ID Name Description ProductID CreatedAt} inventoryOnDate(date:$date){ TotalQuantity TotalCost TotalPnL TotalPnLPct Products{ ProductID Product TotalQuantity TotalValue TotalCost TotalPnL TotalPnLPct Grades{ ProductID Product Grade TotalQuantity AverageCost TotalCostBasis MarketPrice MarketValue UnrealizedPnL } } } }",
    "variables": { "date": "2026-01-11" }
  }'
```

```bash
# GraphQL day-only transactions with per-grade PnL
curl -s -X POST http://localhost:8080/api/graphql \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query($date:String!){ inventory(date:$date){ Date TotalBought TotalSold TotalDayPnL Grades{ ProductID Product GradeID Grade BoughtQty BoughtAvgCost SoldQty SoldAvgPrice DayPnL } } }",
    "variables": { "date": "2026-01-11" }
  }'
```

### 1. List Grades
View all available spice grades.
```bash
curl -X GET http://localhost:8080/api/grades \
  -H "Authorization: Bearer <USER_TOKEN>"
```

### 2. Add Purchase Lot (Stock In)
Record a purchase of stock.
```bash
# First, you need a Product ID and a Grade ID
PRODUCT_ID="<UUID-FROM-LIST-PRODUCTS>"
GRADE_ID="<UUID-FROM-LIST-GRADES>"

curl -X POST http://localhost:8080/api/lots \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2023-10-25",
    "product_id": "'$PRODUCT_ID'",
    "grade_id": "'$GRADE_ID'",
    "quantity_kg": 100.0,
    "unit_cost": 1200.0
  }'
```

### 3. Add Sale Transaction (Stock Out)
Record a sale of stock.
```bash
curl -X POST http://localhost:8080/api/sales \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2023-10-27",
    "product_id": "'$PRODUCT_ID'",
    "grade_id": "'$GRADE_ID'",
    "quantity_kg": 50.0,
    "unit_price": 1600.0
  }'
```

### 4. Get Current Inventory & P&L
Get the current stock status, valuation, and P&L based on today's market price.
```bash
curl -X GET http://localhost:8080/api/inventory/current \
  -H "Authorization: Bearer <USER_TOKEN>"
```
**Response shape (example):**
```json
{
  "products": [
    {
      "product_id": "<UUID>",
      "product": "Cardamom",
      "grades": [
        {
          "product_id": "<UUID>",
          "product": "Cardamom",
          "grade": "Cardamom Extra Bold",
          "total_quantity_kg": 50,
          "average_cost_per_kg": 1200,
          "market_price_per_kg": 1500.5,
          "market_value": 75025,
          "total_cost_basis": 60000,
          "unrealized_pnl": 15025
        }
      ],
      "total_quantity_kg": 50,
      "total_value": 75025,
      "total_cost": 60000,
      "total_pnl": 15025,
      "total_pnl_pct": 25.04
    }
  ],
  "total_quantity_kg": 50,
  "total_value": 75025,
  "total_cost": 60000,
  "total_pnl": 15025,
  "total_pnl_pct": 25.04
}
```

### 5. Get Inventory History
Get stock status as it was on a specific past date.
```bash
curl -X GET "http://localhost:8080/api/inventory/on-date?date=2023-10-26" \
  -H "Authorization: Bearer <USER_TOKEN>"
```

### 6. Get Daily Prices
View market prices for a specific date.
```bash
curl -X GET http://localhost:8080/api/prices/2023-10-27 \
  -H "Authorization: Bearer <USER_TOKEN>"
```
Returns:
```json
{
  "date": "2023-10-27",
  "prices": [
    { "date": "2023-10-27", "product_id": "<UUID>", "grade_id": "<UUID>", "price_per_kg": 1500.5 }
  ]
}
```

### 7. Get Single Price (by Product + Grade)
```bash
curl -X GET http://localhost:8080/api/prices/2023-10-27/<PRODUCT_ID>/<GRADE_ID> \
  -H "Authorization: Bearer <USER_TOKEN>"
```
Returns:
```json
{
  "date": "2023-10-27",
  "product_id": "<PRODUCT_ID>",
  "grade_id": "<GRADE_ID>",
  "price_per_kg": 1500.5
}
```
