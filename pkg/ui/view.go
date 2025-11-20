package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the chat UI - this is called whenever the model changes
func (m ChatModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Create styles
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("57")).
		Padding(0, 1)

	// Header with app title and status
	header := headerStyle.Render("🗨️  P2P Chat - " + m.status)

	// Chat area (messages)
	chatArea := m.renderChatArea()

	// Peer list (sidebar)
	peerList := m.renderPeerList()

	// Input area
	inputArea := m.renderInputArea()

	// Help text
	helpText := "Press '?' for help • Tab to switch focus • Ctrl+C to quit"

	// Layout everything
	mainArea := lipgloss.JoinHorizontal(
		lipgloss.Top,
		chatArea,
		peerList,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainArea,
		inputArea,
		helpText,
	)
}

// renderChatArea renders the message history
func (m ChatModel) renderChatArea() string {
	if len(m.messages) == 0 {
		return "No messages yet. Start chatting!"
	}

	var messageStrings []string
	for _, msg := range m.messages {
		timestamp := msg.Timestamp.Format("15:04")
		messageStr := fmt.Sprintf("[%s] %s: %s", timestamp, msg.Username, msg.Content)
		messageStrings = append(messageStrings, messageStr)
	}

	return strings.Join(messageStrings, "\n")
}

// renderPeerList renders the connected peers sidebar
func (m ChatModel) renderPeerList() string {
	if len(m.peers) == 0 {
		return "No peers connected"
	}

	var peerStrings []string
	peerStrings = append(peerStrings, "Connected Peers:")

	for _, peer := range m.peers {
		status := "🔴"
		if peer.Status == "connected" {
			status = "🟢"
		}
		peerStr := fmt.Sprintf("%s %s", status, peer.Username)
		peerStrings = append(peerStrings, peerStr)
	}

	return strings.Join(peerStrings, "\n")
}

// renderInputArea renders the text input field
func (m ChatModel) renderInputArea() string {
	return fmt.Sprintf("\nYour message: %s", m.input.View())
}
