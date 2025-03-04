package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Config holds the configuration for the bridge
type Config struct {
	LocalAddress      string `json:"local_address"`
	TargetServerIP    string `json:"target_server_ip"`
	TargetServerPort  int    `json:"target_server_port"`
	BroadcastInterval int    `json:"broadcast_interval"`
	ServerName        string `json:"server_name"`
	Debug             bool   `json:"debug"`
}

// Global variables
var (
	config     Config
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
}

func main() {
	flag.Parse()

	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set up logging
	if config.Debug {
		log.Println("Debug mode enabled")
		log.Printf("Configuration: %+v", config)
	}

	// Start the bridge service
	log.Printf("Starting Minecraft bridge to %s:%d", config.TargetServerIP, config.TargetServerPort)
	log.Printf("Server name: %s", config.ServerName)

	// Use a wait group to manage goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Set up a channel to handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start UDP listener for LAN broadcast responses
	go func() {
		defer wg.Done()
		startUDPListener()
	}()

	// Start broadcasting service presence
	go func() {
		defer wg.Done()
		broadcastService()
	}()

	// Start TCP proxy server in a separate goroutine
	go startTCPProxy()

	// Wait for termination signal
	<-sigChan
	log.Println("Received termination signal, shutting down...")
	
	// Wait for goroutines to finish (they won't actually finish due to infinite loops,
	// but this sets up the structure for proper shutdown in the future)
	wg.Wait()
	log.Println("Server shutdown complete")
}

func loadConfig() error {
	// Default configuration
	config = Config{
		LocalAddress:      "0.0.0.0",
		TargetServerIP:    "127.0.0.1",
		TargetServerPort:  19132,
		BroadcastInterval: 5,
		ServerName:        "Minecraft Server",
		Debug:             false,
	}

	// Try to load from file
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if config exists in the config directory
			if configPath == "config.json" {
				altConfigPath := "config/config.json"
				altFile, altErr := os.Open(altConfigPath)
				if altErr == nil {
					log.Printf("Loading config from %s", altConfigPath)
					decoder := json.NewDecoder(altFile)
					err = decoder.Decode(&config)
					altFile.Close()
					if err == nil {
						return nil
					}
					log.Printf("Error parsing config from %s: %v", altConfigPath, err)
				}
			}
			
			log.Println("Config file not found, creating default config")
			saveConfig()
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("Error parsing config: %v", err)
		return err
	}
	
	log.Printf("Loaded configuration from %s", configPath)
	return nil
}

func saveConfig() error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

func broadcastService() {
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:19132")
	if err != nil {
		log.Fatalf("Failed to resolve broadcast address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	log.Println("Broadcasting server presence...")

	for {
		packet := createBroadcastPacket()

		if config.Debug {
			log.Printf("Broadcasting packet (%d bytes)", len(packet))
		}

		_, err = conn.Write(packet)
		if err != nil {
			log.Printf("Broadcast error: %v", err)
		}

		time.Sleep(time.Duration(config.BroadcastInterval) * time.Second)
	}
}

func createBroadcastPacket() []byte {
	// Minecraft Bedrock uses specific format for LAN broadcasts
	// See: https://wiki.vg/Raknet_Protocol#Unconnected_Ping
	
	// Unconnected Ping packet structure
	magic := []byte{0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE, 0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78}
	
	// Current timestamp as int64
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	timestampBytes := make([]byte, 8)
	
	// Convert timestamp to bytes (little endian)
	for i := 0; i < 8; i++ {
		timestampBytes[i] = byte(timestamp >> (i * 8))
	}
	
	// Random client GUID
	guid := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	
	// Server ID string with proper MOTD format
	// Format: MCPE;<motd>;<protocol>;<version>;<players>;<max players>;<server id>;<subtitle>;<gamemode>;<default=1>;<port>;<port v6>
	serverID := fmt.Sprintf("MCPE;%s;527;1.19.0;0;10;%d;Bedrock Server;Survival;1;%d;%d",
		config.ServerName,
		0, // Server unique ID
		config.TargetServerPort,
		config.TargetServerPort,
	)
	
	// Build the packet
	packet := []byte{0x1C} // Unconnected Ping packet ID
	packet = append(packet, timestampBytes...)
	packet = append(packet, magic...)
	packet = append(packet, guid...)
	
	// Add server ID string with its length as a short
	serverIDBytes := []byte(serverID)
	serverIDLen := uint16(len(serverIDBytes))
	packet = append(packet, byte(serverIDLen), byte(serverIDLen>>8))
	packet = append(packet, serverIDBytes...)
	
	return packet
}

func startUDPListener() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:19132", config.LocalAddress))
	if err != nil {
		log.Fatalf("Failed to resolve listen address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to start UDP listener: %v", err)
	}
	defer conn.Close()

	log.Printf("UDP listener started on %s:19132", config.LocalAddress)

	buffer := make([]byte, 1500)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("UDP read error: %v", err)
			continue
		}

		if config.Debug {
			log.Printf("Received %d bytes from %s", n, clientAddr.String())
		}

		go handleUDPPacket(conn, clientAddr, buffer[:n])
	}
}

