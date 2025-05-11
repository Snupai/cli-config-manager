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

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

# Map architecture to release arch
case "$ARCH" in
    "x86_64"|"amd64")
        ARCH="x86_64"
        ;;
    "aarch64"|"arm64")
        ARCH="arm64"
        ;;
    *)
        print_color "$RED" "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Map OS to release OS
case "$OS" in
    "Linux"|"linux")
        OS="Linux"
        ;;
    "Darwin"|"darwin")
        OS="Darwin"
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

# Download the archive (always, since that's what is available)
ARCHIVE_URL="https://github.com/Snupai/cli-config-manager/releases/download/${LATEST_VERSION}/cli-config-manager-${OS}-${ARCH}.tar.gz"
ARCHIVE_PATH="$TEMP_DIR/archive.tar.gz"

print_color "$YELLOW" "Downloading archive..."
if curl -L "$ARCHIVE_URL" -o "$ARCHIVE_PATH" --fail --progress-bar; then
    print_color "$YELLOW" "Extracting archive..."
    tar -xzf "$ARCHIVE_PATH" -C "$TEMP_DIR"
    # Find the dotman binary in the extracted files
    if [ -f "$TEMP_DIR/dotman" ]; then
        BINARY_PATH="$TEMP_DIR/dotman"
    else
        # Try to find it in a subdirectory if present
        BINARY_PATH=$(find "$TEMP_DIR" -type f -name dotman | head -n 1)
        if [ -z "$BINARY_PATH" ]; then
            print_color "$RED" "dotman binary not found in the archive."
            exit 1
        fi
    fi
else
    print_color "$RED" "Failed to download archive from $ARCHIVE_URL."
    exit 1
fi

# Verify the binary
if [ ! -s "$BINARY_PATH" ]; then
    print_color "$RED" "Downloaded file is empty"
    exit 1
fi

# Check if it's a valid binary
if ! file "$BINARY_PATH" | grep -q "ELF\|Mach-O"; then
    print_color "$RED" "Downloaded file is not a valid binary"
    cat "$BINARY_PATH"
    exit 1
fi

# Make it executable
chmod +x "$BINARY_PATH"

# Install the binary
print_color "$YELLOW" "Installing dotman..."
if [ "$(id -u)" -eq 0 ]; then
    # If running as root, install to /usr/local/bin
    mv "$BINARY_PATH" /usr/local/bin/dotman
else
    # If not running as root, install to ~/.local/bin
    mkdir -p ~/.local/bin
    mv "$BINARY_PATH" ~/.local/bin/dotman
    
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