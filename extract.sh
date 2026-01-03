#!/bin/bash
# Extract and verify pipewire-go archive

ARCHIVE="$1"

if [ -z "$ARCHIVE" ]; then
    echo "Usage: $0 <archive.tar.gz>"
    exit 1
fi

if [ ! -f "$ARCHIVE" ]; then
    echo "Error: File not found: $ARCHIVE"
    exit 1
fi

echo "Extracting: $ARCHIVE"
tar -xzf "$ARCHIVE"

PROJECT_NAME=$(tar -tzf "$ARCHIVE" | head -1 | cut -d'/' -f1)
cd "$PROJECT_NAME" || exit 1

echo "✅ Extracted to: $PROJECT_NAME"
echo ""
echo "📋 Verifying structure..."

# Check essential files
REQUIRED_FILES=(
    "go.mod"
    "README.md"
    "ARCHITECTURE.md"
    "IMPLEMENTATION_GUIDE.md"
    "CONTRIBUTING.md"
    "spa/pod.go"
    "core/connection.go"
    "verbose/logger.go"
    "client/client.go"
    "examples/basic_connect.go"
)

MISSING=0
for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ MISSING: $file"
        MISSING=$((MISSING + 1))
    fi
done

if [ $MISSING -eq 0 ]; then
    echo ""
    echo "✅ All files present!"
    echo ""
    echo "Next steps:"
    echo "  1. cd $PROJECT_NAME"
    echo "  2. go mod tidy"
    echo "  3. CGO_ENABLED=0 go build ./..."
    echo "  4. go test ./..."
    echo "  5. make help"
else
    echo ""
    echo "⚠️  $MISSING files missing!"
    exit 1
fi
