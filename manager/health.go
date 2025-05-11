package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Error     error     `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Severity  string    `json:"severity"` // "info", "warning", "error"
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

	// Check for file conflicts
	results = append(results, m.checkFileConflicts())

	// Check for outdated configurations
	results = append(results, m.checkOutdatedConfigs())

	// Check for disk space
	results = append(results, m.checkDiskSpace())

	// Check for file changes
	results = append(results, m.checkFileChanges())

	// Save health check results
	if err := m.saveHealthCheckResults(results); err != nil {
		fmt.Printf("Warning: Failed to save health check results: %v\n", err)
	}

	// Print results
	hasErrors := false
	for _, result := range results {
		icon := "✅"
		if result.Error != nil {
			hasErrors = true
			icon = "❌"
		} else if result.Severity == "warning" {
			icon = "⚠️"
		}
		fmt.Printf("%s %s: %s\n", icon, result.Status, result.Message)
	}

	if hasErrors {
		return fmt.Errorf("health check found issues")
	}

	return nil
}

// saveHealthCheckResults saves the health check results to a file
func (m *Manager) saveHealthCheckResults(results []HealthCheckResult) error {
	healthDir := filepath.Join(m.config.DotmanDir, "health")
	if err := os.MkdirAll(healthDir, 0755); err != nil {
		return err
	}

	// Create a timestamp for the filename
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := filepath.Join(healthDir, fmt.Sprintf("health-check-%s.json", timestamp))

	// Marshal results to JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
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
			Status:    "Symlink Check",
			Message:   fmt.Sprintf("Error checking symlinks: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	if len(brokenLinks) > 0 {
		return HealthCheckResult{
			Status:    "Symlink Check",
			Message:   fmt.Sprintf("Found %d broken symlinks: %s", len(brokenLinks), strings.Join(brokenLinks, ", ")),
			Error:     fmt.Errorf("broken symlinks found"),
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	return HealthCheckResult{
		Status:    "Symlink Check",
		Message:   "All symlinks are valid",
		Timestamp: time.Now(),
		Severity:  "info",
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
			Status:    "Permission Check",
			Message:   fmt.Sprintf("Error checking permissions: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	if len(invalidPerms) > 0 {
		return HealthCheckResult{
			Status:    "Permission Check",
			Message:   fmt.Sprintf("Found %d files with invalid permissions: %s", len(invalidPerms), strings.Join(invalidPerms, ", ")),
			Error:     fmt.Errorf("invalid permissions found"),
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	return HealthCheckResult{
		Status:    "Permission Check",
		Message:   "All files have correct permissions",
		Timestamp: time.Now(),
		Severity:  "info",
	}
}

// checkGitStatus checks the git repository status
func (m *Manager) checkGitStatus() HealthCheckResult {
	if !m.isGitRepo() {
		return HealthCheckResult{
			Status:    "Git Status",
			Message:   "Not a git repository",
			Error:     fmt.Errorf("not a git repository"),
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	// Check for uncommitted changes
	statusCmd := exec.Command("git", "-C", m.config.DotmanDir, "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return HealthCheckResult{
			Status:    "Git Status",
			Message:   fmt.Sprintf("Error checking git status: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	if len(output) > 0 {
		return HealthCheckResult{
			Status:    "Git Status",
			Message:   "Found uncommitted changes",
			Error:     fmt.Errorf("uncommitted changes found"),
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	// Check if remote is configured
	remoteCmd := exec.Command("git", "-C", m.config.DotmanDir, "remote", "get-url", "origin")
	if err := remoteCmd.Run(); err != nil {
		return HealthCheckResult{
			Status:    "Git Status",
			Message:   "No remote repository configured",
			Error:     fmt.Errorf("no remote repository"),
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	return HealthCheckResult{
		Status:    "Git Status",
		Message:   "Repository is clean and properly configured",
		Timestamp: time.Now(),
		Severity:  "info",
	}
}

// checkBackupIntegrity checks the integrity of backups
func (m *Manager) checkBackupIntegrity() HealthCheckResult {
	backupsDir := filepath.Join(m.config.DotmanDir, "backups")
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		return HealthCheckResult{
			Status:    "Backup Check",
			Message:   "No backups directory found",
			Error:     fmt.Errorf("no backups directory"),
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	var invalidBackups []string

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return HealthCheckResult{
			Status:    "Backup Check",
			Message:   fmt.Sprintf("Error reading backups directory: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
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
			Status:    "Backup Check",
			Message:   fmt.Sprintf("Found %d invalid backups: %s", len(invalidBackups), strings.Join(invalidBackups, ", ")),
			Error:     fmt.Errorf("invalid backups found"),
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	return HealthCheckResult{
		Status:    "Backup Check",
		Message:   "All backups are valid",
		Timestamp: time.Now(),
		Severity:  "info",
	}
}

// checkFileConflicts checks for potential file conflicts
func (m *Manager) checkFileConflicts() HealthCheckResult {
	var conflicts []string

	err := filepath.Walk(m.config.ConfigsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(m.config.ConfigsDir, path)
		if err != nil {
			return err
		}

		homePath := filepath.Join(m.config.HomeDir, relPath)
		if _, err := os.Lstat(homePath); err == nil {
			// File exists in home directory
			if linkPath, err := os.Readlink(homePath); err != nil {
				// Not a symlink, potential conflict
				conflicts = append(conflicts, relPath)
			} else if linkPath != path {
				// Symlink points to wrong location
				conflicts = append(conflicts, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return HealthCheckResult{
			Status:    "Conflict Check",
			Message:   fmt.Sprintf("Error checking conflicts: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	if len(conflicts) > 0 {
		return HealthCheckResult{
			Status:    "Conflict Check",
			Message:   fmt.Sprintf("Found %d potential conflicts: %s", len(conflicts), strings.Join(conflicts, ", ")),
			Error:     fmt.Errorf("conflicts found"),
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	return HealthCheckResult{
		Status:    "Conflict Check",
		Message:   "No conflicts found",
		Timestamp: time.Now(),
		Severity:  "info",
	}
}

// checkOutdatedConfigs checks for outdated configuration files
func (m *Manager) checkOutdatedConfigs() HealthCheckResult {
	var outdated []string

	err := filepath.Walk(m.config.ConfigsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file hasn't been modified in the last 30 days
		if time.Since(info.ModTime()) > 30*24*time.Hour {
			relPath, _ := filepath.Rel(m.config.ConfigsDir, path)
			outdated = append(outdated, relPath)
		}
		return nil
	})

	if err != nil {
		return HealthCheckResult{
			Status:    "Outdated Check",
			Message:   fmt.Sprintf("Error checking outdated files: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	if len(outdated) > 0 {
		return HealthCheckResult{
			Status:    "Outdated Check",
			Message:   fmt.Sprintf("Found %d potentially outdated files: %s", len(outdated), strings.Join(outdated, ", ")),
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	return HealthCheckResult{
		Status:    "Outdated Check",
		Message:   "No outdated files found",
		Timestamp: time.Now(),
		Severity:  "info",
	}
}

// checkDiskSpace checks available disk space
func (m *Manager) checkDiskSpace() HealthCheckResult {
	var stat syscall.Statfs_t
	err := syscall.Statfs(m.config.DotmanDir, &stat)
	if err != nil {
		return HealthCheckResult{
			Status:    "Disk Space",
			Message:   fmt.Sprintf("Error checking disk space: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	// Calculate available space in GB
	availableGB := float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024 * 1024)

	if availableGB < 1 {
		return HealthCheckResult{
			Status:    "Disk Space",
			Message:   fmt.Sprintf("Low disk space: %.2f GB available", availableGB),
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	return HealthCheckResult{
		Status:    "Disk Space",
		Message:   fmt.Sprintf("Sufficient disk space: %.2f GB available", availableGB),
		Timestamp: time.Now(),
		Severity:  "info",
	}
}

// checkFileChanges checks for uncommitted file changes
func (m *Manager) checkFileChanges() HealthCheckResult {
	if !m.isGitRepo() {
		return HealthCheckResult{
			Status:    "File Changes",
			Message:   "Not a git repository",
			Timestamp: time.Now(),
			Severity:  "info",
		}
	}

	cmd := exec.Command("git", "-C", m.config.DotmanDir, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return HealthCheckResult{
			Status:    "File Changes",
			Message:   fmt.Sprintf("Error checking file changes: %v", err),
			Error:     err,
			Timestamp: time.Now(),
			Severity:  "error",
		}
	}

	if len(output) > 0 {
		return HealthCheckResult{
			Status:    "File Changes",
			Message:   "Found uncommitted changes",
			Timestamp: time.Now(),
			Severity:  "warning",
		}
	}

	return HealthCheckResult{
		Status:    "File Changes",
		Message:   "No uncommitted changes",
		Timestamp: time.Now(),
		Severity:  "info",
	}
}
