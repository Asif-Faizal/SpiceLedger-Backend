# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS deps
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY --from=deps /src/go.mod /src/go.sum ./
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/base-debian12 AS release
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /out/migrate /app/migrate
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]

FROM golang:1.25-alpine AS dlv-builder
RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM golang:1.25-alpine AS debug-builder
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go mod download && \
    go build -gcflags "all=-N -l" -o /out/server ./cmd/server

FROM alpine:3.19 AS debug
WORKDIR /src
COPY --from=debug-builder /out/server /src/server
COPY --from=dlv-builder /go/bin/dlv /usr/local/bin/dlv
EXPOSE 8080 2345
CMD ["dlv","exec","/src/server","--listen=:2345","--headless","--api-version=2","--log","--accept-multiclient","--continue"]
