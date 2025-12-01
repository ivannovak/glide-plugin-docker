package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ivannovak/glide-plugin-docker/pkg/version"
	"github.com/ivannovak/glide/v3/pkg/plugin/sdk/v2"
)

// Config defines the plugin's type-safe configuration.
// Users configure this in .glide.yml under plugins.docker
type Config struct {
	// DefaultProfile is the default Docker Compose profile to use
	DefaultProfile string `json:"defaultProfile" yaml:"defaultProfile"`

	// AutoDetectFiles enables automatic detection of compose files
	AutoDetectFiles bool `json:"autoDetectFiles" yaml:"autoDetectFiles"`

	// ProjectName overrides the default project name
	ProjectName string `json:"projectName" yaml:"projectName"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		AutoDetectFiles: true,
	}
}

// DockerPlugin implements the SDK v2 Plugin interface for Docker integration
type DockerPlugin struct {
	v2.BasePlugin[Config]
}

// New creates a new Docker plugin instance
func New() *DockerPlugin {
	return &DockerPlugin{}
}

// Metadata returns plugin information
func (p *DockerPlugin) Metadata() v2.Metadata {
	return v2.Metadata{
		Name:        "docker",
		Version:     version.Version,
		Author:      "Glide Team",
		Description: "Docker and Docker Compose integration for Glide",
		License:     "MIT",
		Homepage:    "https://github.com/ivannovak/glide-plugin-docker",
		Tags:        []string{"docker", "compose", "containers"},
		Capabilities: v2.Capabilities{
			RequiresDocker: true,
		},
	}
}

// Configure is called with the type-safe configuration
func (p *DockerPlugin) Configure(ctx context.Context, config Config) error {
	return p.BasePlugin.Configure(ctx, config)
}

// Commands returns the list of commands this plugin provides
func (p *DockerPlugin) Commands() []v2.Command {
	return []v2.Command{
		{
			Name:        "docker",
			Description: "Pass-through to docker compose with automatic file resolution",
			Category:    "containers",
			Aliases:     []string{"d"},
			Visibility:  "project-only",
			Handler:     v2.SimpleCommandHandler(p.executeDockerCommand),
		},
	}
}

// Init is called once after plugin load
func (p *DockerPlugin) Init(ctx context.Context) error {
	return nil
}

// HealthCheck verifies Docker is available
func (p *DockerPlugin) HealthCheck(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker not available: %w", err)
	}
	return nil
}

// executeDockerCommand executes docker compose commands
func (p *DockerPlugin) executeDockerCommand(ctx context.Context, req *v2.ExecuteRequest) (*v2.ExecuteResponse, error) {
	workDir := req.WorkingDir
	if workDir == "" {
		workDir = "."
	}

	// Build docker compose command
	cmdParts := []string{"compose"}

	// Add project name if configured
	config := p.Config()
	if config.ProjectName != "" {
		cmdParts = append(cmdParts, "-p", config.ProjectName)
	}

	// Add default profile if configured
	if config.DefaultProfile != "" {
		cmdParts = append(cmdParts, "--profile", config.DefaultProfile)
	}

	// Auto-detect and add compose files (if enabled)
	if config.AutoDetectFiles {
		composeFiles := p.findComposeFiles(workDir)
		for _, file := range composeFiles {
			cmdParts = append(cmdParts, "-f", file)
		}
	}

	// Add user arguments
	cmdParts = append(cmdParts, req.Args...)

	// Execute docker compose
	cmd := exec.CommandContext(ctx, "docker", cmdParts...)
	cmd.Dir = workDir

	// Set environment - start with parent environment
	cmd.Env = os.Environ()
	// Override/add custom environment variables
	for k, v := range req.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// For interactive commands (exec, run), connect stdio
	if p.isInteractiveCommand(req.Args) {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &v2.ExecuteResponse{
				ExitCode: 1,
				Error:    err.Error(),
			}, nil
		}
	}

	return &v2.ExecuteResponse{
		ExitCode: exitCode,
		Output:   string(output),
	}, nil
}

// findComposeFiles finds docker-compose files in the project
func (p *DockerPlugin) findComposeFiles(workDir string) []string {
	var files []string

	// Common compose file names to check
	fileNames := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}

	for _, name := range fileNames {
		filePath := filepath.Join(workDir, name)
		if _, err := os.Stat(filePath); err == nil {
			files = append(files, name)
		}
	}

	// Check for override file
	overrideNames := []string{
		"docker-compose.override.yml",
		"docker-compose.override.yaml",
		"compose.override.yml",
		"compose.override.yaml",
	}

	for _, name := range overrideNames {
		filePath := filepath.Join(workDir, name)
		if _, err := os.Stat(filePath); err == nil {
			files = append(files, name)
		}
	}

	return files
}

// isInteractiveCommand checks if the command requires interactive TTY
func (p *DockerPlugin) isInteractiveCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Check for explicit interactive flags
	for _, arg := range args {
		if arg == "-it" || arg == "-i" || arg == "--interactive" {
			return true
		}
	}

	// Check for commands that are typically interactive
	switch args[0] {
	case "exec":
		// exec is interactive if ending with shell command
		if len(args) > 2 {
			lastArg := args[len(args)-1]
			if lastArg == "bash" || lastArg == "sh" || lastArg == "/bin/bash" || lastArg == "/bin/sh" {
				return true
			}
		}
	case "run":
		return true
	}

	return false
}
