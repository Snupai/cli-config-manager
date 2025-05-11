package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigDoc represents documentation for a configuration file
type ConfigDoc struct {
	Path         string    `json:"path"`
	Description  string    `json:"description"`
	LastUpdated  time.Time `json:"last_updated"`
	Tags         []string  `json:"tags"`
	Dependencies []string  `json:"dependencies"`
	Notes        string    `json:"notes"`
}

// GenerateDocs generates documentation for all managed configuration files
func (m *Manager) GenerateDocs() error {
	docsDir := filepath.Join(m.config.DotmanDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %v", err)
	}

	// Generate main README
	if err := m.generateMainReadme(docsDir); err != nil {
		return fmt.Errorf("failed to generate main README: %v", err)
	}

	// Generate individual config docs
	if err := m.generateConfigDocs(docsDir); err != nil {
		return fmt.Errorf("failed to generate config docs: %v", err)
	}

	return nil
}

// generateMainReadme generates the main README file
func (m *Manager) generateMainReadme(docsDir string) error {
	readmePath := filepath.Join(docsDir, "README.md")

	// Get list of all managed files
	files, err := m.ListFiles()
	if err != nil {
		return err
	}

	// Generate README content
	var content strings.Builder
	content.WriteString("# Dotman Configuration Documentation\n\n")
	content.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	content.WriteString("## Managed Configuration Files\n\n")
	for _, file := range files {
		content.WriteString(fmt.Sprintf("- [%s](%s.md)\n", file, file))
	}

	content.WriteString("\n## Quick Start\n\n")
	content.WriteString("1. Clone this repository\n")
	content.WriteString("2. Run `dotman link` to create symbolic links\n")
	content.WriteString("3. Run `dotman check` to verify your configuration\n\n")

	content.WriteString("## Maintenance\n\n")
	content.WriteString("- Run `dotman check` regularly to monitor configuration health\n")
	content.WriteString("- Use `dotman backup` before making significant changes\n")
	content.WriteString("- Keep your configuration up to date with `dotman update`\n")

	return os.WriteFile(readmePath, []byte(content.String()), 0644)
}

// generateConfigDocs generates documentation for individual configuration files
func (m *Manager) generateConfigDocs(docsDir string) error {
	return filepath.Walk(m.config.ConfigsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(m.config.ConfigsDir, path)
		if err != nil {
			return err
		}

		// Create documentation
		doc := ConfigDoc{
			Path:         relPath,
			LastUpdated:  info.ModTime(),
			Tags:         m.detectConfigTags(path),
			Dependencies: m.detectDependencies(path),
		}

		// Generate markdown documentation
		docPath := filepath.Join(docsDir, relPath+".md")
		if err := m.writeConfigDoc(docPath, doc); err != nil {
			return err
		}

		// Save JSON metadata
		jsonPath := filepath.Join(docsDir, relPath+".json")
		if err := m.saveConfigMetadata(jsonPath, doc); err != nil {
			return err
		}

		return nil
	})
}

// detectConfigTags detects relevant tags for a configuration file
func (m *Manager) detectConfigTags(path string) []string {
	var tags []string
	ext := filepath.Ext(path)
	base := filepath.Base(path)

	// Add file type tag
	switch ext {
	case ".rc", ".conf", ".config":
		tags = append(tags, "configuration")
	case ".sh", ".bash", ".zsh":
		tags = append(tags, "shell")
	case ".vim", ".vimrc":
		tags = append(tags, "vim")
	case ".gitconfig":
		tags = append(tags, "git")
	}

	// Add application-specific tags
	if strings.Contains(base, "i3") {
		tags = append(tags, "i3")
	}
	if strings.Contains(base, "tmux") {
		tags = append(tags, "tmux")
	}
	if strings.Contains(base, "nvim") {
		tags = append(tags, "neovim")
	}

	return tags
}

// detectDependencies detects dependencies for a configuration file
func (m *Manager) detectDependencies(path string) []string {
	var deps []string
	content, err := os.ReadFile(path)
	if err != nil {
		return deps
	}

	// Look for common dependency patterns
	contentStr := string(content)
	if strings.Contains(contentStr, "require") {
		deps = append(deps, "lua")
	}
	if strings.Contains(contentStr, "plugin") {
		deps = append(deps, "vim-plug")
	}
	if strings.Contains(contentStr, "source") {
		deps = append(deps, "shell")
	}

	return deps
}

// writeConfigDoc writes markdown documentation for a configuration file
func (m *Manager) writeConfigDoc(path string, doc ConfigDoc) error {
	var content strings.Builder

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	content.WriteString(fmt.Sprintf("# %s\n\n", doc.Path))
	content.WriteString(fmt.Sprintf("Last Updated: %s\n\n", doc.LastUpdated.Format("2006-01-02 15:04:05")))

	if len(doc.Tags) > 0 {
		content.WriteString("## Tags\n\n")
		for _, tag := range doc.Tags {
			content.WriteString(fmt.Sprintf("- %s\n", tag))
		}
		content.WriteString("\n")
	}

	if len(doc.Dependencies) > 0 {
		content.WriteString("## Dependencies\n\n")
		for _, dep := range doc.Dependencies {
			content.WriteString(fmt.Sprintf("- %s\n", dep))
		}
		content.WriteString("\n")
	}

	if doc.Description != "" {
		content.WriteString("## Description\n\n")
		content.WriteString(doc.Description + "\n\n")
	}

	if doc.Notes != "" {
		content.WriteString("## Notes\n\n")
		content.WriteString(doc.Notes + "\n\n")
	}

	return os.WriteFile(path, []byte(content.String()), 0644)
}

// saveConfigMetadata saves JSON metadata for a configuration file
func (m *Manager) saveConfigMetadata(path string, doc ConfigDoc) error {
	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