func handleUDPPacket(conn *net.UDPConn, clientAddr *net.UDPAddr, packet []byte) {
	if len(packet) == 0 {
		return
	}

	// Check if this is a ping packet (ID 0x01) and respond with our broadcast packet
	if len(packet) > 0 && packet[0] == 0x01 {
		if config.Debug {
			log.Printf("Received ping packet from %s, sending LAN broadcast response", clientAddr.String())
		}
		
		// Send our broadcast packet directly to the client
		response := createBroadcastPacket()
		_, err := conn.WriteToUDP(response, clientAddr)
		if err != nil {
			log.Printf("Failed to send ping response to client: %v", err)
		}
		return
	}

	// For all other packets, forward to the target server
	targetAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.TargetServerIP, config.TargetServerPort))
	if err != nil {
		log.Printf("Failed to resolve target address: %v", err)
		return
	}

	targetConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		log.Printf("Failed to connect to target server: %v", err)
		return
	}
	defer targetConn.Close()

	if config.Debug {
		log.Printf("Forwarding %d bytes to target server %s:%d", len(packet), config.TargetServerIP, config.TargetServerPort)
	}

	_, err = targetConn.Write(packet)
	if err != nil {
		log.Printf("Failed to forward packet to target: %v", err)
		return
	}

	// Set a deadline for reading the response
	targetBuffer := make([]byte, 1500)
	targetConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	// Read and forward all responses within the deadline
	for {
		n, _, err := targetConn.ReadFromUDP(targetBuffer)
		if err != nil {
			// If we got a timeout, that's normal - just exit the loop
			if strings.Contains(err.Error(), "timeout") {
				if config.Debug {
					log.Printf("Read timeout - finished processing responses")
				}
				break
			}
			
			log.Printf("Failed to read response from target: %v", err)
			break
		}

		if config.Debug {
			log.Printf("Received %d bytes from target, forwarding to client", n)
		}

		_, err = conn.WriteToUDP(targetBuffer[:n], clientAddr)
		if err != nil {
			log.Printf("Failed to send response to client: %v", err)
			break
		}
	}
}

func startTCPProxy() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:19132", config.LocalAddress))
	if err != nil {
		log.Fatalf("Failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP proxy started on %s:19132", config.LocalAddress)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept TCP connection: %v", err)
			continue
		}

		log.Printf("New TCP connection from %s", client.RemoteAddr())
		go handleTCPConnection(client)
	}
}

func handleTCPConnection(client net.Conn) {
	defer client.Close()

	target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", config.TargetServerIP, config.TargetServerPort))
	if err != nil {
		log.Printf("Failed to connect to target server: %v", err)
		return
	}
	defer target.Close()

	go func() {
		defer client.Close()
		defer target.Close()
		copyData(target, client)
	}()

	copyData(client, target)
}

func copyData(dst, src net.Conn) {
	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			return
		}

		_, err = dst.Write(buffer[:n])
		if err != nil {
			return
		}
	}
}