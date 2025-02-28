package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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

	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if config.Debug {
		log.Println("Debug mode enabled")
		log.Printf("Configuration: %+v", config)
	}

	log.Printf("Starting Minecraft LAN Bridge for PS4 -> %s:%d", config.TargetServerIP, config.TargetServerPort)

	go startUDPListener()
	go broadcastService()
	startTCPProxy()
}

// Loads the configuration from a file or creates a default config
func loadConfig() error {
	config = Config{
		LocalAddress:      "0.0.0.0",
		TargetServerIP:    "192.168.1.213", // Change to your Minecraft server IP
		TargetServerPort:  19132,
		BroadcastInterval: 5,
		ServerName:        "Epic World",
		Debug:             true,
	}

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

// Saves the configuration file
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

// Broadcasts a fake LAN server announcement to trick the PS4
func broadcastService() {
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:19132")
	if err != nil {
		log.Fatalf("Failed to resolve broadcast address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("Failed to create UDP broadcast connection: %v", err)
	}
	defer conn.Close()

	log.Println("Broadcasting LAN server presence...")

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

// Creates a fake LAN server broadcast packet
func createBroadcastPacket() []byte {
	magic := []byte{0xFE}
	serverInfo := []byte(config.ServerName)
	portBytes := []byte(strconv.Itoa(config.TargetServerPort))

	packet := append(magic, serverInfo...)
	packet = append(packet, ';')
	packet = append(packet, portBytes...)

	return packet
}

// Listens for UDP queries from the PS4 and forwards them to the real server
func startUDPListener() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:19132", config.LocalAddress))
	if err != nil {
		log.Fatalf("Failed to resolve UDP listener address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to start UDP listener: %v", err)
	}
	defer conn.Close()

	log.Printf("Listening for UDP queries on %s:19132", config.LocalAddress)

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
	targetAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.TargetServerIP, config.TargetServerPort))
	if err != nil {
		log.Printf("Failed to resolve target server address: %v", err)
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
		log.Printf("Failed to forward UDP packet: %v", err)
		return
	}

	targetBuffer := make([]byte, 1500)
	n, _, err := targetConn.ReadFromUDP(targetBuffer)
	if err != nil {
		log.Printf("Failed to read response from target: %v", err)
		return
	}

	_, err = conn.WriteToUDP(targetBuffer[:n], clientAddr)
	if err != nil {
		log.Printf("Failed to send response to client: %v", err)
	}
}

// Starts a TCP proxy to forward connections to the external server
func startTCPProxy() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:19132", config.LocalAddress))
	if err != nil {
		log.Fatalf("Failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP proxy running on %s:19132", config.LocalAddress)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept TCP connection: %v", err)
			continue
		}

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

	go func() { copyData(target, client) }()
	copyData(client, target)
}

func copyData(dst, src net.Conn) {
	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			return
		}
		dst.Write(buffer[:n])
	}
}
