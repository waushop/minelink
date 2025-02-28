package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
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
	log.Printf("Starting PS4 Minecraft bridge to %s:%d", config.TargetServerIP, config.TargetServerPort)

	// Start UDP listener for LAN broadcast responses
	go startUDPListener()

	// Start broadcasting service presence
	go broadcastService()

	// Start TCP proxy server
	startTCPProxy()
}

func loadConfig() error {
	// Default configuration
	config = Config{
		LocalAddress:      "0.0.0.0",
		TargetServerIP:    "127.0.0.1",
		TargetServerPort:  19132,
		BroadcastInterval: 5,
		ServerName:        "PS4 Minecraft Bridge",
		Debug:             false,
	}

	// Try to load from file
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Config file not found, creating default config")
			saveConfig()
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&config)
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
	magic := []byte{0xFE}
	serverInfo := []byte(config.ServerName)
	portBytes := []byte(strconv.Itoa(config.TargetServerPort))

	packet := append(magic, serverInfo...)
	packet = append(packet, ';')
	packet = append(packet, portBytes...)

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

	_, err = targetConn.Write(packet)
	if err != nil {
		log.Printf("Failed to forward packet to target: %v", err)
		return
	}

	targetBuffer := make([]byte, 1500)
	targetConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := targetConn.ReadFromUDP(targetBuffer)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			if config.Debug {
				log.Printf("Timeout waiting for response from target server")
			}
		} else {
			log.Printf("Failed to read response from target: %v", err)
		}
		return
	}

	_, err = conn.WriteToUDP(targetBuffer[:n], clientAddr)
	if err != nil {
		log.Printf("Failed to send response to client: %v", err)
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