#!/bin/bash
# Install CLI tools to ~/.local/bin

set -e

echo "Building PipeWire CLI tools..."
make clean build

echo "Installing to ~/.local/bin..."
mkdir -p ~/.local/bin

for tool in pw-list pw-monitor pw-info pw-connect; do
    echo "Installing $tool..."
    cp bin/$tool ~/.local/bin/
    chmod +x ~/.local/bin/$tool
done

echo "Done! Add ~/.local/bin to your PATH if not already done:"
echo "  export PATH=~/.local/bin:\$PATH"
