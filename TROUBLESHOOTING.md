# Troubleshooting Guide

This guide helps resolve common issues when using pipewire-go.

## Connection Issues

### "Connection refused" error

**Symptoms:**
```
Error: Could not connect to PipeWire daemon
```

**Causes:**
1. PipeWire daemon is not running
2. Wrong socket path
3. Permissions issue
4. Using remote/SSH connection without X11 forwarding

**Solutions:**

1. **Check if PipeWire is running:**
```bash
pips pw-cli info core
# or
systemctl --user status pipewire
```

2. **Find the correct socket path:**
```bash
# Default socket
echo $XDG_RUNTIME_DIR/pipewire-0
# Usually: /run/user/1000/pipewire-0

# List available sockets
ls -la $XDG_RUNTIME_DIR/pipewire*
```

3. **Check permissions:**
```bash
# Socket should be readable and writable
ls -la /run/user/1000/pipewire-0
# Should show: srw-rw---- your_user your_user

# Add yourself to audio group if needed
sudo usermod -a -G audio $USER
```

4. **Start PipeWire manually:**
```bash
# Start the daemon
pipewire &

# Or as service
systemctl --user start pipewire
systemctl --user enable pipewire
```

### Connection hangs/timeout

**Symptoms:**
```
Program hangs when calling NewClient()
```

**Causes:**
1. PipeWire daemon is hung
2. Socket exists but daemon not responding
3. Networking issues (remote connection)

**Solutions:**

1. **Restart PipeWire:**
```bash
systemctl --user restart pipewire
```

2. **Force kill and restart:**
```bash
killall -9 pipewire
killall -9 pipewire-pulse
sleep 1
systemctl --user start pipewire
```

3. **Check daemon logs:**
```bash
# View recent logs
journalctl --user -u pipewire -n 50

# Real-time logs
journalctl --user -u pipewire -f
```

## Audio Not Working

### "No nodes found"

**Symptoms:**
```
List shows 0 nodes
```

**Causes:**
1. No audio devices configured
2. Device drivers not loaded
3. PipeWire not fully initialized

**Solutions:**

1. **Wait for initialization:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := conn.WaitUntilReady(ctx); err != nil {
    log.Fatal("PipeWire not ready", err)
}
// Then list nodes
```

2. **Load audio device:**
```bash
# Check available devices
aplay -L
alsamixer

# Start PulseAudio compatibility layer
pipewire-pulse &
```

3. **Verify device drivers:**
```bash
# List ALSA devices
cat /proc/asound/devices

# Check kernel logs
dmesg | grep -i audio
```

### "Cannot create link"

**Symptoms:**
```
Error: Failed to create link between ports
```

**Causes:**
1. Incompatible port formats
2. Port already connected
3. Invalid port specification
4. Permissions issue

**Solutions:**

1. **Check port formats:**
```go
supported := port.GetSupportedFormats()
if len(supported) == 0 {
    fmt.Println("Port has no supported formats")
}

// Check compatibility
if !source.CanConnectTo(sink) {
    fmt.Println("Ports incompatible")
}
```

2. **Check port type:**
```go
// Must match direction
if source.IsOutput() && sink.IsInput() {
    // OK to connect
}
if source.Type() != sink.Type() {
    // Mismatch (e.g., Audio to MIDI)
}
```

3. **Use example to debug:**
```bash
# Show all ports
go run examples/list_nodes.go -verbose

# Try creating link
go run examples/create_link.go -source "Source" -sink "Sink"
```

## Performance Issues

### "Slow event processing"

**Symptoms:**
```
High CPU usage
Delayed event handling
```

**Causes:**
1. Event listener doing heavy work
2. Allocating memory in event handler
3. Mutex contention
4. Too many event listeners

**Solutions:**

1. **Offload work from event handler:**
```go
// BAD - doing heavy work in handler
conn.RegisterEventListener(EventTypeNodeAdded, func(e Event) error {
    // DON'T: Heavy computation
    complexAnalysis(e)
    return nil
})

// GOOD - queue for processing
queue := make(chan Event, 100)
conn.RegisterEventListener(EventTypeNodeAdded, func(e Event) error {
    queue <- e  // Non-blocking queue
    return nil
})

// Process in separate goroutine
go func() {
    for e := range queue {
        complexAnalysis(e)
    }
}()
```

2. **Reduce listener count:**
```go
// AVOID: Too many listeners
for eventType := range allEventTypes {
    conn.RegisterEventListener(eventType, handler)
}

// GOOD: Filter at source
for _, eventType := range importantEvents {
    conn.RegisterEventListener(eventType, handler)
}
```

3. **Use context for cleanup:**
```go
ctx, cancel := context.WithCancel(context.Background())

// Register listener
listenerID := conn.RegisterEventListener(...)

