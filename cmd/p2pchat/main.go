package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"p2pchat/pkg/chat"
	"p2pchat/pkg/ui"
	"strconv"
	"strings"
	"time"

	"p2pchat/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	DefaultUsername      = "" // Empty to trigger interactive prompt
	DefaultPort          = 0  // 0 to trigger automatic assignment
	DefaultMulticastAddr = "224.0.0.1:9999"
	PortRangeStart       = 8080 // Start of automatic port range
	PortRangeEnd         = 8999 // End of automatic port range
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

	fmt.Printf("🚀 Starting P2P Chat...\n")
	fmt.Printf("   👤 Username: %s\n", config.Username)
	fmt.Printf("   🔌 Port: %d\n", config.Port)
	fmt.Printf("   📡 Discovery: %s\n", config.MulticastAddr)
	if config.Debug {
		fmt.Printf("   🔍 Debug: Enabled (logging to p2pchat-debug.log)\n")
	}
	fmt.Printf("\n🔄 Initializing services...\n")

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

	fmt.Printf("✅ Ready! Starting chat interface...\n\n")

	if _, err := program.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}

func parseArgs() *Config {
	var (
		username  = flag.String("username", DefaultUsername, "Username for chat (interactive prompt if not provided)")
		port      = flag.Int("port", DefaultPort, "TCP port for peer connections (auto-assigned if not provided)")
		multicast = flag.String("multicast", DefaultMulticastAddr, "Multicast address for peer discovery")
		debug     = flag.Bool("debug", false, "Enable debug logging")
		help      = flag.Bool("help", false, "Show help message")
		h         = flag.Bool("h", false, "Show help message (shorthand)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "P2P Chat - IRC-style peer-to-peer chat system\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Simple usage (interactive prompts):\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                                    # Interactive mode\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -username alice                    # Specify username, auto-assign port\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -username alice -port 8080         # Full manual configuration\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -debug                             # Interactive mode with debug logging\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nStatus: Production Ready (Day 8) ✅\n")
	}

	flag.Parse()

	if *help || *h {
		flag.Usage()
		os.Exit(0)
	}

	config := &Config{
		Username:      *username,
		Port:          *port,
		MulticastAddr: *multicast,
		Debug:         *debug,
	}

	// Interactive configuration if needed
	config = enhanceConfigInteractively(config)

	// Validate final port
	if config.Port < 1024 || config.Port > 65535 {
		fmt.Fprintf(os.Stderr, "Error: Port must be between 1024 and 65535\n")
		os.Exit(1)
	}

	return config
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

// enhanceConfigInteractively prompts user for missing configuration
func enhanceConfigInteractively(config *Config) *Config {
	// Interactive username prompt if not provided
	if config.Username == "" {
		config.Username = promptForUsername()
	}

	// Auto-assign port if not provided
	if config.Port == 0 {
		config.Port = findAvailablePort()
		fmt.Printf("🔌 Auto-assigned port: %d\n", config.Port)
	}

	return config
}

// promptForUsername interactively asks for username
func promptForUsername() string {
	fmt.Printf("🌟 Welcome to P2P Chat - Decentralized IRC-style chat!\n")
	fmt.Printf("📖 No server needed - connect directly with peers on your network\n\n")
	fmt.Printf("👤 Enter your username: ")

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Error reading input: %v\n", err)
			os.Exit(1)
		}

		username := strings.TrimSpace(input)
		if username == "" {
			fmt.Printf("❌ Username cannot be empty. Please try again: ")
			continue
		}

		// Basic validation
		if len(username) > 20 {
			fmt.Printf("❌ Username too long (max 20 characters). Please try again: ")
			continue
		}

		if strings.ContainsAny(username, " \t\n\r") {
			fmt.Printf("❌ Username cannot contain spaces. Please try again: ")
			continue
		}

		fmt.Printf("✅ Username set: %s\n", username)
		return username
	}
}

// findAvailablePort automatically finds an available port in the range
func findAvailablePort() int {
	for port := PortRangeStart; port <= PortRangeEnd; port++ {
		if isPortAvailable(port) {
			return port
		}
	}

	// Fallback to system-assigned port if range is exhausted
	fmt.Printf("⚠️  Preferred port range (8080-8999) full, using system-assigned port...\n")
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to find available port: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port
}

// isPortAvailable checks if a port is available for TCP binding
func isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}
