package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"p2pchat/internal/peer"
	"p2pchat/pkg/logger"
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

	// Connection retry
	retryTicker *time.Ticker

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ConnectionState represents the current state of a peer connection
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota // Not connected
	StateConnecting                          // Attempting to connect
	StateConnected                           // Successfully connected
	StateFailed                              // Connection failed, will retry
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// PeerConnection represents a TCP connection to a single peer
type PeerConnection struct {
	PeerID      string
	Username    string
	Address     *net.TCPAddr
	Conn        net.Conn
	State       ConnectionState
	LastSeen    time.Time
	LastAttempt time.Time
	RetryCount  int
	SendChan    chan *Message // Channel for outgoing messages
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewConnectionManager creates a new TCP connection manager
func NewConnectionManager(peerID, username string, port int) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConnectionManager{
		localPeerID:   peerID,
		localUsername: username,
		localPort:     port,
		connections:   make(map[string]*PeerConnection),
		retryTicker:   time.NewTicker(10 * time.Second),
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
	logger.Debug("🔌 TCP listener started on port %d", cm.localPort)

	// Accept incoming connections
	cm.wg.Add(1)
	go cm.acceptConnections()

	// Start connection retry loop
	cm.wg.Add(1)
	go cm.connectionRetryLoop()

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
				tcpListener.SetDeadline(time.Now().Add(5 * time.Second))
			}

			conn, err := cm.listener.Accept()
			if err != nil {
				// Check if it's a timeout (expected) vs real error
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // This is expected, check for cancellation and retry
				}
				logger.Error("❌ Error accepting connection: %v", err)
				continue
			}

			logger.Debug("📞 Incoming connection from %s", conn.RemoteAddr())

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
		logger.Error("❌ Failed to read peer identification: %v", err)
		conn.Close() // Close on error only
		return
	}

	// Parse the identification message
	var msg Message
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		logger.Error("❌ Failed to parse peer identification: %v", err)
		conn.Close() // Close on error only
		return
	}

	// Check if we already have a connection entry for this peer
	cm.connMutex.Lock()
	existing := cm.connections[msg.SenderID]
	var peerConn *PeerConnection

	if existing != nil {
		// Update existing connection with new socket
		existing.Conn = conn
		existing.State = StateConnected
		existing.LastSeen = time.Now()
		existing.Address = conn.RemoteAddr().(*net.TCPAddr)
		peerConn = existing

	} else {
		// Create new peer connection
		peerConn = &PeerConnection{
			PeerID:   msg.SenderID,
			Username: msg.Username,
			Address:  conn.RemoteAddr().(*net.TCPAddr),
			Conn:     conn,
			State:    StateConnected,
			LastSeen: time.Now(),
			SendChan: make(chan *Message, 100), // Buffer for outgoing messages
		}
		peerConn.ctx, peerConn.cancel = context.WithCancel(cm.ctx)
		cm.connections[msg.SenderID] = peerConn

	}
	cm.connMutex.Unlock()

	logger.Debug("✅ Peer connected: %s (%s)", peerConn.Username, peerConn.PeerID)

	// Start message handling goroutines
	cm.wg.Add(2)
	go cm.handlePeerMessages(peerConn, reader)
	go cm.handlePeerSending(peerConn)
}

// ConnectToPeer establishes an outgoing TCP connection to a discovered peer
func (cm *ConnectionManager) ConnectToPeer(p *peer.Peer) error {
	// Check if already connected or connecting
	cm.connMutex.RLock()
	existing := cm.connections[p.ID]
	cm.connMutex.RUnlock()

	if existing != nil && (existing.State == StateConnected || existing.State == StateConnecting) {

		return nil // Already connected or connecting
	}

	// Leader election: Only connect if peer ID is smaller
	// This prevents duplicate connections and race conditions
	if cm.localPeerID >= p.ID {
		logger.Debug("⏳ Waiting for %s to connect to us (peer ID ordering)", p.Username)
		return nil
	}

	// Create or update peer connection entry
	cm.connMutex.Lock()
	if existing == nil {
		existing = &PeerConnection{
			PeerID:   p.ID,
			Username: p.Username,
			Address:  p.Address,
			State:    StateDisconnected,
			SendChan: make(chan *Message, 100),
		}
		existing.ctx, existing.cancel = context.WithCancel(cm.ctx)
		cm.connections[p.ID] = existing
	}
	cm.connMutex.Unlock()

	// Attempt connection
	return cm.attemptConnection(existing)
}

