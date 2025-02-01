package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	bedrockPort    = 19132 // Default Minecraft Bedrock port
	proxyPort      = 19133 // Port for the proxy to listen on
	serverAddress  = "YOUR_EXTERNAL_SERVER_IP:19132" // Change to your actual server IP
	broadcastIP    = "255.255.255.255:19132"
	broadcastDelay = 5 * time.Second // Send LAN broadcast every 5 seconds
)

// LAN announcement packet structure
var lanAnnouncement = []byte{
	0x1c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x4d, 0x43, 0x50, 0x45, 0x3b, 0x4c, 0x41, 0x4e, 0x3b, 0x32, 0x33, 0x31, 0x35, 0x39,
	0x37, 0x32, 0x34, 0x37, 0x38, 0x30, 0x39, 0x38, 0x32, 0x31, 0x32, 0x3b,
	0x48, 0x61, 0x63, 0x6b, 0x65, 0x64, 0x20, 0x42, 0x65, 0x64, 0x72, 0x6f, 0x63, 0x6b, 0x3b,
	0x31, 0x39, 0x31, 0x33, 0x32, 0x3b, 0x31, 0x3b,
}

func main() {
	fmt.Println("Minecraft Bedrock Proxy started...")

	// Start broadcasting fake LAN game
	go broadcastFakeLAN()

	// Start UDP proxy server
	proxy, err := net.ListenUDP("udp", &net.UDPAddr{Port: proxyPort})
	if err != nil {
		fmt.Println("Error starting proxy:", err)
		os.Exit(1)
	}
	defer proxy.Close()
	fmt.Printf("Proxy listening on UDP port %d, forwarding to %s\n", proxyPort, serverAddress)

	buffer := make([]byte, 4096)

	for {
		n, addr, err := proxy.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from client:", err)
			continue
		}

		// Forward the packet to the actual server
		go forwardPacket(buffer[:n], addr)
	}
}

// Broadcasts a fake LAN game
func broadcastFakeLAN() {
	broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP)
	if err != nil {
		fmt.Println("Error resolving broadcast address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, broadcastAddr)
	if err != nil {
		fmt.Println("Error opening broadcast connection:", err)
		return
	}
	defer conn.Close()

	for {
		_, err := conn.Write(lanAnnouncement)
		if err != nil {
			fmt.Println("Error broadcasting LAN announcement:", err)
		}
		fmt.Println("Broadcasted fake LAN server...")
		time.Sleep(broadcastDelay)
	}
}

// Forwards a received packet to the real Bedrock server
func forwardPacket(data []byte, clientAddr *net.UDPAddr) {
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting to Bedrock server:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error forwarding packet:", err)
		return
	}

	// Read response from server and send it back to client
	buffer := make([]byte, 4096)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error reading response from server:", err)
		return
	}

	// Send response back to client
	clientConn, err := net.DialUDP("udp", nil, clientAddr)
	if err != nil {
		fmt.Println("Error opening connection to client:", err)
		return
	}
	defer clientConn.Close()

	_, err = clientConn.Write(buffer[:n])
	if err != nil {
		fmt.Println("Error sending response to client:", err)
	}
}