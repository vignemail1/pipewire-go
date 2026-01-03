# Contributing to PipeWire Go Library

## Code of Conduct

Respectez les autres contributeurs. Les discussions techniques doivent rester constructives.

## Avant de commencer

### Vérifier les issues existantes
Consultez les GitHub Issues pour voir si quelque chose est déjà en cours.

### Discussions majeures
Pour les changements majeurs (nouvelles interfaces, breaking changes), ouvrez une issue d'abord pour discussion.

## Processus de Contribution

### 1. Fork et Clone

```bash
git clone https://github.com/vignemail1/pipewire-go.git
cd pipewire-go
```

### 2. Créer une branche

```bash
git checkout -b feature/ma-fonctionnalite
# ou
git checkout -b fix/mon-bug-fix
```

### 3. Développement

#### Code Quality Standards

**Format:**
```bash
go fmt ./...
```

**Linting:**
```bash
go vet ./...
# Et si disponible:
golangci-lint run
```

**Tests:**
```bash
go test ./... -v
go test ./... -cover
```

**Zero CGO (obligatoire):**
```bash
CGO_ENABLED=0 go build ./...
CGO_ENABLED=0 go test ./...
```

#### Style Guide

- **Nommage:** 
  - Packages: minuscule, court (`spa`, `core`, `client`)
  - Types: PascalCase (`Node`, `Port`, `Link`)
  - Méthodes/Functions: PascalCase pour exports, camelCase pour privés
  - Constants: ALL_CAPS pour les constantes POD types

- **Documentation:**
  - Chaque export public a un commentaire
  - Commentaires au-dessus des types/fonctions
  - Exemples dans les commentaires pour les APIs complexes

```go
// Node represents a PipeWire audio/video node in the graph.
// A node can have multiple ports (input/output) connected to other nodes.
type Node struct {
    ID    uint32
    Name  string
    ...
}

// GetPorts returns all ports of this node.
// The direction parameter can be PortDirectionInput or PortDirectionOutput.
func (n *Node) GetPorts(direction PortDirection) []*Port {
    ...
}
```

- **Error handling:**
  - Toujours wrapper les erreurs avec contexte
  - Utiliser `fmt.Errorf("context: %w", err)`
  - Pas de panic dans la lib (sauf init)

```go
// ✅ Good
if err := conn.SendMessage(id, opcode, payload); err != nil {
    return fmt.Errorf("failed to send message to node %d: %w", id, err)
}

// ❌ Bad
if err != nil {
    panic(err)  // Never in library code
}
```

#### Testing

Chaque changement doit avoir des tests:

```bash
# Créer test_test.go pour chaque module
go test ./spa -v
go test ./core -v
go test ./client -v

# Coverage report
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Exemple de test:

```go
// spa/pod_test.go
package spa

import "testing"

func TestParseInt(t *testing.T) {
    // Build a POD with an int
    builder := NewPODBuilder()
    if err := builder.WriteInt(42); err != nil {
        t.Fatalf("Failed to write int: %v", err)
    }
    
    // Parse it back
    parser := NewPODParser(builder.Bytes())
    pod, err := parser.ParsePOD()
    if err != nil {
        t.Fatalf("Failed to parse: %v", err)
    }
    
    intPod, ok := pod.(*IntPOD)
    if !ok {
        t.Fatalf("Expected *IntPOD, got %T", pod)
    }
    
    if intPod.Value != 42 {
        t.Errorf("Expected 42, got %d", intPod.Value)
    }
}
```

#### Documentation

Mettre à jour la doc si approprié:

- **README.md** - Pour changements d'API utilisateur
- **ARCHITECTURE.md** - Pour changements d'architecture
- **Godoc comments** - Pour toutes les changes de code
- **Examples** - Pour nouvelles features

### 4. Commit

Utiliser des messages clairs:

```bash
# ✅ Good commit messages
git commit -m "feat: add Port direction filtering to Node"
git commit -m "fix: handle EOF in POD parser correctly"
git commit -m "docs: improve verbose logging documentation"
git commit -m "test: add coverage for Object POD parsing"

# ❌ Bad commit messages
git commit -m "fixes"
git commit -m "update stuff"
git commit -m "WIP"
```

Format de message recommandé (conventional commits):
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat` - Nouvelle feature
- `fix` - Bug fix
- `docs` - Documentation
- `test` - Tests
- `perf` - Performance
- `refactor` - Refactoring sans changement fonctionnel

Exemple:

