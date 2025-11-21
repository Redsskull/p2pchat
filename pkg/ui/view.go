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

	// Header with app title, status, and error display
	statusText := m.status
	if statusText == "" {
		statusText = "Ready to chat"
	}

	headerContent := "🗨️  P2P Chat - " + statusText

	// Add error display if there's an error
	if m.lastError != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Padding(0, 1).
			Bold(true)

		errorMsg := errorStyle.Render("⚠️ " + m.lastError)
		headerContent += " | " + errorMsg
	}

	header := headerStyle.Render(headerContent)

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
		welcomeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Align(lipgloss.Center)

		welcome := welcomeStyle.Render("✨ Welcome to P2P Chat! ✨\n\n" +
			"💬 Start typing to send messages\n" +
			"🔍 Other users on your network will appear automatically\n" +
			"📜 Use ↑↓ arrows to scroll through chat history\n" +
			"⌨️  Press Tab to switch focus areas")

		return welcome
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

	// Build the message strings for our viewport with beautiful colors and text wrapping
	var messageStrings []string
	chatWidth := m.width*3/4 - 4 // Account for borders and padding

	for i := startIndex; i < endIndex; i++ {
		msg := m.messages[i]
		timestamp := msg.Timestamp.Format("15:04")

		// Create styled timestamp
		timestampStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		styledTimestamp := timestampStyle.Render(fmt.Sprintf("[%s]", timestamp))

		// Color-code messages by type and user
		var wrappedLines []string
		switch msg.Type {
		case MessageTypeJoin:
			joinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true) // Green
			messageStr := fmt.Sprintf("%s %s", styledTimestamp, joinStyle.Render(fmt.Sprintf("→ %s joined", msg.Username)))
			wrappedLines = []string{messageStr}
		case MessageTypeLeave:
			leaveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true) // Red
			messageStr := fmt.Sprintf("%s %s", styledTimestamp, leaveStyle.Render(fmt.Sprintf("← %s left", msg.Username)))
			wrappedLines = []string{messageStr}
		case MessageTypeSystem:
			systemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true) // Orange
			messageStr := fmt.Sprintf("%s %s", styledTimestamp, systemStyle.Render(fmt.Sprintf("* %s", msg.Content)))
			wrappedLines = []string{messageStr}
		default:
			// Assign consistent colors to users based on username hash
			userColor := m.getUserColor(msg.Username)
			usernameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(userColor)).Bold(true)
			contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

			styledUsername := usernameStyle.Render(msg.Username)
			prefix := fmt.Sprintf("%s %s: ", styledTimestamp, styledUsername)

			// Wrap long messages intelligently
			wrappedLines = m.wrapMessage(prefix, msg.Content, chatWidth, contentStyle)
		}

		// Add all wrapped lines
		messageStrings = append(messageStrings, wrappedLines...)

		// Add subtle visual separator between different users' messages
		if i < endIndex-1 {
			nextMsg := m.messages[i+1]
			if msg.Username != nextMsg.Username && msg.Type == MessageTypeChat && nextMsg.Type == MessageTypeChat {
				separator := lipgloss.NewStyle().
					Foreground(lipgloss.Color("237")).
					Render("  ┈")
				messageStrings = append(messageStrings, separator)
			}
		}
	}

	result := strings.Join(messageStrings, "\n")

	// Add beautiful scroll indicators if needed
	if m.maxScrollOffset > 0 {
		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true).
			Align(lipgloss.Center)

		if m.scrollOffset > 0 {
			scrollIndicator := scrollStyle.Render(fmt.Sprintf("\n\n🔼 Viewing older messages (%d/%d messages up) 🔼\nPress ↓ or End to see latest messages",
				m.scrollOffset, m.maxScrollOffset))
			result += scrollIndicator
		} else {
			scrollIndicator := scrollStyle.Render("\n\n📍 Latest messages (live updates enabled)")
			result += scrollIndicator
		}
	}

	return result
}

