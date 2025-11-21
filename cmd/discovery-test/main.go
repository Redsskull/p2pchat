package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"p2pchat/internal/peer"
	"p2pchat/pkg/discovery"
)

func main() {
	// Parse command line flags
	username := flag.String("username", "user", "Your username")
	port := flag.Int("port", 8080, "TCP port for connections")
	multicast := flag.String("multicast", "224.0.0.1:9999", "Multicast address")
	flag.Parse()

	fmt.Printf("🧪 Discovery Test - User: %s, Port: %d\n", *username, *port)
	fmt.Printf("📡 Multicast: %s\n", *multicast)
	fmt.Println("Press Ctrl+C to quit")
	fmt.Println()

	// Generate peer ID for consistency
	peerID := fmt.Sprintf("%s_%d", *username, time.Now().Unix()%1000)

	// Create discovery service
	service, err := discovery.NewDiscoveryService(peerID, *username, *port, *multicast)
	if err != nil {
		log.Fatalf("Failed to create discovery service: %v", err)
	}

	// Set up event handlers
	service.SetPeerEventHandlers(
		// On peer join
		func(p *peer.Peer) {
			fmt.Printf("🎉 PEER JOINED: %s (%s) at %s\n",
				p.Username, p.ID, p.Address)
		},
		// On peer leave
		func(p *peer.Peer) {
			fmt.Printf("👋 PEER LEFT: %s (%s)\n",
				p.Username, p.ID)
		},
	)

	// Start the service
	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start discovery service: %v", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Status updates every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				peers := service.GetOnlinePeers()
				fmt.Printf("\n📊 STATUS: %d peers online\n", len(peers))
				for _, p := range peers {
					fmt.Printf("   • %s (%s) - %s\n",
						p.Username, p.Status, p.Address)
				}
				fmt.Println()
			}
		}
	}()

	// Interactive commands (optional)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "peers" || text == "p" {
				peers := service.GetAllPeers()
				fmt.Printf("\n🌐 All peers (%d):\n", len(peers))
				for _, p := range peers {
					fmt.Printf("   • %s (%s) - %s - last seen: %s\n",
						p.Username, p.Status, p.Address,
						time.Since(p.LastSeen).Round(time.Second))
				}
				fmt.Println()
			} else if text == "help" || text == "h" {
				fmt.Println("\n📖 Commands:")
				fmt.Println("   peers, p  - Show all peers")
				fmt.Println("   help, h   - Show this help")
				fmt.Println("   Ctrl+C    - Quit")
				fmt.Println()
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Printf("\n🛑 Shutting down...\n")

	// Graceful shutdown
	cancel()
	if err := service.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Printf("✅ Discovery test completed\n")
}
