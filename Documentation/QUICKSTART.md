# Quick Start - PipeWire Go Library

## 🚀 Démarrer en 5 minutes

### 1. Vérifier l'installation Go
```bash
go version  # ≥ 1.21
```

### 2. Structure du projet
```
pipewire-go/
├── go.mod                      # ✅ Créé
├── README.md                   # ✅ Créé
├── ARCHITECTURE.md             # ✅ Créé
├── IMPLEMENTATION_GUIDE.md     # ✅ Créé
├── CONTRIBUTING.md             # ✅ Créé
├── DELIVERABLES.md            # ✅ Créé (ce fichier)
│
├── spa/                        # SPA/POD Format
│   ├── pod.go                  # ✅ Implémenté (400 lignes)
│   └── pod_test.go            # À faire
│
├── core/                       # Protocole bas niveau
│   ├── connection.go           # ✅ Implémenté (350 lignes)
│   └── connection_test.go      # À faire
│
├── client/                     # Client API (À FAIRE)
│   ├── client.go              # TODO Phase 2
│   ├── core.go                # TODO Phase 2
│   ├── registry.go            # TODO Phase 2
│   ├── node.go                # TODO Phase 2
│   ├── port.go                # TODO Phase 2
│   ├── link.go                # TODO Phase 2
│   ├── types.go               # TODO Phase 2
│   └── properties.go           # TODO Phase 2
│
├── verbose/                    # Logging
│   ├── logger.go               # ✅ Implémenté (350 lignes)
│   └── logger_test.go          # À faire
│
└── examples/                   # Exemples
    ├── basic_connect.go        # ✅ Créé
    ├── list_devices.go         # À faire
    └── audio_routing.go        # À faire
```

### 3. Tester le code existant
```bash
# Compiler sans CGO
CGO_ENABLED=0 go build ./...

# Vérifier pas d'erreurs
go vet ./...

# (Optional) Format
go fmt ./...

# Tester example (si PipeWire tourne)
CGO_ENABLED=0 go run example_basic_connect.go
```

### 4. Vérifier PipeWire
```bash
# Vérifier que le daemon tourne
systemctl --user status pipewire

# Si arrêté, démarrer
systemctl --user start pipewire

# Vérifier avec pw-cli
pw-cli info core
# Doit afficher les infos du core
```

### 5. Lire la doc
```
1. README.md          - Vue d'ensemble (20 min)
2. ARCHITECTURE.md    - Design (15 min)
3. IMPLEMENTATION_GUIDE.md - Plan (10 min)
```

---

## 📝 À faire immédiatement

### Phase 2 : Client API (2-3 semaines)

**Priority 1: Core Infrastructure**
```go
// client/types.go - Types communs
type AudioFormat struct { ... }
type NodeState string
type NodeDirection string
type PortType uint32

// client/client.go - Main API
type Client struct { ... }
func NewClient(socketPath string) (*Client, error)
func (c *Client) Close() error
func (c *Client) GetRegistry() *Registry
```

**Priority 2: Registry & Discovery**
```go
// client/registry.go
type Registry struct { ... }
func (r *Registry) ListAll() []*GlobalObject
func (r *Registry) BindObject(id uint32) (interface{}, error)

type GlobalObject struct {
    ID          uint32
    Type        string
    Version     uint32
    Properties  map[string]string
}
```

**Priority 3: Basic Proxies**
```go
// client/core.go
type Core struct { ... }
func (c *Core) Ping() error
func (c *Core) GetVersion() string

// client/node.go
type Node struct { ... }
func (n *Node) Name() string
func (n *Node) GetPorts() []*Port

// client/port.go
type Port struct { ... }
func (p *Port) IsConnected() bool

// client/link.go
type Link struct { ... }
func (l *Link) IsActive() bool
```

**Priority 4: Opérations audio routing**
```go
// client/client.go extensions
func (c *Client) CreateLink(out *Port, in *Port) (*Link, error)
func (c *Client) RemoveLink(link *Link) error
func (c *Client) ChangeRoute(from *Node, to *Node) error
```

### Testing Strategy

```bash
# Pour chaque nouveau module:
go test ./client -v
go test ./client -cover

# Target: >80% coverage
# Utiliser des mocks pour Socket si nécessaire
```

### Documentation

```bash
# Après chaque module:
godoc -http=:6060

# Vérifier apparition nouvelle API
# Vérifier Godoc comments présents
```

---

## 🎯 Priorité Implémentation

### Semaine 1
- [ ] `client/types.go` - Types essentiels
- [ ] `client/client.go` - Client API main
- [ ] `client/registry.go` - Object discovery
- [ ] Tests unitaires pour chacun

