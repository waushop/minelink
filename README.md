# MineLink 🎮
Simple and effective, linking Bedrock players across networks

## Purpose
The purpose of this project is to provide a simple and effective proxy for Minecraft Bedrock Edition, allowing players to connect across different networks. The proxy broadcasts a fake LAN game and forwards packets between clients and the actual server.

## Features
- Broadcasts a fake LAN game to allow clients to discover the server
- Forwards packets between clients and the actual server
- Configurable server address, proxy port, and broadcast IP using environment variables

## Configuration
To configure the proxy, you can set the following environment variables:
- `SERVER_ADDRESS`: The address of the actual Minecraft Bedrock server (default: `YOUR_EXTERNAL_SERVER_IP:19132`)
- `PROXY_PORT`: The port for the proxy to listen on (default: `19133`)
- `BROADCAST_IP`: The IP address for broadcasting the fake LAN game (default: `255.255.255.255:19132`)

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
3. Initialize the Go module:
   ```
   go mod init minelink
   ```
4. Set the environment variables as needed:
   ```
   export SERVER_ADDRESS="your_server_address:19132"
   export PROXY_PORT="19133"
   export BROADCAST_IP="255.255.255.255:19132"
   ```
5. Build and run the proxy:
   ```
   go build -o minelink
   ./minelink
   ```

## Running the Proxy with Docker
To build and run the proxy using Docker, follow these steps:
1. Clone the repository:
   ```
   git clone https://github.com/waushop/minelink.git
   ```
2. Change to the project directory:
   ```
   cd minelink
   ```
3. Build the Docker image:
   ```
   docker build -t minelink .
   ```
4. Run the Docker container:
   ```
   docker run -e SERVER_ADDRESS="your_server_address:19132" -e PROXY_PORT="19133" -e BROADCAST_IP="255.255.255.255:19132" -p 19133:19133 minelink
   ```

## Running the Proxy with Docker Compose
To build and run the proxy using Docker Compose, follow these steps:
1. Clone the repository:
   ```
   git clone https://github.com/waushop/minelink.git
   ```
2. Change to the project directory:
   ```
   cd minelink
   ```
3. Create a `.env` file with the following content:
   ```
   SERVER_ADDRESS=your_server_address:19132
   PROXY_PORT=19133
   BROADCAST_IP=255.255.255.255:19132
   ```
4. Start the service using Docker Compose:
   ```
   docker-compose up --build
   ```

## Contributing
We welcome contributions to the project! If you would like to contribute, please follow these steps:
1. Fork the repository on GitHub.
2. Create a new branch for your feature or bugfix.
3. Make your changes and commit them with clear and descriptive commit messages.
4. Push your changes to your forked repository.
5. Create a pull request to the main repository, describing your changes and the problem they solve.

Thank you for your contributions!