// Cleanup when done
defer cancel()
defer conn.UnregisterEventListener(listenerID)
```

## Memory Leaks

### "Memory usage grows over time"

**Causes:**
1. Event listeners not unregistered
2. Goroutines not exiting
3. Cached data not cleared

**Solutions:**

1. **Always cleanup listeners:**
```go
// Register
id := conn.RegisterEventListener(eventType, handler)

// Cleanup when done
defer conn.UnregisterEventListener(id)
```

2. **Cancel contexts:**
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // Always cancel

// Pass context to blocking operations
if err := conn.WaitUntilReady(ctx); err != nil {
    return err
}
```

3. **Close connection:**
```go
conn, _ := client.NewClient(socket, logger)
defer conn.Close()  // Always close
```

## Concurrency Issues

### "Panic: concurrent map access"

**Cause:**
Concurrent access to shared data without proper synchronization

**Solution:**
The library is thread-safe, but your code must be:

```go
// WRONG - race condition
var results []*Node
for _, node := range conn.GetNodes() {
    go func() {
        // Accessing node concurrently without synchronization
        results = append(results, node)  // RACE!
    }()
}

// RIGHT - synchronized access
var mu sync.Mutex
var results []*Node
for _, node := range conn.GetNodes() {
    go func(n *Node) {
        mu.Lock()
        results = append(results, n)
        mu.Unlock()
    }(node)
}
```

### "Deadlock"

**Causes:**
1. Holding locks while waiting for library calls
2. Library calls trying to acquire same locks

**Solution:**
```go
// WRONG - deadlock potential
var mu sync.Mutex
mu.Lock()
// Library call that needs mutex internally
node := conn.GetNodeByID(1)  // DEADLOCK!
mu.Unlock()

// RIGHT - don't hold locks during library calls
var mu sync.Mutex
node := conn.GetNodeByID(1)
mu.Lock()
// Use node safely
mu.Unlock()
```

## Error Handling

### Type-Safe Error Checking

```go
link, err := conn.CreateLink(source, sink, nil)
if err != nil {
    // Check error type
    switch err.(type) {
    case *client.ValidationError:
        fmt.Println("Invalid parameters")
    case *client.ProtocolError:
        fmt.Println("Communication failed")
    case *client.IncompatibleError:
        fmt.Println("Ports incompatible")
    case *client.TimeoutError:
        fmt.Println("Operation timed out")
    case *client.NotFoundError:
        fmt.Println("Resource not found")
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

### Error Wrapping and Context

```go
if err := node.SetParam(paramID, flags, value); err != nil {
    return fmt.Errorf("failed to set node parameter %d: %w", paramID, err)
}
```

## Best Practices

### Always Use Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := conn.WaitUntilReady(ctx); err != nil {
    log.Fatal("Connection not ready")
}
```

### Check Before Using

```go
ports := node.GetPorts()
if len(ports) == 0 {
    fmt.Println("No ports available")
    return
}

supported := port.GetSupportedFormats()
if len(supported) == 0 {
    fmt.Println("Port has no supported formats")
    return
}
```

### Graceful Degradation

```go
conn, err := client.NewClient(socket, logger)
if err != nil {
    if os.Getenv("REQUIRE_PIPEWIRE") == "1" {
        log.Fatal("PipeWire not available")
    }
    // Continue with fallback behavior
    fmt.Println("PipeWire not available, using fallback")
    return nil
}
defer conn.Close()
```

### Logging for Debugging

```go
import "github.com/vignemail1/pipewire-go/verbose"

logger := verbose.NewLogger(verbose.LogLevelDebug, true)
conn, err := client.NewClient(socket, logger)
// Enable detailed logging for debugging
```

## Debugging Tips

### Enable Verbose Logging

```bash
# Run examples with verbose output
go run examples/list_nodes.go -verbose
go run examples/monitor_graph.go -verbose
```

### Use PipeWire Tools

```bash
# List nodes
pw-dump /etc/pipewire/wireplumber.conf

# Monitor graph
pw-mon

# List devices
pactl list short sinks
pactl list short sources
```

### Check System Logs

```bash
# PipeWire logs
journalctl --user -u pipewire -f

# System audio logs
dmesg | tail -20
journalctl -f | grep -i audio
```

### Reproduce with Example

When reporting issues, try:

```bash
# Clear reproduction
go run examples/list_nodes.go -verbose

# Check if it's library or user code
go test -v ./...
```

## Getting More Help

1. Check [examples](./examples/)
2. Read [API documentation](./README.md#api-documentation)
3. Review [issue tracker](https://github.com/vignemail1/pipewire-go/issues)
4. Check [PipeWire docs](https://pipewire.org/)
5. Create [new issue](https://github.com/vignemail1/pipewire-go/issues/new) with details

---

**Last Updated:** January 7, 2026
