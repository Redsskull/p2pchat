package chat

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Message represents a chat message that flows between peers via TCP
// This is different from DiscoveryMessage which handles peer announcements via UDP
type Message struct {
	// Core identification
	ID       string      `json:"id"`        // Unique message ID to prevent duplicates
	Type     MessageType `json:"type"`      // What kind of message this is
	SenderID string      `json:"sender_id"` // Peer ID of the sender
	Username string      `json:"username"`  // Display name of sender

	// Message content
	Content   string    `json:"content"`   // The actual message text (for chat messages)
	Timestamp time.Time `json:"timestamp"` // When this message was created
	Sequence  uint64    `json:"sequence"`  // Message ordering within sender's stream

	// Optional metadata
	RoomID   string         `json:"room_id,omitempty"`  // Future: support multiple rooms
	Metadata map[string]any `json:"metadata,omitempty"` // Future: extensibility
}

// MessageType defines the different kinds of messages in the chat protocol
type MessageType string

const (
	// Core chat messages
	MessageTypeChat      MessageType = "chat"      // Regular text message: "Hello everyone!"
	MessageTypeJoin      MessageType = "join"      // User joined: "Alice joined the chat"
	MessageTypeLeave     MessageType = "leave"     // User left: "Alice left the chat"
	MessageTypeHeartbeat MessageType = "heartbeat" // Keep-alive: used for connection health

	// Future message types I might add:
	// MessageTypeTyping   MessageType = "typing"    // "Alice is typing..."
	// MessageTypeFile     MessageType = "file"      // File transfer
	// MessageTypeReaction MessageType = "reaction"  // Message reactions
)

// NewChatMessage creates a regular chat message
func NewChatMessage(senderID, username, content string, sequence uint64) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeChat,
		SenderID:  senderID,
		Username:  username,
		Content:   content,
		Timestamp: time.Now(),
		Sequence:  sequence,
		RoomID:    "general", // Default room for now
	}
}

// NewJoinMessage creates a user join notification
func NewJoinMessage(senderID, username string, sequence uint64) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeJoin,
		SenderID:  senderID,
		Username:  username,
		Content:   fmt.Sprintf("%s joined the chat", username),
		Timestamp: time.Now(),
		Sequence:  sequence,
		RoomID:    "general",
	}
}

// NewLeaveMessage creates a user leave notification
func NewLeaveMessage(senderID, username string, sequence uint64) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeLeave,
		SenderID:  senderID,
		Username:  username,
		Content:   fmt.Sprintf("%s left the chat", username),
		Timestamp: time.Now(),
		Sequence:  sequence,
		RoomID:    "general",
	}
}

// NewHeartbeatMessage creates a connection heartbeat
func NewHeartbeatMessage(senderID, username string, sequence uint64) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeHeartbeat,
		SenderID:  senderID,
		Username:  username,
		Content:   "", // Heartbeats don't need content
		Timestamp: time.Now(),
		Sequence:  sequence,
		RoomID:    "general",
	}
}

// Serialization methods

// ToJSON serializes the message for network transmission
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON deserializes a message from network data
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Validate required fields
	if msg.ID == "" {
		return nil, fmt.Errorf("message missing required ID field")
	}
	if msg.SenderID == "" {
		return nil, fmt.Errorf("message missing required sender_id field")
	}
	if msg.Type == "" {
		return nil, fmt.Errorf("message missing required type field")
	}

	return &msg, nil
}

// Utility methods

// IsUserVisible returns true if this message should be shown to users
// (heartbeats are typically hidden from the UI)
func (m *Message) IsUserVisible() bool {
	return m.Type != MessageTypeHeartbeat
}

// IsRecent checks if message is within acceptable time window
// This helps reject very old messages that might be replayed
func (m *Message) IsRecent(maxAge time.Duration) bool {
	return time.Since(m.Timestamp) <= maxAge
}

// String returns a human-readable representation for debugging/logging
func (m *Message) String() string {
	switch m.Type {
	case MessageTypeChat:
		return fmt.Sprintf("[%s] %s: %s",
			m.Timestamp.Format("15:04:05"), m.Username, m.Content)
	case MessageTypeJoin:
		return fmt.Sprintf("[%s] *** %s joined",
			m.Timestamp.Format("15:04:05"), m.Username)
	case MessageTypeLeave:
		return fmt.Sprintf("[%s] *** %s left",
			m.Timestamp.Format("15:04:05"), m.Username)
	case MessageTypeHeartbeat:
		return fmt.Sprintf("[%s] <heartbeat from %s>",
			m.Timestamp.Format("15:04:05"), m.Username)
	default:
		return fmt.Sprintf("[%s] <%s from %s>",
			m.Timestamp.Format("15:04:05"), m.Type, m.Username)
	}
}

// Helper functions

// generateMessageID creates a unique ID for each message
// This helps with duplicate detection and message tracking
func generateMessageID() string {
	// Generate 8 random bytes and convert to hex
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// IsValidMessageType checks if a message type is supported
func IsValidMessageType(msgType MessageType) bool {
	switch msgType {
	case MessageTypeChat, MessageTypeJoin, MessageTypeLeave, MessageTypeHeartbeat:
		return true
	default:
		return false
	}
}
