# P2P Chat Architecture Notes

## Project Overview
IRC-style peer-to-peer chat system in Go with terminal UI.

**Timeline**: Nov 16-30, 2025  
**Target**: Portfolio project demonstrating distributed systems and P2P networking

---

## Architecture Decisions

### 1. Peer Discovery Strategy
**Decision**: [x] Multicast UDP / [ ] DHT / [ ] Bootstrap Nodes  
**Reasoning**: 
- Multicast working only locally does turn me off. But, it is also the simplest, and a good learning tool for the very first time doing this.
- Perfect for LAN-based chat demonstration and portfolio purposes
- Can be extended later with DHT for internet-wide functionality
- Low complexity allows focus on P2P concepts and UI polish

**Implementation Notes**:
- Multicast address: `224.0.0.1:9999` (standard local network multicast)
- Beacon interval: 5 seconds for peer announcements
- Peer timeout: 15 seconds (3 missed beacons)
- Discovery message format:
```json
{
  "type": "discover",
  "peer_id": "uuid",
  "address": "192.168.1.100:8080",
  "timestamp": "2025-11-16T10:30:00Z"
}
```

### 2. Network Communication
**Discovery**: UDP Multicast  
**Messaging**: TCP connections  
**Topology**: [ x] Full Mesh / [ ] Star / [ ] Ring

**Connection Strategy**:
- Each peer maintains TCP connections to all discovered peers
- Incoming connections accepted on random available port
- Connection attempts with exponential backoff on failure
- Graceful connection cleanup when peers leave

### 3. Message Protocol
**Format**: [ x] JSON / [ ] Binary / [ ] Text-based  
**Inspiration**: IRC Protocol (RFC 1459)

**Message Structure**:
```json
{
  "type": "chat|join|leave|ping|pong",
  "sender": "peer_id", 
  "content": "message content",
  "timestamp": "2025-11-16T10:30:00Z",
  "sequence": 123,
  "target": "general" // room/channel (future)
}
```

**Message Types**:
- `chat`: Regular chat message
- `join`: Peer joined network  
- `leave`: Peer leaving network
- `ping`: Health check request
- `pong`: Health check response

### 4. Terminal UI Framework
**Decision**: [x] Bubble Tea / [ ] TView  
**Reasoning**:
- I choose bubble tea because TView is widget based and I have design nightmares from Fyne
- Elm Architecture provides cleaner state management for complex UI interactions
- Better separation of concerns between UI logic and business logic
- More modern approach that's impressive in portfolio context

**UI Layout Design**:
```
┌─────────────────────────────────────────────┐
│ P2P Chat - Connected Peers: 3               │
├─────────────────────────────────┬───────────┤
│ Chat Messages                   │ Users     │
│ [10:30] alice: Hello everyone!  │ • alice   │
│ [10:31] bob: Hey there!         │ • bob     │
│ [10:32] charlie: What's up?     │ • charlie │
│                                 │ • you     │
│                                 │           │
├─────────────────────────────────┴───────────┤
│ > Type your message...                      │
└─────────────────────────────────────────────┘
```

---

## Research Notes

### Peer Discovery Patterns

#### Multicast UDP
**Pros**:
- Simple implementation - no complex routing or discovery protocols
- Works immediately on local networks without configuration
- Low latency for peer discovery (single broadcast)
- Easy to debug and test with network tools

**Cons**:
- Limited to local network segments (LAN only)
- Blocked by many routers and firewalls
- Doesn't scale beyond ~50-100 peers efficiently
- No persistence - peers must actively announce presence

**Implementation Complexity**: Low  
**Suitable for**: LAN-based chat, 2-10 peers

#### DHT (Distributed Hash Table)
**Key Concepts**:
- Node IDs and XOR distance metric
- k-buckets for routing table
- Self-organizing network

**Pros**:
- Scales to millions of peers across the internet
- Self-organizing and fault-tolerant
- No single point of failure
- Efficient logarithmic lookup performance

**Cons**:
- Complex implementation requiring deep understanding of distributed systems
- Higher latency for peer discovery (multiple network hops)
- Requires sophisticated error handling and edge case management
- Overkill for small group chat applications

**Implementation Complexity**: High  
**Suitable for**: Large-scale P2P systems (1000+ peers)

#### Bootstrap Nodes
**How it works**:
- Well-known servers that maintain lists of active peers
- New peers contact bootstrap nodes to get initial peer list
- Bootstrap nodes can provide introductions between peers

**Pros**:
- Simple to implement and understand
- Works across the internet reliably
- Good balance between complexity and functionality
- Fallback option when other discovery methods fail

