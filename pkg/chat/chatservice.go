package chat

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"p2pchat/internal/peer"
	"p2pchat/pkg/discovery"
)

// ChatService is the main service that coordinates discovery and chat messaging
// This is where UDP discovery meets TCP chat - the magic integration layer!
type ChatService struct {
	// Identity
	peerID   string
	username string
	port     int

	// Core services
	discovery   *discovery.DiscoveryService
	connections *ConnectionManager

	// Message handling
	messageSequence  uint64        // Atomic counter for message ordering
	incomingMessages chan *Message // Channel for UI to receive messages

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewChatService creates a new integrated chat service
func NewChatService(peerID, username string, port int, multicastAddr string) (*ChatService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create discovery service
	discoveryService, err := discovery.NewDiscoveryService(username, port, multicastAddr)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create discovery service: %w", err)
	}

	// Create connection manager
	connectionManager := NewConnectionManager(peerID, username, port)

	service := &ChatService{
		peerID:           peerID,
		username:         username,
		port:             port,
		discovery:        discoveryService,
		connections:      connectionManager,
		incomingMessages: make(chan *Message, 100), // Buffer incoming messages for UI
		ctx:              ctx,
		cancel:           cancel,
	}

	// Set up the beautiful integration between discovery and connections
	service.setupIntegration()

	return service, nil
}

// setupIntegration is where the magic happens - UDP discovery feeds TCP connections!
func (cs *ChatService) setupIntegration() {
	// When discovery finds a new peer, automatically connect via TCP
	cs.discovery.SetPeerEventHandlers(
		// On peer join - this is the UDP→TCP bridge!
		func(p *peer.Peer) {
			log.Printf("🎉 Discovery found peer: %s (%s) - connecting via TCP...", p.Username, p.ID)

			// Convert UDP discovery into TCP connection
			err := cs.connections.ConnectToPeer(p)
			if err != nil {
				log.Printf("❌ Failed to connect to peer %s: %v", p.Username, err)
			} else {
				log.Printf("✅ TCP connection established with %s!", p.Username)

				// Send a join message to let them know we're here
				joinMsg := NewJoinMessage(cs.peerID, cs.username, cs.nextSequence())
				cs.connections.SendToPeer(p.ID, joinMsg)
			}
		},

		// On peer leave - handle disconnections gracefully
		func(p *peer.Peer) {
			log.Printf("👋 Peer left discovery: %s (%s)", p.Username, p.ID)
			// TCP connection will timeout naturally, but I could force disconnect here
		},
	)

	// Handle incoming TCP messages
	cs.connections.SetMessageHandler(func(msg *Message, fromPeerID string) {
		log.Printf("📨 Received message from %s: %s", msg.Username, msg.Content)

		// Forward message to UI (this is how messages reach the human!)
		select {
		case cs.incomingMessages <- msg:
			// Message delivered to UI
		default:
			// UI message buffer full - this shouldn't happen in normal use
			log.Printf("⚠️ UI message buffer full, dropping message from %s", msg.Username)
		}
	})

}

// Start begins the chat service - this starts both UDP discovery and TCP listening
func (cs *ChatService) Start() error {
	log.Printf("🚀 Starting chat service for %s on port %d", cs.username, cs.port)

	// Start UDP discovery
	if err := cs.discovery.Start(); err != nil {
		return fmt.Errorf("failed to start discovery: %w", err)
	}
	log.Printf("📡 UDP discovery started - looking for peers...")

	// Start TCP connection manager
	if err := cs.connections.Start(); err != nil {
		cs.discovery.Stop()
		return fmt.Errorf("failed to start connections: %w", err)
	}
	log.Printf("🔌 TCP listener started - ready for peer connections...")

	log.Printf("✅ Chat service fully started! Ready for human conversations! 💬")
	return nil
}

// SendMessage sends a chat message to all connected peers
// This is the function that makes human-to-human communication happen!
func (cs *ChatService) SendMessage(content string) error {
	if content == "" {
		return fmt.Errorf("cannot send empty message")
	}

	// Create the message
	msg := NewChatMessage(cs.peerID, cs.username, content, cs.nextSequence())

	log.Printf("📤 Sending message to all peers: %s", content)

	// Broadcast to all connected peers - this is the magic moment!
	cs.connections.Broadcast(msg)

	// Also add to our own message stream for the UI
	select {
	case cs.incomingMessages <- msg:
		// Our own message appears in our UI too
	default:
		// Buffer full - very unlikely
		log.Printf("⚠️ Failed to add own message to UI buffer")
	}

	return nil
}

