# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod ./
# Only try to copy go.sum if it exists
COPY go.sum* ./
RUN go mod download || true

COPY . .
RUN go build -o minelink main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/minelink /app/minelink

# Expose Minecraft Bedrock Server Ports
EXPOSE 19132/udp
EXPOSE 19132/tcp

# Volume for Config File
VOLUME ["/app/config"]

# Ensure the binary has execution permission
RUN chmod +x /app/minelink

# Set entrypoint
CMD ["/app/minelink", "--config", "/app/config/config.json"]