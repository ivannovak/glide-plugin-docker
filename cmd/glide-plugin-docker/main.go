package main

import (
	"os"

	"github.com/ivannovak/glide-plugin-docker/internal/plugin"
	sdk "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
)

func main() {
	// Initialize the Docker gRPC plugin
	dockerPlugin := plugin.NewGRPCPlugin()

	// Run the plugin using the SDK
	if err := sdk.RunPlugin(dockerPlugin); err != nil {
		os.Exit(1)
	}
}
