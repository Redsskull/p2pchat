package main

import (
	"fmt"
	"net"
	"testing"
)

func TestIsPortAvailable(t *testing.T) {
	// Test that we can check port availability
	// Use a known unavailable port by binding to it first
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()

	// Get the port number
	addr := listener.Addr().(*net.TCPAddr)
	busyPort := addr.Port

	// Test that the busy port is not available
	if isPortAvailable(busyPort) {
		t.Errorf("Port %d should not be available (it's in use)", busyPort)
	}

	// Test some port that should be available (very high port number)
	availablePort := 65432
	if !isPortAvailable(availablePort) {
		t.Logf("Port %d is not available (this might be normal)", availablePort)
	}
}

func TestFindAvailablePort(t *testing.T) {
	// Test that findAvailablePort returns a valid port
	port := findAvailablePort()

	if port < 1024 || port > 65535 {
		t.Errorf("Port %d is out of valid range (1024-65535)", port)
	}

	// Test that the returned port is actually available
	if !isPortAvailable(port) {
		t.Errorf("findAvailablePort returned port %d which is not available", port)
	}
}

func TestPortRangeLogic(t *testing.T) {
	// Test that our port range constants are valid
	if PortRangeStart < 1024 {
		t.Errorf("PortRangeStart (%d) should be >= 1024", PortRangeStart)
	}

	if PortRangeEnd <= PortRangeStart {
		t.Errorf("PortRangeEnd (%d) should be > PortRangeStart (%d)", PortRangeEnd, PortRangeStart)
	}

	if PortRangeEnd > 65535 {
		t.Errorf("PortRangeEnd (%d) should be <= 65535", PortRangeEnd)
	}

	// Test that the range is reasonable (not too small)
	rangeSize := PortRangeEnd - PortRangeStart + 1
	if rangeSize < 10 {
		t.Errorf("Port range size (%d) seems too small for multiple users", rangeSize)
	}
}

func TestMultiplePortAssignments(t *testing.T) {
	var listeners []net.Listener
	var ports []int
	portSeen := make(map[int]bool) // Friend suggested this today. I didn't even know about maps.

	for i := 0; i < 5; i++ {
		port := findAvailablePort()

		// Check for duplicates
		if portSeen[port] {
			t.Errorf("Got duplicate port %d", port)
		}
		portSeen[port] = true

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			t.Logf("Could not bind to port %d: %v", port, err)
			continue
		}
		listeners = append(listeners, listener)
		ports = append(ports, port)
	}

	// Clean up
	for _, listener := range listeners {
		listener.Close()
	}

	if len(ports) == 0 {
		t.Error("Could not get any available ports")
	}

	t.Logf("Got %d unique ports: %v", len(ports), ports)
}

func TestUsernameValidation(t *testing.T) {
	// Test getDefaultUsername function
	username := getDefaultUsername()

	if username == "" {
		t.Error("getDefaultUsername should never return empty string")
	}

	if len(username) > 20 {
		t.Errorf("Default username '%s' is too long (%d chars, max 20)", username, len(username))
	}
}
