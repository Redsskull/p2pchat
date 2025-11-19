package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"p2pchat/internal/peer"
)

// ConnectionManager handles TCP connections to all discovered peers
type ConnectionManager struct {
	// Configuration
	localPeerID   string
	localUsername string
	localPort     int

	// Connection management
	connections map[string]*PeerConnection // peerID -> connection
	connMutex   sync.RWMutex               // Protects connections map

	// Networking
	listener net.Listener // TCP listener for incoming connections

	// Message handling
	messageHandler func(*Message, string) // Callback for incoming messages

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// PeerConnection represents a TCP connection to a single peer
type PeerConnection struct {
	PeerID    string
	Username  string
	Address   *net.TCPAddr
	Conn      net.Conn
	Connected bool
	LastSeen  time.Time
	SendChan  chan *Message // Channel for outgoing messages
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewConnectionManager creates a new TCP connection manager
func NewConnectionManager(peerID, username string, port int) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConnectionManager{
		localPeerID:   peerID,
		localUsername: username,
		localPort:     port,
		connections:   make(map[string]*PeerConnection),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins listening for incoming TCP connections
func (cm *ConnectionManager) Start() error {
	// Start TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cm.localPort))
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}

	cm.listener = listener
	log.Printf("🔌 TCP listener started on port %d", cm.localPort)

	// Accept incoming connections
	cm.wg.Add(1)
	go cm.acceptConnections()

	return nil
}

// acceptConnections handles incoming TCP connections from other peers
func (cm *ConnectionManager) acceptConnections() {
	defer cm.wg.Done()

	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
			// Set a timeout so we can check for context cancellation
			if tcpListener, ok := cm.listener.(*net.TCPListener); ok {
				tcpListener.SetDeadline(time.Now().Add(time.Second))
			}

			conn, err := cm.listener.Accept()
			if err != nil {
				// Check if it's a timeout (expected) vs real error
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // This is expected, check for cancellation and retry
				}
				log.Printf("❌ Error accepting connection: %v", err)
				continue
			}

			log.Printf("📞 Incoming connection from %s", conn.RemoteAddr())

			// Handle the new connection in a goroutine
			cm.wg.Add(1)
			go cm.handleIncomingConnection(conn)
		}
	}
}

// handleIncomingConnection processes a new incoming TCP connection
func (cm *ConnectionManager) handleIncomingConnection(conn net.Conn) {
	defer cm.wg.Done()
	// Note: Do NOT defer conn.Close() here - ownership transfers to peer connection

	// Read the first message to identify the peer
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("❌ Failed to read peer identification: %v", err)
		conn.Close() // Close on error only
		return
	}

	// Parse the identification message
	var msg Message
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		log.Printf("❌ Failed to parse peer identification: %v", err)
		conn.Close() // Close on error only
		return
	}

	// Create peer connection
	peerConn := &PeerConnection{
		PeerID:    msg.SenderID,
		Username:  msg.Username,
		Address:   conn.RemoteAddr().(*net.TCPAddr),
		Conn:      conn,
		Connected: true,
		LastSeen:  time.Now(),
		SendChan:  make(chan *Message, 100), // Buffer for outgoing messages
	}
	peerConn.ctx, peerConn.cancel = context.WithCancel(cm.ctx)

	// Register the connection
	cm.connMutex.Lock()
	cm.connections[peerConn.PeerID] = peerConn
	cm.connMutex.Unlock()

	log.Printf("✅ Peer connected: %s (%s)", peerConn.Username, peerConn.PeerID)

	// Start message handling goroutines
	cm.wg.Add(2)
	go cm.handlePeerMessages(peerConn, reader)
	go cm.handlePeerSending(peerConn)
}

// ConnectToPeer establishes an outgoing TCP connection to a discovered peer
func (cm *ConnectionManager) ConnectToPeer(p *peer.Peer) error {
	// Check if already connected
	cm.connMutex.RLock()
	existing := cm.connections[p.ID]
	cm.connMutex.RUnlock()

	if existing != nil && existing.Connected {
		return nil // Already connected
	}

	// Leader election: Only connect if our peer ID is smaller
	// This prevents duplicate connections and race conditions
	if cm.localPeerID >= p.ID {
		log.Printf("⏳ Waiting for %s to connect to us (peer ID ordering)", p.Username)
		return nil
	}

	log.Printf("🔗 Connecting to peer %s (%s) at %s", p.Username, p.ID, p.Address)

	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", p.Address.String(), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", p.Address, err)
	}

	// Create peer connection
	peerConn := &PeerConnection{
		PeerID:    p.ID,
		Username:  p.Username,
		Address:   p.Address,
		Conn:      conn,
		Connected: true,
		LastSeen:  time.Now(),
		SendChan:  make(chan *Message, 100),
	}
	peerConn.ctx, peerConn.cancel = context.WithCancel(cm.ctx)

	// Register the connection
	cm.connMutex.Lock()
	cm.connections[p.ID] = peerConn
	cm.connMutex.Unlock()

	// Send identification message
	identMsg := NewJoinMessage(cm.localPeerID, cm.localUsername, 0)
	identJSON, _ := identMsg.ToJSON()

	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(string(identJSON) + "\n")
	if err != nil {
		peerConn.Connected = false
		conn.Close()
		return fmt.Errorf("failed to send identification: %w", err)
	}
	err = writer.Flush()
	if err != nil {
		peerConn.Connected = false
		conn.Close()
		return fmt.Errorf("failed to flush identification: %w", err)
	}
	log.Printf("✅ Connected to peer: %s (%s)", p.Username, p.ID)

	// Start message handling
	reader := bufio.NewReader(conn)
	cm.wg.Add(2)
	go cm.handlePeerMessages(peerConn, reader)
	go cm.handlePeerSending(peerConn)

	return nil
}

