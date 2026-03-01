FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files (proxy uses standard library only, but go build needs go.mod)
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY proxy ./proxy
COPY util ./util

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/proxy-server ./proxy/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/proxy-server .

EXPOSE 80

CMD ["./proxy-server"]
