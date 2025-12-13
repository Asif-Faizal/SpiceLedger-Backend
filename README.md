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
Returns a JWT token. Use this token in the `Authorization` header for subsequent requests.
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
  "token": "eyJhbGciOiJIUz..."
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

### 2. Create Grade
Add a new spice grade to the system.
```bash
curl -X POST http://localhost:8080/api/grades \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cardamom Extra Bold",
    "description": "8mm+ pods, green color"
  }'
```

### 3. Set Daily Price
Set the market price for a specific grade on a specific date.
```bash
curl -X POST http://localhost:8080/api/prices \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2023-10-27",
    "grade": "Cardamom Extra Bold",
    "price": 1500.50
  }'
```

---

## User / Inventory Endpoints
**Requires:** `Authorization: Bearer <USER_TOKEN>` (Admins can also access these)

### 1. List Grades
View all available spice grades.
```bash
curl -X GET http://localhost:8080/api/grades \
  -H "Authorization: Bearer <USER_TOKEN>"
```

### 2. Add Purchase Lot (Stock In)
Record a purchase of stock.
```bash
# First, you need a Grade ID (get from /api/grades)
GRADE_ID="<UUID-FROM-LIST-GRADES>"

curl -X POST http://localhost:8080/api/lots \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2023-10-25",
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
