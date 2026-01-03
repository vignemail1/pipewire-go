# PipeWire Go Library - Guide d'Implémentation Complet

## Phase 1 : Fondations (COMPLÉTÉ)

Les modules essentiels ont été créés :

### 1. **spa/pod.go** - Parseur/Builder SPA/POD
- ✅ Implémentation complète du format binaire POD
- ✅ Support de tous les types primitifs (Bool, Int, Long, Float, Double, String, Bytes)
- ✅ Support des types conteneurs (Struct, Object, Array, Choice)
- ✅ Parsing avec gestion d'erreurs robuste
- ✅ Construction de POD avec builder pattern
- ✅ Alignement 8-bytes automatique
- ✅ Zéro allocations excessives

**Utilisation :**
```go
// Parser
parser := spa.NewPODParser(binaryData)
pod, err := parser.ParsePOD()

// Builder
builder := spa.NewPODBuilder()
builder.WriteInt(42)
builder.WriteString("test")
binary := builder.Bytes()
```

### 2. **core/connection.go** - Gestion Socket Unix
- ✅ Connexion Unix domain socket
- ✅ Auto-découverte du socket PipeWire
- ✅ Lecture/écriture asynchrone avec goroutines
- ✅ Parsing des messages avec buffering intelligent
- ✅ Gestion des événements avec callback registry
- ✅ Synchronisation thread-safe avec mutex
- ✅ Timeouts de lecture configurables

**Utilisation :**
```go
conn, err := core.Dial("/run/pipewire-0", logger)
defer conn.Close()

// Envoyer un message
conn.SendMessage(objectID, methodID, podPayload)

// Écouter les événements
conn.RegisterEventHandler(objectID, func(msg *core.Message) error {
    // Traiter le message
    return nil
})
```

### 3. **verbose/logger.go** - Logging Verbose
- ✅ Niveaux de verbosité (Error, Warn, Info, Debug)
- ✅ Timestamps optionnels
- ✅ Dump hexadécimal avec ASCII
- ✅ MessageDumper pour logs structurés
- ✅ Callbacks pour intégration custom
- ✅ Thread-safe

**Utilisation :**
```go
logger := verbose.NewLogger(verbose.LogLevelDebug, true)
logger.Debugf("Message: %v", value)
logger.DumpBinary("data", binaryContent)

dumper := verbose.NewMessageDumper(logger)
dumper.DumpMethodCall(0, "ping", 0, args)
```

---

## Phase 2 : Client API Haute Niveau (À FAIRE)

### Structure à implémenter

#### **client/client.go** - Point d'entrée principal
```go
type Client struct {
    conn     *core.Connection
    logger   *verbose.Logger
    
    // Caches d'objets
    core     *Core
    client   *ClientProxy
    registry *Registry
    
    // Tables d'objets connus
    nodes    map[uint32]*Node
    ports    map[uint32]*Port
    links    map[uint32]*Link
    devices  map[uint32]*Device
}

// Connexion et gestion de cycle de vie
func NewClient(socketPath string, opts ...Option) (*Client, error)
func (c *Client) Close() error
func (c *Client) IsConnected() bool

// Découverte d'objets
func (c *Client) GetCore() *Core
func (c *Client) GetRegistry() *Registry
func (c *Client) ListObjects(filter string) []*GlobalObject
func (c *Client) GetNodeByID(id uint32) *Node
func (c *Client) GetNodeByName(name string) *Node
func (c *Client) GetPort(id uint32) *Port
func (c *Client) GetLink(id uint32) *Link

// Opérations d'audio routing
func (c *Client) CreateLink(output *Port, input *Port, props map[string]string) (*Link, error)
func (c *Client) RemoveLink(link *Link) error
func (c *Client) ChangeRoute(from *Node, to *Node) error

// Event listeners
func (c *Client) OnNodeAdded(callback NodeCallback)
func (c *Client) OnNodeRemoved(callback NodeCallback)
func (c *Client) OnPortAdded(callback PortCallback)
func (c *Client) OnLinkAdded(callback LinkCallback)
func (c *Client) OnLinkRemoved(callback LinkCallback)
```

#### **client/core.go** - Proxy Core
```go
type Core struct {
    id      uint32
    conn    *core.Connection
    props   map[string]string
    version uint32
}

// Opérations synchrones
func (c *Core) Ping() error
func (c *Core) GetProperties() map[string]string
func (c *Core) GetVersion() string
```

