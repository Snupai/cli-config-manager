package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"cli-config-manager/config"
	"cli-config-manager/manager"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "dotman",
	Short: "A better dotfile manager",
	Long: `dotman is a modern dotfile manager that helps you manage your configuration files
across different environments. It provides a simple way to organize, version control,
and deploy your dotfiles.

dotman helps you:
- Manage your dotfiles across multiple machines
- Keep your configuration files in version control
- Automatically create symbolic links
- Handle conflicts safely
- Create and manage GitHub repositories for your configs

For more information about a command, use 'dotman help <command>'.`,
	Version: fmt.Sprintf("dotman version %s (commit: %s, built: %s)", version, commit, date),
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new dotfile repository",
	Long: `Initialize a new dotfile repository in your home directory.

This command will:
1. Create a .dotman directory in your home folder
2. Ask if you want to use an existing repository
   - If yes: You'll be prompted to enter the repository URL
   - If no: You'll be asked for a new repository name
3. Create a private GitHub repository (if creating new)
4. Initialize git and push the initial commit (if creating new)
5. Link all configuration files

Examples:
  # Create a new repository
  dotman init

  # Use an existing repository
  dotman init  # Then choose 'y' and enter the URL when prompted`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New()
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		if err := cfg.EnsureDirectories(); err != nil {
			fmt.Printf("Error creating directories: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Initialized dotman repository at:", cfg.DotmanDir)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to use an existing repository? (y/N): ")
		useExisting, _ := reader.ReadString('\n')
		useExisting = strings.TrimSpace(strings.ToLower(useExisting))

		m := manager.New(cfg)

		if useExisting == "y" {
			fmt.Print("Enter the repository URL (e.g., github.com/user/repo.git): ")
			repoURL, _ := reader.ReadString('\n')
			repoURL = strings.TrimSpace(repoURL)

			// Add https:// if not present
			if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
				repoURL = "https://" + repoURL
			}

			if err := m.InitializeFromExistingRepo(repoURL); err != nil {
				fmt.Printf("Error initializing from existing repository: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully initialized from repository: %s\n", repoURL)
		} else {
			// Ask for repository name
			fmt.Print("Enter GitHub repository name (press Enter to use 'configs'): ")
			repoName, _ := reader.ReadString('\n')
			repoName = strings.TrimSpace(repoName)

			// Use "configs" as default if no name provided
			if repoName == "" {
				repoName = "configs"
			}

			if err := m.InitializeGitRepo(repoName); err != nil {
				fmt.Printf("Error initializing git repository: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully created and initialized GitHub repository: %s\n", repoName)
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a new configuration file to manage",
	Long: `Add a new configuration file to be managed by dotman.

This command will:
1. Copy the specified file to the dotman repository
2. Create necessary directories if they don't exist
3. Prepare the file for management

The file path can be absolute or relative to your home directory.

Examples:
  dotman add ~/.bashrc
  dotman add ~/.config/i3/config
  dotman add .vimrc`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New()
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		m := manager.New(cfg)
		if err := m.AddFile(args[0]); err != nil {
			fmt.Printf("Error adding file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added %s to managed files\n", args[0])
	},
}

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link all managed configuration files",
	Long: `Create symbolic links for all managed configuration files.

This command will:
1. Scan the dotman repository for managed files
2. Create symbolic links in their original locations
3. Handle any existing files or links safely
4. Show which files were linked

Use this command after:
- Initializing a new repository
- Pulling changes from remote
- Adding new files

Example:
  dotman link`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New()
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		m := manager.New(cfg)
		if err := m.Link(); err != nil {
			fmt.Printf("Error linking files: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully linked all managed files")
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed configuration files",
	Long: `Display a list of all configuration files currently being managed by dotman.

This command will:
1. Scan the dotman repository
2. Show all files being managed
3. Display their relative paths

Use this command to:
- Check which files are being managed
- Verify your configuration
- Plan your next changes

Example:
  dotman list`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New()
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		m := manager.New(cfg)
		files, err := m.ListFiles()
		if err != nil {
			fmt.Printf("Error listing files: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No files are currently being managed")
			return
		}

		fmt.Println("Managed files:")
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}
	},
}

