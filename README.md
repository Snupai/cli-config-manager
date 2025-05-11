# dotman - A Better Dotfile Manager

dotman is a modern dotfile manager that helps you manage your configuration files across different environments. It provides a simple way to organize, version control, and deploy your dotfiles.

## Features

- Simple and intuitive command-line interface
- Automatic symbolic link management
- Version control friendly
- Safe file operations with conflict detection
- Easy to use and maintain
- Git integration for version control
- Automatic GitHub repository creation
- List managed files
- Initialize from existing repository
- Shell completion support (bash, zsh, fish, powershell)

## Installation

### Quick Install

```bash
# One-line installation
curl -sSL https://raw.githubusercontent.com/Snupai/cli-config-manager/main/install.sh | bash
```

### Manual Installation

```bash
# Download the script
curl -sSL https://raw.githubusercontent.com/Snupai/cli-config-manager/main/install.sh -o install.sh
# Make it executable
chmod +x install.sh
# Run it
./install.sh
```

### From Source

```bash
# Clone the repository
git clone https://github.com/Snupai/cli-config-manager.git
cd cli-config-manager

# Build the project
go build

# Install the binary (optional)
sudo mv dotman /usr/local/bin/
```

## Prerequisites

- Git configured with your name and email:
  ```bash
  git config --global user.name "Your Name"
  git config --global user.email "your.email@example.com"
  ```
- GitHub CLI (gh) installed and authenticated for automatic repository creation

## Usage

### Initialize a new dotfile repository

```bash
dotman init
```

This will:
1. Create a `.dotman` directory in your home folder
2. Ask if you want to use an existing repository
   - If yes: Enter the repository URL (e.g., github.com/user/repo.git)
   - If no: Enter a new repository name (press Enter to use 'configs' as the default name)
3. Create a private GitHub repository (if creating new)
4. Initialize git and push the initial commit (if creating new)
5. Link all configuration files

### Add a configuration file

```bash
dotman add ~/.bashrc
```

This will copy your `.bashrc` file to the dotman repository and prepare it for management.

### List managed files

```bash
dotman list
```

This will show all files currently being managed by dotman.

### Link all managed files

```bash
dotman link
```

This will create symbolic links for all managed files in their original locations.

### Commit and push changes

```bash
dotman commit "Your commit message"
```

This will commit all changes in the dotman repository and push them to the remote repository.

### Update from remote repository

```bash
dotman update
```

This will pull the latest changes from the remote repository and relink all files.

### Upgrade dotman

```bash
dotman upgrade
```

This will:
1. Check for a newer version of dotman
2. Download and install the new version if available
3. Preserve your configuration and managed files

### Check version

```bash
dotman version
```

This will show the current version, commit hash, and build date.

### Shell Completion

dotman supports shell completion for bash, zsh, fish, and PowerShell. The installation script will automatically set up completions for your shell.

To manually set up completions:

```bash
# Bash
source <(dotman completion bash)

# Zsh
source <(dotman completion zsh)

# Fish
dotman completion fish > ~/.config/fish/completions/dotman.fish

# PowerShell
dotman completion powershell > dotman.ps1
```

## Example Workflow

### New Setup

1. Initialize the repository:
   ```bash
   dotman init
   # Press Enter to use 'configs' as the repository name
   ```

2. Add your configuration files:
   ```bash
   dotman add ~/.bashrc
   dotman add ~/.vimrc
   dotman add ~/.config/i3/config
   ```

3. Check managed files:
   ```bash
   dotman list
   ```

4. Link all files:
   ```bash
   dotman link
   ```

5. Commit and push your changes:
   ```bash
   dotman commit "Initial configuration"
   ```

### Using on Another Machine

1. Initialize from existing repository:
   ```bash
   dotman init
   # Choose 'y' when asked about existing repository
   # Enter your repository URL (e.g., github.com/user/configs.git)
   ```

2. Your configuration files will be automatically linked!

## Directory Structure

```
~/.dotman/
├── configs/
│   ├── .bashrc
│   ├── .vimrc
│   └── .config/
│       └── i3/
│           └── config
├── .git/
└── .gitignore
```

## Versioning

dotman follows [Semantic Versioning](https://semver.org/). The version number is automatically injected during the build process.

To check your current version:
```bash
dotman version
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 