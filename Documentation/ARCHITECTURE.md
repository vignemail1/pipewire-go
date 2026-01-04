# PipeWire Go Library - Architecture & Design

## Vue d'ensemble

Cette librairie implémente un client PipeWire entièrement en Go sans CGO, permettant l'accès complet aux capacités de la socket Unix PipeWire pour des applications d'audio routing, monitoring et contrôle.

### Principes de Design

1. **Zero CGO** - Aucune dépendance C, binaires statiquement compilables
2. **Pure Socket Communication** - Communication directe via socket Unix avec protocole natif
3. **Layered Architecture** - Trois couches : protocol (bas niveau) → client API (haut niveau) → user code
4. **Verbose by Design** - Mode debug détaillé à chaque niveau
5. **Type Safe** - Typage fort avec interfaces pour extensibilité

---

## Architecture Détaillée

### Couche 1 : Protocol (Package `core` et `spa`)

#### Responsabilités
- Communication bas niveau avec socket Unix
- Marshalling/unmarshalling POD
- Gestion des messages asynchrones
- Dispatch des événements

#### Composants

**`spa/pod.go`** - Format SPA/POD
```
POD = Plain Old Data, format binaire compact

Tous les messages PipeWire utilisent POD pour le payload.

Structure POD:
┌──────────────┬──────────────┬──────────────────────────┐
│ Size (4B)    │ Type (4B)    │ Payload (size-8 bytes)   │
│ uint32 LE    │ uint32 LE    │ padded to 8 bytes        │
└──────────────┴──────────────┴──────────────────────────┘

Types supportés:
- Primitifs: Bool, Id, Int, Long, Float, Double, String, Bytes
- Conteneurs: Struct, Object, Array, Choice
- Spécialisés: Rectangle, Fraction, Pointer, Fd, Sequence

Parser pattern:
  pod, err := parser.ParsePOD()  // Retourne PODValue interface
  if intVal, ok := pod.(*IntPOD); ok { ... }

Builder pattern:
  builder := NewPODBuilder()
  builder.WriteInt(42)
  builder.WriteString("test")
  binary := builder.Bytes()
```

**`core/connection.go`** - Gestion Socket

```
Connection = liaison socket + read/write loops + event dispatch

Architecture async:
┌─────────────────────────────────────┐
│ socket.Read()                       │
│   ↓ (read loop goroutine)           │
│ readBuf.Write()                     │
│   ↓                                 │
│ parseMessage() → Message            │
│   ↓                                 │
│ dispatchEvent() → handlers          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ user: SendMessage()                 │
│   ↓                                 │
│ marshallMessage() → binary           │
│   ↓                                 │
│ writeChan <- data                   │
│   ↓ (write loop goroutine)          │
│ socket.Write()                      │
└─────────────────────────────────────┘

Thread-safety: mutex pour tous les partagés (messageMap, eventHandlers)
```

### Couche 2 : Client API (Package `client`)

#### Responsabilités
- Proxies pour les objets distants (Core, Registry, Node, Port, Link)
- Découverte d'objets via Registry
- Cache et synchronisation d'état
- Callbacks pour state changes

#### Object Model

```
PipeWire Object Graph:

┌─────────────────────────────────────────────────────┐
│ Core (id=0)                                         │
│  └─ Registry (global object list)                   │
│      ├─ Node[1] (Playback device)                   │
│      │   ├─ Port[5] (FL)                            │
│      │   ├─ Port[6] (FR)                            │
│      │   └─ ...                                     │
│      ├─ Node[2] (Capture device)                    │
│      │   ├─ Port[10] (Mic)                          │
│      │   └─ ...                                     │
│      └─ Link[20] (Node[1].Port[5] → Node[2].Port[10])
└─────────────────────────────────────────────────────┘

Chaque objet a:
- ID unique (uint32)
- Type interface (Node, Port, Link, etc.)
- Version de l'interface
- Propriétés (map[string]string)
- Listener pour changements
```

