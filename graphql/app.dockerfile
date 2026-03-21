FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o graphql-service ./graphql/cmd/graph

FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/graphql-service .

# Create non-root user
RUN adduser -D user
USER user

# Command to run
CMD ["./graphql-service"]
