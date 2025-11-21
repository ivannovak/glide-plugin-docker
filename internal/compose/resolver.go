package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/internal/context"
)

// ComposeResolver handles Docker Compose file resolution
type ComposeResolver struct {
	ctx *context.ProjectContext
}

// NewComposeResolver creates a new compose file resolver
func NewComposeResolver(ctx *context.ProjectContext) *ComposeResolver {
	return &ComposeResolver{
		ctx: ctx,
	}
}

// ResolveComposeFiles determines which compose files to use based on context
func (cr *ComposeResolver) ResolveComposeFiles() ([]string, error) {
	var composeFiles []string

	// Find the main compose file
	mainFile, err := cr.findMainComposeFile()
	if err != nil {
		return nil, fmt.Errorf("failed to find docker-compose file: %w", err)
	}
	composeFiles = append(composeFiles, mainFile)

	// Find override file based on development mode
	overrideFile := cr.findOverrideFile()
	if overrideFile != "" {
		composeFiles = append(composeFiles, overrideFile)
	}

	// Find environment-specific overrides
	envOverride := cr.findEnvironmentOverride()
	if envOverride != "" {
		composeFiles = append(composeFiles, envOverride)
	}

	return composeFiles, nil
}

// findMainComposeFile locates the primary docker-compose file
func (cr *ComposeResolver) findMainComposeFile() (string, error) {
	// Try various compose file names in order of preference
	variants := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}

	// Determine search path based on location
	searchPath := cr.getSearchPath()

	for _, variant := range variants {
		fullPath := filepath.Join(searchPath, variant)
		if fileExists(fullPath) {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("no docker-compose file found in %s", searchPath)
}

// findOverrideFile locates the override file based on development mode
func (cr *ComposeResolver) findOverrideFile() string {
	// In multi-worktree mode, override is in the root directory
	if cr.ctx.DevelopmentMode == context.ModeMultiWorktree {
		// The override file should be at ../docker-compose.override.yml from vcs/
		// or ../../docker-compose.override.yml from worktrees/*/
		var overridePath string

		switch cr.ctx.Location {
		case context.LocationRoot:
			overridePath = filepath.Join(cr.ctx.ProjectRoot, "docker-compose.override.yml")
		case context.LocationMainRepo:
			overridePath = filepath.Join(cr.ctx.ProjectRoot, "docker-compose.override.yml")
		case context.LocationWorktree:
			overridePath = filepath.Join(cr.ctx.ProjectRoot, "docker-compose.override.yml")
		}

		if fileExists(overridePath) {
			return overridePath
		}

		// Try with .yaml extension
		overridePath = strings.Replace(overridePath, ".yml", ".yaml", 1)
		if fileExists(overridePath) {
			return overridePath
		}
	} else {
		// In single-repo mode, override is in the same directory as main compose file
		searchPath := cr.getSearchPath()

		// Try standard override file names
		overrideVariants := []string{
			"docker-compose.override.yml",
			"docker-compose.override.yaml",
			"compose.override.yml",
			"compose.override.yaml",
		}

		for _, variant := range overrideVariants {
			fullPath := filepath.Join(searchPath, variant)
			if fileExists(fullPath) {
				return fullPath
			}
		}
	}

	return ""
}

// findEnvironmentOverride looks for environment-specific compose files
func (cr *ComposeResolver) findEnvironmentOverride() string {
	// Check for environment-specific override (e.g., docker-compose.local.yml)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	searchPath := cr.getSearchPath()
	envVariants := []string{
		fmt.Sprintf("docker-compose.%s.yml", env),
		fmt.Sprintf("docker-compose.%s.yaml", env),
		fmt.Sprintf("compose.%s.yml", env),
		fmt.Sprintf("compose.%s.yaml", env),
	}

	for _, variant := range envVariants {
		fullPath := filepath.Join(searchPath, variant)
		if fileExists(fullPath) {
			return fullPath
		}
	}

	return ""
}

// getSearchPath returns the directory to search for compose files
func (cr *ComposeResolver) getSearchPath() string {
	switch cr.ctx.Location {
	case context.LocationRoot:
		// In root, check vcs/ directory
		return filepath.Join(cr.ctx.ProjectRoot, "vcs")
	case context.LocationMainRepo:
		// In main repo, use working directory
		return cr.ctx.WorkingDir
	case context.LocationWorktree:
		// In worktree, use working directory
		return cr.ctx.WorkingDir
	default:
		// Default to working directory
		return cr.ctx.WorkingDir
	}
}

// BuildComposeCommand constructs a docker-compose command with proper -f flags
func (cr *ComposeResolver) BuildComposeCommand(args ...string) []string {
	cmd := []string{"compose"}

	// Add compose files
	for _, file := range cr.ctx.ComposeFiles {
		cmd = append(cmd, "-f", file)
	}

	// Add user arguments
	cmd = append(cmd, args...)

	return cmd
}

// GetRelativeComposeFiles returns compose files as relative paths from working directory
func (cr *ComposeResolver) GetRelativeComposeFiles() []string {
	var relativeFiles []string

	for _, file := range cr.ctx.ComposeFiles {
		relPath, err := filepath.Rel(cr.ctx.WorkingDir, file)
		if err != nil {
			// If we can't get relative path, use absolute
			relativeFiles = append(relativeFiles, file)
		} else {
			relativeFiles = append(relativeFiles, relPath)
		}
	}

	return relativeFiles
}

// ValidateComposeFiles ensures all resolved compose files exist
func (cr *ComposeResolver) ValidateComposeFiles() error {
	if len(cr.ctx.ComposeFiles) == 0 {
		return fmt.Errorf("no compose files configured")
	}

	for _, file := range cr.ctx.ComposeFiles {
		if !fileExists(file) {
			return fmt.Errorf("compose file does not exist: %s", file)
		}
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists
// func dirExists(path string) bool {
// 	info, err := os.Stat(path)
// 	if err != nil {
// 		return false
// 	}
// 	return info.IsDir()
// }