#### Proxies

Un proxy représente un objet serveur:

```go
type Core struct {
    id      uint32           // Always 0
    conn    *core.Connection // Reference to socket
    props   map[string]string
    version uint32
}

type Node struct {
    ID      uint32
    Type    string
    Props   map[string]string
    Ports   []*Port
    Links   []*Link
    conn    *core.Connection
}

type Port struct {
    ID         uint32
    Name       string
    Direction  PortDirection // Input/Output
    Type       PortType      // Audio/Midi/Video
    Props      map[string]string
    ParentNode *Node
    conn       *core.Connection
}

type Link struct {
    ID        uint32
    Output    *Port
    Input     *Port
    Props     map[string]string
    conn      *core.Connection
}
```

Les proxies:
- Cach localement l'état de l'objet distant
- Synchronisent via listeners sur événements serveur
- Permettent des opérations (SetParam, Remove, etc.)

### Couche 3 : User Code

L'utilisateur interagit principalement avec:

```go
// Connexion principal
client, _ := client.NewClient("/run/pipewire-0")

// Découverte
nodes := client.GetRegistry().ListNodes()

// Operations
link, _ := client.CreateLink(outPort, inPort, nil)

// Monitoring
client.OnNodeAdded(func(n *Node) {
    fmt.Printf("Node: %s\n", n.Name())
})
```

---

## Protocol Flow - Initialisation

Quand le client se connecte:

```
Time │ Client                          Server
     │
  1  │─ connect unix socket ─────────→│ accept
     │
  2  │─ core.hello() ──────────────→│ handshake
     │  {version, core_id, client_id}  │
     │
  3  │←─────────── core.welcome() ──│ ok
     │  {version, server_id, props}  │
     │
  4  │─ core.get_registry() ────────→│ bind registry
     │
  5  │←───── registry.global events ──│ all objects
     │  node 5: "Playback", v3        │
     │  node 6: "Capture", v3         │
     │  port 10: "Speaker FL", v3     │
     │  ...                           │
     │
  6  │─ registry.bind() ──────────→│ proxy node
     │
  7  │←─── node.param / node.info ──│ node data
     │
     │ [Ready to use]
```

---

## Design Patterns

### 1. Proxy Pattern

Chaque objet distant a un proxy local:

```go
// Proxy Node encapsule object distant id=123
node := client.GetNode(123)

// Appel méthode:
node.SetParam(paramID, value)
  ↓
conn.SendMessage(node.ID, methodID, marshalledValue)
  ↓
server processes message
  ↓
server sends event back
  ↓
handleEvent updates node.Props

// Tout transparent pour l'user
```

### 2. Observer Pattern

État synchronisé via listeners:

```go
// User register callback
client.OnNodeAdded(func(n *Node) { ... })

// Server sends event
conn.RegisterEventHandler(registryID, func(msg *core.Message) {
    if msg.OpCode == globalAddEvent {
        newNode := parseNode(msg.Payload)
        client.nodeAddedCallbacks.call(newNode)
    }
})
```

### 3. Builder Pattern

Construction de messages:

```go
// Low level
builder := spa.NewPODBuilder()
builder.WriteInt(42)
binary := builder.Bytes()
conn.SendMessage(objID, methodID, binary)

// High level
node.SetParam(paramID, &AudioFormat{...})
  ↓ internally uses POD builder
```

---

## Threading Model

### Single-threaded vs Concurrent

**Connection goroutines (internal):**
```
main goroutine                Read goroutine
    │                              │
    ├─ SendMessage() ──────→ writeChan ──→ socket.Write()
    │                              │
    └─ RegisterHandler()   socket.Read() → readBuf
                                   │
                            parseMessage()
                                   │
                            dispatchEvent()
                                   │
                          user's callback
```

**Critical sections (mutex-protected):**
- messageMap (pending requests)
- eventHandlers (listener registry)
- POD parser state

