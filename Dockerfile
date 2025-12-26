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

FROM gcr.io/distroless/base-debian12 AS release
WORKDIR /app
COPY --from=builder /out/server /app/server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]

FROM golang:1.25-alpine AS debug
WORKDIR /src
RUN apk add --no-cache git
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go mod download && \
    go build -gcflags "all=-N -l" -o /src/server ./cmd/server
EXPOSE 8080 2345
CMD ["dlv","exec","/src/server","--listen=:2345","--headless","--api-version=2","--log","--accept-multiclient","--continue"]