**Cons**:
- Creates dependency on bootstrap node availability
- Single point of failure if only one bootstrap node
- Requires maintaining and hosting bootstrap infrastructure

### P2P Protocol Analysis

#### BitTorrent DHT Operations
1. **PING**: Check peer availability
2. **FIND_NODE**: Locate peers near target ID
3. **GET_PEERS**: Find peers with specific data
4. **ANNOUNCE**: Declare data availability

**Key Insights for Chat**:
- Health checking (PING/PONG) is essential for detecting disconnected peers
- Announce pattern useful for broadcasting user status changes
- Parallel queries show importance of concurrent network operations

#### Kademlia Concepts
- **Distance Metric**: XOR of node IDs
- **Routing**: Logarithmic routing table
- **Lookup**: Parallel queries for efficiency

**Relevance to Chat**:
- XOR distance metric could be used for consistent peer ordering
- Self-organizing properties inspire resilient network design
- Logarithmic scaling concepts applicable to future DHT implementation

### IRC Protocol Analysis

#### Message Format Pattern
```
:prefix COMMAND params :trailing
```

#### Essential Commands
- `PRIVMSG`: Send message to user/channel
- `JOIN/PART`: Join/leave channel
- `NICK`: Change nickname  
- `QUIT`: Disconnect
- `PING/PONG`: Keep-alive

#### Adapted for P2P Chat
```json
// Instead of IRC's ":nick!user@host PRIVMSG #channel :message"
{
  "type": "chat",
  "sender": "alice_uuid", 
  "content": "Hello everyone!",
  "timestamp": "2025-11-16T10:30:00Z"
}
```

### TUI Framework Comparison

#### Bubble Tea
**Architecture**: The Elm Architecture (Model-Update-View)
```go
type model struct {
    messages []Message
    peers    []Peer
    input    string
}
```

**Pros**:
- Clean functional architecture prevents state management bugs
- Excellent documentation and active community
- Built-in input handling and event system
- Modern approach that showcases advanced Go patterns

**Cons**:
- Steeper learning curve for developers used to imperative UI
- Less direct control over rendering specifics
- Smaller ecosystem compared to more established frameworks

#### TView
**Architecture**: Widget-based immediate mode
```go
app := tview.NewApplication()
chatView := tview.NewTextView()
inputField := tview.NewInputField()
```

**Pros**:
- Mature and battle-tested with large community
- Rich set of pre-built widgets and layouts
- More intuitive for developers with GUI experience
- Direct control over widget behavior and appearance

**Cons**:
- Widget-based approach can lead to complex state management
- Less modern architectural patterns
- Potential for callback hell in complex UIs
- More imperative style that's harder to test

---

## System Architecture

### Components
1. **Peer Discovery Service**: UDP multicast listener/broadcaster
2. **Connection Manager**: TCP connection handling
3. **Message Router**: Route messages between peers
4. **Terminal UI**: Display chat and handle input
5. **Peer Registry**: Track connected peers

### Data Flow
```
[User Input] → [Terminal UI] → [Message Router] → [Connection Manager] → [Network]
     ↑                                                                        ↓
[Terminal UI] ← [Message Router] ← [Connection Manager] ← [Peer Discovery] ← [Network]
```

### Concurrency Design
- **Main Goroutine**: Terminal UI event loop
- **Discovery Goroutine**: UDP multicast listener
- **Connection Goroutines**: One per TCP connection
- **UI Update Goroutine**: Handle incoming messages

---

## Implementation Plan

### Phase 1: Discovery (Days 2-3) ✅ COMPLETED
- [x] UDP multicast beacon system (with syscall optimization for Arch Linux)
- [x] Peer registry with timeout mechanism (10s stale, 30s offline)
- [x] Peer discovery service with beacon/receive/cleanup goroutines
- [x] Event-driven peer join/leave handling
- [x] Multi-instance testing framework
- [x] Real-time peer discovery working between multiple peers

**Day 2 Achievement**: Successfully built complete P2P peer discovery system with:
- Automatic peer detection via UDP multicast (224.0.0.1:9999)
- Thread-safe peer registry with timeout management
- Professional logging and status reporting
- Graceful shutdown with leave message broadcasting
- Verified working across multiple terminal instances

### Phase 2: Messaging (Days 4-5)  
- [ ] TCP connection establishment between discovered peers
- [ ] JSON chat message protocol (separate from discovery messages)
- [ ] TCP message routing and broadcasting
- [ ] Message history storage and ordering