#### **client/registry.go** - Registry Global
```go
type Registry struct {
    id       uint32
    conn     *core.Connection
    
    // Cache des objets globaux
    objects  map[uint32]*GlobalObject
    
    // Listeners
    listeners []RegistryListener
}

type GlobalObject struct {
    ID          uint32
    Type        string        // "Node", "Port", "Link", etc.
    Version     uint32
    Properties  map[string]string
}

// Opérations
func (r *Registry) ListAll() []*GlobalObject
func (r *Registry) BindObject(id uint32, iface string) (interface{}, error)
func (r *Registry) OnAdded(callback func(*GlobalObject))
func (r *Registry) OnRemoved(callback func(uint32))
```

#### **client/node.go** - Proxy Node
```go
type Node struct {
    ID          uint32
    Type        string
    Props       map[string]string
    Ports       []*Port
    Links       []*Link
    
    conn        *core.Connection
}

// État et propriétés
func (n *Node) Name() string
func (n *Node) Direction() NodeDirection // Playback/Capture/Duplex
func (n *Node) State() NodeState        // Suspended/Idle/Running/Error
func (n *Node) SampleRate() uint32
func (n *Node) Channels() uint32
func (n *Node) ClassID() string

// Gestion des ports
func (n *Node) GetPorts() []*Port
func (n *Node) GetPort(name string) *Port
func (n *Node) GetPortsByDirection(dir PortDirection) []*Port

// Paramètres SPA
func (n *Node) SetParam(paramID uint32, value spa.PODValue) error
func (n *Node) GetParam(paramID uint32) (spa.PODValue, error)
func (n *Node) EnumParams(paramID uint32) ([]spa.PODValue, error)
```

#### **client/port.go** - Proxy Port
```go
type Port struct {
    ID          uint32
    Name        string
    Direction   PortDirection // Input/Output
    Type        PortType      // Audio/Midi/Video
    Props       map[string]string
    ParentNode  *Node
    
    conn        *core.Connection
}

type PortDirection uint32
const (
    PortDirectionInput  PortDirection = iota
    PortDirectionOutput
)

type PortType uint32
const (
    PortTypeAudio PortType = iota
    PortTypeMidi
    PortTypeVideo
)

// État
func (p *Port) IsConnected() bool
func (p *Port) GetConnectedPorts() []*Port

// Format
func (p *Port) GetSupportedFormats() ([]AudioFormat, error)
func (p *Port) GetCurrentFormat() (*AudioFormat, error)
```

#### **client/link.go** - Proxy Link
```go
type Link struct {
    ID          uint32
    InputPort   *Port
    OutputPort  *Port
    Props       map[string]string
    
    conn        *core.Connection
}

// État
func (l *Link) IsActive() bool
func (l *Link) GetFormat() (*AudioFormat, error)

// Modification
func (l *Link) UpdateProperties(props map[string]string) error
func (l *Link) Remove() error
```

#### **client/types.go** - Types communs
```go
type AudioFormat struct {
    Format      string // "S16LE", "S32LE", "F32LE", etc.
    SampleRate  uint32 // 44100, 48000, 96000, etc.
    Channels    uint32
    ChannelMask string // "FL,FR" for stereo
}

type NodeDirection string
const (
    NodeDirectionPlayback NodeDirection = "playback"
    NodeDirectionCapture  NodeDirection = "capture"
    NodeDirectionDuplex   NodeDirection = "duplex"
)

type NodeState string
const (
    NodeStateError     NodeState = "error"
    NodeStateSuspended NodeState = "suspended"
    NodeStateIdle      NodeState = "idle"
    NodeStateRunning   NodeState = "running"
)
```

---

## Phase 3 : Protocol Implementation Détaillée (À FAIRE)

### Interfaces PipeWire à supporter (par priorité MVP)

#### **Niveau Core (Obligatoire)**
- `pw.Core` - Objet racine (id=0)
  - Methods: ping, hello, sync, get_registry, error, done
  - Events: info, done, error

- `pw.Client` - Représentation client (id=1)
  - Methods: error, update_properties
  - Events: info, permissions, done

- `pw.Registry` - Découverte des objets
  - Methods: bind
  - Events: global, global_remove

#### **Niveau Audio Routing (MVP)**
- `pw.Node` - Nœud audio/video
  - Methods: add_listener, add_port, remove_port, set_param, etc.
  - Events: info, param, ports_changed

