# P2P Chat

A peer-to-peer IRC-style chat system built in Go with a terminal interface. Connect directly with other users on your local network without needing a central server.

## Features

- **Decentralized**: No server required - peers connect directly to each other
- **Auto-discovery**: Automatically finds other chat users on your local network
- **Terminal UI**: Clean, responsive chat interface built with Bubble Tea
- **Real-time messaging**: Instant message delivery between connected peers
- **Network resilient**: Handles peers joining and leaving gracefully

## Quick Start

```bash
# Clone and build
git clone <repository-url>
cd p2pchat
go build -o p2pchat cmd/p2pchat/main.go

# Start chatting (Terminal 1)
./p2pchat -username alice -port 8080

# Start second user (Terminal 2) 
./p2pchat -username bob -port 8081

# Debug mode (logs to file)
./p2pchat -username charlie -port 8082 -debug
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
-username string    Your display name in chat
-port int          TCP port for peer connections (default: 8080)
-multicast string  Multicast address for discovery (default: 224.0.0.1:9999)
-debug             Enable debug logging
-help              Show help message
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



## Development Status

**COMPLETE: Production-Quality P2P Chat Application! 🚀**

This project successfully demonstrates enterprise-grade distributed systems engineering with a professional terminal user interface. The complete P2P chat system is fully implemented and verified working with real human-to-human communication.

**Core Networking Achievements:**
- ✅ Full mesh P2P networking (every peer connects to every peer)
- ✅ Automatic peer discovery via UDP multicast
- ✅ Real-time message broadcasting verified across 3+ peers
- ✅ Connection retry with exponential backoff
- ✅ Leader election preventing race conditions
- ✅ Production-quality error handling and state management

**Terminal UI Achievements:**
- ✅ Professional Bubble Tea terminal interface using MVU architecture
- ✅ Real-time message display with clean formatting and timestamps
- ✅ Live peer status indicators with connection state visualization
- ✅ Event-driven UI updates: P2P network events automatically refresh interface
- ✅ Complete logging system overhaul with silent mode for clean user experience
- ✅ Seamless integration between UDP discovery + TCP messaging and terminal UI
- ✅ Verified working: Alice ↔ Bob real-time terminal chat sessions

## Technical Highlights

- **Distributed Systems**: Production P2P mesh networking with leader election and fault tolerance
- **Network Programming**: UDP multicast discovery + TCP reliable messaging with retry logic  
- **Terminal UI Development**: Modern Bubble Tea framework with MVU (Model-View-Update) architecture
- **Event-Driven Architecture**: Seamless P2P network events → UI updates via Commands pattern
- **Concurrent Programming**: Advanced goroutines, channels, contexts, and mutex coordination
- **Modern Go**: Clean architecture, proper error handling, centralized logging system
- **Real P2P Achievement**: Verified Alice ↔ Bob real-time chat with professional terminal interface

## Current Limitations

- **LAN Only**: Uses multicast UDP which doesn't route across the internet
- **Mesh Scaling**: Full mesh topology doesn't scale beyond ~20-30 peers
- **Message History**: Limited to current session, no scrolling for long conversations
- **No Persistence**: Messages aren't saved when you disconnect

## Future Enhancements

### Short-term (UI Polish)
- Scrollable message history with pagination
- Message history persistence to files
- Color-coded messages by user
- Chat commands (/users, /quit, /help, /nick)
- Window resizing support and responsive layouts

### Long-term (Network Features)
- DHT-based discovery for internet-wide connectivity
- Message encryption for privacy
- Chat rooms and channels
- File transfer capabilities
- Performance optimizations for larger peer groups

## License

MIT License - see LICENSE file for details

## Contributing

This project welcomes feedback and contributions. Please see the architecture documentation and current development status before contributing.
