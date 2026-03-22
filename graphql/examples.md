# GraphQL Example Queries and Mutations

The GraphQL service is running at `http://localhost:8081/graphql`. 

> [!IMPORTANT]
> In the GraphQL Playground, you must provide headers in the **HTTP HEADERS** tab (bottom left) as a **JSON object**:
```json
{
 "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjoiM0JBR2w0M3hkUjBWdzBaeExreFNFdzhla0F1IiwidXNlcl90eXBlIjoiYWRtaW4iLCJlbWFpbCI6ImFkbWluQHNwaWNlLmNvbSIsImV4cCI6MTc3NDEwNzE4MywibmJmIjoxNzc0MTA3MTIzLCJpYXQiOjE3NzQxMDcxMjN9.JTHop2crI3Wry1mtEioc8WON-Gec2mCobIxhZYP0AAM"
}
```

## Queries

### List Products with Grades and Prices
```graphql
query GetProducts($date: String) {
  products(date: $date) {
    id
    name
    category
    description
    status
    grades {
      id
      name
      description
      status
      price
    }
  }
}
```
**Variables:**
```json
{
  "date": "2026-03-21",
  "search":"car"
}
```

### Get Position for a Specific Grade
```graphql
query GetGradePosition($spiceGradeId: ID!) {
  getGradePosition(spiceGradeId: $spiceGradeId) {
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
**Variables:**
```json
{
  "spiceGradeId": "grade-1"
}
```

### Get All Portfolio Positions
```graphql
query GetPositions {
  getPositions {
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

### List Transactions for a Specific Grade
```graphql
query ListGradeTransactions($spiceGradeId: ID!, $skip: Int, $take: Int) {
  listGradeTransactions(spiceGradeId: $spiceGradeId, skip: $skip, take: $take) {
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
**Variables:**
```json
{
  "spiceGradeId": "grade-1",
  "skip": 0,
  "take": 10
}
```

### List All User Transactions
```graphql
query ListTransactions($skip: Int, $take: Int) {
  listTransactions(skip: $skip, take: $take) {
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
**Variables:**
```json
{
  "skip": 0,
  "take": 10
}
```

## Mutations

### Create or Update Product
```graphql
mutation CreateProduct($input: CreateProductInput!) {
  createProduct(input: $input) {
    id
    name
    category
    description
    status
  }
}
```
**Variables:**
Create
```json
{
  "input": {
    "id": "",
    "name": "Turmeric",
    "category": "spice",
    "description": "Premium Quality Turmeric"
  }
}
```

Update
```json
{
  "input": {
    "id": "prod-1",
    "name": "Turmeric",
    "category": "spice",
    "description": "Premium Quality Turmeric"
  }
}
```

### Create or Update Grade
```graphql
mutation CreateGrade($input: CreateGradeInput!) {
  createGrade(input: $input) {
    id
    productId
    name
    description
    status
  }
}
```
**Variables:**
Create
```json
{
  "input": {
    "id": "",
    "productId": "prod-1",
    "name": "A Grade",
    "description": "Superior quality"
  }
}

Update
```json
{
  "input": {
    "id": "grade-1",
    "productId": "prod-1",
    "name": "A Grade",
    "description": "Superior quality"
  }
}
```

### Create or Update Daily Price
```graphql
mutation CreateDailyPrice($input: CreateDailyPriceInput!) {
  createDailyPrice(input: $input) {
    id
    productId
    gradeId
    price
    date
    time
  }
}
```
**Variables:**
Create
```json
{
  "input": {
    "id": "",
    "productId": "prod-1",
    "gradeId": "grade-1",
    "price": 150.5,
    "date": "2026-03-21",
    "time": "10:00:00"
  }
}

Update
```json
{
  "input": {
    "id": "price-1",
    "productId": "prod-1",
    "gradeId": "grade-1",
    "price": 150.5,
    "date": "2026-03-21",
    "time": "10:00:00"
  }
}
```

### Buy Spice Grade
```graphql
mutation Buy($spiceGradeId: ID!, $quantity: Float!, $price: Float!, $tradeDate: String) {
  buy(spiceGradeId: $spiceGradeId, quantity: $quantity, price: $price, tradeDate: $tradeDate) {
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
**Variables:**
```json
{
  "spiceGradeId": "grade-1",
  "quantity": 100.5,
  "price": 155.25,
  "tradeDate": "2026-03-21"
}
```

### Sell Spice Grade
```graphql
mutation Sell($spiceGradeId: ID!, $quantity: Float!, $price: Float!, $tradeDate: String) {
  sell(spiceGradeId: $spiceGradeId, quantity: $quantity, price: $price, tradeDate: $tradeDate) {
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
**Variables:**
```json
{
  "spiceGradeId": "grade-1",
  "quantity": 50.0,
  "price": 160.0,
  "tradeDate": "2026-03-22"
}
```

## Authentication Handling
The service includes a middleware that extracts the `Authorization` header from the HTTP request and forwards it to the gRPC backend using a client interceptor.

1. **GraphQL Middleware**: Extracts the token and stores it in the `context.Context`.
2. **gRPC Client Interceptor**: Automatically adds the token to the outgoing gRPC metadata.
3. **gRPC Backend**: Validates the token using the existing `AuthInterceptor` in the `control` service.