// attemptConnection tries to establish a TCP connection to a peer
func (cm *ConnectionManager) attemptConnection(peerConn *PeerConnection) error {
	peerConn.State = StateConnecting
	peerConn.LastAttempt = time.Now()

	logger.Debug("🔗 Connecting to peer %s (%s) at %s (attempt %d)",
		peerConn.Username, peerConn.PeerID, peerConn.Address, peerConn.RetryCount+1)

	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", peerConn.Address.String(), 5*time.Second)
	if err != nil {
		peerConn.State = StateFailed
		peerConn.RetryCount++
		logger.Error("❌ Failed to connect to peer %s: %v (will retry)", peerConn.Username, err)
		return fmt.Errorf("failed to connect to %s: %w", peerConn.Address, err)
	}

	// Update connection
	peerConn.Conn = conn
	peerConn.State = StateConnected
	peerConn.LastSeen = time.Now()
	peerConn.RetryCount = 0

	// Send identification message
	identMsg := NewJoinMessage(cm.localPeerID, cm.localUsername, 0)
	identJSON, _ := identMsg.ToJSON()

	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(string(identJSON) + "\n")
	if err != nil {
		peerConn.State = StateFailed
		conn.Close()
		return fmt.Errorf("failed to send identification: %w", err)
	}
	err = writer.Flush()
	if err != nil {
		peerConn.State = StateFailed
		conn.Close()
		return fmt.Errorf("failed to flush identification: %w", err)
	}

	logger.Debug("✅ Connected to peer: %s (%s)", peerConn.Username, peerConn.PeerID)

	// Start message handling
	reader := bufio.NewReader(conn)
	cm.wg.Add(2)
	go cm.handlePeerMessages(peerConn, reader)
	go cm.handlePeerSending(peerConn)

	return nil
}

// connectionRetryLoop periodically retries failed connections
func (cm *ConnectionManager) connectionRetryLoop() {
	defer cm.wg.Done()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-cm.retryTicker.C:
			cm.retryFailedConnections()
		}
	}
}

// retryFailedConnections attempts to reconnect to failed peers
func (cm *ConnectionManager) retryFailedConnections() {
	cm.connMutex.RLock()
	var failedPeers []*PeerConnection
	for _, peerConn := range cm.connections {
		if peerConn.State == StateFailed {
			// Exponential backoff: wait longer after each failure
			backoffDelay := time.Duration(1<<uint(min(peerConn.RetryCount, 6))) * time.Second // Max 64s
			if time.Since(peerConn.LastAttempt) > backoffDelay {
				failedPeers = append(failedPeers, peerConn)
			}
		}
	}
	cm.connMutex.RUnlock()

	// Retry failed connections
	for _, peerConn := range failedPeers {
		// Only retry if we should initiate the connection (leader election)
		if cm.localPeerID < peerConn.PeerID {
			go cm.attemptConnection(peerConn)
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
			// Set read timeout - longer for interactive chat
			peerConn.Conn.SetReadDeadline(time.Now().Add(2 * time.Minute))

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					logger.Debug("📞 Peer %s disconnected", peerConn.Username)
				} else {
					logger.Error("❌ Error reading from peer %s: %v", peerConn.Username, err)
				}
				return
			}

			// Parse the message
			msg, err := FromJSON([]byte(line))
			if err != nil {
				logger.Error("❌ Invalid message from peer %s: %v", peerConn.Username, err)
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
				logger.Error("❌ Failed to serialize message: %v", err)
				continue
			}

			// Send message
			peerConn.Conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
			_, err = writer.WriteString(string(jsonData) + "\n")
			if err != nil {
				logger.Error("❌ Failed to send message to peer %s: %v", peerConn.Username, err)
				return
			}

			err = writer.Flush()
			if err != nil {
				logger.Error("❌ Failed to flush message to peer %s: %v", peerConn.Username, err)
				return
			}
		}
	}
}

// Broadcast sends a message to all connected peers
func (cm *ConnectionManager) Broadcast(msg *Message) {
	cm.connMutex.RLock()
	defer cm.connMutex.RUnlock()

	connectedCount := 0
	for _, peerConn := range cm.connections {
		if peerConn.State == StateConnected {
			connectedCount++
		}
	}

	logger.Debug("📡 Broadcasting message to %d connected peers", connectedCount)

	for peerID, peerConn := range cm.connections {
		if peerConn.State != StateConnected {
			continue
		}

		select {
		case peerConn.SendChan <- msg:
			// Message queued successfully
		default:
			// Send channel full, peer might be slow or disconnected
			logger.Error("⚠️ Send queue full for peer %s, skipping message", peerID)
		}
	}
}

// SendToPeer sends a message to a specific peer
func (cm *ConnectionManager) SendToPeer(peerID string, msg *Message) error {
	cm.connMutex.RLock()
	peerConn, exists := cm.connections[peerID]
	cm.connMutex.RUnlock()

	if !exists || peerConn.State != StateConnected {
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
	peerConn.State = StateFailed
	peerConn.cancel()
	if peerConn.Conn != nil {
		peerConn.Conn.Close()
		peerConn.Conn = nil
	}

	logger.Debug("❌ Peer disconnected: %s (%s) - will retry connection", peerConn.Username, peerConn.PeerID)
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
		if peerConn.State == StateConnected {
			peers = append(peers, peerID)
		}
	}
	return peers
}

// Stop shuts down the connection manager gracefully
func (cm *ConnectionManager) Stop() error {
	logger.Debug("🛑 Stopping connection manager...")

	// Cancel all operations
	cm.cancel()

	// Stop retry ticker
	if cm.retryTicker != nil {
		cm.retryTicker.Stop()
	}

	// Close listener
	if cm.listener != nil {
		cm.listener.Close()
	}

	// Close all peer connections
	cm.connMutex.Lock()
	for _, peerConn := range cm.connections {
		peerConn.cancel()
		if peerConn.Conn != nil {
			peerConn.Conn.Close()
		}
	}
	cm.connMutex.Unlock()

	// Wait for all goroutines to finish
	cm.wg.Wait()

	logger.Debug("✅ Connection manager stopped")
	return nil
}