// handlePeerMessages reads incoming messages from a peer connection
func (cm *ConnectionManager) handlePeerMessages(peerConn *PeerConnection, reader *bufio.Reader) {
	defer cm.wg.Done()
	defer cm.disconnectPeer(peerConn)

	for {
		select {
		case <-peerConn.ctx.Done():
			return
		default:
			// Set read timeout
			peerConn.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					log.Printf("📞 Peer %s disconnected", peerConn.Username)
				} else {
					log.Printf("❌ Error reading from peer %s: %v", peerConn.Username, err)
				}
				return
			}

			// Parse the message
			msg, err := FromJSON([]byte(line))
			if err != nil {
				log.Printf("❌ Invalid message from peer %s: %v", peerConn.Username, err)
				continue
			}

			// Update last seen
			peerConn.LastSeen = time.Now()

			// Handle the message
			if cm.messageHandler != nil {
				cm.messageHandler(msg, peerConn.PeerID)
			}
		}
	}
}

// handlePeerSending sends outgoing messages to a peer connection
func (cm *ConnectionManager) handlePeerSending(peerConn *PeerConnection) {
	defer cm.wg.Done()

	writer := bufio.NewWriter(peerConn.Conn)

	for {
		select {
		case <-peerConn.ctx.Done():
			return
		case msg := <-peerConn.SendChan:
			// Serialize message
			jsonData, err := msg.ToJSON()
			if err != nil {
				log.Printf("❌ Failed to serialize message: %v", err)
				continue
			}

			// Send message
			peerConn.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err = writer.WriteString(string(jsonData) + "\n")
			if err != nil {
				log.Printf("❌ Failed to send message to peer %s: %v", peerConn.Username, err)
				return
			}

			err = writer.Flush()
			if err != nil {
				log.Printf("❌ Failed to flush message to peer %s: %v", peerConn.Username, err)
				return
			}
		}
	}
}

// Broadcast sends a message to all connected peers
func (cm *ConnectionManager) Broadcast(msg *Message) {
	cm.connMutex.RLock()
	defer cm.connMutex.RUnlock()

	log.Printf("📡 Broadcasting message to %d peers", len(cm.connections))

	for peerID, peerConn := range cm.connections {
		if !peerConn.Connected {
			continue
		}

		select {
		case peerConn.SendChan <- msg:
			// Message queued successfully
		default:
			// Send channel full, peer might be slow or disconnected
			log.Printf("⚠️ Send queue full for peer %s, skipping message", peerID)
		}
	}
}

// SendToPeer sends a message to a specific peer
func (cm *ConnectionManager) SendToPeer(peerID string, msg *Message) error {
	cm.connMutex.RLock()
	peerConn, exists := cm.connections[peerID]
	cm.connMutex.RUnlock()

	if !exists || !peerConn.Connected {
		return fmt.Errorf("peer %s not connected", peerID)
	}

	select {
	case peerConn.SendChan <- msg:
		return nil
	default:
		return fmt.Errorf("send queue full for peer %s", peerID)
	}
}

// disconnectPeer handles peer disconnection cleanup
func (cm *ConnectionManager) disconnectPeer(peerConn *PeerConnection) {
	peerConn.Connected = false
	peerConn.cancel()
	if peerConn.Conn != nil {
		peerConn.Conn.Close()
	}

	log.Printf("❌ Peer disconnected: %s (%s)", peerConn.Username, peerConn.PeerID)
}

// SetMessageHandler sets the callback for incoming messages
func (cm *ConnectionManager) SetMessageHandler(handler func(*Message, string)) {
	cm.messageHandler = handler
}

// GetConnectedPeers returns a list of currently connected peers
func (cm *ConnectionManager) GetConnectedPeers() []string {
	cm.connMutex.RLock()
	defer cm.connMutex.RUnlock()

	peers := make([]string, 0, len(cm.connections))
	for peerID, peerConn := range cm.connections {
		if peerConn.Connected {
			peers = append(peers, peerID)
		}
	}
	return peers
}

// Stop shuts down the connection manager gracefully
func (cm *ConnectionManager) Stop() error {
	log.Printf("🛑 Stopping connection manager...")

	// Cancel all operations
	cm.cancel()

	// Close listener
	if cm.listener != nil {
		cm.listener.Close()
	}

	// Close all peer connections
	cm.connMutex.Lock()
	for _, peerConn := range cm.connections {
		peerConn.cancel()
		peerConn.Conn.Close()
	}
	cm.connMutex.Unlock()

	// Wait for all goroutines to finish
	cm.wg.Wait()

	log.Printf("✅ Connection manager stopped")
	return nil
}
