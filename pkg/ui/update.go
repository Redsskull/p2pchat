package ui

import (
	"p2pchat/pkg/chat"

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

			// Add to our message history
			m.messages = append(m.messages, displayMsg)

			// Keep only last 1000 messages to prevent memory issues
			if len(m.messages) > 1000 {
				m.messages = m.messages[1:]
				// Adjust scroll offset if we removed messages from the beginning
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			}

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
		} else {
			m.status = msg.Status
		}

	// Handle periodic ticks
	case struct{}: // Our tick message
		// Refresh peer list periodically
		cmds = append(cmds, UpdatePeers(m.chatService))
	}

	// CRITICAL FIX: Always restart message listening
	// This ensures we never stop listening for P2P messages
	cmds = append(cmds, ListenForMessages(m.chatService))

	return m, tea.Batch(cmds...)
}

// handleKeyPress processes keyboard input with scroll support
func (m ChatModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "enter":
		if m.focused == FocusInput && m.input.Value() != "" {
			// Send the message!
			content := m.input.Value()
			m.input.SetValue("") // Clear input
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

	// SCROLLING CONTROLS - New for Day 6!
	case "up", "k":
		if m.focused == FocusMessages || m.focused == FocusInput {
			m.scrollUp(1)
		}

	case "down", "j":
		if m.focused == FocusMessages || m.focused == FocusInput {
			m.scrollDown(1)
		}

	case "pgup":
		m.scrollUp(5) // Fast scroll up

	case "pgdown":
		m.scrollDown(5) // Fast scroll down

	case "home":
		// Go to oldest messages (top)
		m.scrollOffset = m.maxScrollOffset
		m.autoScroll = false

	case "end":
		// Go to newest messages (bottom)
		m.scrollToBottom()

	case "?":
		m.showHelp = !m.showHelp

	default:
		// Let the input component handle typing when focused
		if m.focused == FocusInput {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
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
		display[i] = PeerDisplay{
			Username: peer.Username,
			Status:   peer.ConnectionState,
			Address:  peer.Address,
			LastSeen: peer.LastSeen,
		}
	}
	return display
}
