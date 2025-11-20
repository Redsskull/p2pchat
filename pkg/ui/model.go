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
	messages []DisplayMessage // All chat messages to show
	peers    []PeerDisplay    // Connected peers to show in sidebar
	input    textinput.Model  // Text input component for typing

	// Window dimensions (Bubble Tea will tell  when terminal resizes)
	width  int
	height int

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
	FocusMessages                  // User is scrolling through messages
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
		chatService: chatService,
		messages:    []DisplayMessage{},
		peers:       []PeerDisplay{},
		input:       input,
		focused:     FocusInput,
		showHelp:    false,
	}
}

// Init returns initial commands when the app starts
func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(
		ListenForMessages(m.chatService), // Start listening for P2P messages
		UpdatePeers(m.chatService),       // Get initial peer list
		PeriodicPeerUpdate(),             // Start periodic peer updates
	)
}