### Semaine 2
- [ ] `client/core.go` - Core proxy
- [ ] `client/node.go` - Node proxy  
- [ ] Tests et exemples

### Semaine 3
- [ ] `client/port.go` - Port proxy
- [ ] `client/link.go` - Link proxy
- [ ] Audio routing operations
- [ ] Integration tests

### Semaine 4+
- [ ] Documentation complète
- [ ] Examples compilables
- [ ] Release v0.1.0 MVP
- [ ] Commencer TUI/GUI client

---

## 📚 Ressources déjà incluses

Pour chaque partie :

### Code
- spa/pod.go - Parseur/builder POD (✅ prêt)
- core/connection.go - Socket (✅ prêt)
- verbose/logger.go - Logging (✅ prêt)
- example_basic_connect.go - Test POD (✅ prêt)

### Documentation
- README.md - Guide utilisateur complet
- ARCHITECTURE.md - Design détaillé
- IMPLEMENTATION_GUIDE.md - Plan exact avec interfaces
- CONTRIBUTING.md - Pour contributeurs
- go.mod - Module declaration

### Plan
- Phases 1-5 décrites
- Interfaces définies dans IMPLEMENTATION_GUIDE
- Checklist complète

---

## 🔍 Avant de coder

Vérifier:

1. **PipeWire tourne**
   ```bash
   pw-cli info core
   # ✅ Affiche les infos
   ```

2. **Go compile sans CGO**
   ```bash
   CGO_ENABLED=0 go build ./...
   # ✅ Pas d'erreur
   ```

3. **Socket accessible**
   ```bash
   ls -l /run/pipewire-0
   # ✅ Fichier socket existe
   ```

4. **Docs lisibles**
   ```bash
   godoc -http=:6060
   # ✅ Voir les APIs existantes
   ```

---

## 💡 Tips d'implémentation

### Lors du codage de nouveaux proxies

1. **Suivre le pattern Proxy**
   ```go
   type NewObject struct {
       ID   uint32
       conn *core.Connection
       // ... state
   }
   ```

2. **Implémenter les methods**
   ```go
   func (o *NewObject) DoSomething() error {
       // Send message via conn
       // Parse response
       // Return result
   }
   ```

3. **Ajouter event handler**
   ```go
   func (o *NewObject) onEvent(msg *core.Message) error {
       // Handle server events
       // Update state
       return nil
   }
   ```

4. **Tester**
   ```go
   func TestNewObject(t *testing.T) {
       // Create mock or use real daemon
       // Test methods
   }
   ```

### Debugging avec verbose

```go
// Activer le verbose complet
logger := verbose.NewLogger(verbose.LogLevelDebug, true)

// Voir tous les messages
conn, _ := core.Dial(socketPath, logger)

// Output sera très détaillé :
// [15:04:05.123] [DEBUG] Sending message: ObjectID=0, OpCode=0
// [15:04:05.124] [DEBUG] Received event from server...
```

### Tester contre daemon réel

```bash
# Terminal 1: Run your client
./your-client

# Terminal 2: Monitor avec pw-cli
watch 'pw-cli ls'

# Terminal 3: Try other clients
pw-top
pw-dump
```

---

## 📞 Problèmes Communs

### "Socket not found"
```bash
# Vérifier daemon running
systemctl --user status pipewire

# Vérifier socket
ls -l /run/pipewire-0
```

### "Permission denied"
```bash
# Check user in audio group
groups $USER | grep audio
# Add if needed:
sudo usermod -aG audio $USER
```

### "Build fails with CGO"
```bash
# Vérifier variable
export CGO_ENABLED=0
go build ./...
```

### "Pod parsing fails"
```bash
// Enable verbose logging
logger := verbose.NewLogger(verbose.LogLevelDebug, true)
// Check DumpBinary output
logger.DumpBinary("raw data", binaryContent)
```

---

## ✅ Checklist Complétude

Avant de déclarer la Phase 2 complète:

- [ ] Tous les types définis
- [ ] Tous les proxies implémentés
- [ ] Unit tests >80% coverage
- [ ] Godoc complet
- [ ] Examples compilables et runables
- [ ] Zero CGO verification
- [ ] ARCHITECTURE.md updated
- [ ] CONTRIBUTING.md respected

---

## 🎉 Vous êtes prêt!

Les fondations sont là. Le reste c'est du coding steady selon le plan.

Bon courage! 🚀

Questions? Voir README.md ou ARCHITECTURE.md.
