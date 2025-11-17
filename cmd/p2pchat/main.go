package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	DefaultUsername      = "user"
	DefaultPort          = 8080
	DefaultMulticastAddr = "224.0.0.1:9999"
)

type Config struct {
	Username      string
	Port          int
	MulticastAddr string
	Debug         bool
}

func main() {
	config := parseArgs()

	if config.Debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Starting P2P Chat with config: %+v", config)
	}

	fmt.Printf("🚀 P2P Chat starting...\n")
	fmt.Printf("   Username: %s\n", config.Username)
	fmt.Printf("   Port: %d\n", config.Port)
	fmt.Printf("   Multicast: %s\n", config.MulticastAddr)
	fmt.Printf("   Debug: %v\n", config.Debug)
	fmt.Printf("\n")

	// TODO: Initialize components
	// 1. Start peer discovery service
	// 2. Start connection manager
	// 3. Start message router
	// 4. Launch terminal UI

	fmt.Println("📋 TODO: Implementation starts Day 2!")
	fmt.Println("   - Peer discovery via UDP multicast")
	fmt.Println("   - TCP connection management")
	fmt.Println("   - Bubble Tea terminal interface")

	// For now, just show I can parse args and exit
	fmt.Printf("\n✅ Day 1 complete! Architecture designed, project structured.\n")
}

func parseArgs() *Config {
	var (
		username  = flag.String("username", getDefaultUsername(), "Username for chat (default: system username or 'user')")
		port      = flag.Int("port", DefaultPort, "TCP port for peer connections")
		multicast = flag.String("multicast", DefaultMulticastAddr, "Multicast address for peer discovery")
		debug     = flag.Bool("debug", false, "Enable debug logging")
		help      = flag.Bool("help", false, "Show help message")
		h         = flag.Bool("h", false, "Show help message (shorthand)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "P2P Chat - IRC-style peer-to-peer chat system\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -username alice -port 8080\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -debug -multicast 224.0.0.2:9999\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nProject Status: Day 1 - Architecture & Design Complete\n")
		fmt.Fprintf(os.Stderr, "Next: Day 2 - Peer Discovery Implementation\n")
	}

	flag.Parse()

	if *help || *h {
		flag.Usage()
		os.Exit(0)
	}

	// Validate port range
	if *port < 1024 || *port > 65535 {
		fmt.Fprintf(os.Stderr, "Error: Port must be between 1024 and 65535\n")
		os.Exit(1)
	}

	return &Config{
		Username:      *username,
		Port:          *port,
		MulticastAddr: *multicast,
		Debug:         *debug,
	}
}

func getDefaultUsername() string {
	if username := os.Getenv("USER"); username != "" {
		return username
	}
	if username := os.Getenv("USERNAME"); username != "" {
		return username
	}

	// Try to get hostname as fallback
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}

	// Generate a simple random username as last resort
	pid := os.Getpid()
	return "user" + strconv.Itoa(pid%1000)
}
