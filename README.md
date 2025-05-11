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
- Shell completion support (bash, zsh, fish)
- Health monitoring and documentation generation
- Backup and restore functionality

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

## Shell Completion Setup

After installation, set up shell completion for your shell:

### Zsh
Add this line to your `~/.zshrc`:
```bash
eval "$(dotman completion zsh)"
```

### Bash
Add this line to your `~/.bashrc`:
```bash
source <(dotman completion bash)
```

### Fish
```bash
dotman completion fish > ~/.config/fish/completions/dotman.fish
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

### Add a configuration file

```bash
dotman add ~/.bashrc
```

This will:
1. Copy your file to the dotman repository
2. Create a symbolic link in the original location
3. Add and commit the file to git

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

### Commit changes

```bash
dotman commit "Your commit message"
```

This will commit all changes in the dotman repository.

### Push changes

```bash
dotman push
```

This will push committed changes to the remote repository.

### Update from remote repository

```bash
dotman update
```

This will pull the latest changes from the remote repository and relink all files.

### Remove a file from management

```bash
dotman remove ~/.bashrc
```

This will:
1. Copy the file back to its original location
2. Remove the symbolic link
3. Remove the file from git tracking
4. Commit the removal

### Health Check

```bash
dotman check
```

This will:
1. Check for broken symbolic links
2. Verify file permissions
3. Check git repository status
4. Verify backup integrity
5. Check for file conflicts
6. Check for outdated configurations
7. Monitor disk space
8. Check for uncommitted changes

### Generate Documentation

```bash
dotman docs
```

This will:
1. Create a main README with an overview of all configurations
2. Generate individual documentation for each configuration file
3. Detect and document dependencies and tags
4. Save metadata in JSON format

### Backup and Restore

```bash
# Create a backup
dotman backup ~/.bashrc

# List available backups
dotman restore

# Restore a specific backup
dotman restore 2024-02-20-123456
```

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

3. Push your changes:
   ```bash
   dotman push
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
├── configs/          # Your configuration files
├── backups/          # Backup files
├── health/           # Health check results
├── docs/             # Generated documentation
├── .git/
└── .gitignore
```

## Contributing

Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting any contributions. We welcome all forms of contributions, including:

- Bug reports
- Feature requests
- Documentation improvements
- Code contributions

## License

This project is licensed under the MIT License - see the LICENSE file for details. 