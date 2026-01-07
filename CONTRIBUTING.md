# Contributing to pipewire-go

Thank you for your interest in contributing to pipewire-go! This document provides guidelines and instructions for contributing.

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please be respectful of other contributors and users.

## How to Contribute

### Reporting Bugs

Before reporting a bug, please:

1. Check the [issue tracker](https://github.com/vignemail1/pipewire-go/issues) to see if it's already reported
2. Try to reproduce the issue
3. Gather information:
   - Go version (`go version`)
   - Operating system
   - PipeWire version
   - Minimal code to reproduce

When creating a bug report:

```markdown
## Description
Clear description of what isn't working.

## Steps to Reproduce
1. Step 1
2. Step 2
3. ...

## Expected Behavior
What should happen.

## Actual Behavior
What actually happens.

## Environment
- Go version: 1.21
- OS: Linux
- PipeWire version: 0.3.x

## Code Example
// Minimal reproduction code
```

### Suggesting Features

Feature requests are welcome! Please:

1. Check if the feature has been requested before
2. Describe the use case
3. Explain why this feature would be useful
4. Provide examples if possible

### Pull Requests

#### Before You Start

1. Check the [roadmap](CHANGELOG.md#roadmap) to understand priorities
2. Look at [open issues](https://github.com/vignemail1/pipewire-go/issues) for current work
3. Fork the repository
4. Create a feature branch from `main`

```bash
git checkout -b feature/your-feature-name
```

#### Code Standards

All code must:

- [ ] Follow Go conventions and [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [ ] Pass `go fmt` formatting
- [ ] Pass `go vet ./...` linting
- [ ] Pass `golangci-lint run ./...` (if available)
- [ ] Have 85%+ test coverage
- [ ] Include proper error handling
- [ ] Have clear comments for exported symbols
- [ ] Use descriptive variable and function names

#### Testing Requirements

All changes must include tests:

```bash
# Run tests locally
go test -v ./...

# Check coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Minimum coverage expectations:
- New code: 85%+
- Modified code: Maintain or improve coverage
- Bug fixes: Include test that would fail without the fix

#### Commit Messages

Write clear, descriptive commit messages:

```
Add feature: Brief description

Detailed explanation of what changed and why.
Referrence relevant issues: Fixes #123
```

Format:
- First line: "[type]: Brief description" (50 chars max)
- Types: feat, fix, docs, style, refactor, perf, test, ci
- Body: Detailed explanation (72 chars per line)
- Reference issues: "Fixes #123", "Related to #456"

#### Documentation

Update documentation for:

- New public APIs (GoDoc comments)
- Changed behavior (README.md)
- New examples (if adding major features)
- API changes (CHANGELOG.md)

#### Pull Request Process

1. Update CHANGELOG.md with your changes
2. Ensure all tests pass: `go test ./...`
3. Check coverage: `go test -cover ./...`
4. Verify code quality: `go vet ./...`
5. Create a descriptive PR with:
   - Clear title
   - Description of changes
   - Reference to related issues
   - Any breaking changes noted

Example PR description:

```markdown
## Description
Implements feature X to enable use case Y.

## Changes
- Added `NewFunction()` to client package
- Updated `ExistingType` with new field
- Added 12 tests covering new functionality

## Testing
- All tests pass locally
- Coverage: 87% (was 85%)
- Manual testing with real PipeWire: OK

## Breaking Changes
None

## Closes
Closes #123
```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Optional: golangci-lint for enhanced linting

### Setup

```bash
# Clone repository
git clone https://github.com/vignemail1/pipewire-go.git
cd pipewire-go

# Install dependencies
go mod download

# Run tests
go test ./...

# Install linter (optional)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Project Structure

```
pipewire-go/
├── client/           # Client API
│   ├── *_test.go   # Client tests
│   └── *.go        # Client implementation
├── core/             # Protocol layer
│   ├── *_test.go   # Core tests
│   └── *.go        # Core implementation
├── spa/              # SPA/POD serialization
├── verbose/          # Logging
├── examples/         # Working examples
├── version.go        # Version management
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
└── LICENSE
```

## Key Concepts

### Nodes
Represent audio devices or applications. Key properties:
- Name, description, state
- Sample rate, channel count
- Parameters (format, latency, etc.)

### Ports
Represent connection points on nodes. Key properties:
- Direction (input/output)
- Type (audio, MIDI, etc.)
- Supported formats
- Connection status

### Links
Represent audio connections between ports. Properties:
- Source and destination ports
- Custom properties

### Events
Notifications of graph changes:
- Node added/removed
- Port added/removed/modified
- Link created/destroyed

## Testing Guidelines

### Unit Tests

Test individual functions:

```go
func TestNodeName(t *testing.T) {
    node := &client.Node{}
    node.Props = map[string]string{"node.name": "Test"}
    
    if node.Name() != "Test" {
        t.Errorf("Expected 'Test', got '%s'", node.Name())
    }
}
```

### Table-Driven Tests

For multiple scenarios:

```go
func TestPortDirection(t *testing.T) {
    tests := []struct {
        name      string
        direction client.PortDirection
        expected  string
    }{
        {"input", client.PortDirectionInput, "In"},
        {"output", client.PortDirectionOutput, "Out"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test code
        })
    }
}
```

### Integration Tests

Test against real PipeWire:

```bash
# Run integration tests (skips if no daemon)
go test -v -tags=integration ./...
```

## Performance Considerations

When optimizing:

1. Benchmark before and after: `go test -bench=. -benchmem`
2. Profile with pprof for significant changes
3. Document performance impact in PR
4. Ensure thread-safety for concurrent access

## Documentation

### GoDoc Standards

All public types and functions must have:

```go
// GetNodes returns all nodes in the graph.
// Returns empty slice if no nodes available.
func (c *Client) GetNodes() []*Node {
    // ...
}
```

### README Updates

Update README.md for:
- New major features
- API changes
- Breaking changes
- Important bug fixes

### CHANGELOG Updates

Update CHANGELOG.md following [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
### Added
- New feature description

### Fixed
- Bug fix description
```

## Release Process

Releases are managed by maintainers:

1. Update version in `version.go`
2. Update CHANGELOG.md
3. Create git tag: `git tag vX.Y.Z`
4. Push: `git push origin vX.Y.Z`
5. Create GitHub Release with changelog

## Getting Help

- Check [existing issues](https://github.com/vignemail1/pipewire-go/issues)
- Ask in issue discussions
- Review [examples](./examples/) for patterns
- Check [PipeWire documentation](https://pipewire.org/)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

All contributors will be recognized in the README contributors section.

Thank you for contributing! 🙋
