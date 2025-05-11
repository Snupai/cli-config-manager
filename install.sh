#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print with color
print_color() {
    color=$1
    message=$2
    echo -e "${color}${message}${NC}"
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map architecture to Go architecture
case "$ARCH" in
    "x86_64")
        ARCH="amd64"
        ;;
    "aarch64")
        ARCH="arm64"
        ;;
    *)
        print_color "$RED" "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Map OS to Go OS
case "$OS" in
    "linux")
        OS="linux"
        ;;
    "darwin")
        OS="darwin"
        ;;
    *)
        print_color "$RED" "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Get the latest release version
print_color "$YELLOW" "Fetching latest release version..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/Snupai/cli-config-manager/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    print_color "$RED" "Failed to fetch latest version"
    exit 1
fi

print_color "$GREEN" "Latest version: $LATEST_VERSION"

# Download URL
DOWNLOAD_URL="https://github.com/Snupai/cli-config-manager/releases/download/${LATEST_VERSION}/dotman-${OS}-${ARCH}"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Download the binary
print_color "$YELLOW" "Downloading dotman..."
if ! curl -L "$DOWNLOAD_URL" -o "$TEMP_DIR/dotman" --fail --progress-bar; then
    print_color "$RED" "Failed to download binary from $DOWNLOAD_URL"
    print_color "$YELLOW" "Checking alternative download URL..."
    # Try alternative URL format
    ALT_URL="https://github.com/Snupai/cli-config-manager/releases/download/${LATEST_VERSION}/cli-config-manager-${OS}-${ARCH}"
    if ! curl -L "$ALT_URL" -o "$TEMP_DIR/dotman" --fail --progress-bar; then
        print_color "$RED" "Failed to download binary from alternative URL"
        exit 1
    fi
fi

# Verify the binary
if [ ! -s "$TEMP_DIR/dotman" ]; then
    print_color "$RED" "Downloaded file is empty"
    exit 1
fi

# Check if it's a valid binary
if ! file "$TEMP_DIR/dotman" | grep -q "ELF\|Mach-O"; then
    print_color "$RED" "Downloaded file is not a valid binary"
    cat "$TEMP_DIR/dotman"
    exit 1
fi

# Make it executable
chmod +x "$TEMP_DIR/dotman"

# Install the binary
print_color "$YELLOW" "Installing dotman..."
if [ "$(id -u)" -eq 0 ]; then
    # If running as root, install to /usr/local/bin
    mv "$TEMP_DIR/dotman" /usr/local/bin/dotman
else
    # If not running as root, install to ~/.local/bin
    mkdir -p ~/.local/bin
    mv "$TEMP_DIR/dotman" ~/.local/bin/dotman
    
    # Add ~/.local/bin to PATH if it's not already there
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        print_color "$YELLOW" "Adding ~/.local/bin to your PATH..."
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
        print_color "$GREEN" "Please restart your shell or run 'source ~/.bashrc' or 'source ~/.zshrc' to update your PATH"
    fi
fi

# Install shell completions
print_color "$YELLOW" "Installing shell completions..."

# Detect shell
SHELL_NAME=$(basename "$SHELL")

case "$SHELL_NAME" in
    "bash")
        if [ "$(id -u)" -eq 0 ]; then
            dotman completion bash > /etc/bash_completion.d/dotman
        else
            mkdir -p ~/.local/share/bash-completion/completions
            dotman completion bash > ~/.local/share/bash-completion/completions/dotman
        fi
        ;;
    "zsh")
        mkdir -p ~/.zsh/completion
        dotman completion zsh > ~/.zsh/completion/_dotman
        echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
        echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
        ;;
    "fish")
        mkdir -p ~/.config/fish/completions
        dotman completion fish > ~/.config/fish/completions/dotman.fish
        ;;
esac

print_color "$GREEN" "dotman has been installed successfully!"
print_color "$GREEN" "You can now use the 'dotman' command"

# Print version
print_color "$YELLOW" "Installed version:"
dotman --version 