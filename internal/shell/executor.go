package shell

import (
	"strings"

	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/internal/shell"
)

// DockerExecutorProvider implements the ExecutorProvider interface
type DockerExecutorProvider struct{}

// NewDockerExecutorProvider creates a new Docker executor provider
func NewDockerExecutorProvider() *DockerExecutorProvider {
	return &DockerExecutorProvider{}
}

// Name returns the unique identifier for this executor
func (p *DockerExecutorProvider) Name() string {
	return "docker"
}

// CanHandle returns true if this executor can handle the given command
func (p *DockerExecutorProvider) CanHandle(cmd *shell.Command) bool {
	if cmd == nil {
		return false
	}

	// Handle docker and docker-compose commands
	if cmd.Name == "docker" {
		// Check if it's a compose command
		if len(cmd.Args) > 0 && cmd.Args[0] == "compose" {
			return true
		}
		return true
	}

	// Handle legacy docker-compose command
	if cmd.Name == "docker-compose" {
		return true
	}

	return false
}

// CreateExecutor creates a new executor instance for Docker commands
func (p *DockerExecutorProvider) CreateExecutor(options shell.Options) shell.CommandExecutor {
	// For now, return the default executor
	// The DockerExecutor from internal/shell/docker.go will be migrated here
	// once we integrate the plugin system with the shell package
	return shell.NewExecutor(options)
}

// SupportsCommand checks if a specific command is supported
func (p *DockerExecutorProvider) SupportsCommand(command string) bool {
	dockerCommands := []string{
		"compose", "ps", "up", "down", "exec", "logs", "build",
		"pull", "push", "restart", "stop", "start", "rm", "run",
		"images", "volume", "network", "system", "info", "version",
	}

	cmd := strings.ToLower(command)
	for _, supported := range dockerCommands {
		if cmd == supported {
			return true
		}
	}

	return false
}

// GetContextExtensionName returns the name of the context extension this executor uses
func (p *DockerExecutorProvider) GetContextExtensionName() string {
	return "docker"
}

// RequiresContext returns true if this executor needs project context
func (p *DockerExecutorProvider) RequiresContext() bool {
	return true
}

// ValidateContext checks if the context is valid for this executor
func (p *DockerExecutorProvider) ValidateContext(ctx *context.ProjectContext) error {
	// Check if Docker extension is available
	if ctx.Extensions == nil {
		return nil // Context extensions not initialized, skip validation
	}

	dockerExt := ctx.Extensions["docker"]
	if dockerExt == nil {
		return nil // Docker not detected, but that's okay
	}

	// Context is valid
	return nil
}
