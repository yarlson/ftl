<p>&nbsp;</p>
<p align="center">
  <img src="assets/logo.svg" alt="FTL logo">
</p>
<p>&nbsp;</p>

# FTL: faster than light deployment

FTL is a deployment tool that reduces complexity for projects that don't require extensive orchestration infrastructure. It provides automated deployment to cloud providers like Hetzner, DigitalOcean, Linode, and custom servers without the overhead of CI/CD pipelines or container orchestration platforms.

## Core Features

- Single YAML configuration file with environment variable substitution
- Zero-downtime deployments
- Automatic SSL/TLS certificate management
- Docker-based deployment with built-in health checks
- Integrated Nginx reverse proxy
- Multi-provider support (Hetzner, DigitalOcean, Linode, custom servers)

## Installation

1. Via Homebrew (macOS and Linux)

```bash
brew tap yarlson/ftl
brew install ftl
```

2. Download from GitHub releases

```bash
curl -L https://github.com/yarlson/ftl/releases/latest/download/ftl_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv ftl /usr/local/bin/
```

3. Or build from source

```bash
go install github.com/yarlson/ftl@latest
```

## Usage

1. Create configuration file:

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com

servers:
  - host: my-project.example.com
    port: 22
    user: my-project
    ssh_key: ~/.ssh/id_rsa

services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
      interval: 10s
      timeout: 5s
      retries: 3
    routes:
      - path: /
        strip_prefix: false

dependencies:
  - name: postgres
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_USER=${POSTGRES_USER:-postgres}
      - POSTGRES_DB=${POSTGRES_DB:-app}

volumes:
  - postgres_data
```

Environment variables in the configuration can be:

- Required: `${VAR_NAME}` - Must be set in the environment
- Optional with default: `${VAR_NAME:-default_value}` - Uses default if not set

2. Initialize server:

```bash
ftl setup
```

3. Deploy application:

```bash
ftl deploy
```

## How It Works

FTL manages deployments through these main components:

### Server Setup (`ftl setup`)

- Installs required packages (Docker, basic tools)
- Configures firewall rules
- Sets up user permissions
- Initializes Docker networks

### Deployment Process (`ftl deploy`)

1. Connects to configured servers via SSH
2. Pulls specified Docker images
3. Starts new containers with health checks
4. Configures Nginx reverse proxy
5. Manages SSL/TLS certificates via ACME
6. Performs zero-downtime container replacement
7. Cleans up unused resources

## Use Cases

### Suitable For

- Web applications with straightforward deployment needs
- Projects requiring automated SSL and reverse proxy setup
- Small to medium services running on single or multiple servers
- Teams seeking to minimize deployment infrastructure
- Applications requiring environment-specific configurations

### Not Designed For

- Complex microservice architectures requiring service mesh
- Systems needing advanced orchestration features
- Multi-region deployment coordination
- Specialized compliance environments

## Configuration Options

### Basic Structure

```yaml
project:
  name: string # Project identifier (required)
  domain: string # Primary domain (required, must be FQDN)
  email: string # Contact email (required, valid email format)

servers:
  - host: string # Server hostname/IP (required, FQDN or IP)
    port: int # SSH port (required, 1-65535)
    user: string # SSH user (required)
    ssh_key: string # Path to SSH key file (required)

services:
  - name: string # Service identifier (required)
    image: string # Docker image (required)
    port: int # Container port (required, 1-65535)
    path: string # Service path (default: "./" )
    command: string # Override container command
    entrypoint: [string] # Override container entrypoint
    health_check:
      path: string # Health check endpoint
      interval: duration # Time between checks
      timeout: duration # Check timeout
      retries: int # Number of retries
    routes:
      - path: string # Route path prefix (required)
        strip_prefix: bool # Strip prefix from requests
    volumes: [string] # Volume mappings (format: "volume:path")

dependencies:
  - name: string # Dependency name (required)
    image: string # Docker image (required)
    volumes: [string] # Volume mappings (format: "volume:path")
    env: # Environment variables
      KEY: value

volumes: [string] # Named volumes list
```

### Environment Variable Substitution

FTL supports two forms of environment variable substitution in the configuration:

1. Required variables: `${VAR_NAME}`

   - Must be present in the environment
   - Deployment fails if variable is not set

2. Variables with defaults: `${VAR_NAME:-default_value}`
   - Uses the environment variable if set
   - Falls back to the default value if not set

### Advanced Options

- Custom health check configurations
- Volume management
- Environment variable handling
- Service dependencies
- Custom routing rules

## Development

```bash
# Clone repository
git clone https://github.com/yarlson/ftl.git

# Install dependencies
cd ftl
go mod download

# Run tests
go test ./...
```

## Contributing

Contributions are welcome. Please ensure:

- Code follows project style
- Tests pass and new tests are added
- Documentation is updated

## License

[MIT License](LICENSE)
