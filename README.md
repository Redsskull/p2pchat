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

# Start chatting
./p2pchat -username alice

# Or with custom settings
./p2pchat -username bob -port 8080 -debug
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

The P2P chat system uses a two-phase approach: **UDP discovery** followed by **TCP messaging**.

```
   Alice's Computer          Network          Bob's Computer
        │                     │                     │
        │ "I'm Alice!"        │                     │
        │ ═══════════════════►│ UDP Multicast       │
        │                     │ ═══════════════════►│ "Oh, Alice exists!"
        │                     │        "I'm Bob!"   │  
        │ UDP Multicast       │ ◄═══════════════════│
        │ ◄═══════════════════│                     │ "Oh, Bob exists!"
        │                     │                     │
        │ "Let's chat, Bob!"  │                     │
        │ ────────────────────┼ TCP Connection      │
        │                     │ ────────────────────┼► "Alice wants to chat!"
        │                     │      "Hi Alice!"    │
        │ TCP Connection      │ ◄───────────────────┤
        │ ◄───────────────────┼                     │ "Bob says hi!"
```

**Phase 1: Discovery (UDP Multicast)**
- Peers broadcast their presence on the local network
- Everyone discovers everyone else automatically

**Phase 2: Messaging (TCP)**
- Direct, reliable connections established between all peer pairs
- Chat messages flow over these stable TCP connections



## Development Status

This is an active development project demonstrating distributed systems concepts and modern Go practices. The core P2P networking and terminal UI functionality is implemented and working.

## Technical Highlights

- **Distributed Systems**: Demonstrates peer-to-peer networking, consensus, and fault tolerance
- **Network Programming**: UDP multicast discovery, TCP connection management
- **Concurrent Programming**: Goroutines and channels for non-blocking network I/O
- **Modern Go**: Clean architecture, proper error handling, comprehensive testing
- **Terminal UIs**: Event-driven programming with Bubble Tea

## Limitations

- **LAN Only**: Uses multicast UDP which doesn't route across the internet
- **Mesh Scaling**: Full mesh topology doesn't scale beyond ~20-30 peers
- **No Persistence**: Messages aren't saved when you disconnect

## Future Enhancements

- DHT-based discovery for internet-wide connectivity
- Message encryption for privacy
- Chat rooms and channels
- File transfer capabilities
- Message history persistence

## License

MIT License - see LICENSE file for details

## Contributing

This project welcomes feedback and contributions. Please see the architecture documentation and current development status before contributing.
