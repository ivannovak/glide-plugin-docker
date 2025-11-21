package resolver

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide-plugin-docker/internal/compose"
)

// Resolver provides high-level Docker resolution functionality
type Resolver struct {
	ctx             *context.ProjectContext
	composeResolver *compose.ComposeResolver
}

// NewResolver creates a new Docker resolver
func NewResolver(ctx *context.ProjectContext) *Resolver {
	return &Resolver{
		ctx:             ctx,
		composeResolver: compose.NewComposeResolver(ctx),
	}
}

// Resolve performs complete Docker environment resolution
func (r *Resolver) Resolve() error {
	// Resolve compose files
	composeFiles, err := r.composeResolver.ResolveComposeFiles()
	if err != nil {
		return err // Return the error instead of ignoring it
	}

	r.ctx.ComposeFiles = composeFiles

	// Set ComposeOverride field for compatibility
	r.ctx.ComposeOverride = r.GetOverrideFile()

	// Check Docker daemon status
	r.ctx.DockerRunning = r.checkDockerDaemon()

	// Get container status if Docker is running
	if r.ctx.DockerRunning && len(r.ctx.ComposeFiles) > 0 {
		r.ctx.ContainersStatus = r.getContainerStatus()
	}

	return nil
}

// checkDockerDaemon verifies if Docker daemon is running
func (r *Resolver) checkDockerDaemon() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

// getContainerStatus retrieves status of all containers
func (r *Resolver) getContainerStatus() map[string]context.ContainerStatus {
	status := make(map[string]context.ContainerStatus)

	// Build docker-compose ps command
	args := r.composeResolver.BuildComposeCommand("ps", "--format", "json")
	cmd := exec.Command("docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return status
	}

	// Parse the output (simplified for now)
	// In a real implementation, we'd parse the JSON output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// This is a simplified implementation
		// Real implementation would parse JSON and extract actual status
		if strings.Contains(line, "running") {
			status["app"] = context.ContainerStatus{
				Name:   "app",
				Status: "running",
			}
		}
	}

	return status
}

// GetComposeCommand returns a docker-compose command with resolved files
func (r *Resolver) GetComposeCommand(args ...string) []string {
	return r.composeResolver.BuildComposeCommand(args...)
}

// GetComposeFiles returns the resolved compose files
func (r *Resolver) GetComposeFiles() []string {
	return r.ctx.ComposeFiles
}

// GetRelativeComposeFiles returns compose files as relative paths
func (r *Resolver) GetRelativeComposeFiles() []string {
	return r.composeResolver.GetRelativeComposeFiles()
}

// ValidateSetup ensures Docker environment is properly configured
func (r *Resolver) ValidateSetup() error {
	// Check if Docker is installed
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker is not installed or not in PATH")
	}

	// Check if Docker daemon is running
	if !r.ctx.DockerRunning {
		return fmt.Errorf("docker daemon is not running")
	}

	// Validate compose files
	if len(r.ctx.ComposeFiles) == 0 {
		return fmt.Errorf("no compose files configured")
	}

	return r.composeResolver.ValidateComposeFiles()
}

// GetComposeProjectName returns the project name for docker-compose
func (r *Resolver) GetComposeProjectName() string {
	// Get base project name from project root
	baseName := filepath.Base(r.ctx.ProjectRoot)

	// In multi-worktree mode, use worktree name if available
	if r.ctx.DevelopmentMode == context.ModeMultiWorktree && r.ctx.WorktreeName != "" {
		// Sanitize worktree name for docker-compose
		name := strings.ReplaceAll(r.ctx.WorktreeName, "/", "-")
		name = strings.ReplaceAll(name, "_", "-")
		return fmt.Sprintf("%s-%s", baseName, name)
	}

	// Default to base project name
	return baseName
}

// GetDockerNetwork returns the Docker network name for the project
func (r *Resolver) GetDockerNetwork() string {
	return fmt.Sprintf("%s_default", r.GetComposeProjectName())
}

// IsDockerizedProject checks if the current project uses Docker
func (r *Resolver) IsDockerizedProject() bool {
	return len(r.ctx.ComposeFiles) > 0
}

// GetOverrideFile returns the path to the override file if it exists
func (r *Resolver) GetOverrideFile() string {
	// Check if any of the compose files is an override
	for _, file := range r.ctx.ComposeFiles {
		if strings.Contains(file, "override") {
			return file
		}
	}
	return ""
}

// ResolveForMode resolves Docker setup for a specific development mode
func (r *Resolver) ResolveForMode(mode context.DevelopmentMode) error {
	// Temporarily set mode for resolution
	originalMode := r.ctx.DevelopmentMode
	r.ctx.DevelopmentMode = mode

	// Resolve with the specified mode
	err := r.Resolve()

	// Restore original mode if there was an error
	if err != nil {
		r.ctx.DevelopmentMode = originalMode
	}

	return err
}
