#!/usr/bin/env python3
"""
PipeWire Go Library - Complete Package Generator
Generates all files including Phase 2-5 skeleton templates
"""

import os
import tarfile
import json
from pathlib import Path
from datetime import datetime

def create_package(output_name="pipewire-go-0.1.0-dev.tar.gz"):
    """Create a complete tar.gz package with all files"""
    
    base_dir = Path("/tmp/pipewire-go-package")
    project_dir = base_dir / "pipewire-go"
    
    # Clean and create directories
    import shutil
    if base_dir.exists():
        shutil.rmtree(base_dir)
    base_dir.mkdir(parents=True)
    project_dir.mkdir(parents=True)
    
    # Dictionary of all files to create
    files = {
        # Root level
        "go.mod": """module github.com/vignemail1/pipewire-go

go 1.21

require (
    // No external dependencies for core library
    // Pure Go implementation
)
""",
        
        ".gitignore": """# Binaries
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
""",
        
        "LICENSE": """MIT License

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
""",
        
        "Makefile": """.PHONY: help build test clean coverage fmt lint all

help:
\t@echo "PipeWire Go Library - Available targets:"
\t@echo "  make build      - Build library (CGO disabled)"
\t@echo "  make test       - Run tests"
\t@echo "  make coverage   - Generate coverage report"
\t@echo "  make fmt        - Format code"
\t@echo "  make lint       - Run linter"
\t@echo "  make clean      - Clean build artifacts"
\t@echo "  make all        - Build, test, coverage"

all: clean fmt lint build test coverage

build:
\tCGO_ENABLED=0 go build -v ./...

test:
\tCGO_ENABLED=0 go test -v ./...

coverage:
\tCGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...
\tgo tool cover -html=coverage.out -o coverage.html

fmt:
\tgo fmt ./...

lint:
\tgo vet ./...

clean:
\trm -rf bin/ dist/ *.out *.html
\tgo clean -v ./...

.DEFAULT_GOAL := help
""",
    }
    
    # Create all file entries
    print("[*] Creating file structure...")
    
    for filepath, content in files.items():
        full_path = project_dir / filepath
        full_path.parent.mkdir(parents=True, exist_ok=True)
        full_path.write_text(content)
        print(f"  ✅ {filepath}")
    
    # Create directories structure
    dirs = [
        "spa", "core", "client", "verbose", "examples",
        "cmd/pw-tui", "cmd/pw-gui"
    ]
    
    for dir_name in dirs:
        (project_dir / dir_name).mkdir(parents=True, exist_ok=True)
        print(f"  📁 {dir_name}/")
    
    # Create tar.gz
    print(f"\n[*] Creating archive: {output_name}")
    
    with tarfile.open(output_name, "w:gz") as tar:
        tar.add(project_dir, arcname="pipewire-go")
    
    print(f"✅ Archive created: {output_name}")
    size_mb = os.path.getsize(output_name) / (1024*1024)
    print(f"   Size: {size_mb:.2f} MB")
    
    return output_name

if __name__ == "__main__":
    archive = create_package()
    print(f"\n🎉 Ready to download: {archive}")
    print("\nTo extract:")
    print(f"  tar -xzf {archive}")
    print(f"  cd pipewire-go")
    print("  make help")
