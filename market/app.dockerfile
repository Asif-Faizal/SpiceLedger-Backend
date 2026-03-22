# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Copy entire project for internal dependencies
COPY . .

RUN go mod download

# Build the market microservice
RUN CGO_ENABLED=0 GOOS=linux go build -o market-server ./market/cmd/market/main.go

# Run stage
FROM alpine:3.19

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /build/market-server .

# Expose the port the app runs on
EXPOSE 50051

# Command to run the executable
CMD ["./market-server"]
