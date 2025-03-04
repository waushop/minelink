# MineLink 🎮
A bridge for connecting Minecraft Bedrock Edition console/mobile clients to a Minecraft Bedrock server

## Purpose
The purpose of this project is to provide a simple and effective proxy for Minecraft Bedrock Edition, allowing players to connect across different networks. The proxy broadcasts a fake LAN game and forwards packets between clients and the actual server.

## Features
- Broadcasts a fake LAN game to allow clients to discover the server in the Friends tab
- Forwards UDP and TCP packets between clients and the actual server
- Fully configurable through JSON configuration file
- Graceful shutdown on interrupt

## Configuration
Edit `config/config.json` to match your setup:

```json
{
  "local_address": "0.0.0.0",
  "target_server_ip": "192.168.1.213",
  "target_server_port": 19132,
  "broadcast_interval": 5,
  "server_name": "Epic World",
  "debug": true
}
```

- `local_address`: IP address to bind to (0.0.0.0 = all interfaces)
- `target_server_ip`: The IP address of your actual Bedrock server
- `target_server_port`: The port your Bedrock server is running on (default 19132)
- `broadcast_interval`: How often to broadcast the server's presence in seconds
- `server_name`: The name shown in the console friends tab
- `debug`: Whether to enable debug logging

## Running the Proxy
To run the proxy, follow these steps:
1. Clone the repository:
   ```
   git clone https://github.com/waushop/minelink.git
   ```
2. Change to the project directory:
   ```
   cd minelink
   ```
3. Edit `config/config.json` to match your setup
4. Build and run the proxy:
   ```
   go build -o minelink
   ./minelink
   ```

You can specify a custom config path:
```
./minelink -config /path/to/config.json
```

## Running with Docker
To build and run the proxy using Docker, follow these steps:
1. Clone the repository:
   ```
   git clone https://github.com/waushop/minelink.git
   ```
2. Change to the project directory:
   ```
   cd minelink
   ```
3. Edit `config/config.json` to match your setup
4. Build the Docker image:
   ```
   docker build -t minelink .
   ```
5. Run the Docker container:
   ```
   docker run -p 19132:19132/udp -p 19132:19132/tcp -v $(pwd)/config:/app/config minelink
   ```

## Running with Docker Compose
To build and run the proxy using Docker Compose, follow these steps:
1. Clone the repository:
   ```
   git clone https://github.com/waushop/minelink.git
   ```
2. Change to the project directory:
   ```
   cd minelink
   ```
3. Edit `config/config.json` to match your setup
4. Start the service using Docker Compose:
   ```
   docker-compose up -d
   ```

## Contributing
We welcome contributions to the project! If you would like to contribute, please follow these steps:
1. Fork the repository on GitHub.
2. Create a new branch for your feature or bugfix.
3. Make your changes and commit them with clear and descriptive commit messages.
4. Push your changes to your forked repository.
5. Create a pull request to the main repository, describing your changes and the problem they solve.

Thank you for your contributions!
