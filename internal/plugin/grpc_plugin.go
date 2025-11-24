package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
)

// GRPCPlugin implements the gRPC GlidePluginServer interface
type GRPCPlugin struct {
	*v1.BasePlugin
}

// NewGRPCPlugin creates a new gRPC-based Docker plugin
func NewGRPCPlugin() *GRPCPlugin {
	metadata := &v1.PluginMetadata{
		Name:        "docker",
		Version:     "1.0.0",
		Author:      "Glide Team",
		Description: "Docker and Docker Compose integration for Glide",
		Homepage:    "https://github.com/ivannovak/glide-plugin-docker",
		License:     "MIT",
		Tags:        []string{"docker", "compose", "containers"},
		Aliases:     []string{"d"},
		Namespaced:  false,
	}

	p := &GRPCPlugin{
		BasePlugin: v1.NewBasePlugin(metadata),
	}

	// Register Docker commands
	p.registerCommands()

	return p
}

// registerCommands registers all Docker-related commands
func (p *GRPCPlugin) registerCommands() {
	// Main docker command - passthrough to docker compose
	p.RegisterCommand("docker", v1.NewSimpleCommand(
		&v1.CommandInfo{
			Name:        "docker",
			Description: "Pass-through to docker compose with automatic file resolution",
			Category:    "containers",
			Aliases:     []string{"d"},
			Visibility:  "project-only",
		},
		p.executeDockerCommand,
	))
}

// executeDockerCommand executes docker compose commands
func (p *GRPCPlugin) executeDockerCommand(ctx context.Context, req *v1.ExecuteRequest) (*v1.ExecuteResponse, error) {
	workDir := req.WorkDir
	if workDir == "" {
		workDir = "."
	}

	// Build docker compose command
	cmdParts := []string{"compose"}

	// Auto-detect and add compose files
	composeFiles := p.findComposeFiles(workDir)
	for _, file := range composeFiles {
		cmdParts = append(cmdParts, "-f", file)
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
			return &v1.ExecuteResponse{
				Success:  false,
				ExitCode: 1,
				Error:    err.Error(),
			}, nil
		}
	}

	return &v1.ExecuteResponse{
		Success:  exitCode == 0,
		ExitCode: int32(exitCode),
		Stdout:   output,
	}, nil
}

// findComposeFiles finds docker-compose files in the project
func (p *GRPCPlugin) findComposeFiles(workDir string) []string {
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
func (p *GRPCPlugin) isInteractiveCommand(args []string) bool {
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

