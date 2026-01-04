# PipeWire Go Library (pipewire-go)

Librairie Go pure pour interagir avec PipeWire via socket Unix, sans CGO, avec accès complet aux capacités du protocole.

Ce projet a été totalement vibe codé avec Perplexity Labs.

/!\ Projet en cours de développement /!\, il n'a pas encore été testé

## Table des matières

- [Objectives](#objectives)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [API Overview](#api-overview)
- [Protocol Implementation](#protocol-implementation)
- [Verbose Mode & Debugging](#verbose-mode--debugging)
- [Examples](#examples)
- [Contributing](#contributing)

## Objectives

Cette librairie offre :

✅ **Zéro CGO** - Pure Go, compilable statiquement, sans dépendances C  
✅ **Complète** - Accès à tout ce qui est possible via la socket PipeWire  
✅ **Robuste** - Implémentation basée sur le code source PipeWire  
✅ **Debuggable** - Mode verbose pour diagnostique de chaque action  
✅ **Documentée** - Code commenté et doc complète pour clients Go  
✅ **Extensible** - Base solide pour TUI/GUI audio routing  

## Architecture

### Structure Modulaire

```text
pipewire-go/
├── core/              # Logique centrale du protocole
│   ├── connection.go  # Gestion socket Unix
│   ├── protocol.go    # Marshalling/unmarshalling
│   └── types.go       # Types PipeWire
├── spa/               # Implémentation SPA/POD
│   ├── pod.go         # Parseur/builder POD
│   ├── types.go       # Types SPA
│   └── audio.go       # Format audio spécifique
├── client/            # Client API haute niveau
│   ├── client.go      # Connexion au daemon
│   ├── registry.go    # Registry d'objets
│   ├── core.go        # Proxy Core
│   ├── node.go        # Proxy Node
│   ├── port.go        # Proxy Port
│   ├── link.go        # Proxy Link
│   └── properties.go  # Gestion des propriétés
├── verbose/           # Mode verbose & logging
│   ├── logger.go      # Système de logging
│   └── dumper.go      # Dump binaires, POD, etc
└── examples/          # Exemples d'utilisation
    ├── basic_connect.go
    ├── list_devices.go
    └── audio_routing.go
```

### Flow de Communication

```
Client Go
    │
    ├─► Socket Unix (/run/pipewire-0)
    │
    ├─► Connection Manager
    │   ├─ Envoi: Marshalling (POD)
    │   └─ Réception: Unmarshalling (POD)
    │
    ├─► Protocol Handler
    │   ├─ Methods (Client → Server)
    │   └─ Events (Server → Client)
    │
    └─► Object Proxies
        ├─ Core (id=0)
        ├─ Client (id=1)
        ├─ Registry
        ├─ Nodes
        ├─ Ports
        ├─ Links
        └─ ...
```

## Quick Start

### Installation

```bash
go get github.com/vignemail1/pipewire-go
```

### Usage Basique

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/vignemail1/pipewire-go/client"
    "github.com/vignemail1/pipewire-go/verbose"
)

func main() {
    // Configuration du logger verbose
    logger := verbose.NewLogger(verbose.LogLevelDebug, true)
    
    // Connexion au daemon PipeWire
    conn, err := client.Connect(client.DefaultSocketPath, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Lister tous les nodes
    nodes, err := conn.ListNodes()
    if err != nil {
        log.Fatal(err)
    }
    
    for _, node := range nodes {
        fmt.Printf("Node %d: %s\n", node.ID, node.Props["node.name"])
    }
}
```

## Core Concepts

### 1. Connexion (Connection)

- Établit la liaison socket Unix avec le daemon
- Gère l'event loop asynchrone
- Reçoit/envoie les messages au format POD

### 2. Proxies

Les proxies représentent des objets distants sur le serveur :

```go
// Core: id=0, l'objet racine
core := conn.GetCore()

// Client: id=1, représente ce client
client := conn.GetClient()

// Registry: pour découvrir les objets globaux
registry := conn.GetRegistry()

// Nodes, Ports, Links, etc. créés dynamiquement
node := conn.GetNode(nodeID)
```

### 3. Events & Methods

- **Methods** : Requêtes du client vers le serveur
- **Events** : Notifications du serveur vers le client

```go
// Method: lier deux ports
link, err := node.Connect(outputPort, inputPort)

// Event: écouteur pour les changements de graph
conn.OnNodeAdded(func(node *client.Node) {
    fmt.Printf("Node ajouté: %s\n", node.Name)
})
```

### 4. Propriétés (Properties)

Les objets ont des propriétés clé-valeur :

```go
props := node.Properties()
format := props["audio.format"]         // S16LE, F32LE, etc.
channels := props["audio.channels"]     // 2, 6, 8, ...
sampleRate := props["audio.rate"]       // 44100, 48000, ...
```

### 5. Types de Ports

```go
// Audio Input
inputPort := node.GetPort("in_0", client.PortTypeAudio, client.PortDirectionInput)

// Audio Output
outputPort := node.GetPort("out_0", client.PortTypeAudio, client.PortDirectionOutput)
```

## API Overview

### Client Connection

```go
type Client struct {
    // ... private fields
}

// Connect au daemon PipeWire
func Connect(socketPath string, logger *verbose.Logger) (*Client, error)

// Disconnecter et cleanup
func (c *Client) Close() error

// Obtenir les proxies principaux
func (c *Client) GetCore() *Core
func (c *Client) GetClient() *ClientProxy
func (c *Client) GetRegistry() *Registry

// Opérations haute niveau
func (c *Client) ListNodes() ([]*Node, error)
func (c *Client) ListPorts() ([]*Port, error)
func (c *Client) ListLinks() ([]*Link, error)
func (c *Client) GetNodeByName(name string) (*Node, error)
func (c *Client) CreateLink(out *Port, in *Port, props map[string]string) (*Link, error)
func (c *Client) RemoveLink(link *Link) error

// Event listeners
func (c *Client) OnNodeAdded(callback func(*Node))
func (c *Client) OnNodeRemoved(callback func(*Node))
func (c *Client) OnPortAdded(callback func(*Port))
func (c *Client) OnLinkAdded(callback func(*Link))
```

### Registry

```go
type Registry struct {
    // ... private fields
}

// Tous les objets globaux découverts
func (r *Registry) AllObjects() []*GlobalObject

// Filtrer par type
func (r *Registry) ObjectsByType(typeStr string) []*GlobalObject

// Observer les changements
func (r *Registry) OnAdded(callback func(*GlobalObject))
func (r *Registry) OnRemoved(callback func(id uint32))
```

### Node (Audio/Video Node)

```go
type Node struct {
    ID           uint32
    Type         string
    Props        map[string]string
    // ... internal state
}

// Propriétés d'accès
func (n *Node) Name() string
func (n *Node) Direction() string       // "playback" ou "capture"
func (n *Node) State() string           // "suspended", "running", etc
func (n *Node) SampleRate() uint32
func (n *Node) Channels() uint32

// Gestion des ports
func (n *Node) GetPorts() ([]*Port, error)
func (n *Node) GetPort(name string) (*Port, error)

// Paramètres
func (n *Node) GetParams(paramID uint32) ([]spa.POD, error)
func (n *Node) SetParam(paramID uint32, flags uint32, pod spa.POD) error
```

### Port

```go
type Port struct {
    ID          uint32
    Direction   PortDirection  // Input/Output
    Type        PortType       // Audio/Video/Midi
    Name        string
    Props       map[string]string
    ParentNode  *Node
}

// État du port
func (p *Port) IsConnected() bool
func (p *Port) GetLinks() ([]*Link, error)

// Format supportés
func (p *Port) GetSupportedFormats() ([]AudioFormat, error)
```

### Link (Connexion entre ports)

```go
type Link struct {
    ID     uint32
    Output *Port
    Input  *Port
    Props  map[string]string
}

// État et paramètres
func (l *Link) IsActive() bool
func (l *Link) GetFormat() (*AudioFormat, error)
func (l *Link) SetFormat(format *AudioFormat) error
```

## Protocol Implementation

### SPA/POD Format

La librairie implémente natif le format SPA/POD (Simple Plugin API / Plain Old Data) :

```
POD Structure (binary):
┌─────────────────────────────────────────────────┐
│ Size (uint32)  │ Type (uint32)  │ Payload ... │
├─────────────────────────────────────────────────┤
│ 4 bytes        │ 4 bytes        │ (size-8)    │
│ little-endian  │ little-endian  │ padded x8   │
└─────────────────────────────────────────────────┘

Types:
  0x00: None
  0x01: Bool
  0x02: Id
  0x03: Int
  0x04: Long
  0x05: Float
  0x06: Double
  0x07: String
  0x08: Bytes
  0x09: Rectangle
  0x0a: Fraction
  0x0b: Bitmap
  0x0c: Array
  0x0d: Struct
  0x0e: Object
  0x0f: Choice
  0x10: Pointer
  0x11: Fd
  0x12: Sequence
```

### Parsing POD

```go
package spa

type PODParser struct {
    data []byte
    pos  uint32
}

// Créer un parser
parser := NewPODParser(binaryData)

// Parser les éléments
intVal, err := parser.ParseInt()
floatVal, err := parser.ParseFloat()
str, err := parser.ParseString()

// Parser des structures complexes
pod, err := parser.ParsePOD()  // POD générique

// Itérer sur un objet POD
if obj, ok := pod.(*ObjectPOD); ok {
    for _, prop := range obj.Properties {
        // ...
    }
}
```

### Building POD

```go
builder := NewPODBuilder(buffer, bufferSize)

// Ajouter des valeurs primitives
builder.WriteInt(42)
builder.WriteFloat(3.14)
builder.WriteString("hello")

// Structures complexes
frame := builder.PushObject()
builder.WriteProp("audio.format", "S16LE")
builder.WriteProp("audio.rate", "48000")
builder.PopFrame(frame)

// Récupérer le POD binaire compilé
binary := builder.Bytes()
```

### Message Protocol

```
Message Native:

Handshake (Client → Server):
  Core Method: "ping"
  Payload: []
  
Response (Server → Client):
  Core Event: "pong"
  Payload: []

Method Call:
  object_id: uint32
  method_id: uint32
  signature: string (optionnel)
  args: POD[]

Event:
  object_id: uint32
  event_id: uint32
  args: POD[]
```

## Verbose Mode & Debugging

### Configuration du Logger

```go
logger := verbose.NewLogger(
    verbose.LogLevelDebug,  // LogLevelError, Info, Debug
    true,                    // includeTimestamps
)

// Ou utiliser les defaults
logger := verbose.DefaultLogger()
logger.SetLevel(verbose.LogLevelDebug)
```

### Types d'Output

```
[DEBUG] Connection: Sending message on socket /run/pipewire-0
[DEBUG]   Object ID: 0 (Core)
[DEBUG]   Method: ping
[DEBUG]   POD dump:
[DEBUG]     Size: 8 bytes
[DEBUG]     Type: 0x0d (Struct)
[DEBUG]     Content: (empty struct)
[DEBUG] 
[DEBUG] Connection: Received event from server
[DEBUG]   Object ID: 0 (Core)
[DEBUG]   Event: pong
[DEBUG]   Response time: 1.2ms
[DEBUG]   POD dump: ...
```

### Dump Binaires

```go
// Dump d'un buffer POD
logger.DumpPOD("Received POD", podBuffer)

// Dump d'une structure binaire complète
logger.DumpBinary("Raw socket data", rawData)

// Dump avec adresses mémoire et ASCII
logger.DumpHex("Message content", data, offset, length)
```

### Events de Logging

```go
logger.OnSend(func(objID uint32, methodID uint32, pod []byte) {
    // Callback quand quelque chose est envoyé
})

logger.OnReceive(func(objID uint32, eventID uint32, pod []byte) {
    // Callback quand quelque chose est reçu
})

logger.OnError(func(err error, context string) {
    // Callback en cas d'erreur
})
```

## Examples

### 1. Lister tous les nodes et ports

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/vignemail1/pipewire-go/client"
    "github.com/vignemail1/pipewire-go/verbose"
)

func main() {
    logger := verbose.NewLogger(verbose.LogLevelInfo, true)
    conn, err := client.Connect(client.DefaultSocketPath, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    nodes, err := conn.ListNodes()
    if err != nil {
        log.Fatal(err)
    }
    
    for _, node := range nodes {
        fmt.Printf("Node %d: %s\n", node.ID, node.Name())
        fmt.Printf("  Type: %s\n", node.Type)
        fmt.Printf("  State: %s\n", node.State())
        
        ports, _ := node.GetPorts()
        for _, port := range ports {
            fmt.Printf("  └─ Port %d: %s (%s)\n",
                port.ID, port.Name, port.Direction)
        }
    }
}
```

### 2. Créer un lien entre deux ports

```go
func linkAudioPorts(conn *client.Client, 
                   nodeAName string, 
                   nodeBName string) (*client.Link, error) {
    nodeA, err := conn.GetNodeByName(nodeAName)
    if err != nil {
        return nil, err
    }
    
    nodeB, err := conn.GetNodeByName(nodeBName)
    if err != nil {
        return nil, err
    }
    
    // Trouver le port output de A
    portsA, _ := nodeA.GetPorts()
    var outputPort *client.Port
    for _, p := range portsA {
        if p.Direction == client.PortDirectionOutput {
            outputPort = p
            break
        }
    }
    
    // Trouver le port input de B
    portsB, _ := nodeB.GetPorts()
    var inputPort *client.Port
    for _, p := range portsB {
        if p.Direction == client.PortDirectionInput {
            inputPort = p
            break
        }
    }
    
    if outputPort == nil || inputPort == nil {
        return nil, fmt.Errorf("ports non trouvés")
    }
    
    return conn.CreateLink(outputPort, inputPort, nil)
}
```

### 3. Monitorer les changements du graph

```go
func monitorGraph(conn *client.Client) {
    conn.OnNodeAdded(func(node *client.Node) {
        fmt.Printf("✨ Node ajouté: %s (id=%d)\n", node.Name(), node.ID)
    })
    
    conn.OnNodeRemoved(func(nodeID uint32) {
        fmt.Printf("🗑️  Node supprimé: id=%d\n", nodeID)
    })
    
    conn.OnLinkAdded(func(link *client.Link) {
        fmt.Printf("🔗 Lien créé: %s → %s\n",
            link.Output.Name, link.Input.Name)
    })
    
    // Rester actif
    select {}
}
```

## Contributing

Les contributions sont bienvenues ! Consulter CONTRIBUTING.md pour les guidelines.

### Checklist de Développement

- [ ] Code sans CGO (testable avec `CGO_ENABLED=0`)
- [ ] Tests unitaires (>80% coverage)
- [ ] Documentation complète (godoc)
- [ ] Mode verbose testé
- [ ] Exemples compilables et testés

### Tester la Compilation

```bash
CGO_ENABLED=0 go build ./...
CGO_ENABLED=0 go test ./...
```

---

**Licence**: MIT  
**Status**: MVP en développement
