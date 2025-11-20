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

	// Handle window resizing
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Resize input component too
		m.input.Width = msg.Width - 4

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

// handleKeyPress processes keyboard input
func (m ChatModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "enter":
		// Send the message!
		if m.focused == FocusInput && m.input.Value() != "" {
			content := m.input.Value()
			m.input.SetValue("") // Clear input

			// Send message via your ChatService
			return m, SendMessageCmd(m.chatService, content)
		}

	case "tab":
		// Switch focus between input and peer list
		if m.focused == FocusInput {
			m.focused = FocusPeers
			m.input.Blur()
		} else {
			m.focused = FocusInput
			m.input.Focus()
		}

	case "?":
		m.showHelp = !m.showHelp

	default:
		// Let the input component handle typing
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
