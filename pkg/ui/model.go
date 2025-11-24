package ui

import (
	"time"

	"p2pchat/pkg/chat"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ChatModel represents the entire state of the chat TUI application
// This is your "single source of truth" - everything the UI needs to know
type ChatModel struct {
	// Core chat functionality
	chatService *chat.ChatService

	// UI State
	messages    []DisplayMessage // All chat messages to show
	peers       []PeerDisplay    // Connected peers to show in sidebar
	input       textinput.Model  // Text input component for typing
	maxMessages int              // Maximum messages to keep in UI (performance optimization)

	// Scroll state for message history
	scrollOffset    int  // How many messages scrolled up from bottom (0 = at bottom)
	maxScrollOffset int  // Maximum valid scroll offset
	autoScroll      bool // Whether to auto-scroll to new messages (default: true)

	// Window dimensions (Bubble Tea will know when terminal resizes)
	width          int // Total window width
	height         int // Total window height
	chatAreaHeight int // Height available for messages (calculated)

	// UI behavior state
	focused  FocusArea // Which part of UI has focus
	showHelp bool      // Whether to show help panel

	// Status and errors
	status    string // Current status message
	lastError string // Last error to display
}

// DisplayMessage represents a message formatted for display in the UI
type DisplayMessage struct {
	Content   string
	Username  string
	Timestamp time.Time
	Type      MessageType // chat, join, leave, system
	Style     string      // Color/style info
}

// PeerDisplay represents peer info formatted for the sidebar
type PeerDisplay struct {
	Username string
	Status   string // "connected", "connecting", "offline"
	Address  string
	LastSeen time.Time
}

// FocusArea represents which part of the UI currently has focus
type FocusArea int

const (
	FocusInput    FocusArea = iota // User is typing a message
	FocusPeers                     // User is browsing peer list
	FocusMessages                  // User is scrolling through messages - NEW!
)

// MessageType for styling different kinds of messages
type MessageType int

const (
	MessageTypeChat MessageType = iota
	MessageTypeJoin
	MessageTypeLeave
	MessageTypeSystem
	MessageTypeError
)

// NewChatModel creates a new chat model with your existing ChatService
func NewChatModel(chatService *chat.ChatService) ChatModel {
	input := textinput.New()
	input.Placeholder = "Type a message..."
	input.Focus()

	return ChatModel{
		chatService:     chatService,
		messages:        []DisplayMessage{},
		peers:           []PeerDisplay{},
		input:           input,
		maxMessages:     500,  // UI limit lower than backend (1000) for performance
		scrollOffset:    0,    // Start at bottom
		maxScrollOffset: 0,    // No messages yet
		autoScroll:      true, // Auto-scroll to new messages
		focused:         FocusInput,
		showHelp:        false,
	}
}

// Scroll Management Methods

// scrollUp moves the viewport up (showing older messages)
func (m *ChatModel) scrollUp(lines int) {
	m.scrollOffset += lines
	if m.scrollOffset > m.maxScrollOffset {
		m.scrollOffset = m.maxScrollOffset
	}
	// When user manually scrolls, disable auto-scroll
	if m.scrollOffset > 0 {
		m.autoScroll = false
	}
}

// scrollDown moves the viewport down (showing newer messages)
func (m *ChatModel) scrollDown(lines int) {
	m.scrollOffset -= lines
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
		// When back at bottom, re-enable auto-scroll
		m.autoScroll = true
	}
}

// scrollToBottom immediately goes to newest messages
func (m *ChatModel) scrollToBottom() {
	m.scrollOffset = 0
	m.autoScroll = true
}

// updateScrollBounds calculates maximum scroll offset based on messages and viewport
func (m *ChatModel) updateScrollBounds() {
	if m.chatAreaHeight <= 0 {
		m.maxScrollOffset = 0
		return
	}

	totalMessages := len(m.messages)
	visibleMessages := m.chatAreaHeight

	if totalMessages <= visibleMessages {
		m.maxScrollOffset = 0
	} else {
		m.maxScrollOffset = totalMessages - visibleMessages
	}

	// Ensure current scroll position is still valid
	if m.scrollOffset > m.maxScrollOffset {
		m.scrollOffset = m.maxScrollOffset
	}
}

// Init returns initial commands when the app starts
func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(
		LoadMessageHistory(m.chatService), // NEW: Load existing message history
		ListenForMessages(m.chatService),  // Start listening for P2P messages
		UpdatePeers(m.chatService),        // Get initial peer list
		PeriodicPeerUpdate(),              // Start periodic peer updates
	)
}

// addMessage adds a message to the UI with performance optimization
func (m *ChatModel) addMessage(msg DisplayMessage) {
	m.messages = append(m.messages, msg)

	// Performance optimization: limit UI messages
	if len(m.messages) > m.maxMessages {
		// Remove oldest 20% of messages to avoid frequent cleanup
		removeCount := m.maxMessages / 5
		copy(m.messages, m.messages[removeCount:])
		m.messages = m.messages[:len(m.messages)-removeCount]

		// Adjust scroll offset to maintain position
		if m.scrollOffset > removeCount {
			m.scrollOffset -= removeCount
		} else {
			m.scrollOffset = 0
		}
	}

	m.updateScrollBounds()
}
