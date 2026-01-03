#!/bin/bash
# Build script for pipewire-go library - generates complete archive

set -e

PROJECT_NAME="pipewire-go"
VERSION="0.1.0-dev"
ARCHIVE_NAME="${PROJECT_NAME}-${VERSION}.tar.gz"
BUILD_DIR="/tmp/${PROJECT_NAME}-build"

echo "================================"
echo "PipeWire Go Library - Build Script"
echo "================================"
echo "Project: $PROJECT_NAME"
echo "Version: $VERSION"
echo "Output: $ARCHIVE_NAME"
echo ""

# Clean previous build
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/$PROJECT_NAME"

cd "$BUILD_DIR/$PROJECT_NAME"

# Create directory structure
echo "Creating directory structure..."
mkdir -p spa core client verbose examples cmd/pw-tui cmd/pw-gui

# Files list - these will be created from the provided content
echo "Setting up files..."

# Root level documentation files
cat > go.mod << 'EOF'
module github.com/vignemail1/pipewire-go

go 1.21

require (
	// No external dependencies for core library
	// Pure Go implementation
)
EOF

cat > .gitignore << 'EOF'
# Binaries
*.exe
*.exe~
*.dll
*.so
*.so.*
*.dylib

# Test binaries
*.test

# Output
*.out
coverage.out
coverage.html

# IDE
.vscode/
.idea/
*.swp
*.swo

# Build
dist/
build/
*.tar.gz

# OS
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db
EOF

cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2025 PipeWire Go Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

cat > Makefile << 'EOF'
.PHONY: help build test clean coverage fmt lint all

help:
	@echo "PipeWire Go Library - Available targets:"
	@echo "  make build      - Build library (CGO disabled)"
	@echo "  make test       - Run tests"
	@echo "  make coverage   - Generate coverage report"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linter"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make all        - Build, test, coverage"
	@echo ""
	@echo "Environment:"
	@echo "  CGO_ENABLED=0 is enforced for all builds"

all: clean fmt lint build test coverage

build:
	CGO_ENABLED=0 go build -v ./...
	CGO_ENABLED=0 go build -o ./bin/pw-tui ./cmd/pw-tui
	CGO_ENABLED=0 go build -o ./bin/pw-gui ./cmd/pw-gui

test:
	CGO_ENABLED=0 go test -v ./...

coverage:
	CGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

fmt:
	go fmt ./...

lint:
	go vet ./...

clean:
	rm -rf bin/ dist/ *.out *.html coverage.*
	go clean -v ./...

.DEFAULT_GOAL := help
EOF

# Create README files for each major package
cat > spa/README.md << 'EOF'
# SPA/POD Format Package

Implements the SPA (Simple Plugin API) and POD (Plain Old Data) binary format used by PipeWire.

## Components

- `pod.go` - Parser and builder for POD structures
- `types.go` - (to be created) POD type definitions and constants
- `audio.go` - (to be created) Audio-specific POD structures

## Usage

```go
// Parsing
parser := spa.NewPODParser(data)
pod, err := parser.ParsePOD()

// Building
builder := spa.NewPODBuilder()
builder.WriteInt(42)
builder.WriteString("test")
binary := builder.Bytes()
```

## References

- PipeWire SPA/POD documentation
- pipewire/spa/pod.h header file
EOF

cat > core/README.md << 'EOF'
# Core Protocol Package

Low-level PipeWire protocol implementation.

## Components

- `connection.go` - Unix socket management and message I/O
- `protocol.go` - (to be created) Native protocol implementation
- `types.go` - (to be created) Protocol types and constants
- `errors.go` - (to be created) Protocol error types

## Usage

```go
conn, err := core.Dial("/run/pipewire-0", logger)
defer conn.Close()

conn.SendMessage(objectID, methodID, podPayload)
conn.RegisterEventHandler(objectID, handler)
```

## Architecture

- Async read/write loops with goroutines
- Event dispatch with callback registry
- Thread-safe message handling
EOF

cat > client/README.md << 'EOF'
# Client API Package

High-level client API for PipeWire applications.

## Components

- `client.go` - Main Client type and lifecycle
- `core.go` - (to be created) Core proxy
- `registry.go` - (to be created) Registry and object discovery
- `node.go` - (to be created) Node proxy
- `port.go` - (to be created) Port proxy
- `link.go` - (to be created) Link proxy
- `types.go` - (to be created) Common types
- `properties.go` - (to be created) Property handling

## Usage

```go
client, err := client.NewClient("")
defer client.Close()

nodes := client.ListNodes()
link, err := client.CreateLink(output, input, nil)
```

## Event Callbacks

```go
client.OnNodeAdded(func(n *Node) {
    fmt.Printf("Node: %s\n", n.Name())
})
```
EOF

cat > verbose/README.md << 'EOF'
# Verbose Logging Package

Debug and logging utilities for PipeWire protocol analysis.

## Components

- `logger.go` - Leveled logging with formatting
- `dumper.go` - (to be created) Binary and POD dumping

## Usage

```go
logger := verbose.NewLogger(verbose.LogLevelDebug, true)
logger.Debugf("Message: %v", value)
logger.DumpBinary("data", binaryContent)

dumper := verbose.NewMessageDumper(logger)
dumper.DumpMethodCall(objID, "methodName", methodID, args)
```

## Log Levels

- Silent: No output
- Error: Only errors
- Warn: Warnings and errors
- Info: General information
- Debug: All details including binary dumps
EOF

echo "✅ Documentation created"
echo ""
echo "Build complete!"
echo "Archive would be at: $BUILD_DIR/$ARCHIVE_NAME"
echo ""
echo "To create the actual archive, run:"
echo "  tar -czf $ARCHIVE_NAME -C $BUILD_DIR $PROJECT_NAME/"
echo ""
echo "The archive will be in the current directory."
