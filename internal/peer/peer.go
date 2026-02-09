package peer

import (
	"net"
	"time"
)

// Peer represents a chat participant in the network
type Peer struct {
	ID       string       // Unique identifier (could be username + random suffix)
	Username string       // Display name
	Address  *net.TCPAddr // IP and port for TCP connections
	LastSeen time.Time    // When we last heard from this peer
	Status   PeerStatus   // Current status
}

// PeerStatus represents the current state of a peer
type PeerStatus int

const (
	PeerStatusUnknown PeerStatus = iota
	PeerStatusOnline             // Actively sending heartbeats
	PeerStatusStale              // Haven't heard from recently, but not timed out
	PeerStatusOffline            // Timed out, consider disconnected
)

// String returns human-readable status
func (s PeerStatus) String() string {
	switch s {
	case PeerStatusOnline:
		return "online"
	case PeerStatusStale:
		return "stale"
	case PeerStatusOffline:
		return "offline"
	default:
		return "unknown"
	}
}

// IsAlive returns true if peer should be considered active
func (p *Peer) IsAlive() bool {
	return p.Status == PeerStatusOnline || p.Status == PeerStatusStale
}

// UpdateLastSeen marks the peer as recently active
func (p *Peer) UpdateLastSeen() {
	p.LastSeen = time.Now()
	p.Status = PeerStatusOnline
}

// CheckTimeout updates peer status based on time since last contact
func (p *Peer) CheckTimeout(staleThreshold, offlineThreshold time.Duration) {
	elapsed := time.Since(p.LastSeen)

	switch {
	case elapsed > offlineThreshold:
		p.Status = PeerStatusOffline
	case elapsed > staleThreshold:
		p.Status = PeerStatusStale
	default:
		p.Status = PeerStatusOnline
	}
}
