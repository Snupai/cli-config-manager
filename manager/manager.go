package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cli-config-manager/config"
)

// Manager handles dotfile operations
type Manager struct {
	config *config.Config
}

// New creates a new Manager instance
func New(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// ListFiles returns a list of all managed files
func (m *Manager) ListFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(m.config.ConfigsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the configs directory itself
		if info.IsDir() {
			return nil
		}

		// Get relative path from configs directory
		relPath, err := filepath.Rel(m.config.ConfigsDir, path)
		if err != nil {
			return err
		}

		files = append(files, relPath)
		return nil
	})

	return files, err
}

// InitializeFromExistingRepo initializes the dotman directory from an existing GitHub repository
func (m *Manager) InitializeFromExistingRepo(repoURL string) error {
	// Check if git is configured
	gitUserCmd := exec.Command("git", "config", "user.name")
	gitEmailCmd := exec.Command("git", "config", "user.email")

	userName, err := gitUserCmd.Output()
	if err != nil {
		return fmt.Errorf("git user.name not configured. Please run: git config --global user.name 'Your Name'")
	}

	userEmail, err := gitEmailCmd.Output()
	if err != nil {
		return fmt.Errorf("git user.email not configured. Please run: git config --global user.email 'your.email@example.com'")
	}

	// Check if the directory is empty
	entries, err := os.ReadDir(m.config.DotmanDir)
	if err != nil {
		return fmt.Errorf("error reading dotman directory: %v", err)
	}

	if len(entries) > 0 {
		return fmt.Errorf("dotman directory is not empty. Please remove existing files or use a different directory")
	}

	// Clone the repository with verbose output
	fmt.Printf("Cloning repository: %s\n", repoURL)
	cloneCmd := exec.Command("git", "clone", repoURL, m.config.DotmanDir)
	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error cloning repository: %v\nOutput: %s", err, string(output))
	}
	fmt.Printf("Repository cloned successfully\n")

	// Create configs directory if it doesn't exist
	configsDir := filepath.Join(m.config.DotmanDir, "configs")
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		return fmt.Errorf("error creating configs directory: %v", err)
	}

	// Update .gitignore to include configs directory
	gitignorePath := filepath.Join(m.config.DotmanDir, ".gitignore")
	gitignoreContent := []byte("# Ignore everything in this directory\n*\n# Except this file\n!.gitignore\n!configs/\n")
	if err := os.WriteFile(gitignorePath, gitignoreContent, 0644); err != nil {
		return fmt.Errorf("error updating .gitignore: %v", err)
	}

	// Configure git for this repository
	configCmds := []struct {
		args []string
		desc string
	}{
		{[]string{"config", "user.name", strings.TrimSpace(string(userName))}, "Setting user name"},
		{[]string{"config", "user.email", strings.TrimSpace(string(userEmail))}, "Setting user email"},
	}

	for _, cmd := range configCmds {
		fmt.Printf("%s...\n", cmd.desc)
		gitCmd := exec.Command("git", append([]string{"-C", m.config.DotmanDir}, cmd.args...)...)
		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("error %s: %v", cmd.desc, err)
		}
	}

	// Add and commit the configs directory
	fmt.Println("Adding configs directory...")
	addCmd := exec.Command("git", "-C", m.config.DotmanDir, "add", "configs", ".gitignore")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("error adding configs directory: %v", err)
	}

	fmt.Println("Committing changes...")
	commitCmd := exec.Command("git", "-C", m.config.DotmanDir, "commit", "-m", "Add configs directory")
	if err := commitCmd.Run(); err != nil {
		// If there's nothing to commit, that's fine
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			fmt.Println("No changes to commit")
		} else {
			return fmt.Errorf("error committing configs directory: %v", err)
		}
	}

	// Push the changes
	fmt.Println("Pushing changes...")
	pushCmd := exec.Command("git", "-C", m.config.DotmanDir, "push")
	if err := pushCmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to push changes: %v\n", err)
	}

	fmt.Println("Repository initialized successfully. You can now start adding configuration files.")
	return nil
}

