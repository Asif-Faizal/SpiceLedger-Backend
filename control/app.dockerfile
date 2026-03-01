# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Copy entire project for internal dependencies
COPY . .

RUN go mod download

# Build the control microservice
RUN CGO_ENABLED=0 GOOS=linux go build -o control-server ./control/cmd/control/main.go

# Run stage
FROM alpine:3.19

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /build/control-server .

# Expose the port the app runs on
EXPOSE 50051

# Command to run the executable
CMD ["./control-server"]
