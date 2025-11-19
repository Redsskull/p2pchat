package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"p2pchat/pkg/chat"
)

func main() {
	// Parse command line flags
	username := flag.String("username", getDefaultUsername(), "Your username")
	port := flag.Int("port", 8080, "TCP port for connections")
	multicast := flag.String("multicast", "224.0.0.1:9999", "Multicast address")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Generate a simple peer ID
	peerID := fmt.Sprintf("%s_%d", *username, time.Now().Unix()%1000)

	fmt.Printf("\n🎉 P2P Chat Test - Let's Make Humans Talk!\n")
	fmt.Printf("   👤 Username: %s\n", *username)
	fmt.Printf("   🆔 Peer ID: %s\n", peerID)
	fmt.Printf("   🔌 Port: %d\n", *port)
	fmt.Printf("   📡 Multicast: %s\n", *multicast)
	fmt.Printf("\n💬 Instructions:\n")
	fmt.Printf("   • Type messages and press Enter to send\n")
	fmt.Printf("   • Type 'status' to see connected peers\n")
	fmt.Printf("   • Type 'quit' or press Ctrl+C to exit\n")
	fmt.Printf("   • Run another instance in a different terminal to chat!\n")
	fmt.Print("\n" + strings.Repeat("═", 60) + "\n")

	// Create chat service
	service, err := chat.NewChatService(peerID, *username, *port, *multicast)
	if err != nil {
		log.Fatalf("❌ Failed to create chat service: %v", err)
	}

	// Start the service
	if err := service.Start(); err != nil {
		log.Fatalf("❌ Failed to start chat service: %v", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Start message display goroutine
	go displayMessages(service, ctx)

	// Start status updates
	go statusUpdates(service, ctx)

	// Interactive input loop - this is where the human types!
	go handleUserInput(service, ctx, cancel)

	// Wait for shutdown
	select {
	case <-sigChan:
		fmt.Printf("\n\n🛑 Shutting down gracefully...\n")
	case <-ctx.Done():
		fmt.Printf("\n\n👋 Chat session ended.\n")
	}

	// Graceful shutdown
	cancel()
	if err := service.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Printf("✅ P2P Chat Test completed. Thanks for chatting! 💬\n")
}

// displayMessages shows incoming chat messages - this is where humans see each other's words!
func displayMessages(service *chat.ChatService, ctx context.Context) {
	messages := service.GetMessages()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-messages:
			if !ok {
				return // Channel closed
			}

			// Display the message to the human
			displayMessage(msg)
		}
	}
}

// displayMessage formats and shows a single message
func displayMessage(msg *chat.Message) {
	timestamp := msg.Timestamp.Format("15:04:05")

	switch msg.Type {
	case chat.MessageTypeChat:
		// This is the magic moment - human words appearing on screen!
		fmt.Printf("💬 [%s] %s: %s\n", timestamp, msg.Username, msg.Content)
	case chat.MessageTypeJoin:
		fmt.Printf("🎉 [%s] %s joined the chat!\n", timestamp, msg.Username)
	case chat.MessageTypeLeave:
		fmt.Printf("👋 [%s] %s left the chat\n", timestamp, msg.Username)
	case chat.MessageTypeHeartbeat:
		// Don't show heartbeats to users
		return
	default:
		fmt.Printf("📨 [%s] %s: %s\n", timestamp, msg.Username, msg.Content)
	}

	// Show prompt again
	fmt.Printf("💭 You: ")
}

// statusUpdates shows periodic status information
func statusUpdates(service *chat.ChatService, ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status := service.GetStatus()
			peers := service.GetConnectedPeers()

			fmt.Printf("\n📊 Status Update:\n")
			fmt.Printf("   🔍 Discovered peers: %d\n", status.DiscoveredPeers)
			fmt.Printf("   🔗 Connected peers: %d\n", status.ConnectedPeers)
			fmt.Printf("   📤 Messages sent: %d\n", status.MessagesSent)

			if len(peers) > 0 {
				fmt.Printf("   👥 Connected to:\n")
				for _, peer := range peers {
					statusIcon := "🔴" // Not connected
					if peer.Connected {
						statusIcon = "🟢" // Connected
					}
					fmt.Printf("      %s %s (%s)\n", statusIcon, peer.Username, peer.Address)
				}
			}
			fmt.Printf("💭 You: ")
		}
	}
}