// InitializeGitRepo initializes a git repository and creates it on GitHub
func (m *Manager) InitializeGitRepo(repoName string) error {
	// Check if git is configured
	gitUserCmd := exec.Command("git", "config", "user.name")
	gitEmailCmd := exec.Command("git", "config", "user.email")

	_, err := gitUserCmd.Output()
	if err != nil {
		return fmt.Errorf("git user.name not configured. Please run: git config --global user.name 'Your Name'")
	}

	_, err = gitEmailCmd.Output()
	if err != nil {
		return fmt.Errorf("git user.email not configured. Please run: git config --global user.email 'your.email@example.com'")
	}

	// Initialize git repository
	initCmd := exec.Command("git", "-C", m.config.DotmanDir, "init")
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("error initializing git repository: %v", err)
	}

	// Create .gitignore
	gitignorePath := filepath.Join(m.config.DotmanDir, ".gitignore")
	gitignoreContent := []byte("# Ignore everything in this directory\n*\n# Except this file\n!.gitignore\n!configs/\n")
	if err := os.WriteFile(gitignorePath, gitignoreContent, 0644); err != nil {
		return fmt.Errorf("error creating .gitignore: %v", err)
	}

	// Create README.md
	readmePath := filepath.Join(m.config.DotmanDir, "README.md")
	readmeContent := []byte("# Dotman Managed Dotfiles\n\nThis is my dotman-managed dotfiles repository.")
	if err := os.WriteFile(readmePath, readmeContent, 0644); err != nil {
		return fmt.Errorf("error creating README.md: %v", err)
	}

	// Create repository on GitHub using gh CLI (public by default)
	createRepoCmd := exec.Command("gh", "repo", "create", repoName, "--public", "--source", m.config.DotmanDir, "--remote", "origin")
	if err := createRepoCmd.Run(); err != nil {
		return fmt.Errorf("error creating GitHub repository: %v. Make sure you have the GitHub CLI (gh) installed and are authenticated", err)
	}

	// Add and commit initial files
	addCmd := exec.Command("git", "-C", m.config.DotmanDir, "add", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("error adding files: %v", err)
	}

	commitCmd := exec.Command("git", "-C", m.config.DotmanDir, "commit", "-m", "Initial commit")
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("error committing files: %v", err)
	}

	// Set the default branch to main
	branchCmd := exec.Command("git", "-C", m.config.DotmanDir, "branch", "-M", "main")
	if err := branchCmd.Run(); err != nil {
		return fmt.Errorf("error setting default branch: %v", err)
	}

	// Try to push with retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		pushCmd := exec.Command("git", "-C", m.config.DotmanDir, "push", "-u", "origin", "main")
		if err := pushCmd.Run(); err != nil {
			if i == maxRetries-1 {
				// On last retry, try to get more detailed error information
				output, _ := pushCmd.CombinedOutput()
				return fmt.Errorf("error pushing to GitHub after %d attempts: %v\nOutput: %s", maxRetries, err, string(output))
			}
			// Wait a bit before retrying
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		// Push succeeded
		return nil
	}

	return fmt.Errorf("failed to push to GitHub after %d attempts", maxRetries)
}

// AddFile adds a new file to be managed
func (m *Manager) AddFile(filePath string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", absPath)
	}

	// Get relative path from home directory
	relPath, err := filepath.Rel(m.config.HomeDir, absPath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %v", err)
	}

	// Create target directory in configs
	targetDir := filepath.Join(m.config.ConfigsDir, filepath.Dir(relPath))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("error creating target directory: %v", err)
	}

	// Copy file to configs directory
	targetPath := filepath.Join(m.config.ConfigsDir, relPath)
	if err := copyFile(absPath, targetPath); err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	// Create parent directories for the symlink if they don't exist
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("error creating parent directories: %v", err)
	}

	// Remove existing file/link if it exists
	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("error removing existing file: %v", err)
	}

	// Create symbolic link
	if err := os.Symlink(targetPath, absPath); err != nil {
		return fmt.Errorf("error creating symbolic link: %v", err)
	}

	fmt.Printf("Added and linked: %s -> %s\n", absPath, targetPath)

	// Add and commit the file
	fmt.Println("Committing changes...")

	// First, ensure the file is tracked by git
	addCmd := exec.Command("git", "-C", m.config.DotmanDir, "add", "-f", targetPath)
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error adding file to git: %v\nOutput: %s", err, string(output))
	}

	// Check if there are any changes to commit
	statusCmd := exec.Command("git", "-C", m.config.DotmanDir, "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("error checking git status: %v", err)
	}

	if len(output) == 0 {
		fmt.Println("No changes to commit")
		return nil
	}

	commitMsg := fmt.Sprintf("Add %s", relPath)
	commitCmd := exec.Command("git", "-C", m.config.DotmanDir, "commit", "-m", commitMsg)
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error committing file: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// Link creates symbolic links for all managed files
func (m *Manager) Link() error {
	return filepath.Walk(m.config.ConfigsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from configs directory
		relPath, err := filepath.Rel(m.config.ConfigsDir, path)
		if err != nil {
			return err
		}

		// Create target path in home directory
		targetPath := filepath.Join(m.config.HomeDir, relPath)

		// Create parent directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Remove existing file/link if it exists
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}

		// Create symbolic link
		if err := os.Symlink(path, targetPath); err != nil {
			return err
		}

		fmt.Printf("Linked: %s -> %s\n", targetPath, path)
		return nil
	})
}

