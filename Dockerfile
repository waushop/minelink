# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download || true

COPY . .
RUN go build -o minelink main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/minelink /app/minelink
COPY --from=builder /app/config /app/config

# Expose Minecraft Bedrock Server Ports
EXPOSE 19132/udp
EXPOSE 19132/tcp

# Volume for Config File
VOLUME ["/app/config"]

# Ensure the binary has execution permission
RUN chmod +x /app/minelink

# Add necessary libraries
RUN apk add --no-cache libc6-compat

# Set entrypoint
CMD ["/app/minelink", "--config", "/app/config/config.json"]