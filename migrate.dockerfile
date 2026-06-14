FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /build/migrate .
COPY --from=builder /build/migrations ./migrations

ENTRYPOINT ["./migrate"]
CMD ["up"]
