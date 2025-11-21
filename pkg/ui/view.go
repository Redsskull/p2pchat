package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the chat UI - this is called whenever the model changes
// IMPORTANT: View functions are READ-ONLY - never modify the model here!
func (m ChatModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Create styles
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("57")).
		Padding(0, 1).
		Width(m.width)

	chatStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Height(m.chatAreaHeight).
		Width(m.width * 3 / 4) // Chat takes 75% of width

	peerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Height(m.chatAreaHeight).
		Width(m.width / 4) // Peers take 25% of width

	// Header with app title and status
	statusText := m.status
	if statusText == "" {
		statusText = "Ready to chat"
	}
	header := headerStyle.Render("🗨️  P2P Chat - " + statusText)

	// Chat area (messages) - now with scrolling!
	chatContent := m.renderChatArea()
	chatArea := chatStyle.Render(chatContent)

	// Peer list (sidebar)
	peerContent := m.renderPeerList()
	peerList := peerStyle.Render(peerContent)

	// Input area
	inputArea := m.renderInputArea()

	// Help text with context-sensitive instructions
	helpText := m.renderHelpText()

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

// renderChatArea renders the message history with scrolling support
func (m ChatModel) renderChatArea() string {
	if len(m.messages) == 0 {
		return "No messages yet. Start chatting!\n\nWhen peers connect, messages will appear here.\nUse ↑↓ arrows to scroll through history."
	}

	// Use the chatAreaHeight calculated in Update()
	availableHeight := m.chatAreaHeight
	if availableHeight <= 2 {
		availableHeight = 5 // Minimum reasonable size
	}

	totalMessages := len(m.messages)

	// Determine which messages to show based on scroll position
	var startIndex, endIndex int

	if totalMessages <= availableHeight {
		// All messages fit on screen
		startIndex = 0
		endIndex = totalMessages
	} else {
		// Show a window of messages based on scroll position
		// scrollOffset = 0 means show latest (bottom)
		// scrollOffset > 0 means show older messages

		endIndex = totalMessages - m.scrollOffset
		startIndex = endIndex - availableHeight

		// Safety bounds
		if startIndex < 0 {
			startIndex = 0
			endIndex = availableHeight
		}
		if endIndex > totalMessages {
			endIndex = totalMessages
			startIndex = totalMessages - availableHeight
		}
	}

	// Build the message strings for our viewport
	var messageStrings []string
	for i := startIndex; i < endIndex; i++ {
		msg := m.messages[i]
		timestamp := msg.Timestamp.Format("15:04")

		// Color-code messages by type
		var messageStr string
		switch msg.Type {
		case MessageTypeJoin:
			messageStr = fmt.Sprintf("[%s] → %s joined", timestamp, msg.Username)
		case MessageTypeLeave:
			messageStr = fmt.Sprintf("[%s] ← %s left", timestamp, msg.Username)
		case MessageTypeSystem:
			messageStr = fmt.Sprintf("[%s] * %s", timestamp, msg.Content)
		default:
			messageStr = fmt.Sprintf("[%s] %s: %s", timestamp, msg.Username, msg.Content)
		}

		messageStrings = append(messageStrings, messageStr)
	}

	result := strings.Join(messageStrings, "\n")

	// Add scroll indicators if needed
	if m.maxScrollOffset > 0 {
		scrollIndicator := ""
		if m.scrollOffset > 0 {
			scrollIndicator = fmt.Sprintf("\n\n↑ Viewing older messages (%d/%d) - Press ↓ or End for latest",
				m.scrollOffset, m.maxScrollOffset)
		} else {
			scrollIndicator = "\n\n● Latest messages (auto-scroll enabled)"
		}
		result += scrollIndicator
	}

	return result
}

// renderPeerList renders the connected peers sidebar with status indicators
func (m ChatModel) renderPeerList() string {
	var peerStrings []string
	peerStrings = append(peerStrings, "🌐 Connected Peers")
	peerStrings = append(peerStrings, "")

	if len(m.peers) == 0 {
		peerStrings = append(peerStrings, "No peers connected")
		peerStrings = append(peerStrings, "")
		peerStrings = append(peerStrings, "Waiting for peers...")
		peerStrings = append(peerStrings, "Make sure other users")
		peerStrings = append(peerStrings, "are on the same network")
		return strings.Join(peerStrings, "\n")
	}

	for _, peer := range m.peers {
		var statusIcon string

		switch peer.Status {
		case "connected":
			statusIcon = "🟢"
		case "connecting":
			statusIcon = "🟡"
		default:
			statusIcon = "🔴"
		}

		peerStr := fmt.Sprintf("%s %s", statusIcon, peer.Username)
		peerStrings = append(peerStrings, peerStr)
	}

	// Add connection stats
	connected := 0
	for _, peer := range m.peers {
		if peer.Status == "connected" {
			connected++
		}
	}

	peerStrings = append(peerStrings, "")
	peerStrings = append(peerStrings, fmt.Sprintf("📊 %d/%d active", connected, len(m.peers)))

	return strings.Join(peerStrings, "\n")
}

// renderInputArea renders the text input field
func (m ChatModel) renderInputArea() string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(m.width - 2)

	focusIndicator := ""
	if m.focused == FocusInput {
		focusIndicator = "● "
	}

	content := fmt.Sprintf("%sType message: %s", focusIndicator, m.input.View())
	return inputStyle.Render(content)
}

// renderHelpText provides context-sensitive help
func (m ChatModel) renderHelpText() string {
	var help string

	switch m.focused {
	case FocusInput:
		help = "Enter: send message • Tab: switch focus • ↑↓: scroll messages • Ctrl+C: quit"
	case FocusMessages:
		help = "↑↓: scroll • PgUp/PgDn: fast scroll • Home/End: top/bottom • Tab: switch focus"
	case FocusPeers:
		help = "Tab: switch focus • ↑↓: scroll messages • Enter: focus input • Ctrl+C: quit"
	default:
		help = "Tab: switch focus • Enter: send message • Ctrl+C: quit"
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(help)
}