- `pw.Port` - Port d'un nœud
  - Methods: add_listener, query_objects, enum_params, subscribe_params, etc.
  - Events: info, param, registered, unregistered

- `pw.Link` - Liaison entre ports
  - Methods: add_listener, set_param
  - Events: info, param

#### **Niveau Permissions (Avancé)**
- `pw.Permission` - Contrôle d'accès (facultatif MVP)

### Structure des messages Protocol Native

Basé sur l'étude du source PipeWire (`module-protocol-native.c`):

```
Connection Handshake:
┌─ Client → Server ─────────────────────┐
│ core.hello()                          │
│  payload: client_version, client_id   │
└───────────────────────────────────────┘

┌─ Server → Client ─────────────────────┐
│ core.welcome() ou core.error()        │
│  payload: server_version, server_id   │
└───────────────────────────────────────┘

Method Call Format:
┌─────────────────────────────────────────────┐
│ uint32: Object ID                           │
│ uint32: Method ID                           │
│ uint32: Sequence (for request-reply)        │
│ POD[]: Arguments (variable length)          │
└─────────────────────────────────────────────┘

Event Format:
┌─────────────────────────────────────────────┐
│ uint32: Object ID                           │
│ uint32: Event ID                            │
│ uint32: Sequence (matches request)          │
│ POD[]: Arguments (variable length)          │
└─────────────────────────────────────────────┘
```

---

## Phase 4 : Tests et Documentation (À FAIRE)

### Tests unitaires à couvrir

```bash
# Tests POD parsing/building
go test ./spa -v -cover

# Tests connection
go test ./core -v -cover

# Tests client API
go test ./client -v -cover

# Couverture globale
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Documentation à générer

```bash
# Godoc local
godoc -http=:6060

# API documentation
go doc ./...

# Examples compilables et testés
go build ./examples/...
```

---

## Phase 5 : Clients Haute Niveau (Future)

Une fois la lib stable, créer des clients :

### TUI Client (Terminal UI)
```bash
cd cmd/pw-tui
go build .
./pw-tui
```

Fonctionnalités :
- Vue graphe interactif (nodes, ports, links)
- Routage audio drag-and-drop
- Monitoring latence/CPU
- Gestion paramètres

### GUI Client (GTK)
```bash
cd cmd/pw-gui
go build .
./pw-gui
```

---

## Checklist Développement

### Code Quality
- [ ] `CGO_ENABLED=0 go build ./...` - Compile sans CGO
- [ ] `go vet ./...` - Pas de warning vet
- [ ] `golangci-lint run` - Lint clean
- [ ] `go fmt ./...` - Formatted
- [ ] Tests >80% coverage

### Documentation
- [ ] README.md complet
- [ ] Godoc sur tous les exports
- [ ] Examples compilables
- [ ] ARCHITECTURE.md détaillé
- [ ] Protocol notes.md

### Protocol Compliance
- [ ] Respecte spec PipeWire native protocol
- [ ] Basé sur code source PipeWire analysé
- [ ] Gère handshake correctement
- [ ] Events/Methods mappés à la source

---

## Resources pour l'implémentation

### Source Code PipeWire
```
https://gitlab.freedesktop.org/pipewire/pipewire/

Fichiers clés:
- src/pipewire/protocol-native.c  - Protocol implementation
- src/modules/module-protocol-native.c - Native module
- src/pipewire/keys.h - Property names
- include/pipewire/*.h - API headers
```

### Documentation PipeWire
```
https://docs.pipewire.org/

Sections importantes:
- Native Protocol
- SPA POD
- PipeWire Library
- API Reference
```

### Testing
```
# Connect with pw-cli for comparison
pw-cli info core

# Monitor protocol traffic
strace -e trace=network pw-cli info core

# Analyze with Wireshark (Unix socket support)
```

---

## Next Steps

1. **Implement Phase 2** (Client API)
   - Commencer par Core, Registry
   - Puis Node, Port, Link
   - Tests unitaires pour chaque

2. **Test suite**
   - Unit tests pour POD parser
   - Integration tests contre daemon réel
   - Example programs compilables

3. **Documentation**
   - Godoc complète
   - Guide utilisateur détaillé
   - Architecture document

4. **Cleanup & Release**
   - Code review
   - Performance profiling
   - Release v0.1.0 MVP
