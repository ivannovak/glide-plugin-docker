package plugin

import (
	"context"
	"os/exec"

	glidectx "github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide-plugin-docker/internal/resolver"
)

// DockerDetector implements the SDK ContextExtension interface for Docker detection
type DockerDetector struct{}

// NewDockerDetector creates a new Docker detector
func NewDockerDetector() *DockerDetector {
	return &DockerDetector{}
}

// Name returns the unique identifier for this extension
func (d *DockerDetector) Name() string {
	return "docker"
}

// Detect analyzes the project environment and returns Docker-specific context data
func (d *DockerDetector) Detect(ctx context.Context, projectRoot string) (interface{}, error) {
	// Create a temporary context for detection
	// We need this to use the existing docker resolver logic
	tempCtx := &glidectx.ProjectContext{
		ProjectRoot: projectRoot,
		WorkingDir:  projectRoot,
		Extensions:  make(map[string]interface{}),
	}

	// Detect development mode (needed for compose file resolution)
	// For now, we'll use a simple heuristic
	tempCtx.DevelopmentMode = d.detectDevelopmentMode(projectRoot)
	tempCtx.Location = glidectx.LocationProject

	// Use Docker resolver for detection
	r := resolver.NewResolver(tempCtx)
	if err := r.Resolve(); err != nil {
		// If resolution fails, Docker is not available
		return nil, nil
	}

	// Build the extension data structure matching the compatibility layer field names
	result := map[string]interface{}{
		"docker_running":   tempCtx.DockerRunning,
		"compose_files":    tempCtx.ComposeFiles,
		"compose_override": tempCtx.ComposeOverride,
	}

	// Include container status if available
	if tempCtx.ContainersStatus != nil && len(tempCtx.ContainersStatus) > 0 {
		result["containers_status"] = tempCtx.ContainersStatus
	}

	return result, nil
}

// Merge combines this extension's data with existing extension data
func (d *DockerDetector) Merge(existing interface{}, new interface{}) (interface{}, error) {
	// For Docker, we simply prefer the new data over existing
	// In the future, we might want more sophisticated merging
	if new != nil {
		return new, nil
	}
	return existing, nil
}

// detectDevelopmentMode is a simple helper to detect development mode
// This is a temporary implementation until the context detector is refactored
func (d *DockerDetector) detectDevelopmentMode(projectRoot string) glidectx.DevelopmentMode {
	// Check if this is a multi-worktree setup
	// A multi-worktree setup has vcs/ and worktrees/ directories
	// For now, return single-repo as default
	return glidectx.ModeSingleRepo
}

// checkDockerDaemon verifies if Docker daemon is running
func checkDockerDaemon() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}
