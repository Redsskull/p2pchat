package components

import "time"

// PeerListComponent handles the peer sidebar
type PeerListComponent struct {
	Peers  []PeerDisplay
	Width  int
	Height int
}

// DisplayMessage placeholder for component
type DisplayMessage struct {
	Content   string
	Username  string
	Timestamp time.Time
}

// PeerDisplay placeholder for component
type PeerDisplay struct {
	Username string
	Status   string
}

// TODO: Implement peer list component
// For now, this is just a placeholder
