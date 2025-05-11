package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status  string
	Message string
	Error   error
}

// HealthCheck performs various checks on the dotfile configuration
func (m *Manager) HealthCheck() error {
	var results []HealthCheckResult

	// Check for broken symlinks
	results = append(results, m.checkBrokenSymlinks())

	// Check file permissions
	results = append(results, m.checkFilePermissions())

	// Check git repository status
	results = append(results, m.checkGitStatus())

	// Check backup integrity
	results = append(results, m.checkBackupIntegrity())

	// Print results
	hasErrors := false
	for _, result := range results {
		if result.Error != nil {
			hasErrors = true
			fmt.Printf("❌ %s: %s\n", result.Status, result.Message)
		} else {
			fmt.Printf("✅ %s: %s\n", result.Status, result.Message)
		}
	}

	if hasErrors {
		return fmt.Errorf("health check found issues")
	}

	return nil
}

// checkBrokenSymlinks checks for broken symbolic links
func (m *Manager) checkBrokenSymlinks() HealthCheckResult {
	var brokenLinks []string

	err := filepath.Walk(m.config.ConfigsDir, func(path string, info os.FileInfo, err error) error {
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

		// Check if the symlink exists in home directory
		homePath := filepath.Join(m.config.HomeDir, relPath)
		if _, err := os.Lstat(homePath); os.IsNotExist(err) {
			brokenLinks = append(brokenLinks, relPath)
		}
		return nil
	})

	if err != nil {
		return HealthCheckResult{
			Status:  "Symlink Check",
			Message: fmt.Sprintf("Error checking symlinks: %v", err),
			Error:   err,
		}
	}

	if len(brokenLinks) > 0 {
		return HealthCheckResult{
			Status:  "Symlink Check",
			Message: fmt.Sprintf("Found %d broken symlinks: %s", len(brokenLinks), strings.Join(brokenLinks, ", ")),
			Error:   fmt.Errorf("broken symlinks found"),
		}
	}

	return HealthCheckResult{
		Status:  "Symlink Check",
		Message: "All symlinks are valid",
	}
}

// checkFilePermissions checks file permissions
func (m *Manager) checkFilePermissions() HealthCheckResult {
	var invalidPerms []string

	err := filepath.Walk(m.config.ConfigsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is readable
		if info.Mode()&0400 == 0 {
			relPath, _ := filepath.Rel(m.config.ConfigsDir, path)
			invalidPerms = append(invalidPerms, relPath)
		}
		return nil
	})

	if err != nil {
		return HealthCheckResult{
			Status:  "Permission Check",
			Message: fmt.Sprintf("Error checking permissions: %v", err),
			Error:   err,
		}
	}

	if len(invalidPerms) > 0 {
		return HealthCheckResult{
			Status:  "Permission Check",
			Message: fmt.Sprintf("Found %d files with invalid permissions: %s", len(invalidPerms), strings.Join(invalidPerms, ", ")),
			Error:   fmt.Errorf("invalid permissions found"),
		}
	}

	return HealthCheckResult{
		Status:  "Permission Check",
		Message: "All files have correct permissions",
	}
}

// checkGitStatus checks the git repository status
func (m *Manager) checkGitStatus() HealthCheckResult {
	if !m.isGitRepo() {
		return HealthCheckResult{
			Status:  "Git Status",
			Message: "Not a git repository",
			Error:   fmt.Errorf("not a git repository"),
		}
	}

	// Check for uncommitted changes
	statusCmd := exec.Command("git", "-C", m.config.DotmanDir, "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return HealthCheckResult{
			Status:  "Git Status",
			Message: fmt.Sprintf("Error checking git status: %v", err),
			Error:   err,
		}
	}

	if len(output) > 0 {
		return HealthCheckResult{
			Status:  "Git Status",
			Message: "Found uncommitted changes",
			Error:   fmt.Errorf("uncommitted changes found"),
		}
	}

	// Check if remote is configured
	remoteCmd := exec.Command("git", "-C", m.config.DotmanDir, "remote", "get-url", "origin")
	if err := remoteCmd.Run(); err != nil {
		return HealthCheckResult{
			Status:  "Git Status",
			Message: "No remote repository configured",
			Error:   fmt.Errorf("no remote repository"),
		}
	}

	return HealthCheckResult{
		Status:  "Git Status",
		Message: "Repository is clean and properly configured",
	}
}

// checkBackupIntegrity checks the integrity of backups
func (m *Manager) checkBackupIntegrity() HealthCheckResult {
	backupsDir := filepath.Join(m.config.DotmanDir, "backups")
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		return HealthCheckResult{
			Status:  "Backup Check",
			Message: "No backups directory found",
			Error:   fmt.Errorf("no backups directory"),
		}
	}

	var invalidBackups []string

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return HealthCheckResult{
			Status:  "Backup Check",
			Message: fmt.Sprintf("Error reading backups directory: %v", err),
			Error:   err,
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		backupDir := filepath.Join(backupsDir, entry.Name())
		metadataPath := filepath.Join(backupDir, "metadata.json")
		contentPath := filepath.Join(backupDir, "content")

		// Check if both metadata and content exist
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			invalidBackups = append(invalidBackups, entry.Name())
			continue
		}
		if _, err := os.Stat(contentPath); os.IsNotExist(err) {
			invalidBackups = append(invalidBackups, entry.Name())
			continue
		}
	}

	if len(invalidBackups) > 0 {
		return HealthCheckResult{
			Status:  "Backup Check",
			Message: fmt.Sprintf("Found %d invalid backups: %s", len(invalidBackups), strings.Join(invalidBackups, ", ")),
			Error:   fmt.Errorf("invalid backups found"),
		}
	}

	return HealthCheckResult{
		Status:  "Backup Check",
		Message: "All backups are valid",
	}
}
