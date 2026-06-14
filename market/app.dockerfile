# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o market-server ./market/cmd/market/main.go

# Run stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /build/market-server .

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app
USER app

EXPOSE 50052

CMD ["./market-server"]
