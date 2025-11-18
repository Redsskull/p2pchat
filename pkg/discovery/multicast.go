package discovery

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

const (
	DefaultMulticastAddr = "224.0.0.1:9999"
	MaxMessageSize       = 1024 // bytes
)

// MulticastService handles UDP multicast for peer discovery
type MulticastService struct {
	multicastAddr *net.UDPAddr
	conn          *net.UDPConn
	localAddr     *net.UDPAddr
}

// NewMulticastService creates a new multicast service
func NewMulticastService(multicastAddress string) (*MulticastService, error) {
	// Parse the multicast address
	addr, err := net.ResolveUDPAddr("udp", multicastAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid multicast address %s: %w", multicastAddress, err)
	}

	// Validate it's actually a multicast address
	if !addr.IP.IsMulticast() {
		return nil, fmt.Errorf("address %s is not a multicast address", addr.IP)
	}

	return &MulticastService{
		multicastAddr: addr,
	}, nil
}

// Start begins listening for multicast messages
func (ms *MulticastService) Start() error {
	// Listen on the multicast address
	conn, err := net.ListenMulticastUDP("udp", nil, ms.multicastAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on multicast address: %w", err)
	}

	ms.conn = conn

	// Get our local address for logging
	ms.localAddr = conn.LocalAddr().(*net.UDPAddr)

	// ENABLE MULTICAST LOOPBACK - This is the key fix!
	rawConn, err := ms.conn.SyscallConn()
	if err == nil {
		rawConn.Control(func(fd uintptr) {
			// Enable multicast loopback on the socket
			syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_LOOP, 1)
			// Set TTL to 1 (local network only)
			syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_TTL, 1)
		})
	}

	fmt.Printf("🔊 Multicast service listening on %s (local: %s)\n",
		ms.multicastAddr, ms.localAddr)

	return nil
}

// Stop closes the multicast connection
func (ms *MulticastService) Stop() error {
	if ms.conn != nil {
		err := ms.conn.Close()
		ms.conn = nil // Prevent double-close
		return err
	}
	return nil
}

// Send broadcasts a message to all peers on the network
func (ms *MulticastService) Send(message *DiscoveryMessage) error {
	if ms.conn == nil {
		return fmt.Errorf("multicast service not started")
	}

	// Convert message to JSON
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Check message size
	if len(data) > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", len(data), MaxMessageSize)
	}

	// Send to multicast group
	_, err = ms.conn.WriteToUDP(data, ms.multicastAddr)
	if err != nil {
		return fmt.Errorf("failed to send multicast message: %w", err)
	}

	fmt.Printf("📤 Sent: %s (%d bytes)\n", message.String(), len(data))
	return nil
}

// ReceiveWithTimeout listens for incoming discovery messages with custom timeout
func (ms *MulticastService) ReceiveWithTimeout(timeout time.Duration) (*DiscoveryMessage, *net.UDPAddr, error) {
	if ms.conn == nil {
		return nil, nil, fmt.Errorf("multicast service not started")
	}

	// Set custom read timeout
	ms.conn.SetReadDeadline(time.Now().Add(timeout))

	// Read from network
	buffer := make([]byte, MaxMessageSize)
	n, senderAddr, err := ms.conn.ReadFromUDP(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil, fmt.Errorf("read timeout after %v: %w", timeout, err)
		}
		return nil, nil, fmt.Errorf("failed to read multicast message: %w", err)
	}

	// Parse the JSON message
	message, err := FromJSON(buffer[:n])
	if err != nil {
		return nil, senderAddr, fmt.Errorf("failed to parse message from %s: %w", senderAddr, err)
	}

	fmt.Printf("📥 Received: %s from %s (%d bytes)\n",
		message.String(), senderAddr, n)

	return message, senderAddr, nil
}

// Receive listens for incoming discovery messages with default timeout
func (ms *MulticastService) Receive() (*DiscoveryMessage, *net.UDPAddr, error) {
	return ms.ReceiveWithTimeout(5 * time.Second) // Production timeout
}

// GetLocalAddr returns local UDP address
func (ms *MulticastService) GetLocalAddr() *net.UDPAddr {
	return ms.localAddr
}
