# Architecture Guide

This document describes the internal architecture of pipewire-go and how components interact.

## Overview

pipewire-go is a client library for PipeWire audio server with three main layers:

```
┌────────────────────────┐
│   Application Layer (user code)     │
└────────────────────────┘
           ┃
┌────────────────────────┐
│   Client API Layer               │
│   (Client, Node, Port, Link)      │
└────────────────────────┘
           ┃
┌────────────────────────┐
│   Protocol Layer (core package)   │
│   (Connection, Message, Types)    │
└────────────────────────┘
           ┃
┌────────────────────────┐
│   System Layer (Unix socket)      │
│   (PipeWire daemon)                │
└────────────────────────┘
```

## Package Structure

### 1. Client Package (`client/`)

High-level API for users.

#### Key Types

```go
// Client - Main connection to PipeWire
type Client struct {
    connection  *core.Connection
    ctx         context.Context
    nodes       map[uint32]*Node
    ports       map[uint32]*Port
    links       map[uint32]*Link
    listeners   map[EventType][]EventListener
    mu          sync.RWMutex
    // ...
}

// Node - Audio device or application
type Node struct {
    ID          uint32
    Type        string
    Props       map[string]string
    // ...
}

// Port - Connection point on a node
type Port struct {
    ID          uint32
    NodeID      uint32
    direction   PortDirection
    portType    PortType
    formats     []*Format
    // ...
}

// Link - Audio connection between ports
type Link struct {
    ID          uint32
    SourceID    uint32
    DestID      uint32
    properties  map[string]string
    // ...
}
```

#### Responsibilities

1. **Graph Management**
   - Maintain node, port, and link cache
   - Handle graph notifications
   - Provide query/enumeration APIs

2. **Event Handling**
   - Register event listeners
   - Dispatch events to listeners
   - Process async notifications

3. **Connection Lifecycle**
   - NewClient() - establish connection
   - WaitUntilReady() - synchronize initialization
   - Close() - graceful shutdown

### 2. Core Package (`core/`)

Low-level protocol implementation.

#### Key Types

```go
// Connection - Socket communication with daemon
type Connection struct {
    socket      net.Conn
    msgChan     chan *Message
    doneChan    chan struct{}
    mu          sync.Mutex
    // ...
}

// Message - Protocol message
type Message struct {
    Type        MessageType
    ObjectID    uint32
    MethodID    uint32
    Signature   string
    Args        []interface{}
    // ...
}

// Type Definitions
type (
    NodeState     uint32
    PortDirection uint32
    PortType      uint32
    ParamID       uint32
)
```

#### Responsibilities

1. **Socket Communication**
   - Read/write to Unix socket
   - Handle message framing
   - Manage receive loop

2. **Message Processing**
   - Parse protocol messages
   - Marshal/unmarshal data
   - Handle protocol versioning

3. **Type System**
   - Define PipeWire types
   - Provide type IDs
   - Handle type conversions

### 3. SPA Package (`spa/`)

Serialization protocol for PipeWire Object Architecture (POD).

#### Key Concepts

- **POD** - Plain Old Data format for serialization
- **SPA** - Serialization Protocol for Audio

#### Types

- Basic: Int, String, Bytes, Boolean
- Structured: Rectangle, Fraction, Object
- Collections: Arrays, Structs

### 4. Verbose Package (`verbose/`)

Logging utilities for debugging.

#### Log Levels

- `LogLevelError` - Only errors
- `LogLevelWarn` - Warnings and errors
- `LogLevelInfo` - General information
- `LogLevelDebug` - Detailed debugging
- `LogLevelTrace` - Very verbose tracing

## Data Flow

### Connection Establishment

```
1. NewClient(socket, logger)
   │
   ├→ Create Connection
   │   │
   │   ├→ Open Unix socket
   │   ├→ Send version request
   │   ├→ Receive version reply
   │   └→ Start receive loop
   │
   ├→ Create Client struct
   │   ├→ Initialize node/port/link maps
   │   ├→ Create event dispatcher
   │   └→ Start event loop
   │
   ├→ Request initial graph state
   │   ├→ Send GetRegistry request
   │   ├→ Receive object list
   │   └→ Cache all objects
   │
   └→ Return Client ready for use

2. WaitUntilReady(ctx)
   │
   ├→ Check if sync event received
   │   ├→ If yes: Return immediately
   │   └→ If no: Wait for sync event
   │
   └→ Respect context timeout
```

### Message Processing

```
Socket Read Loop (core/connection.go)
   │
   ├→ Read from socket
   │   ├→ Read message size
   │   ├→ Read message data
   │   └→ Unmarshal POD data
   │
   ├→ Parse message fields
   │   ├→ Extract object ID
   │   ├→ Extract method/event ID
   │   └→ Extract parameters
   │
   ├→ Route to handler
   │   ├→ CoreListener for core events
   │   ├→ Object listener if registered
   │   └→ Event channel for client
   │
   └→ Continue reading
```

