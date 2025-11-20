package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"p2pchat/pkg/chat"
	"p2pchat/pkg/ui"
	"strconv"
	"time"

	"p2pchat/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
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

	// Set up logging
	if config.Debug {
		// Debug mode: log to file so it doesn't interfere with TUI
		err := logger.ToFile("p2pchat-debug.log")
		if err != nil {
			log.Printf("Failed to create log file, logging to stderr: %v", err)
		}
	} else {
		// Normal TUI mode: silent logging
		logger.Silent()
	}

	fmt.Printf("🚀 P2P Chat starting...\n")
	fmt.Printf("   Username: %s\n", config.Username)
	fmt.Printf("   Port: %d\n", config.Port)
	fmt.Printf("   Multicast: %s\n", config.MulticastAddr)

	// Create and start services...
	peerID := fmt.Sprintf("%s_%d", config.Username, time.Now().Unix()%1000)
	chatService, err := chat.NewChatService(peerID, config.Username, config.Port, config.MulticastAddr)
	if err != nil {
		log.Fatalf("Failed to create chat service: %v", err)
	}

	if err := chatService.Start(); err != nil {
		log.Fatalf("Failed to start chat service: %v", err)
	}
	defer chatService.Stop()

	// Start TUI
	model := ui.NewChatModel(chatService)
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
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