```
feat(client): add Node.GetMetadata() method

Implement retrieving node metadata properties (channel positions, etc.)
from the server. This allows audio routing clients to understand the
channel layout of audio devices.

Closes #123
```

### 5. Push et Pull Request

```bash
git push origin feature/ma-fonctionnalite
```

Ouvrir une PR sur GitHub avec:

- **Description claire** - Qu'est-ce qui change et pourquoi
- **Related issues** - Lier les issues avec "Closes #123"
- **Testing instructions** - Comment tester le changement
- **Screenshots/logs** - Si applicable (surtout pour verbose output)

Template PR:

```markdown
## Description
Brève description de ce que ce PR fait.

## Related Issues
Closes #123

## How to Test
1. Compile: `go build ./...`
2. Run tests: `go test ./...`
3. Check verbose output: `./example_basic_connect`

## Checklist
- [ ] Tests ajoutés/mis à jour
- [ ] Documentation mise à jour
- [ ] Code formaté (`go fmt`)
- [ ] Zero CGO verified (`CGO_ENABLED=0 go build`)
- [ ] No warnings with `go vet`
```

## Areas for Contribution

### High Priority (Phase 2)

- [ ] `client/client.go` - Main client API
- [ ] `client/core.go` - Core proxy implementation
- [ ] `client/registry.go` - Object registry
- [ ] `client/node.go` - Audio node proxies
- [ ] `client/port.go` - Port proxies
- [ ] `client/link.go` - Link proxies
- [ ] Unit tests for all above
- [ ] Integration tests with real daemon

### Medium Priority (Phase 3)

- [ ] Advanced POD parsing (Object, Array, Sequence)
- [ ] Protocol message types (enums)
- [ ] Permission handling
- [ ] Session manager support

### Low Priority (Phase 4+)

- [ ] TUI client implementation
- [ ] GUI client (GTK) implementation
- [ ] Performance monitoring
- [ ] Network bridging

## Development Tools

### Debugging

```bash
# Set logger to debug level
logger := verbose.NewLogger(verbose.LogLevelDebug, true)

# Run your code with verbose output
./your-binary 2>&1 | tee debug.log

# Analyze with pw-cli for comparison
pw-cli info core
pw-cli ls -l

# Monitor with strace
strace -e trace=write,read ./your-binary

# Monitor daemon
pw-top
```

### Testing Against Real Daemon

```bash
# Start daemon if not running
systemctl --user start pipewire

# Verify it's running
pw-cli info core

# Run integration tests
go test -v ./client -integration

# Clean up test artifacts
./cleanup-test-sockets.sh
```

### Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Trace
go test -trace=trace.out ./...
go tool trace trace.out
```

## Documentation Standards

### Godoc

Chaque export public doit avoir:

```go
// Package name - brief description
// 
// This package provides ...
// 
// Example usage:
//  
//  client, _ := NewClient(socketPath)
//  nodes := client.GetRegistry().ListNodes()
//
package client

// TypeName - description (one liner)
// 
// More detailed explanation if needed.
// Properties and methods listed below.
type TypeName struct {
    Field1 string // Description
    Field2 uint32 // Description
}

// MethodName description
// 
// Additional details about behavior, error conditions, etc.
func (t *TypeName) MethodName(arg Type) (Result, error) {
    ...
}
```

Générer la documentation:

```bash
godoc -http=:6060 &
# Puis visiter http://localhost:6060
```

### Examples

Chaque feature importante doit avoir un exemple compilable:

```bash
examples/
├── basic_connect.go      # ✅ Compilable et runnable
├── list_devices.go
└── audio_routing.go

# Tous testables:
go run ./examples/basic_connect.go
```

## Review Process

Les PRs sont reviewés pour:

1. **Correctness** - Le code fait-il ce qu'il prétend?
2. **Testing** - Couverture suffisante?
3. **Documentation** - Les APIs sont-elles documentées?
4. **Style** - Suit-il les conventions?
5. **Performance** - Y a-t-il des régressions?
6. **Security** - Y a-t-il des problèmes de sécurité?

Les reviewers demanderont des changements si nécessaire. C'est normal, c'est du feedback constructif.

## Questions?

- **Issues** - Pour les bugs et features
- **Discussions** - Pour les questions de design
- **Documentation** - Consultez README.md et ARCHITECTURE.md

## License

En contribuant, vous acceptez que votre code soit sous licence MIT.

---

Merci de contribuer à pipewire-go! 🎉
