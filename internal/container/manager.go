package container

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide-plugin-docker/internal/resolver"
)

// Container represents a Docker container
type Container struct {
	ID        string    `json:"ID"`
	Name      string    `json:"Name"`
	Service   string    `json:"Service"`
	State     string    `json:"State"`
	Status    string    `json:"Status"`
	Health    string    `json:"Health"`
	ExitCode  int       `json:"ExitCode"`
	CreatedAt time.Time `json:"CreatedAt"`
	StartedAt time.Time `json:"StartedAt"`
	Ports     []string  `json:"PublishedPorts"`
	Project   string    `json:"Project"`
}

// ContainerManager manages Docker containers
type ContainerManager struct {
	ctx      *context.ProjectContext
	resolver *resolver.Resolver
}

// NewContainerManager creates a new container manager
func NewContainerManager(ctx *context.ProjectContext) *ContainerManager {
	return &ContainerManager{
		ctx:      ctx,
		resolver: resolver.NewResolver(ctx),
	}
}

// GetStatus retrieves the status of all project containers
func (cm *ContainerManager) GetStatus() ([]Container, error) {
	// Ensure we have compose files
	if len(cm.ctx.ComposeFiles) == 0 {
		return nil, fmt.Errorf("no docker-compose files configured")
	}

	// Build docker-compose ps command with JSON format
	args := cm.resolver.GetComposeCommand("ps", "--format", "json", "--all")
	cmd := exec.Command("docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	// Parse JSON output
	var containers []Container

	// Docker Compose outputs one JSON object per line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var container Container
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			// Try to parse as simple format if JSON fails
			continue
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// GetContainerByService finds a container by service name
func (cm *ContainerManager) GetContainerByService(service string) (*Container, error) {
	containers, err := cm.GetStatus()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		if container.Service == service {
			return &container, nil
		}
	}

	return nil, fmt.Errorf("container for service '%s' not found", service)
}

// Start starts all containers
func (cm *ContainerManager) Start() error {
	args := cm.resolver.GetComposeCommand("up", "-d")
	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start containers: %w\nOutput: %s", err, output)
	}

	return nil
}

// Stop stops all containers
func (cm *ContainerManager) Stop() error {
	args := cm.resolver.GetComposeCommand("down")
	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop containers: %w\nOutput: %s", err, output)
	}

	return nil
}

// Restart restarts specific services or all if no services specified
func (cm *ContainerManager) Restart(services ...string) error {
	args := cm.resolver.GetComposeCommand("restart")
	args = append(args, services...)

	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart containers: %w\nOutput: %s", err, output)
	}

	return nil
}

// Remove removes stopped containers
func (cm *ContainerManager) Remove(removeVolumes bool) error {
	args := cm.resolver.GetComposeCommand("down")
	if removeVolumes {
		args = append(args, "-v")
	}

	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove containers: %w\nOutput: %s", err, output)
	}

	return nil
}

// Logs retrieves logs from containers
func (cm *ContainerManager) Logs(service string, follow bool, tail int) (string, error) {
	args := cm.resolver.GetComposeCommand("logs")

	if follow {
		args = append(args, "-f")
	}

	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	if service != "" {
		args = append(args, service)
	}

	cmd := exec.Command("docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(output), nil
}

// StreamLogs streams logs from containers in real-time
func (cm *ContainerManager) StreamLogs(service string, onLog func(string)) error {
	args := cm.resolver.GetComposeCommand("logs", "-f")

	if service != "" {
		args = append(args, service)
	}

	cmd := exec.Command("docker", args...)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start log streaming: %w", err)
	}

	// Read logs line by line
	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		if err != nil {
			break
		}
		if n > 0 && onLog != nil {
			onLog(string(buf[:n]))
		}
	}

	return cmd.Wait()
}

// Execute runs a command in a container
func (cm *ContainerManager) Execute(service string, command []string, interactive bool) error {
	args := cm.resolver.GetComposeCommand("exec")

	if !interactive {
		args = append(args, "-T")
	}

	args = append(args, service)
	args = append(args, command...)

	cmd := exec.Command("docker", args...)

	if interactive {
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	return cmd.Run()
}

// IsRunning checks if a specific service is running
func (cm *ContainerManager) IsRunning(service string) bool {
	container, err := cm.GetContainerByService(service)
	if err != nil {
		return false
	}

	return container.State == "running"
}

// GetOrphanedContainers finds containers not defined in compose files
func (cm *ContainerManager) GetOrphanedContainers() ([]Container, error) {
	// Get project name
	projectName := cm.resolver.GetComposeProjectName()

	// List all containers with this project label
	cmd := exec.Command("docker", "ps", "-a",
		"--filter", fmt.Sprintf("label=com.docker.compose.project=%s", projectName),
		"--format", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Parse all project containers
	var allContainers []Container
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var container Container
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue
		}
		allContainers = append(allContainers, container)
	}

	// Get current compose services
	currentServices, err := cm.GetComposeServices()
	if err != nil {
		return nil, err
	}

	// Find orphaned containers
	var orphaned []Container
	for _, container := range allContainers {
		isOrphaned := true
		for _, service := range currentServices {
			if container.Service == service {
				isOrphaned = false
				break
			}
		}
		if isOrphaned {
			orphaned = append(orphaned, container)
		}
	}

	return orphaned, nil
}

// RemoveOrphaned removes orphaned containers
func (cm *ContainerManager) RemoveOrphaned() error {
	args := cm.resolver.GetComposeCommand("down", "--remove-orphans")
	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove orphaned containers: %w\nOutput: %s", err, output)
	}

	return nil
}

// GetComposeServices gets list of services defined in compose files
func (cm *ContainerManager) GetComposeServices() ([]string, error) {
	args := cm.resolver.GetComposeCommand("config", "--services")
	cmd := exec.Command("docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get compose services: %w", err)
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")
	return services, nil
}

// Pull pulls the latest images for all services
func (cm *ContainerManager) Pull() error {
	args := cm.resolver.GetComposeCommand("pull")
	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull images: %w\nOutput: %s", err, output)
	}

	return nil
}

// Build builds or rebuilds services
func (cm *ContainerManager) Build(noCache bool) error {
	args := cm.resolver.GetComposeCommand("build")
	if noCache {
		args = append(args, "--no-cache")
	}

	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build services: %w\nOutput: %s", err, output)
	}

	return nil
}

// Scale scales a service to a specific number of instances
func (cm *ContainerManager) Scale(service string, replicas int) error {
	args := cm.resolver.GetComposeCommand("up", "-d", "--scale",
		fmt.Sprintf("%s=%d", service, replicas))

	cmd := exec.Command("docker", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to scale service: %w\nOutput: %s", err, output)
	}

	return nil
}