// GetMessages returns a channel for receiving incoming messages
// The UI reads from this channel to show messages to the human
func (cs *ChatService) GetMessages() <-chan *Message {
	return cs.incomingMessages
}

// GetConnectedPeers returns information about currently connected peers
func (cs *ChatService) GetConnectedPeers() []PeerInfo {
	// Get peers from discovery (UDP - who's announcing)
	discoveredPeers := cs.discovery.GetOnlinePeers()

	// Get peers from connections (TCP - who we're chatting with)
	connectedPeerIDs := cs.connections.GetConnectedPeers()

	// Create a combined view
	peerInfos := make([]PeerInfo, 0, len(discoveredPeers))

	for _, p := range discoveredPeers {
		info := PeerInfo{
			PeerID:     p.ID,
			Username:   p.Username,
			Address:    p.Address.String(),
			Status:     p.Status.String(),
			LastSeen:   p.LastSeen,
			Discovered: true,  // Found via UDP discovery
			Connected:  false, // Default to false
		}

		// Check if we also have a TCP connection
		for _, connectedID := range connectedPeerIDs {
			if connectedID == p.ID {
				info.Connected = true
				break
			}
		}

		peerInfos = append(peerInfos, info)
	}

	return peerInfos
}

// PeerInfo provides a combined view of peer discovery and connection status
type PeerInfo struct {
	PeerID     string
	Username   string
	Address    string
	Status     string // From discovery service
	LastSeen   time.Time
	Discovered bool // Found via UDP discovery
	Connected  bool // Has active TCP connection
}

// nextSequence returns the next message sequence number
func (cs *ChatService) nextSequence() uint64 {
	return atomic.AddUint64(&cs.messageSequence, 1)
}

// NotifyPeerJoin sends a join notification to all peers
func (cs *ChatService) NotifyPeerJoin() {
	joinMsg := NewJoinMessage(cs.peerID, cs.username, cs.nextSequence())
	cs.connections.Broadcast(joinMsg)
}

// NotifyPeerLeave sends a leave notification to all peers
func (cs *ChatService) NotifyPeerLeave() {
	leaveMsg := NewLeaveMessage(cs.peerID, cs.username, cs.nextSequence())
	cs.connections.Broadcast(leaveMsg)
}

// SendHeartbeat sends a heartbeat to all connected peers
func (cs *ChatService) SendHeartbeat() {
	heartbeat := NewHeartbeatMessage(cs.peerID, cs.username, cs.nextSequence())
	cs.connections.Broadcast(heartbeat)
}

// GetStatus returns current service status
func (cs *ChatService) GetStatus() ServiceStatus {
	discoveredPeers := cs.discovery.GetOnlinePeers()
	connectedPeers := cs.connections.GetConnectedPeers()

	return ServiceStatus{
		Username:        cs.username,
		PeerID:          cs.peerID,
		Port:            cs.port,
		DiscoveredPeers: len(discoveredPeers),
		ConnectedPeers:  len(connectedPeers),
		MessagesSent:    cs.messageSequence,
	}
}

// ServiceStatus provides overall service information
type ServiceStatus struct {
	Username        string
	PeerID          string
	Port            int
	DiscoveredPeers int // Found via UDP
	ConnectedPeers  int // Connected via TCP
	MessagesSent    uint64
}

// Stop gracefully shuts down the chat service
func (cs *ChatService) Stop() error {
	log.Printf("🛑 Stopping chat service...")

	// Send leave notification to all peers
	cs.NotifyPeerLeave()

	// Give a moment for leave messages to be sent
	time.Sleep(100 * time.Millisecond)

	// Cancel all operations
	cs.cancel()

	// Stop services in reverse order
	var err error
	if stopErr := cs.connections.Stop(); stopErr != nil {
		log.Printf("Error stopping connections: %v", stopErr)
		err = stopErr
	}

	if stopErr := cs.discovery.Stop(); stopErr != nil {
		log.Printf("Error stopping discovery: %v", stopErr)
		if err == nil {
			err = stopErr
		}
	}

	// Close message channel
	close(cs.incomingMessages)

	// Wait for all goroutines
	cs.wg.Wait()

	log.Printf("✅ Chat service stopped")
	return err
}