// CommitAndPush commits and pushes changes to the remote repository
func (m *Manager) CommitAndPush(message string) error {
	// Check if we're in a git repository
	if !m.isGitRepo() {
		return fmt.Errorf("not a git repository. Please initialize git first")
	}

	// Add all changes
	addCmd := exec.Command("git", "-C", m.config.DotmanDir, "add", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("error adding files: %v", err)
	}

	// Commit changes
	commitCmd := exec.Command("git", "-C", m.config.DotmanDir, "commit", "-m", message)
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}

	// Push changes
	pushCmd := exec.Command("git", "-C", m.config.DotmanDir, "push")
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("error pushing changes: %v", err)
	}

	return nil
}

// Update pulls the latest changes from the remote repository
func (m *Manager) Update() error {
	// Check if we're in a git repository
	if !m.isGitRepo() {
		return fmt.Errorf("not a git repository. Please initialize git first")
	}

	// Pull latest changes
	pullCmd := exec.Command("git", "-C", m.config.DotmanDir, "pull")
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("error pulling changes: %v", err)
	}

	// Relink files after update
	return m.Link()
}

// isGitRepo checks if the dotman directory is a git repository
func (m *Manager) isGitRepo() bool {
	gitDir := filepath.Join(m.config.DotmanDir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, sourceFile, 0644)
}

// BackupMetadata represents the metadata for a backup
type BackupMetadata struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	SymlinkPath  string    `json:"symlink_path,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// Backup represents a complete backup
type Backup struct {
	BackupMetadata
	Content []byte `json:"-"`
}

// BackupFile creates a backup of a managed file
func (m *Manager) BackupFile(filePath string) error {
	// Ensure the backups directory exists
	backupsDir := filepath.Join(m.config.DotmanDir, "backups")
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return fmt.Errorf("failed to create backups directory: %v", err)
	}

	// Read the original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Create backup metadata
	backup := Backup{
		BackupMetadata: BackupMetadata{
			ID:           time.Now().Format("2006-01-02-150405"),
			OriginalPath: filePath,
			Timestamp:    time.Now(),
		},
		Content: content,
	}

	// Check if the file is a symlink
	if linkPath, err := os.Readlink(filePath); err == nil {
		backup.SymlinkPath = linkPath
	}

	// Create backup directory
	backupDir := filepath.Join(backupsDir, backup.ID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Save the file content
	if err := os.WriteFile(filepath.Join(backupDir, "content"), content, 0644); err != nil {
		return fmt.Errorf("failed to save backup content: %v", err)
	}

	// Save the metadata
	metadata, err := json.MarshalIndent(backup.BackupMetadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(filepath.Join(backupDir, "metadata.json"), metadata, 0644); err != nil {
		return fmt.Errorf("failed to save metadata: %v", err)
	}

	return nil
}

// ListBackups returns a list of all available backups
func (m *Manager) ListBackups() ([]BackupMetadata, error) {
	backupsDir := filepath.Join(m.config.DotmanDir, "backups")
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backups directory: %v", err)
	}

	var backups []BackupMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metadataPath := filepath.Join(backupsDir, entry.Name(), "metadata.json")
		metadata, err := os.ReadFile(metadataPath)
		if err != nil {
			continue // Skip backups with missing metadata
		}

		var backup BackupMetadata
		if err := json.Unmarshal(metadata, &backup); err != nil {
			continue // Skip backups with invalid metadata
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

// RestoreBackup restores a file from a backup
func (m *Manager) RestoreBackup(backupID string) error {
	backupsDir := filepath.Join(m.config.DotmanDir, "backups")
	backupDir := filepath.Join(backupsDir, backupID)

	// Read metadata
	metadataPath := filepath.Join(backupDir, "metadata.json")
	metadata, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read backup metadata: %v", err)
	}

	var backup BackupMetadata
	if err := json.Unmarshal(metadata, &backup); err != nil {
		return fmt.Errorf("failed to parse backup metadata: %v", err)
	}

	// Read backup content
	contentPath := filepath.Join(backupDir, "content")
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return fmt.Errorf("failed to read backup content: %v", err)
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(backup.OriginalPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	// Restore the file
	if err := os.WriteFile(backup.OriginalPath, content, 0644); err != nil {
		return fmt.Errorf("failed to restore file: %v", err)
	}

	// Restore symlink if it existed
	if backup.SymlinkPath != "" {
		// Remove existing file/link if it exists
		if err := os.Remove(backup.OriginalPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove existing file: %v", err)
		}

		// Create the symlink
		if err := os.Symlink(backup.SymlinkPath, backup.OriginalPath); err != nil {
			return fmt.Errorf("failed to restore symlink: %v", err)
		}
	}

	return nil
}

// Push pushes committed changes to the remote repository
func (m *Manager) Push() error {
	// Check if we're in a git repository
	if !m.isGitRepo() {
		return fmt.Errorf("not a git repository. Please initialize git first")
	}

	// Push changes
	pushCmd := exec.Command("git", "-C", m.config.DotmanDir, "push")
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("error pushing changes: %v", err)
	}

	return nil
}