### Event Dispatch

```
Event Loop (client/client.go)
   │
   ├→ Receive event from channel
   │
   ├→ Process event
   │   ├→ Update cache (nodes/ports/links)
   │   ├→ Emit typed event
   │   └→ Call user listeners
   │
   ├→ Handle user operations
   │   ├→ CreateLink
   │   ├→ SetParameter
   │   └→ etc.
   │
   └→ Continue loop until closed
```

## Synchronization

### Thread Safety Model

1. **Client-level RWMutex**
   - Protects node/port/link maps
   - Read-lock for queries
   - Write-lock for updates

2. **Connection-level Mutex**
   - Protects socket writes
   - Prevents message interleaving

3. **Context Passing**
   - Propagate cancellation
   - Enforce timeouts

### Concurrency Example

```go
// Multiple goroutines can safely read
go func() {
    nodes := conn.GetNodes()  // Read-lock acquired
    // Use nodes
}()

// Events update cache (write-lock)
// User operations (read-lock)
```

## Event System

### Event Types

```go
const (
    EventTypeNodeAdded EventType = iota
    EventTypeNodeRemoved
    EventTypeNodeUpdated
    EventTypePortAdded
    EventTypePortRemoved
    EventTypePortUpdated
    EventTypeLinkAdded
    EventTypeLinkRemoved
    // ...
)
```

### Event Flow

```
PipeWire Daemon
   │
   ├→ Send node added event
   │
   ├→ Connection.Read()
   │   ├→ Parse message
   │   └→ Send to eventChan
   │
   ├→ Client.eventLoop()
   │   ├→ Receive from eventChan
   │   ├→ Update cache
   │   ├→ Emit typed event
   │   └→ Call listeners
   │
   └→ User Listener (NodeAdded)
      │
      └→ Process node addition
```

## Protocol Design

### Message Format

```
┌────────────────────┐
│ Header                          │
│ ├─── Type: u32              │
│ ├─── ObjectID: u32          │
│ ├─── OpID: u32              │
│ └─── Size: u32              │
├────────────────────┤
│ Body (POD-encoded)             │
│ ├─── Parameter 1           │
│ ├─── Parameter 2           │
│ └─── ...                    │
└────────────────────┘
```

### Handshake

```
Client
   │
   ├→ Send: Hello message
   │   ├─ Version: protocol_version
   │   ├─ Client name
   │   └─ Properties
   │
   ├→ Receive: Hello reply
   │   ├─ Server version
   │   ├─ Server ID
   │   └─ Properties
   │
   └→ Connection established
```

## Performance Considerations

### Caching Strategy

1. **Node/Port/Link Cache**
   - Keep in-memory copy of graph
   - Update on events
   - Avoids repeated queries

2. **Property Caching**
   - Cache node/port properties
   - Reduce protocol messages
   - Invalidate on updates

### Memory Usage

```
Per Node:    ~10KB (name, properties, state)
Per Port:    ~2KB  (direction, type, formats)
Per Link:    ~1KB  (source/dest IDs)
Overhead:    ~1MB  (socket, buffers, listeners)
```

### Optimization Techniques

1. **Batch Operations**
   - Multiple link creations
   - Parameter updates

2. **Event Filtering**
   - Only listen to needed events
   - Reduce callback overhead

3. **Goroutine Management**
   - Offload heavy work
   - Use channels for async operations

## Extension Points

### Adding New Event Types

1. Define EventType in `client/types.go`
2. Add handler in `Connection`
3. Emit in `Client.eventLoop()`
4. Document in README

### Adding New Methods

1. Add protocol message in `core/`
2. Implement in `Client` or object type
3. Add tests in `*_test.go`
4. Document in GoDoc

### Custom Listeners

```go
type CustomListener struct {
    Name string
    // ...
}

func (cl *CustomListener) OnEvent(e Event) error {
    // Custom handling
    return nil
}
```

## Debugging

### Trace Points

```go
// Enable verbose logging
logger := verbose.NewLogger(verbose.LogLevelDebug, true)

// Key trace points:
// - Connection establishment
// - Message send/receive
// - Event dispatch
// - Cache updates
```

### Common Debugging Patterns

```go
// Check graph state
fmt.Printf("Nodes: %d\n", len(conn.GetNodes()))
fmt.Printf("Ports: %d\n", len(conn.GetPorts()))
fmt.Printf("Links: %d\n", len(conn.GetLinks()))

// Trace events
conn.RegisterEventListener(client.EventTypeNodeAdded, func(e client.Event) error {
    logger.Infof("Node added: %v", e.Data)
    return nil
})
```

## Future Enhancements

1. **Connection Pooling** - Multiple concurrent operations
2. **Caching Strategies** - Configurable cache behavior
3. **Plugin System** - Custom event processors
4. **Metrics** - Performance monitoring
5. **Retry Logic** - Automatic reconnection

---

**Last Updated:** January 7, 2026
