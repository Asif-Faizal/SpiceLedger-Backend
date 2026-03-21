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
  "date": "2026-03-21"
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

## Authentication Handling
The service includes a middleware that extracts the `Authorization` header from the HTTP request and forwards it to the gRPC backend using a client interceptor.

1. **GraphQL Middleware**: Extracts the token and stores it in the `context.Context`.
2. **gRPC Client Interceptor**: Automatically adds the token to the outgoing gRPC metadata.
3. **gRPC Backend**: Validates the token using the existing `AuthInterceptor` in the `control` service.
