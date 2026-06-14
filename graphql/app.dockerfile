FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o graphql-service ./graphql/cmd/graph

FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /build/graphql-service .

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app
USER app

EXPOSE 8081

HEALTHCHECK --interval=15s --timeout=5s --retries=3 --start-period=10s \
    CMD curl -f http://localhost:8081/health || exit 1

CMD ["./graphql-service"]
