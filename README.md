# P2P Chat

[![asciicast](https://asciinema.org/a/ykPnzDlq7LGyskLnWRf5NWO1T.svg)](https://asciinema.org/a/ykPnzDlq7LGyskLnWRf5NWO1T)

A peer-to-peer IRC-style chat system built in Go with a terminal interface. Connect directly with other users on your local network without needing a central server.

## Features

- **Decentralized**: No server required - peers connect directly to each other
- **Auto-discovery**: Automatically finds other chat users on your local network
- **Beautiful Terminal UI**: Stunning color-coded chat interface with scrollable history
- **Visual Excellence**: 10-color user palette, intelligent text wrapping, elegant typography
- **Real-time messaging**: Instant message delivery between connected peers with live status
- **Enhanced UX**: Smart error handling, responsive layouts, perfect keyboard navigation
- **Network resilient**: Handles peers joining and leaving gracefully with visual feedback
- **Production Quality**: Thread-safe message storage with duplicate detection

## Quick Start

```bash
# Clone and build
git clone <repository-url>
cd p2pchat
go build -o p2pchat cmd/p2pchat/main.go

# Interactive mode (recommended for first-time users)
./p2pchat
# You'll be prompted for username, port is auto-assigned

# Or specify username directly (port auto-assigned)
./p2pchat -username alice

# Full manual configuration
./p2pchat -username alice -port 8080

# Debug mode with interactive setup
./p2pchat -debug

# Try the interactive demo script
./demo-interactive.sh
```

## How It Works

P2P Chat uses UDP multicast for peer discovery on your local network, then establishes direct TCP connections between peers for reliable messaging. Each peer maintains connections to all other peers in a full mesh topology.

```
[Alice] ←→ [Bob]
   ↕        ↕
[Charlie] ←→ [Dave]
```

## Architecture

- **Peer Discovery**: UDP multicast (224.0.0.1:9999) for finding peers on LAN
- **Messaging**: Direct TCP connections for reliable chat delivery  
- **Protocol**: JSON-based messages inspired by IRC
- **UI**: Terminal interface using Bubble Tea framework
- **Concurrency**: Goroutines handle network I/O without blocking the UI

## Command Line Options

```
-username string    Your display name in chat (interactive prompt if not provided)
-port int          TCP port for peer connections (auto-assigned if not provided)  
-multicast string  Multicast address for discovery (default: 224.0.0.1:9999)
-debug             Enable debug logging to file
-help              Show help message
```

### Usage Examples

**Interactive Mode (Recommended)**
```bash
./p2pchat
# Prompts for username, auto-assigns port
```

**Quick Start with Username**
```bash
./p2pchat -username alice
# Uses specified username, auto-assigns port
```

**Full Manual Configuration**
```bash
./p2pchat -username alice -port 8080
# Specify both username and port
```

**Debug Mode**
```bash
./p2pchat -username alice -debug
# Enable debug logging to p2pchat-debug.log
```

**Help**
```bash
./p2pchat -help
# Show all available options
```

## Interactive Demo Script

I've included a comprehensive demo script that showcases all the new interactive features:

```bash
./demo-interactive.sh
```

The demo script offers 6 different modes:

1. **Interactive Mode** - Experience the full interactive setup
2. **Quick Start** - Username specified, port auto-assigned  
3. **Manual Configuration** - Full control over settings
4. **Debug Mode** - Interactive setup with debug logging
5. **Help Message** - View all available options
6. **Multi-User Simulation** - See automatic port assignment in action

### Running the Demo

```bash
# Make sure you've built the application first
go build -o p2pchat cmd/p2pchat/main.go

# Make the demo script executable (if needed)
chmod +x demo-interactive.sh

# Run the interactive demo
./demo-interactive.sh
```

The demo is perfect for:
- First-time users wanting to explore features
- Demonstrating the app to others
- Testing different configuration options
- Understanding automatic port assignment

### Live Demo

*The asciinema recording at the top of this README shows P2P Chat in action with real users chatting, demonstrating the beautiful terminal UI, automatic peer discovery, real-time messaging, and seamless multi-user experience.*

## Automatic Port Assignment

P2P Chat intelligently handles port assignment to make connecting multiple users effortless:

- **Automatic Range**: Searches ports 8080-8999 for first available port
- **Collision Detection**: Automatically finds free ports when multiple users start simultaneously
- **System Fallback**: Uses system-assigned port if preferred range is exhausted
- **Manual Override**: Command line `-port` flag still works for specific port requirements

This means you can easily start multiple chat instances without worrying about port conflicts:

```bash
# Terminal 1
./p2pchat -username alice
# Auto-assigns port 8080

# Terminal 2 
./p2pchat -username bob  
# Auto-assigns port 8081 (8080 was taken)

# Terminal 3
./p2pchat -username charlie
# Auto-assigns port 8082 (8080, 8081 were taken)
```

## Message Protocol

Messages are JSON-encoded and sent over TCP:

```json
{
  "type": "chat",
  "sender": "alice",
  "content": "Hello everyone!",
  "timestamp": "2025-11-16T10:30:00Z",
  "sequence": 42
}
```

## Requirements

- Go 1.21 or later
- Network access (LAN for peer discovery)
- Terminal with UTF-8 support

## Project Structure

```
p2pchat/
├── cmd/p2pchat/          # Main application
├── pkg/                  # Public packages  
│   ├── discovery/        # Peer discovery
│   ├── chat/            # TCP connections & messaging
│   └── ui/              # Terminal interface
├── internal/            # Private packages
│   └── peer/            # Peer data structures
└── docs/               # Documentation
```

## Network Architecture

The P2P chat system creates a **full mesh network** where every peer connects to every other peer:

```
                    FULL MESH P2P NETWORK
                    
         Alice ●─────────────────● Bob
           │ ╲                 ╱ │
           │   ╲             ╱   │
           │     ╲         ╱     │
           │       ╲     ╱       │
           │         ╲ ╱         │
           │           ╲         │
           │         ╱ ╲         │
           │       ╱     ╲       │
           │     ╱         ╲     │
           │   ╱             ╲   │
           │ ╱                 ╲ │
         Charlie ●─────────────────● 

    Every peer talks to every other peer!
    
    Real-time messages verified:
    • Alice: "Hello I'm Alice!" → Bob ✓ & Charlie ✓  
    • Bob: "Hello I'm Bob!" → Alice ✓ & Charlie ✓
    • Charlie: "Hello I'm Charlie!" → Alice ✓ & Bob ✓
```

**Phase 1: UDP Discovery**
- Automatic peer discovery via multicast (224.0.0.1:9999)
- Any startup order works - true P2P resilience

**Phase 2: TCP Mesh Connections**
- Leader election prevents connection races
- Automatic retry with exponential backoff
- Full mesh: 3 peers = 3 bidirectional connections



## Technical Architecture

P2P Chat implements a sophisticated distributed systems architecture:

- **Distributed Networking**: Full mesh P2P topology where every peer connects directly to every other peer
- **Automatic Discovery**: UDP multicast for finding peers on your local network  
- **Reliable Messaging**: TCP connections ensure message delivery between peers
- **Beautiful Terminal UI**: Modern interface built with Bubble Tea framework
- **Real-time Updates**: Live peer status and instant message delivery
- **Fault Tolerance**: Automatic reconnection and network resilience
- **Resource Efficient**: Lightweight design with minimal CPU and memory usage

## Current Limitations

- **LAN Only**: Uses multicast UDP which doesn't route across the internet
- **Mesh Scaling**: Full mesh topology doesn't scale beyond ~20-30 peers
- **No Persistence**: Messages aren't saved when you disconnect (by design - privacy-focused)
- **Advanced Features**: No encryption, file transfer, or chat rooms (future enhancements)

## Future Enhancements

### Advanced Features (Optional)
- Chat commands (/users, /quit, /help, /nick, /clear)
- Message encryption for privacy
- File transfer capabilities
- Performance optimizations for 50+ peer groups

### Network Expansion (Long-term)
- DHT-based discovery for internet-wide connectivity
- Chat rooms and channels
- Voice chat integration
- Mobile client compatibility

---

## License

MIT License - see LICENSE file for details

## Contributing

This project welcomes feedback and contributions. Please see the architecture documentation before contributing.
