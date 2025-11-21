# Glide Docker Plugin

External Docker plugin for Glide - provides Docker and Docker Compose integration.

## Overview

This plugin provides Docker and Docker Compose functionality for Glide, including:

- Automatic Docker Compose file resolution
- Container management
- Docker daemon health checking
- Interactive command support (exec, logs, etc.)
- Development mode awareness (single-repo vs multi-worktree)

## Installation

### Method 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/ivannovak/glide-plugin-docker
cd glide-plugin-docker

# Build the plugin
make build

# Install to PATH
sudo cp glide-plugin-docker /usr/local/bin/
```

### Method 2: Go Install (when published)

```bash
go install github.com/ivannovak/glide-plugin-docker/cmd/glide-plugin-docker@latest
```

## Usage

Once installed, the plugin provides the `docker` command to Glide:

```bash
# Start containers
glide docker up -d

# Stop containers
glide docker down

# View running containers
glide docker ps

# Execute commands in containers
glide docker exec php bash
glide docker exec -it php bash

# View logs
glide docker logs -f nginx

# Restart services
glide docker restart php

# Rebuild containers
glide docker build --no-cache
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Tidy dependencies
make tidy
```

### Current Status

**Phase 6: External Plugin Extraction** (In Progress)

The plugin structure is complete, but currently relies on Glide's internal packages via a local replace directive. To make this fully standalone:

1. Glide core needs to export necessary context types to public packages
2. Update plugin imports to use public SDK types
3. Remove the local replace directive

This will be completed in a future phase when the public SDK API is finalized.

## License

MIT