// handleUserInput processes what the human types
func handleUserInput(service *chat.ChatService, ctx context.Context, cancel context.CancelFunc) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("💭 You: ")

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			text := strings.TrimSpace(scanner.Text())

			if text == "" {
				fmt.Printf("💭 You: ")
				continue
			}

			// Handle special commands
			switch strings.ToLower(text) {
			case "quit", "exit", "q":
				fmt.Printf("👋 Goodbye!\n")
				cancel()
				return

			case "status", "s":
				showStatus(service)

			case "help", "h":
				showHelp()

			case "peers", "p":
				showPeers(service)

			default:
				// Regular chat message - this is the human-to-human magic!
				err := service.SendMessage(text)
				if err != nil {
					fmt.Printf("❌ Failed to send message: %v\n", err)
				}
			}

			fmt.Printf("💭 You: ")
		}
	}
}

// showStatus displays current service status
func showStatus(service *chat.ChatService) {
	status := service.GetStatus()
	peers := service.GetConnectedPeers()

	fmt.Printf("\n📊 Current Status:\n")
	fmt.Printf("   👤 Username: %s\n", status.Username)
	fmt.Printf("   🆔 Peer ID: %s\n", status.PeerID)
	fmt.Printf("   🔌 Port: %d\n", status.Port)
	fmt.Printf("   🔍 Discovered peers: %d\n", status.DiscoveredPeers)
	fmt.Printf("   🔗 Connected peers: %d\n", status.ConnectedPeers)
	fmt.Printf("   📤 Messages sent: %d\n", status.MessagesSent)

	if len(peers) > 0 {
		fmt.Printf("   👥 Peer Details:\n")
		for _, peer := range peers {
			discoveryIcon := "🔍"
			connectionIcon := "🔴"

			if peer.Discovered {
				discoveryIcon = "📡"
			}
			if peer.Connected {
				connectionIcon = "🔗"
			}

			fmt.Printf("      %s %s %s (%s) - last seen: %s\n",
				discoveryIcon, connectionIcon, peer.Username, peer.Address,
				time.Since(peer.LastSeen).Round(time.Second))
		}
	} else {
		fmt.Printf("   👥 No peers found yet. Run another instance to chat!\n")
	}
	fmt.Printf("\n")
}

// showPeers displays connected peers
func showPeers(service *chat.ChatService) {
	peers := service.GetConnectedPeers()

	fmt.Printf("\n👥 Connected Peers (%d):\n", len(peers))
	if len(peers) == 0 {
		fmt.Printf("   No peers connected yet.\n")
		fmt.Printf("   💡 Tip: Run another instance with a different username!\n")
	} else {
		for i, peer := range peers {
			statusText := "Discovered"
			if peer.Connected {
				statusText = "Connected & Chatting"
			}
			fmt.Printf("   %d. %s (%s) - %s\n", i+1, peer.Username, peer.Address, statusText)
		}
	}
	fmt.Printf("\n")
}

// showHelp displays available commands
func showHelp() {
	fmt.Printf("\n📖 Available Commands:\n")
	fmt.Printf("   💬 Just type a message and press Enter to chat\n")
	fmt.Printf("   📊 'status' or 's' - Show detailed status\n")
	fmt.Printf("   👥 'peers' or 'p' - Show connected peers\n")
	fmt.Printf("   📖 'help' or 'h' - Show this help\n")
	fmt.Printf("   👋 'quit' or 'q' - Exit the chat\n")
	fmt.Printf("   ⚡ Ctrl+C - Quick exit\n")
	fmt.Printf("\n💡 Pro tip: Open multiple terminals to chat with yourself!\n\n")
}

// getDefaultUsername creates a default username
func getDefaultUsername() string {
	if username := os.Getenv("USER"); username != "" {
		return username
	}
	if username := os.Getenv("USERNAME"); username != "" {
		return username
	}

	// Generate a simple default
	pid := os.Getpid()
	return "user" + strconv.Itoa(pid%1000)
}
