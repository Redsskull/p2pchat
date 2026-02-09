package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// DiscoveryMessage represents a peer announcement on the network
type DiscoveryMessage struct {
	Type      MessageType `json:"type"`
	PeerID    string      `json:"peer_id"`
	Username  string      `json:"username"`
	Address   string      `json:"address"` // "192.168.1.100:8080"
	Port      int         `json:"port"`    // TCP port for chat connections
	Timestamp time.Time   `json:"timestamp"`
	Sequence  uint64      `json:"sequence"` // Message counter for ordering
}

// MessageType defines the kind of discovery message
type MessageType string

const (
	MessageTypeAnnounce MessageType = "announce" // "I'm here!"
	MessageTypePing     MessageType = "ping"     // "Are you still there?"
	MessageTypePong     MessageType = "pong"     // "Yes, I'm still here!"
	MessageTypeLeave    MessageType = "leave"    // "I'm going offline"
)

// NewAnnounceMessage creates a peer announcement
func NewAnnounceMessage(peerID, username string, tcpPort int) *DiscoveryMessage {
	return &DiscoveryMessage{
		Type:      MessageTypeAnnounce,
		PeerID:    peerID,
		Username:  username,
		Port:      tcpPort,
		Timestamp: time.Now(),
	}
}

// ToJSON serializes the message for network transmission
func (m *DiscoveryMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON deserializes a message from network data
func FromJSON(data []byte) (*DiscoveryMessage, error) {
	var msg DiscoveryMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetSenderAddr returns the sender's address for TCP connections
func (m *DiscoveryMessage) GetSenderAddr() (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", m.Address)
}

// IsRecent checks if message is within acceptable time window
func (m *DiscoveryMessage) IsRecent(maxAge time.Duration) bool {
	return time.Since(m.Timestamp) <= maxAge
}

// String returns human-readable message info
func (m *DiscoveryMessage) String() string {
	return fmt.Sprintf("[%s] %s@%s (%s)",
		m.Type, m.Username, m.Address, m.Timestamp.Format("15:04:05"))
}
