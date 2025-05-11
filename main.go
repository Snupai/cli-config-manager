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

	"archive/tar"
	"compress/gzip"
	"io"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var verbose bool

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
		fmt.Println("repo: https://github.com/Snupai/cli-config-manager")
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

		currentVersion = strings.TrimPrefix(currentVersion, "v")

		if verbose {
			fmt.Printf("Current version: %s\n", currentVersion)
		}

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

		if verbose {
			fmt.Printf("Latest version: %s\n", latestVersion)
		}

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

		// Determine OS and architecture for archive naming
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		var releaseOS, releaseArch string

		switch goos {
		case "linux":
			releaseOS = "Linux"
		case "darwin":
			releaseOS = "Darwin"
		default:
			fmt.Printf("Unsupported OS: %s\n", goos)
			os.Exit(1)
		}

		switch goarch {
		case "amd64":
			releaseArch = "x86_64"
		case "arm64":
			releaseArch = "arm64"
		default:
			fmt.Printf("Unsupported architecture: %s\n", goarch)
			os.Exit(1)
		}

		archiveName := fmt.Sprintf("cli-config-manager-%s-%s.tar.gz", releaseOS, releaseArch)
		downloadURL := fmt.Sprintf(
			"https://github.com/Snupai/cli-config-manager/releases/download/%s/%s",
			release.TagName,
			archiveName,
		)

		if verbose {
			fmt.Printf("Download URL: %s\n", downloadURL)
		}

		tempDir, err := os.MkdirTemp("", "dotman-upgrade")
		if err != nil {
			fmt.Printf("Error creating temp directory: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tempDir)

		archivePath := filepath.Join(tempDir, archiveName)

		fmt.Println("Downloading new version archive...")
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

		out, err := os.Create(archivePath)
		if err != nil {
			fmt.Printf("Error creating archive file: %v\n", err)
			os.Exit(1)
		}
		if _, err := io.Copy(out, resp.Body); err != nil {
			fmt.Printf("Error saving archive: %v\n", err)
			os.Exit(1)
		}
		out.Close()

		if verbose {
			fmt.Printf("Archive downloaded to: %s\n", archivePath)
		}

		if err := untar(archivePath, tempDir, verbose); err != nil {
			fmt.Printf("Error extracting archive: %v\n", err)
			os.Exit(1)
		}

		dotmanPath := filepath.Join(tempDir, "dotman")
		if _, err := os.Stat(dotmanPath); os.IsNotExist(err) {
			// Try to find it in a subdirectory
			dotmanPath = ""
			err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
				if info != nil && info.Name() == "dotman" && !info.IsDir() {
					dotmanPath = path
					return io.EOF // stop walking
				}
				return nil
			})
			if dotmanPath == "" {
				fmt.Println("dotman binary not found in the archive.")
				os.Exit(1)
			}
		}

		if verbose {
			fmt.Printf("dotman binary found at: %s\n", dotmanPath)
		}

		currentBinary, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting current binary path: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Replacing current binary: %s\n", currentBinary)
		}

		fmt.Println("Installing new version...")
		if err := os.Rename(dotmanPath, currentBinary); err != nil {
			if linkErr, ok := err.(*os.LinkError); ok && strings.Contains(linkErr.Error(), "cross-device link") {
				if verbose {
					fmt.Println("Rename failed due to cross-device link, falling back to copy.")
				}
				if err := copyFile(dotmanPath, currentBinary); err != nil {
					fmt.Printf("Error copying new version: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Printf("Error installing new version: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Successfully upgraded to version %s\n", latestVersion)
	},
}

func untar(src, dest string, verbose bool) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		if verbose {
			fmt.Printf("Extracting: %s\n", target)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			os.Chmod(target, os.FileMode(hdr.Mode))
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
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

	upgradeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for upgrade")

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
