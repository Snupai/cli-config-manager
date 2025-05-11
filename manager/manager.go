package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	_, err := gitUserCmd.Output()
	if err != nil {
		return fmt.Errorf("git user.name not configured. Please run: git config --global user.name 'Your Name'")
	}

	_, err = gitEmailCmd.Output()
	if err != nil {
		return fmt.Errorf("git user.email not configured. Please run: git config --global user.email 'your.email@example.com'")
	}

	// Clone the repository
	cloneCmd := exec.Command("git", "clone", repoURL, m.config.DotmanDir)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	// Link all files
	return m.Link()
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

	// Push to GitHub
	pushCmd := exec.Command("git", "-C", m.config.DotmanDir, "push", "-u", "origin", "main")
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("error pushing to GitHub: %v", err)
	}

	return nil
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
