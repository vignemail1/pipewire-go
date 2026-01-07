# PipeWire-Go API Documentation

Comprehensive API reference and usage guide for the pipewire-go library.

## Table of Contents

1. [Client API](#client-api)
2. [Registry API](#registry-api)
3. [Objects API](#objects-api)
4. [Error Handling](#error-handling)
5. [Examples](#examples)
6. [Patterns](#patterns)

---

## Client API

### NewClient

Create a new PipeWire client connection.

```go
func NewClient(name string) (*Client, error)
```

**Parameters:**
- `name` - Application name (displayed in PipeWire)

**Returns:**
- `*Client` - Connected client
- `error` - Connection error

**Example:**
```go
client, err := client.NewClient("my-audio-app")
if err != nil {
    return err
}
defer client.Disconnect()
```

**Errors:**
- `ConnectionError` - If PipeWire daemon is not running
- `ValidationError` - If name is invalid

---

### Client.Disconnect

Close the connection to PipeWire.

```go
func (c *Client) Disconnect() error
```

**Returns:**
- `error` - Disconnection error

**Example:**
```go
defer client.Disconnect()
```

---

### Client.GetRegistry

Get the registry for accessing PipeWire objects.

```go
func (c *Client) GetRegistry() *Registry
```

**Returns:**
- `*Registry` - The registry object

**Example:**
```go
registry := client.GetRegistry()
nodes := registry.ListNodes()
```

---

## Registry API

### Registry.ListAll

List all objects in the registry.

```go
func (r *Registry) ListAll() []*GlobalObject
```

**Returns:**
- `[]*GlobalObject` - All objects (copy)

**Note:** Returns a copy of the registry - use carefully for large graphs.

**Example:**
```go
allObjects := registry.ListAll()
fmt.Printf("Total objects: %d\n", len(allObjects))
```

---

### Registry.GetObject

Get a specific object by ID.

```go
func (r *Registry) GetObject(id uint32) *GlobalObject
```

**Parameters:**
- `id` - Object ID

**Returns:**
- `*GlobalObject` - The object, or nil if not found

**Example:**
```go
node := registry.GetObject(42)
if node != nil {
    fmt.Printf("Found: %s\n", node.Type)
}
```

---

### Registry.ListByType

List objects of a specific type.

```go
func (r *Registry) ListByType(objType string) []*GlobalObject
```

**Parameters:**
- `objType` - Object type ("Node", "Port", "Link", etc.)

**Returns:**
- `[]*GlobalObject` - Filtered objects

**Example:**
```go
nodes := registry.ListByType("Node")
ports := registry.ListByType("Port")
links := registry.ListByType("Link")
```

---

### Registry.ListNodes / ListPorts / ListLinks

Convenience methods to list specific object types.

```go
func (r *Registry) ListNodes() []*GlobalObject
func (r *Registry) ListPorts() []*GlobalObject
func (r *Registry) ListLinks() []*GlobalObject
```

**Example:**
```go
nodes := registry.ListNodes()
for _, node := range nodes {
    fmt.Printf("Node: %s (ID: %d)\n", node.Name, node.ID)
}
```

---

### Registry.OnGlobalAdded

Register a listener for new objects.

```go
func (r *Registry) OnGlobalAdded(listener RegistryListener)
```

**Parameters:**
- `listener` - Callback function(obj *GlobalObject)

**Example:**
```go
registry.OnGlobalAdded(func(obj *GlobalObject) {
    fmt.Printf("New object: %s (ID: %d)\n", obj.Type, obj.ID)
})
```

---

### Registry.CountObjects / Count

Count objects in registry.

```go
func (r *Registry) CountObjects() int
func (r *Registry) Count(objType string) int
```

**Example:**
```go
total := registry.CountObjects()
nodes := registry.Count("Node")
ports := registry.Count("Port")
```

---

## Objects API

### GlobalObject

Represents a PipeWire object.

```go
type GlobalObject struct {
    ID      uint32
    Type    string
    Version uint32
    Props   map[string]string
}
```

**Fields:**
- `ID` - Unique object identifier
- `Type` - Object type ("Node", "Port", "Link", etc.)
- `Version` - Interface version
- `Props` - Object properties (key-value pairs)

**Example:**
```go
node := registry.GetObject(42)
if node != nil {
    fmt.Printf("Type: %s\n", node.Type)
    fmt.Printf("Props: %v\n", node.Props)
}
```

---

## Error Handling

### Using error.Is

Check for specific error types:

```go
err := client.Connect()
if errors.Is(err, pipewireio.ErrNotConnected) {
    log.Fatal("PipeWire daemon not running")
}
```

### Using error.As

Extract error details:

```go
err := someOperation()
var valErr *pipewireio.ValidationError
if errors.As(err, &valErr) {
    log.Printf("Validation error in field %s: %s", valErr.Field, valErr.Message)
}
```

### Available Error Types

- `ConnectionError` - Connection failures
- `ValidationError` - Invalid input
- `ProtocolError` - Protocol violations
- `TimeoutError` - Operation timeouts
- `FormatNegotiationError` - Format negotiation failures
- `StateError` - Invalid state transitions
- `ParameterError` - Parameter operation failures
- `NotFoundError` - Resource not found
- `PermissionError` - Permission denials

---

## Examples

### Basic Connection

```go
client, err := client.NewClient("my-app")
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer client.Disconnect()

registry := client.GetRegistry()
fmt.Printf("Connected! Total objects: %d\n", registry.CountObjects())
```

### Enumerate Graph

```go
registry := client.GetRegistry()

fmt.Println("=== Nodes ===")
for _, node := range registry.ListNodes() {
    fmt.Printf("%d: %s\n", node.ID, node.Props["node.name"])
}

fmt.Println("\n=== Ports ===")
for _, port := range registry.ListPorts() {
    fmt.Printf("%d: %s\n", port.ID, port.Props["port.name"])
}

fmt.Println("\n=== Links ===")
for _, link := range registry.ListLinks() {
    fmt.Printf("%d: %s -> %s\n", link.ID,
        link.Props["link.output.port"],
        link.Props["link.input.port"])
}
```

### Monitor Graph Changes

```go
registry := client.GetRegistry()

registry.OnGlobalAdded(func(obj *GlobalObject) {
    fmt.Printf("[NEW] %s #%d\n", obj.Type, obj.ID)
})

// Keep running to receive events
select {}
```

---

## Patterns

### Safe Registry Access

```go
// Get and validate
node := registry.GetObject(nodeID)
if node == nil {
    return errors.New("node not found")
}

// Use node
name := node.Props["node.name"]
```

### Event Listener

```go
var nodeCount int
registry.OnGlobalAdded(func(obj *GlobalObject) {
    if obj.Type == "Node" {
        nodeCount++
    }
})
```

### Type-Safe Error Handling

```go
func handleError(err error) {
    switch err := err.(type) {
    case *pipewireio.ValidationError:
        log.Printf("Validation: field=%s msg=%s", err.Field, err.Message)
    case *pipewireio.ConnectionError:
        log.Printf("Connection: %s", err.Reason)
    default:
        log.Printf("Error: %v", err)
    }
}
```

---

## Performance Tips

1. **Cache Registry References**: Don't call `GetRegistry()` repeatedly
2. **Use Typed Filters**: Use `ListNodes()` instead of `ListByType("Node")`
3. **Avoid Full Copies**: Use `GetObject()` for single items, not `ListAll()`
4. **Fast Event Handlers**: Keep `OnGlobalAdded` callbacks fast
5. **Concurrent Access**: Safe from multiple goroutines via RWMutex

---

For more examples, see the `examples/` directory.
