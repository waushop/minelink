# Stage 1: Build the Go binary
FROM golang:1.21 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o minecraft-bridge main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/minecraft-bridge /app/minecraft-bridge

# Expose Minecraft Bedrock Server Ports
EXPOSE 19132/udp
EXPOSE 19132/tcp

# Volume for Config File
VOLUME ["/app/config"]

# Ensure the binary has execution permission
RUN chmod +x /app/minecraft-bridge

# Set entrypoint
CMD ["/app/minecraft-bridge", "--config", "/app/config/config.json"]