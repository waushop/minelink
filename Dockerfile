FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o minecraft-bridge main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/minecraft-bridge .

EXPOSE 19132/udp
EXPOSE 19132/tcp

VOLUME /app/config

CMD ["/app/minecraft-bridge", "--config", "/app/config/config.json"]