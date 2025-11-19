package chat

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageCreation(t *testing.T) {
	// Test chat message creation
	chatMsg := NewChatMessage("peer123", "alice", "Hello everyone!", 42)

	if chatMsg.Type != MessageTypeChat {
		t.Errorf("Expected chat message type, got %s", chatMsg.Type)
	}
	if chatMsg.SenderID != "peer123" {
		t.Errorf("Expected sender peer123, got %s", chatMsg.SenderID)
	}
	if chatMsg.Username != "alice" {
		t.Errorf("Expected username alice, got %s", chatMsg.Username)
	}
	if chatMsg.Content != "Hello everyone!" {
		t.Errorf("Expected content 'Hello everyone!', got %s", chatMsg.Content)
	}
	if chatMsg.Sequence != 42 {
		t.Errorf("Expected sequence 42, got %d", chatMsg.Sequence)
	}
	if chatMsg.ID == "" {
		t.Error("Message ID should not be empty")
	}
}

func TestJoinLeaveMessages(t *testing.T) {
	// Test join message
	joinMsg := NewJoinMessage("peer456", "bob", 1)
	if joinMsg.Type != MessageTypeJoin {
		t.Errorf("Expected join message type, got %s", joinMsg.Type)
	}
	if joinMsg.Content != "bob joined the chat" {
		t.Errorf("Expected join content, got %s", joinMsg.Content)
	}

	// Test leave message
	leaveMsg := NewLeaveMessage("peer456", "bob", 2)
	if leaveMsg.Type != MessageTypeLeave {
		t.Errorf("Expected leave message type, got %s", leaveMsg.Type)
	}
	if leaveMsg.Content != "bob left the chat" {
		t.Errorf("Expected leave content, got %s", leaveMsg.Content)
	}
}

func TestHeartbeatMessage(t *testing.T) {
	heartbeat := NewHeartbeatMessage("peer789", "charlie", 100)

	if heartbeat.Type != MessageTypeHeartbeat {
		t.Errorf("Expected heartbeat message type, got %s", heartbeat.Type)
	}
	if heartbeat.Content != "" {
		t.Errorf("Expected empty content for heartbeat, got %s", heartbeat.Content)
	}
	if heartbeat.IsUserVisible() {
		t.Error("Heartbeat messages should not be user visible")
	}
}

func TestJSONSerialization(t *testing.T) {
	// Create a test message
	original := NewChatMessage("test_peer", "testuser", "Test message content", 123)

	// Serialize to JSON
	jsonData, err := original.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize message: %v", err)
	}

	// Verify it's valid JSON
	var jsonCheck map[string]any
	if err := json.Unmarshal(jsonData, &jsonCheck); err != nil {
		t.Fatalf("Serialized data is not valid JSON: %v", err)
	}

	// Deserialize back to Message
	deserialized, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to deserialize message: %v", err)
	}

	// Compare key fields
	if deserialized.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, deserialized.ID)
	}
	if deserialized.Type != original.Type {
		t.Errorf("Type mismatch: expected %s, got %s", original.Type, deserialized.Type)
	}
	if deserialized.SenderID != original.SenderID {
		t.Errorf("SenderID mismatch: expected %s, got %s", original.SenderID, deserialized.SenderID)
	}
	if deserialized.Content != original.Content {
		t.Errorf("Content mismatch: expected %s, got %s", original.Content, deserialized.Content)
	}
	if deserialized.Sequence != original.Sequence {
		t.Errorf("Sequence mismatch: expected %d, got %d", original.Sequence, deserialized.Sequence)
	}
}

func TestMessageValidation(t *testing.T) {
	// Test valid message
	validMsg := NewChatMessage("peer1", "user1", "Valid message", 1)
	jsonData, _ := validMsg.ToJSON()

	_, err := FromJSON(jsonData)
	if err != nil {
		t.Errorf("Valid message should deserialize without error, got: %v", err)
	}

	// Test invalid JSON
	_, err = FromJSON([]byte("invalid json"))
	if err == nil {
		t.Error("Invalid JSON should return an error")
	}

	// Test missing required fields
	invalidMessages := []string{
		`{"type":"chat","sender_id":"","username":"test","content":"test"}`, // empty sender_id
		`{"type":"","sender_id":"test","username":"test","content":"test"}`, // empty type
		`{"sender_id":"test","username":"test","content":"test"}`,           // missing type
		`{"type":"chat","username":"test","content":"test"}`,                // missing sender_id
	}

	for i, invalidJSON := range invalidMessages {
		_, err := FromJSON([]byte(invalidJSON))
		if err == nil {
			t.Errorf("Invalid message %d should return validation error", i)
		}
	}
}

func TestMessageTypeValidation(t *testing.T) {
	validTypes := []MessageType{
		MessageTypeChat,
		MessageTypeJoin,
		MessageTypeLeave,
		MessageTypeHeartbeat,
	}

	for _, msgType := range validTypes {
		if !IsValidMessageType(msgType) {
			t.Errorf("Message type %s should be valid", msgType)
		}
	}

	// Test invalid type
	if IsValidMessageType("invalid_type") {
		t.Error("Invalid message type should return false")
	}
}

func TestMessageUtilities(t *testing.T) {
	// Test IsRecent
	msg := NewChatMessage("peer1", "user1", "test", 1)

	// Should be recent (just created)
	if !msg.IsRecent(time.Minute) {
		t.Error("Fresh message should be recent")
	}

	// Test with very short window
	if msg.IsRecent(time.Nanosecond) {
		t.Error("Message should not be recent with nanosecond window")
	}

	// Test String() method
	str := msg.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}

	// Test different message types have different string formats
	joinMsg := NewJoinMessage("peer2", "user2", 1)
	if joinMsg.String() == msg.String() {
		t.Error("Different message types should have different string representations")
	}
}
