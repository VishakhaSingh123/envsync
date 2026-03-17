#!/usr/bin/env bash
# EnvSync install script
set -e

BINARY_NAME="envsync"
INSTALL_DIR="/usr/local/bin"

echo "Building EnvSync..."
go build -ldflags="-s -w" -o "$BINARY_NAME" .
echo "✔ Built"

echo "Installing to $INSTALL_DIR/$BINARY_NAME..."
sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
echo "✔ Installed"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✔ EnvSync installed successfully!"
echo ""
echo "Next steps:"
echo "  1. cd your-project"
echo "  2. envsync init"
echo "  3. export ENVSYNC_KEY=\$(openssl rand -base64 32)"
echo "  4. envsync audit --env dev"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
