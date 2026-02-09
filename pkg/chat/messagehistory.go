package chat

import (
	"sort"
	"sync"

	"p2pchat/pkg/logger"
)

// MessageHistory manages chronologically ordered message storage
// This handles in-memory storage, duplicate detection, and efficient retrieval
type MessageHistory struct {
	messages    []*Message      // Chronologically ordered messages
	messageIDs  map[string]bool // Fast duplicate detection O(1) lookup
	maxMessages int             // Maximum messages to keep in memory
	mutex       sync.RWMutex    // Protects concurrent access
}

// NewMessageHistory creates a new message history manager
func NewMessageHistory(maxMessages int) *MessageHistory {
	if maxMessages <= 0 {
		maxMessages = 1000 // Default to 1000 messages
	}

	return &MessageHistory{
		messages:    make([]*Message, 0, maxMessages),
		messageIDs:  make(map[string]bool),
		maxMessages: maxMessages,
	}
}

// AddMessage adds a message to history with duplicate detection and ordering
func (h *MessageHistory) AddMessage(msg *Message) bool {
	if msg == nil {
		return false
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check for duplicates using message ID
	if h.messageIDs[msg.ID] {
		logger.Debug("🔄 Duplicate message detected: %s (ID: %s)", msg.Content, msg.ID)
		return false // Message already exists
	}

	// Skip heartbeat messages from history (they're just for connection health)
	if msg.Type == MessageTypeHeartbeat {
		return false
	}

	// Add to messages and mark as seen
	h.messages = append(h.messages, msg)
	h.messageIDs[msg.ID] = true

	// Sort messages chronologically (important for multi-peer consistency)
	sort.Slice(h.messages, func(i, j int) bool {
		return h.messages[i].Timestamp.Before(h.messages[j].Timestamp)
	})

	// Cleanup old messages if we exceed limit
	h.cleanup()

	logger.Debug("📚 Added message to history: %s (Total: %d)", msg.Content, len(h.messages))
	return true
}

// GetMessages returns all messages, optionally filtered by type
func (h *MessageHistory) GetMessages(messageTypes ...MessageType) []*Message {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if len(messageTypes) == 0 {
		// Return all messages (make a copy to prevent external modification)
		result := make([]*Message, len(h.messages))
		copy(result, h.messages)
		return result
	}

	// Filter by message types
	typeMap := make(map[MessageType]bool)
	for _, msgType := range messageTypes {
		typeMap[msgType] = true
	}

	var filtered []*Message
	for _, msg := range h.messages {
		if typeMap[msg.Type] {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

// GetRecentMessages returns the most recent N messages
func (h *MessageHistory) GetRecentMessages(limit int) []*Message {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	totalMessages := len(h.messages)
	if limit <= 0 || limit >= totalMessages {
		// Return all messages
		result := make([]*Message, totalMessages)
		copy(result, h.messages)
		return result
	}

	// Return the last N messages
	startIndex := totalMessages - limit
	result := make([]*Message, limit)
	copy(result, h.messages[startIndex:])
	return result
}

// GetMessageCount returns the total number of stored messages
func (h *MessageHistory) GetMessageCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.messages)
}

// HasMessage checks if a message ID exists (for duplicate detection)
func (h *MessageHistory) HasMessage(messageID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.messageIDs[messageID]
}

// Clear removes all messages from history
func (h *MessageHistory) Clear() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.messages = h.messages[:0] // Keep capacity but reset length
	h.messageIDs = make(map[string]bool)

	logger.Debug("🗑️ Message history cleared")
}

// cleanup removes oldest messages when exceeding maxMessages limit
// This must be called with mutex already locked!
func (h *MessageHistory) cleanup() {
	if len(h.messages) <= h.maxMessages {
		return // No cleanup needed
	}

	// Calculate how many messages to remove
	excessMessages := len(h.messages) - h.maxMessages

	// Remove oldest messages and their IDs
	for i := 0; i < excessMessages; i++ {
		oldMsg := h.messages[i]
		delete(h.messageIDs, oldMsg.ID)
	}

	// Shift remaining messages to beginning of slice
	copy(h.messages, h.messages[excessMessages:])
	h.messages = h.messages[:h.maxMessages]

	logger.Debug("🧹 Cleaned up %d old messages, %d remaining", excessMessages, len(h.messages))
}

// GetStats returns statistics about the message history
func (h *MessageHistory) GetStats() MessageHistoryStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := MessageHistoryStats{
		TotalMessages: len(h.messages),
		MaxMessages:   h.maxMessages,
		UniqueIDs:     len(h.messageIDs),
	}

	return stats
}

// MessageHistoryStats provides insights into message history
type MessageHistoryStats struct {
	TotalMessages int `json:"total_messages"`
	MaxMessages   int `json:"max_messages"`
	UniqueIDs     int `json:"unique_ids"`
}
