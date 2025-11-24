# glide-plugin-docker

[![CI](https://github.com/ivannovak/glide-plugin-docker/actions/workflows/ci.yml/badge.svg)](https://github.com/ivannovak/glide-plugin-docker/actions/workflows/ci.yml)
[![Semantic Release](https://github.com/ivannovak/glide-plugin-docker/actions/workflows/semantic-release.yml/badge.svg)](https://github.com/ivannovak/glide-plugin-docker/actions/workflows/semantic-release.yml)

Docker and Docker Compose integration plugin for [Glide CLI](https://github.com/ivannovak/glide).

## Overview

This plugin provides Docker and Docker Compose functionality for Glide. When installed, Glide will automatically detect Docker projects and provide intelligent container management with automatic compose file resolution.

## Installation

### From GitHub Releases (Recommended)

```bash
glide plugins install github.com/ivannovak/glide-plugin-docker
```

### From Source

```bash
# Clone the repository
git clone https://github.com/ivannovak/glide-plugin-docker.git
cd glide-plugin-docker

# Build and install (requires Go 1.24+)
make install
```

## What It Detects

The plugin automatically detects Docker projects by looking for:

- **Compose files**: `docker-compose.yml`, `docker-compose.yaml`, `compose.yml`, `compose.yaml`
- **Override files**: `docker-compose.override.yml`, `docker-compose.override.yaml`
- **Dockerfile**: `Dockerfile`, `*.dockerfile`
- **Docker daemon status**: Checks if Docker is running

### Automatic Compose File Resolution

The plugin intelligently finds and uses Docker Compose files:

1. Searches for main compose file (`docker-compose.yml`, `compose.yml`, etc.)
2. Automatically includes override files if present
3. Handles both Docker Compose V1 and V2 syntax

## Available Commands

Once a Docker project is detected, the following command becomes available:

### Container Management
- `docker` (alias: `d`) - Pass-through to `docker compose` with automatic file resolution
  - Automatically detects and applies compose files
  - Supports all docker compose commands and flags
  - Interactive TTY support for exec and run commands

## Configuration

The plugin works out-of-the-box without configuration. However, you can customize behavior in your `.glide.yml`:

```yaml
plugins:
  docker:
    enabled: true
    # Additional configuration options can be added here in the future
```

## Examples

### Basic Docker Operations

```bash
# Start containers in detached mode
glide docker up -d

# Stop and remove containers
glide docker down

# View running containers
glide docker ps

# View all containers (including stopped)
glide docker ps -a

# Restart specific service
glide docker restart nginx

# View service logs
glide docker logs nginx

# Follow logs in real-time
glide docker logs -f php
```

### Interactive Commands

```bash
# Execute command in container
glide docker exec php ls -la

# Interactive shell
glide docker exec -it php bash

# Run one-off command
glide docker run --rm php php -v
```

### Building and Rebuilding

```bash
# Build or rebuild services
glide docker build

# Build with no cache
glide docker build --no-cache

# Build specific service
glide docker build nginx
```

### Advanced Operations

```bash
# Pull latest images
glide docker pull

# Show docker compose configuration
glide docker config

# Validate compose file
glide docker config --quiet

# Scale services
glide docker up -d --scale php=3

# Remove volumes
glide docker down -v
```

### Common Workflows

```bash
# Development workflow
glide docker up -d          # Start services
glide docker logs -f app    # Watch logs
glide docker exec -it app bash  # Enter container
glide docker down           # Stop services

# Debugging workflow
glide docker ps             # Check container status
glide docker logs app       # View logs
glide docker exec app env   # Check environment
glide docker restart app    # Restart service

# Clean rebuild
glide docker down -v        # Stop and remove volumes
glide docker build --no-cache  # Rebuild from scratch
glide docker up -d          # Start fresh
```

## Development

### Prerequisites

- Go 1.24 or higher
- Make (optional, for convenience targets)
- Docker and Docker Compose

### Building

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linters
make lint

# Format code
make fmt
```

### Testing

The plugin includes comprehensive tests for:

- Compose file detection
- Docker daemon status checking
- Command execution
- File resolution logic

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

## How It Works

The plugin provides a complete pass-through to `docker compose`, but adds intelligent features:

1. **Auto-detection**: Finds compose files in your project
2. **File Resolution**: Automatically applies the correct `-f` flags
3. **Interactive Support**: Detects when commands need TTY (exec, run)
4. **Error Handling**: Provides helpful error messages for common issues

All arguments are passed directly to `docker compose` without modification, ensuring full compatibility with native Docker Compose commands.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [Glide](https://github.com/ivannovak/glide) - The main Glide CLI
- [glide-plugin-go](https://github.com/ivannovak/glide-plugin-go) - Go plugin for Glide
- [glide-plugin-node](https://github.com/ivannovak/glide-plugin-node) - Node.js plugin for Glide
- [glide-plugin-php](https://github.com/ivannovak/glide-plugin-php) - PHP plugin for Glide

## Support

- [GitHub Issues](https://github.com/ivannovak/glide-plugin-docker/issues)
- [Glide Documentation](https://github.com/ivannovak/glide#readme)
- [Plugin Development Guide](https://github.com/ivannovak/glide/blob/main/docs/PLUGIN_DEVELOPMENT.md)
