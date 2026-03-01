FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Copy entire project for internal dependencies
COPY . .

RUN go mod download

# Build rest gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/rest-server ./rest/cmd/rest

FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /build/rest-server .

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app
USER app

EXPOSE 8082

HEALTHCHECK --interval=15s --timeout=5s --retries=3 --start-period=10s \
    CMD curl -f http://localhost:8082/health || exit 1

CMD ["./rest-server"]
