package ui

import (
	"fmt"
	"p2pchat/pkg/chat"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all events and returns new state + commands
// This is called every time something happens (key press, network event, etc.)
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Handle keyboard input
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	// Handle window resizing - THIS IS WHERE WE CALCULATE chatAreaHeight!
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate chat area height properly here (not in View!)
		headerHeight := 1
		inputHeight := 3 // Input field + border + padding
		helpHeight := 1
		usedHeight := headerHeight + inputHeight + helpHeight

		m.chatAreaHeight = m.height - usedHeight
		if m.chatAreaHeight < 3 {
			m.chatAreaHeight = 3 // Minimum size
		}

		// Update scroll bounds when window resizes
		m.updateScrollBounds()

		// Resize input component too
		m.input.Width = msg.Width - 8 // Account for borders

	// NEW: Handle loading message history on startup
	case MessageHistoryMsg:
		// Convert chat messages to display messages
		for _, msg := range msg.Messages {
			displayMsg := DisplayMessage{
				Content:   msg.Content,
				Username:  msg.Username,
				Timestamp: msg.Timestamp,
				Type:      convertMessageType(msg.Type),
			}
			m.addMessage(displayMsg)
		}
		m.scrollToBottom()

	// Handle incoming chat messages from your P2P network!
	case IncomingMessageMsg:
		if msg.Message != nil {
			// Convert your chat.Message to DisplayMessage
			displayMsg := DisplayMessage{
				Content:   msg.Message.Content,
				Username:  msg.Message.Username,
				Timestamp: msg.Message.Timestamp,
				Type:      convertMessageType(msg.Message.Type),
			}

			// Add to our message history using optimized function
			m.addMessage(displayMsg)

			// Update scroll bounds with new message
			m.updateScrollBounds()

			// Auto-scroll to new message if we're at the bottom
			if m.autoScroll {
				m.scrollOffset = 0
			}
		}

		// Keep listening for more messages!
		cmds = append(cmds, ListenForMessages(m.chatService))

	// Handle peer updates
	case PeerUpdateMsg:
		m.peers = convertPeersToDisplay(msg.Peers)
		// Schedule next peer update
		cmds = append(cmds, PeriodicPeerUpdate())

	// Handle status updates
	case StatusUpdateMsg:
		if msg.IsError {
			m.lastError = msg.Status
			// Clear error after 5 seconds
			go func() {
				time.Sleep(5 * time.Second)
				m.lastError = ""
			}()
		} else {
			m.status = msg.Status
			// Clear any previous errors on successful status
			m.lastError = ""
		}

	// Handle periodic ticks
	case struct{}: // Our tick message
		// Refresh peer list periodically
		cmds = append(cmds, UpdatePeers(m.chatService))
	}

	// This ensures we never stop listening for P2P messages
	cmds = append(cmds, ListenForMessages(m.chatService))

	return m, tea.Batch(cmds...)
}

// handleChatCommand processes IRC-style chat commands
func (m ChatModel) handleChatCommand(command string) (ChatModel, tea.Cmd) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return m, nil
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help", "/h":
		return m.showHelpMessage()

	case "/users", "/who":
		return m.showUsersList()

	case "/quit", "/q", "/exit":
		return m, tea.Quit

	case "/nick":
		if len(parts) < 2 {
			m.lastError = "Usage: /nick <new_username>"
			return m, nil
		}
		newUsername := parts[1]
		return m.changeUsername(newUsername)

	case "/clear":
		return m.clearMessages()

	default:
		m.lastError = fmt.Sprintf("Unknown command: %s. Type /help for available commands.", cmd)
		return m, nil
	}
}

// showHelpMessage displays available chat commands
func (m ChatModel) showHelpMessage() (ChatModel, tea.Cmd) {
	helpMsg := DisplayMessage{
		Content:   "Available commands:\n/help - Show this help\n/users - List connected users\n/nick <name> - Change username\n/clear - Clear message history\n/quit - Exit chat",
		Username:  "System",
		Timestamp: time.Now(),
		Type:      MessageTypeSystem,
		Style:     "help",
	}

	m.addMessage(helpMsg)
	if m.autoScroll {
		m.scrollToBottom()
	}

	return m, nil
}

// showUsersList displays connected peers
func (m ChatModel) showUsersList() (ChatModel, tea.Cmd) {
	var content string
	if len(m.peers) == 0 {
		content = "No other users connected. Waiting for peers to join..."
	} else {
		var userList strings.Builder
		userList.WriteString("Connected users:\n")
		for _, peer := range m.peers {
			status := "●" // online indicator
			if peer.Status != "connected" {
				status = "◯" // offline indicator
			}
			userList.WriteString(fmt.Sprintf("  %s %s (%s)\n", status, peer.Username, peer.Status))
		}
		content = userList.String()
	}

	userMsg := DisplayMessage{
		Content:   content,
		Username:  "System",
		Timestamp: time.Now(),
		Type:      MessageTypeSystem,
		Style:     "users",
	}

	m.addMessage(userMsg)
	if m.autoScroll {
		m.scrollToBottom()
	}

	return m, nil
}

