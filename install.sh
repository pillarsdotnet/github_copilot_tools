#!/bin/bash
#
# Install github-copilot-tools scripts to ~/.local/bin/
#

set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${HOME}/.local/bin"

echo "Installing GitHub Copilot Review Tools..."
echo ""

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Copy scripts
for script in "$REPO_DIR"/bin/*; do
  script_name=$(basename "$script")
  echo "Installing $script_name..."
  cp "$script" "$INSTALL_DIR/$script_name"
  chmod +x "$INSTALL_DIR/$script_name"
done

echo ""
echo "✓ Installation complete!"
echo ""
echo "Scripts installed to: $INSTALL_DIR"
echo ""
echo "Verify installation with:"
echo "  which copilot-reviews"
echo "  which open-review-threads"
echo "  which resolve-pr-review-threads"
echo "  which hide-pr-review"
echo "  which latest-copilot-review-id"
echo "  which is-sentinel-review"
echo "  which has-post-sentinel-commits"
