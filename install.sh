#!/bin/bash
#
# Install github-copilot-tools scripts to the first writable directory on $PATH
#
# Usage: ./install.sh [INSTALL_DIR]
#
# If INSTALL_DIR is not specified, searches $PATH for the first writable
# directory and installs there. Falls back to ~/.local/bin if no writable
# directory is found on PATH.
#

set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function to check if directory exists and is writable
is_writable() {
  [[ -d "$1" ]] && [[ -w "$1" ]]
}

# Find first writable directory on PATH
find_writable_on_path() {
  IFS=: read -ra PATH_DIRS <<< "$PATH"
  
  for dir in "${PATH_DIRS[@]}"; do
    if is_writable "$dir"; then
      echo "$dir"
      return 0
    fi
  done
  
  return 1
}

# Determine installation directory
if [[ -n "$1" ]]; then
  # Explicit directory provided
  INSTALL_DIR="$1"
  
  if [[ ! -d "$INSTALL_DIR" ]]; then
    echo "Creating directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
  fi
  
  if ! is_writable "$INSTALL_DIR"; then
    echo "Error: Cannot write to $INSTALL_DIR" >&2
    exit 1
  fi
  
  echo "Installing GitHub Copilot Review Tools to: $INSTALL_DIR"
else
  # Search $PATH for first writable directory
  if INSTALL_DIR=$(find_writable_on_path); then
    echo "Found writable directory on \$PATH: $INSTALL_DIR"
  else
    # Fall back to ~/.local/bin
    INSTALL_DIR="${HOME}/.local/bin"
    echo "No writable directories found on \$PATH"
    echo "Using fallback: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
    
    # Warn if not on PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
      echo ""
      echo "⚠️  WARNING: $INSTALL_DIR is not on your \$PATH"
      echo "Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc):"
      echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
      echo ""
    fi
  fi
fi

echo "Installing GitHub Copilot Review Tools..."
echo ""

# Copy scripts
count=0
for script in "$REPO_DIR"/bin/*; do
  script_name=$(basename "$script")
  cp "$script" "$INSTALL_DIR/$script_name"
  chmod +x "$INSTALL_DIR/$script_name"
  echo "✓ $script_name"
  ((++count))
done

echo ""
echo "✓ Installation complete! ($count scripts installed)"
echo ""
echo "Installed to: $INSTALL_DIR"
echo ""

# Verify at least one script is accessible
if command -v copilot-reviews &> /dev/null; then
  echo "✓ Scripts are accessible from \$PATH"
else
  echo "⚠️  Scripts not yet on \$PATH. Update your shell profile or run:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