### Phase 3: Terminal UI (Days 6-7)
- [ ] Bubble Tea chat interface integration
- [ ] User list sidebar showing discovered peers
- [ ] Message input and display areas
- [ ] Real-time UI updates from network events

### Phase 4: Polish (Days 8-15)
- [ ] Enhanced error handling and connection reliability
- [ ] Chat commands (/users, /quit, /help, /nick)
- [ ] Color-coded messages and improved UX
- [ ] Comprehensive documentation and demo recordings

---

## Technical Challenges & Solutions

### Challenge 1: Peer Discovery Reliability ✅ SOLVED
**Problem**: Multicast packets can be lost, leading to incomplete peer discovery
**Solution Implemented**: 
- ✅ Periodic beacon broadcasting every 5 seconds
- ✅ Peer timeout mechanism (10s stale, 30s offline with cleanup)
- ✅ Multicast loopback enabled via syscalls for reliable local testing
- ✅ Graceful leave message broadcasting on shutdown
- ✅ Thread-safe peer registry with proper concurrency handling

### Challenge 2: Message Ordering
**Problem**: Messages from different peers may arrive out of order
**Solution**: 
- Add sequence numbers to messages from each peer
- Implement message buffering with timestamp-based ordering
- Use logical clocks (Lamport timestamps) for consistent ordering
- Display messages in chronological order with clear timestamps

### Challenge 3: Network Partitions
**Problem**: Network splits can isolate groups of peers
**Solution**: 
- Implement connection health monitoring with ping/pong
- Attempt reconnection with exponential backoff
- Show clear connection status in UI
- Gracefully handle partial connectivity scenarios

### Challenge 4: UI Responsiveness  
**Problem**: Network operations could block the terminal interface
**Solution Planned**: 
- Use goroutines for all network operations (✅ implemented for discovery)
- Implement non-blocking message channels between network and UI
- Buffer incoming messages to prevent UI lag
- Use Bubble Tea's command system for async operations

**Discovery Implementation**: Already uses proper goroutine separation:
- ✅ Beacon loop goroutine (periodic announcements)
- ✅ Receive loop goroutine (message listening)  
- ✅ Cleanup loop goroutine (peer timeout management)
- ✅ Context-based coordination for clean shutdown

---

## Future Enhancements

### Short-term (if time permits)
- [ ] Message encryption (simple shared key)
- [ ] File transfer capability
- [ ] Multiple chat rooms/channels
- [ ] Message history persistence

### Long-term (portfolio evolution)
- [ ] DHT-based discovery for internet-wide chat
- [ ] End-to-end encryption with key exchange
- [ ] Voice chat integration
- [ ] Mobile client compatibility

---

## Success Metrics

### Technical
- [x] Peers can discover each other within 5 seconds ✅ ACHIEVED
  - Real-time peer discovery working in < 1 second
  - Tested with Alice/Bob discovering each other instantly
- [ ] Messages delivered to all peers < 100ms (Day 3 - TCP messaging)
- [ ] System handles 10+ peers gracefully (Day 3+ - scaling tests)
- [x] Network operations remain non-blocking ✅ ACHIEVED
  - All discovery operations use separate goroutines
  - Context-based coordination prevents blocking

### Portfolio Impact
- [x] Demonstrates P2P networking knowledge ✅ ACHIEVED
  - UDP multicast implementation with syscall optimization
  - Real peer discovery system working across network
- [x] Shows understanding of distributed systems ✅ ACHIEVED
  - Event-driven architecture with peer join/leave handling
  - Concurrent programming with goroutines and channels
- [x] Clean, maintainable Go code ✅ ACHIEVED
  - Professional project structure with proper separation of concerns
  - Thread-safe implementations with proper error handling
- [x] Compelling demo and documentation ✅ ACHIEVED
  - Multi-terminal peer discovery demonstration
  - Professional logging and status reporting
  - Comprehensive architecture documentation

**Day 2 Success**: Peer discovery system fully functional and demonstrated!

---

**Implementation Status**: 
- [x] Peer discovery patterns researched
- [x] P2P protocols analyzed  
- [x] IRC message format studied
- [x] TUI framework chosen
- [x] Architecture decisions finalized
- [x] **Day 2: Peer discovery system fully implemented and tested**

**Current Status**: Phase 1 (Discovery) complete! Multi-peer discovery working with:
- UDP multicast beacon system with syscall optimization
- Thread-safe peer registry with timeout management  
- Real-time peer join/leave event handling
- Professional logging and status reporting
- Verified working across multiple terminal instances

**Next**: Day 3 - TCP chat messaging between discovered peers