// renderPeerList renders the connected peers sidebar with status indicators
func (m ChatModel) renderPeerList() string {
	var peerStrings []string
	peerStrings = append(peerStrings, "🌐 Connected Peers")
	peerStrings = append(peerStrings, "")

	if len(m.peers) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

		peerStrings = append(peerStrings, emptyStyle.Render("No peers found"))
		peerStrings = append(peerStrings, "")
		peerStrings = append(peerStrings, emptyStyle.Render("🔍 Searching..."))
		peerStrings = append(peerStrings, emptyStyle.Render(""))
		peerStrings = append(peerStrings, emptyStyle.Render("💡 Make sure other"))
		peerStrings = append(peerStrings, emptyStyle.Render("users are running"))
		peerStrings = append(peerStrings, emptyStyle.Render("p2pchat on the same"))
		peerStrings = append(peerStrings, emptyStyle.Render("network"))
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
		focusIndicator = "🖋️  "
	} else {
		focusIndicator = "💭 "
	}

	placeholder := "Type your message..."
	if len(m.peers) == 0 {
		placeholder = "Waiting for peers to connect..."
	}

	content := fmt.Sprintf("%s%s %s", focusIndicator, placeholder, m.input.View())
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

// getUserColor returns a consistent color for each user based on their username
func (m ChatModel) getUserColor(username string) string {
	// Beautiful color palette for users
	colors := []string{
		"39",  // Bright blue
		"203", // Pink
		"148", // Green
		"214", // Orange
		"177", // Purple
		"81",  // Cyan
		"226", // Yellow
		"196", // Red
		"117", // Light blue
		"205", // Magenta
	}

	// Simple hash function to assign consistent colors
	hash := 0
	for _, char := range username {
		hash += int(char)
	}

	return colors[hash%len(colors)]
}

// wrapMessage intelligently wraps long messages with proper indentation
func (m ChatModel) wrapMessage(prefix, content string, maxWidth int, contentStyle lipgloss.Style) []string {
	if maxWidth <= 0 {
		maxWidth = 50 // Fallback width
	}

	// Calculate visible length of prefix (without ANSI color codes)
	visiblePrefix := stripANSI(prefix)
	prefixLen := len(visiblePrefix)

	// If content fits on one line, return it as-is
	if len(content)+prefixLen <= maxWidth {
		styledContent := contentStyle.Render(content)
		return []string{prefix + styledContent}
	}

	var lines []string
	words := strings.Fields(content)
	if len(words) == 0 {
		return []string{prefix}
	}

	// First line with full prefix
	currentLine := ""
	availableWidth := maxWidth - prefixLen

	for _, word := range words {
		// Check if adding this word would exceed the width
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= availableWidth {
			currentLine = testLine
		} else {
			// Current word doesn't fit, start new line
			if currentLine != "" {
				// Finish current line
				styledLine := contentStyle.Render(currentLine)
				lines = append(lines, prefix+styledLine)
				currentLine = word

				// Switch to continuation prefix for subsequent lines
				prefix = strings.Repeat(" ", prefixLen)
				availableWidth = maxWidth - prefixLen
			} else {
				// Single word is too long, force break
				styledWord := contentStyle.Render(word)
				lines = append(lines, prefix+styledWord)
				prefix = strings.Repeat(" ", prefixLen)
			}
		}
	}

	// Add remaining content
	if currentLine != "" {
		styledLine := contentStyle.Render(currentLine)
		lines = append(lines, prefix+styledLine)
	}

	return lines
}

// stripANSI removes ANSI color codes to calculate visible string length
func stripANSI(s string) string {
	// Simple regex to remove ANSI escape sequences
	// This is a basic implementation - for production, consider using a library
	result := ""
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // Skip the '['
		} else if inEscape && (s[i] == 'm' || s[i] == 'K') {
			inEscape = false
		} else if !inEscape {
			result += string(s[i])
		}
	}

	return result
}
