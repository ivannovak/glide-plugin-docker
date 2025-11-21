package plugin

import (
	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/ivannovak/glide/pkg/plugin/sdk"
	"github.com/ivannovak/glide-plugin-docker/internal/commands"
	dockershell "github.com/ivannovak/glide-plugin-docker/internal/shell"
	"github.com/spf13/cobra"
)

// DockerPlugin implements the SDK Plugin interfaces for Docker functionality
type DockerPlugin struct {
	detector *DockerDetector
	executor *dockershell.DockerExecutorProvider
}

// New creates a new Docker plugin instance
func New() *DockerPlugin {
	return &DockerPlugin{
		detector: NewDockerDetector(),
		executor: dockershell.NewDockerExecutorProvider(),
	}
}

// NewDockerPlugin creates a new Docker plugin instance (legacy name)
func NewDockerPlugin() *DockerPlugin {
	return New()
}

// Name returns the plugin identifier
func (p *DockerPlugin) Name() string {
	return "docker"
}

// Version returns the plugin version
func (p *DockerPlugin) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (p *DockerPlugin) Description() string {
	return "Docker and Docker Compose integration for Glide"
}

// Register adds plugin commands to the command tree
func (p *DockerPlugin) Register(root *cobra.Command) error {
	// Get the command definitions from the SDK layer
	cmdDefs := p.ProvideCommands()

	// Convert and register each command with the root
	for _, cmdDef := range cmdDefs {
		if cmdDef != nil {
			cobraCmd := cmdDef.ToCobraCommand()

			// Wrap the command to inject project context
			// We need to do this because plugin commands don't have direct access to the app context
			p.wrapCommandWithContext(cobraCmd, root)

			root.AddCommand(cobraCmd)
		}
	}

	return nil
}

// wrapCommandWithContext wraps a command to inject project context from the root command
func (p *DockerPlugin) wrapCommandWithContext(cmd *cobra.Command, root *cobra.Command) {
	// Store the original RunE
	originalRunE := cmd.RunE
	if originalRunE == nil {
		return
	}

	// Wrap it to inject context
	cmd.RunE = func(c *cobra.Command, args []string) error {
		// Get the root command to access its context
		// The context should be set by the main CLI before execution
		rootCtx := c.Root().Context()
		if rootCtx != nil {
			// Set the context on this command
			c.SetContext(rootCtx)
		}

		// Call the original RunE
		return originalRunE(c, args)
	}
}

// Configure allows plugin-specific configuration
func (p *DockerPlugin) Configure(config map[string]interface{}) error {
	// Docker plugin doesn't require specific configuration yet
	// Future: Could add docker-compose path, default profiles, etc.
	return nil
}

// Metadata returns plugin information
func (p *DockerPlugin) Metadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "docker",
		Version:     "1.0.0",
		Author:      "Glide Team",
		Description: "Docker and Docker Compose integration for Glide",
		Aliases:     []string{},
		Commands: []plugin.CommandInfo{
			{
				Name:        "docker",
				Category:    "Development",
				Description: "Docker and Docker Compose commands",
				Aliases:     []string{"d"},
			},
		},
		BuildTags:  []string{},
		ConfigKeys: []string{"docker"},
	}
}

// ProvideContext returns the context extension for Docker detection
func (p *DockerPlugin) ProvideContext() sdk.ContextExtension {
	return p.detector
}

// ProvideCommands returns the commands provided by this plugin
func (p *DockerPlugin) ProvideCommands() []*sdk.PluginCommandDefinition {
	return []*sdk.PluginCommandDefinition{
		commands.NewDockerCommand(),
	}
}

// ProvideExecutor returns the executor provider for Docker commands
func (p *DockerPlugin) ProvideExecutor() *dockershell.DockerExecutorProvider {
	return p.executor
}