**Client state (thread-safe operations):**
```go
// User can call from multiple goroutines:
go func() { client.GetNode(123) }()
go func() { client.GetNode(456) }()

// Internally protected by mutex
```

---

## Error Handling

### Erreurs possibles

1. **Connection errors**
   - Socket not found
   - Permission denied
   - Daemon not running

2. **Protocol errors**
   - Invalid POD format
   - Unexpected message type
   - Malformed object

3. **Object errors**
   - Object not found
   - Operation not supported
   - Permission denied

### Strategy

```go
if err := conn.SendMessage(...); err != nil {
    switch err {
    case ErrConnectionClosed:
        // Reconnect?
    case ErrInvalidPOD:
        // Protocol error, log and continue
    case ErrTimeout:
        // Server not responding
    }
}
```

---

## Verbose Logging - Architecture

Mode debug pour chaque couche:

```
Layer: core/connection.go
  Logger.Debugf("Sending message: ObjectID=%d, OpCode=%d")
  Logger.DumpBinary("Raw socket data", bytes)

Layer: spa/pod.go
  Logger.Debugf("Parsing POD type=%d", podType)
  Logger.DumpBinary("POD payload", payload)

Layer: client/
  Logger.Debugf("Node added: %s (id=%d)", name, id)
  Dumper.DumpObject(id, type, props)

User:
  logger.SetLevel(LogLevelDebug)
  // All detail visible
```

**Timers:**
```
[15:04:05.234] [DEBUG] Sending message...
[15:04:05.235] [DEBUG] Received event...
Response time: 1ms visible from timestamps
```

---

## Performance Considerations

### Memory
- POD parser: streaming, zero-copy où possible
- Connection: single read buffer, reused
- Caches: LRU for large object lists (future)

### CPU
- Async I/O: non-blocking socket reads
- No polling: event-driven via select/epoll implicit in net.Conn
- Minimal allocations in hot paths

### Latency
- Direct socket communication: microsecond order
- No IPC marshalling overhead: POD is binary-efficient
- Event dispatch: immediate callback invocation

---

## Extension Points

### Pour implémenter support d'objets supplémentaires:

1. **Créer nouveau proxy dans `client/`**
   ```go
   type NewObject struct {
       ID   uint32
       conn *core.Connection
   }
   ```

2. **Implémenter méthodes/événements**
   ```go
   func (o *NewObject) DoSomething(arg Type) error {
       return o.conn.SendMessage(o.ID, methodID, pod)
   }
   
   func (o *NewObject) onEvent(msg *core.Message) error {
       // Handle event
   }
   ```

3. **Ajouter à client registry**
   ```go
   client.newObjects[id] = newObj
   ```

---

## Testing Strategy

### Unit Tests

```
spa/pod_test.go
  - ParsePOD with all types
  - BuildPOD with all types
  - Round-trip parsing

core/connection_test.go
  - Mock socket reading/writing
  - Message parsing
  - Event dispatching

client/node_test.go
  - Proxy creation
  - State synchronization
```

### Integration Tests

```
Integration with real daemon:
  1. Start pipewire daemon
  2. Connect and verify handshake
  3. List actual nodes/ports
  4. Create/remove links
  5. Verify changes via alternate client (pw-cli)
```

### Examples

```
All examples must:
  - Compile with CGO_ENABLED=0
  - Run against real daemon
  - Show expected output
  - Be documentable
```

---

## Future Enhancements

1. **Session Manager Interface**
   - Endpoint support
   - Session control

2. **Performance API**
   - CPU usage monitoring
   - Latency measurement

3. **Advanced Routing**
   - Parameter curves
   - Conditional linking

4. **Networking**
   - Remote PipeWire over TCP
   - Network bridges

5. **Utilities**
   - pw-cli equivalent
   - pw-top equivalent
   - Audio analysis tools