var commitCmd = &cobra.Command{
	Use:   "commit [message]",
	Short: "Commit and push changes to the remote repository",
	Long: `Commit and push your configuration changes to the remote repository.

This command will:
1. Add all changes to git
2. Create a commit with your message
3. Push the changes to the remote repository

Use this command to:
- Save your configuration changes
- Sync changes across machines
- Keep your dotfiles in version control

Examples:
  dotman commit "Update vim configuration"
  dotman commit "Add new i3 workspace settings"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New()
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		m := manager.New(cfg)
		if err := m.CommitAndPush(args[0]); err != nil {
			fmt.Printf("Error committing changes: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully committed and pushed changes")
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull latest changes from the remote repository",
	Long: `Pull the latest changes from the remote repository and update your configuration.

This command will:
1. Pull the latest changes from the remote repository
2. Update all managed files
3. Relink files to their original locations

Use this command to:
- Sync changes from another machine
- Update your configuration
- Get the latest changes

Example:
  dotman update`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New()
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		m := manager.New(cfg)
		if err := m.Update(); err != nil {
			fmt.Printf("Error updating: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully updated and relinked files")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dotman version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade dotman to the latest version",
	Long: `Check for and install the latest version of dotman.

This command will:
1. Check the latest version available on GitHub
2. Compare it with your current version
3. Download and install the new version if available
4. Preserve your configuration and managed files

Examples:
  dotman upgrade  # Check and install updates`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current version
		currentVersion := version
		if currentVersion == "dev" {
			fmt.Println("Cannot check for updates in development version")
			os.Exit(1)
		}

		// Remove 'v' prefix if present
		currentVersion = strings.TrimPrefix(currentVersion, "v")

		// Get latest version from GitHub
		resp, err := http.Get("https://api.github.com/repos/Snupai/cli-config-manager/releases/latest")
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			fmt.Printf("Error parsing release info: %v\n", err)
			os.Exit(1)
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")

		// Compare versions
		if latestVersion == currentVersion {
			fmt.Printf("You are already using the latest version: %s\n", currentVersion)
			return
		}

		fmt.Printf("New version available: %s (current: %s)\n", latestVersion, currentVersion)
		fmt.Print("Do you want to upgrade? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" {
			fmt.Println("Upgrade cancelled")
			return
		}

		// Determine OS and architecture
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		if goarch == "amd64" {
			goarch = "x86_64"
		}

		// Download URL
		downloadURL := fmt.Sprintf(
			"https://github.com/Snupai/cli-config-manager/releases/download/%s/dotman-%s-%s",
			release.TagName,
			goos,
			goarch,
		)

		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "dotman-upgrade")
		if err != nil {
			fmt.Printf("Error creating temp directory: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tempDir)

		// Download new version
		fmt.Println("Downloading new version...")
		resp, err = http.Get(downloadURL)
		if err != nil {
			fmt.Printf("Error downloading new version: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error downloading new version: HTTP %d\n", resp.StatusCode)
			os.Exit(1)
		}

		// Save to temp file
		newBinaryPath := filepath.Join(tempDir, "dotman")
		newBinary, err := os.OpenFile(newBinaryPath, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			fmt.Printf("Error creating new binary file: %v\n", err)
			os.Exit(1)
		}
		defer newBinary.Close()

		if _, err := newBinary.ReadFrom(resp.Body); err != nil {
			fmt.Printf("Error saving new binary: %v\n", err)
			os.Exit(1)
		}

		// Get current binary path
		currentBinary, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting current binary path: %v\n", err)
			os.Exit(1)
		}

		// Replace current binary
		fmt.Println("Installing new version...")
		if err := os.Rename(newBinaryPath, currentBinary); err != nil {
			fmt.Printf("Error installing new version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully upgraded to version %s\n", latestVersion)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(upgradeCmd)

	// Add completion commands
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(dotman completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ dotman completion bash > /etc/bash_completion.d/dotman
  # macOS:
  $ dotman completion bash > /usr/local/etc/bash_completion.d/dotman

Zsh:
  $ source <(dotman completion zsh)

  # To load completions for each session, execute once:
  $ dotman completion zsh > "${fpath[1]}/_dotman"

Fish:
  $ dotman completion fish > ~/.config/fish/completions/dotman.fish

PowerShell:
  PS> dotman completion powershell > dotman.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
