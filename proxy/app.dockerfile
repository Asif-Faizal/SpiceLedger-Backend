FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Copy entire project for internal dependencies
COPY . .

RUN go mod download

# Build proxy server
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/proxy-server ./proxy/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/proxy-server .

EXPOSE 80

CMD ["./proxy-server"]
