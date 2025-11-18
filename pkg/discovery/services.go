package discovery

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"p2pchat/internal/peer"
)

// DiscoveryService coordinates peer discovery via UDP multicast
type DiscoveryService struct {
	// Core components
	multicast *MulticastService
	registry  *PeerRegistry

	// Local peer info
	localPeerID   string
	localUsername string
	localTCPPort  int

	// Configuration
	beaconInterval  time.Duration
	cleanupInterval time.Duration

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(username string, tcpPort int, multicastAddr string) (*DiscoveryService, error) {
	// Create multicast service
	multicast, err := NewMulticastService(multicastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create multicast service: %w", err)
	}

	// Create peer registry
	registry := NewPeerRegistry()

	// Generate unique peer ID
	peerID := fmt.Sprintf("%s_%d", username, time.Now().Unix())

	return &DiscoveryService{
		multicast:       multicast,
		registry:        registry,
		localPeerID:     peerID,
		localUsername:   username,
		localTCPPort:    tcpPort,
		beaconInterval:  5 * time.Second,  // Announce every 5 seconds
		cleanupInterval: 10 * time.Second, // Cleanup every 10 seconds
	}, nil
}

// SetPeerEventHandlers sets callbacks for when peers join/leave
func (ds *DiscoveryService) SetPeerEventHandlers(onJoin, onLeave func(*peer.Peer)) {
	ds.registry.SetEventHandlers(onJoin, onLeave)
}

// Start begins the discovery service
func (ds *DiscoveryService) Start() error {
	// Start multicast listening
	if err := ds.multicast.Start(); err != nil {
		return fmt.Errorf("failed to start multicast: %w", err)
	}

	// Create context for coordinating goroutines
	ds.ctx, ds.cancel = context.WithCancel(context.Background())

	log.Printf("🚀 Discovery service started")
	log.Printf("   Local peer: %s (%s)", ds.localUsername, ds.localPeerID)
	log.Printf("   TCP port: %d", ds.localTCPPort)

	// Start background tasks
	go ds.beaconLoop()
	go ds.receiveLoop()
	go ds.cleanupLoop()

	// Send initial announcement
	if err := ds.sendAnnouncement(); err != nil {
		log.Printf("⚠️  Failed to send initial announcement: %v", err)
	}

	return nil
}

// Stop gracefully shuts down the discovery service
func (ds *DiscoveryService) Stop() error {
	if ds.cancel != nil {
		// Send leave message
		ds.sendLeaveMessage()

		// Stop background tasks
		ds.cancel()

		// Stop multicast
		if err := ds.multicast.Stop(); err != nil {
			return fmt.Errorf("failed to stop multicast: %w", err)
		}

		log.Printf("👋 Discovery service stopped")
	}
	return nil
}

// GetPeers returns current list of discovered peers
func (ds *DiscoveryService) GetAllPeers() []*peer.Peer {
	return ds.registry.GetAllPeers()
}

// GetOnlinePeers returns only online peers
func (ds *DiscoveryService) GetOnlinePeers() []*peer.Peer {
	return ds.registry.GetOnlinePeers()
}

// GetPeerCount returns number of online peers
func (ds *DiscoveryService) GetPeerCount() int {
	return ds.registry.GetPeerCount()
}

// beaconLoop sends periodic announcements
func (ds *DiscoveryService) beaconLoop() {
	ticker := time.NewTicker(ds.beaconInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			log.Printf("🔊 Beacon loop stopping")
			return
		case <-ticker.C:
			if err := ds.sendAnnouncement(); err != nil {
				log.Printf("⚠️  Failed to send beacon: %v", err)
			}
		}
	}
}

// receiveLoop listens for incoming discovery messages
func (ds *DiscoveryService) receiveLoop() {
	for {
		select {
		case <-ds.ctx.Done():
			log.Printf("📡 Receive loop stopping")
			return
		default:
			// Try to receive a message
			msg, senderAddr, err := ds.multicast.ReceiveWithTimeout(1 * time.Second)
			if err != nil {
				// Timeout is normal, continue
				continue
			}

			// Handle the message
			ds.handleDiscoveryMessage(msg, senderAddr)
		}
	}
}

// cleanupLoop periodically removes stale peers
func (ds *DiscoveryService) cleanupLoop() {
	ticker := time.NewTicker(ds.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			log.Printf("🧹 Cleanup loop stopping")
			return
		case <-ticker.C:
			ds.registry.CleanupStalePeers()
		}
	}
}

// sendAnnouncement broadcasts presence
func (ds *DiscoveryService) sendAnnouncement() error {
	msg := NewAnnounceMessage(ds.localPeerID, ds.localUsername, ds.localTCPPort)

	// Set our address (will be overridden by receiver, but good for debugging)
	localAddr := ds.multicast.GetLocalAddr()
	if localAddr != nil {
		msg.Address = fmt.Sprintf("%s:%d", localAddr.IP, ds.localTCPPort)
	}

	return ds.multicast.Send(msg)
}

// sendLeaveMessage announces going offline
func (ds *DiscoveryService) sendLeaveMessage() {
	msg := &DiscoveryMessage{
		Type:      MessageTypeLeave,
		PeerID:    ds.localPeerID,
		Username:  ds.localUsername,
		Port:      ds.localTCPPort,
		Timestamp: time.Now(),
	}

	// Best effort - don't wait for errors
	ds.multicast.Send(msg)

	// Give it a moment to send
	time.Sleep(100 * time.Millisecond)
}

// handleDiscoveryMessage processes incoming discovery messages
func (ds *DiscoveryService) handleDiscoveryMessage(msg *DiscoveryMessage, senderAddr *net.UDPAddr) {
	// Ignore our own messages
	if msg.PeerID == ds.localPeerID {
		return
	}

	// Check message age (ignore very old messages)
	if !msg.IsRecent(30 * time.Second) {
		log.Printf("⏰ Ignoring old message from %s", msg.Username)
		return
	}

	switch msg.Type {
	case MessageTypeAnnounce, MessageTypePing:
		// Add or update peer
		ds.registry.AddOrUpdatePeer(msg, senderAddr)

	case MessageTypeLeave:
		// Remove peer gracefully
		ds.registry.RemovePeer(msg.PeerID)

	case MessageTypePong:
		// Update peer's last seen time
		ds.registry.AddOrUpdatePeer(msg, senderAddr)

	default:
		log.Printf("❓ Unknown message type: %s from %s", msg.Type, msg.Username)
	}
}
