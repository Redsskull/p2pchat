package discovery

import (
	"fmt"
	"net"
	"sync"
	"time"

	"p2pchat/internal/peer"
	"p2pchat/pkg/logger"
)

// PeerRegistry manages the list of discovered peers
type PeerRegistry struct {
	mu    sync.RWMutex
	peers map[string]*peer.Peer // key: peer ID

	// Configuration
	staleTimeout   time.Duration
	offlineTimeout time.Duration

	// Events
	onPeerJoin  func(*peer.Peer)
	onPeerLeave func(*peer.Peer)
}

// NewPeerRegistry creates a new peer registry
func NewPeerRegistry() *PeerRegistry {
	return &PeerRegistry{
		peers:          make(map[string]*peer.Peer),
		staleTimeout:   10 * time.Second,
		offlineTimeout: 30 * time.Second,
	}
}

// SetEventHandlers sets callbacks for peer join/leave events
func (pr *PeerRegistry) SetEventHandlers(onJoin, onLeave func(*peer.Peer)) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.onPeerJoin = onJoin
	pr.onPeerLeave = onLeave
}

// AddOrUpdatePeer adds a new peer or updates existing peer's last seen time
func (pr *PeerRegistry) AddOrUpdatePeer(msg *DiscoveryMessage, senderAddr *net.UDPAddr) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	// Convert UDP address to TCP address for connections
	tcpAddr := &net.TCPAddr{
		IP:   senderAddr.IP,
		Port: msg.Port, // Use the port from the message
	}

	existingPeer, exists := pr.peers[msg.PeerID]

	if exists {
		// Update existing peer
		existingPeer.UpdateLastSeen()
		logger.Debug("📱 Updated peer: %s (%s)", msg.Username, tcpAddr)
	} else {
		// Add new peer
		newPeer := &peer.Peer{
			ID:       msg.PeerID,
			Username: msg.Username,
			Address:  tcpAddr,
			LastSeen: time.Now(),
			Status:   peer.PeerStatusOnline,
		}

		pr.peers[msg.PeerID] = newPeer
		logger.Debug("✅ New peer joined: %s (%s)", msg.Username, tcpAddr)

		// Notify about new peer
		if pr.onPeerJoin != nil {
			pr.onPeerJoin(newPeer)
		}
	}
}

// GetAllPeers returns a copy of all peers
func (pr *PeerRegistry) GetAllPeers() []*peer.Peer {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	peers := make([]*peer.Peer, 0, len(pr.peers))
	for _, p := range pr.peers {
		// Create a copy to avoid race conditions
		peerCopy := *p
		peers = append(peers, &peerCopy)
	}
	return peers
}

// GetOnlinePeers returns only online and stale peers
func (pr *PeerRegistry) GetOnlinePeers() []*peer.Peer {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var onlinePeers []*peer.Peer
	for _, p := range pr.peers {
		if p.IsAlive() {
			peerCopy := *p
			onlinePeers = append(onlinePeers, &peerCopy)
		}
	}
	return onlinePeers
}

// GetPeerCount returns the number of online peers
func (pr *PeerRegistry) GetPeerCount() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	count := 0
	for _, p := range pr.peers {
		if p.IsAlive() {
			count++
		}
	}
	return count
}

// CleanupStaleePeers checks for timed out peers and removes them
func (pr *PeerRegistry) CleanupStalePeers() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	var toRemove []string

	for peerID, p := range pr.peers {
		p.CheckTimeout(pr.staleTimeout, pr.offlineTimeout)

		if p.Status == peer.PeerStatusOffline {
			toRemove = append(toRemove, peerID)
			logger.Debug("🔴 Peer went offline: %s", p.Username)

			// Notify about peer leaving
			if pr.onPeerLeave != nil {
				pr.onPeerLeave(p)
			}
		}
	}

	// Remove offline peers
	for _, peerID := range toRemove {
		delete(pr.peers, peerID)
	}

	if len(toRemove) > 0 {
		logger.Debug("🧹 Cleaned up %d offline peers", len(toRemove))
	}
}

// RemovePeer explicitly removes a peer (for graceful leave)
func (pr *PeerRegistry) RemovePeer(peerID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if p, exists := pr.peers[peerID]; exists {
		delete(pr.peers, peerID)
		logger.Debug("👋 Peer left gracefully: %s", p.Username)

		if pr.onPeerLeave != nil {
			pr.onPeerLeave(p)
		}
	}
}

// String returns a human-readable summary of peers
func (pr *PeerRegistry) String() string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return fmt.Sprintf("Registry: %d peers (%d online)",
		len(pr.peers), pr.GetPeerCount())
}