// changeUsername changes the user's display name
func (m ChatModel) changeUsername(newUsername string) (ChatModel, tea.Cmd) {
	// Validate username
	if len(newUsername) == 0 {
		m.lastError = "Username cannot be empty"
		return m, nil
	}
	if len(newUsername) > 20 {
		m.lastError = "Username too long (max 20 characters)"
		return m, nil
	}
	if strings.ContainsAny(newUsername, " \t\n\r") {
		m.lastError = "Username cannot contain spaces"
		return m, nil
	}

	// Create system message about the change
	changeMsg := DisplayMessage{
		Content:   fmt.Sprintf("You changed your username to: %s", newUsername),
		Username:  "System",
		Timestamp: time.Now(),
		Type:      MessageTypeSystem,
		Style:     "nick",
	}

	m.addMessage(changeMsg)
	if m.autoScroll {
		m.scrollToBottom()
	}

	// Actually change username in chat service
	err := m.chatService.ChangeUsername(newUsername)
	if err != nil {
		m.lastError = fmt.Sprintf("Failed to change username: %v", err)
		return m, nil
	}

	m.status = fmt.Sprintf("Username changed to: %s", newUsername)

	return m, nil
}

// clearMessages clears the message history
func (m ChatModel) clearMessages() (ChatModel, tea.Cmd) {
	m.messages = []DisplayMessage{}
	m.scrollOffset = 0
	m.maxScrollOffset = 0
	m.status = "Message history cleared"

	return m, nil
}

// handleKeyPress processes keyboard input with scroll support
func (m ChatModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "enter":
		if m.focused == FocusInput && m.input.Value() != "" {
			// Get the message content
			content := m.input.Value()

			// Check if it's a command
			if strings.HasPrefix(content, "/") {
				m.input.SetValue("") // Clear input
				return m.handleChatCommand(content)
			}

			// Validate message isn't too long
			if len(content) > 1000 {
				m.lastError = "Message too long (max 1000 characters)"
				return m, nil
			}

			// Check if we have any peers
			if len(m.peers) == 0 {
				m.lastError = "No peers connected - waiting for others to join"
				// Don't clear input, let them try again when peers connect
				return m, nil
			}

			m.input.SetValue("") // Clear input
			m.status = "Sending message..."
			return m, SendMessageCmd(m.chatService, content)
		} else if m.focused != FocusInput {
			// Enter switches to input focus from other areas
			m.focused = FocusInput
			m.input.Focus()
		}

	// TAB: Switch between focus areas
	case "tab":
		switch m.focused {
		case FocusInput:
			m.focused = FocusMessages
			m.input.Blur()
		case FocusMessages:
			m.focused = FocusPeers
		case FocusPeers:
			m.focused = FocusInput
			m.input.Focus()
		}

	// Handle input first when focused - this fixes the 'k' and '?' bug
	default:
		if m.focused == FocusInput {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)

			// Clear any typing-related errors when user starts typing
			if m.input.Value() != "" && m.lastError != "" {
				// Clear certain types of errors when user is actively typing
				if m.lastError == "Message too long (max 1000 characters)" ||
					m.lastError == "No peers connected - waiting for others to join" {
					m.lastError = ""
				}
			}

			return m, cmd
		}

		// SCROLLING CONTROLS - Only when not actively typing
		switch msg.String() {
		case "k", "up":
			if m.focused == FocusMessages {
				m.scrollUp(1)
			}
		case "j", "down":
			if m.focused == FocusMessages {
				m.scrollDown(1)
			}
		case "pgup":
			m.scrollUp(5)
		case "pgdown":
			m.scrollDown(5)
		case "home":
			m.scrollOffset = m.maxScrollOffset
			m.autoScroll = false
		case "end":
			m.scrollToBottom()
		case "?":
			m.showHelp = !m.showHelp
		}
	}

	return m, nil
}

// Helper functions
func convertMessageType(chatType chat.MessageType) MessageType {
	switch chatType {
	case chat.MessageTypeJoin:
		return MessageTypeJoin
	case chat.MessageTypeLeave:
		return MessageTypeLeave
	default:
		return MessageTypeChat
	}
}

func convertPeersToDisplay(peers []chat.PeerInfo) []PeerDisplay {
	display := make([]PeerDisplay, len(peers))
	for i, peer := range peers {
		status := "disconnected"
		if peer.Connected {
			status = "connected"
		}

		display[i] = PeerDisplay{
			Username: peer.Username,
			Status:   status,
			Address:  peer.Address,
			LastSeen: peer.LastSeen,
		}
	}
	return display
}
