package ui

import (
	"p2pchat/pkg/chat"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Event messages that your Update() function handles
type IncomingMessageMsg struct {
	Message *chat.Message
}

// NEW: Message for loading existing chat history
type MessageHistoryMsg struct {
	Messages []*chat.Message
}

type PeerUpdateMsg struct {
	Peers []chat.PeerInfo
}

type StatusUpdateMsg struct {
	Status  string
	IsError bool
}

// Commands that bridge your ChatService to Bubble Tea
func ListenForMessages(chatService *chat.ChatService) tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-chatService.GetMessages():
			return IncomingMessageMsg{Message: msg}
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}
}

// NEW: Load existing message history from ChatService
func LoadMessageHistory(chatService *chat.ChatService) tea.Cmd {
	return func() tea.Msg {
		messages := chatService.GetMessageHistory()
		return MessageHistoryMsg{Messages: messages}
	}
}

func SendMessageCmd(chatService *chat.ChatService, content string) tea.Cmd {
	return func() tea.Msg {
		err := chatService.SendMessage(content)
		if err != nil {
			return StatusUpdateMsg{Status: "Error: " + err.Error(), IsError: true}
		}
		return StatusUpdateMsg{Status: "Message sent", IsError: false}
	}
}

func UpdatePeers(chatService *chat.ChatService) tea.Cmd {
	return func() tea.Msg {
		peers := chatService.GetConnectedPeers()
		return PeerUpdateMsg{Peers: peers}
	}
}

func PeriodicPeerUpdate() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return struct{}{} // This matches your update.go handler
	})
}
