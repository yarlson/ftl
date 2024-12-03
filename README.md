<p>&nbsp;</p>
<p align="center">
  <img src="assets/logo.svg" alt="FTL logo">
</p>
<p>&nbsp;</p>

# FTL: Faster Than Light Deployment

FTL is a deployment tool that reduces complexity for projects that don't require extensive orchestration infrastructure. It provides automated deployment to cloud providers like Hetzner, DigitalOcean, Linode, and custom servers without the overhead of CI/CD pipelines or container orchestration platforms.

## Core Features

- Single YAML configuration file with environment variable substitution
- Zero-downtime deployments
- Automatic SSL/TLS certificate management
- Docker-based deployment with built-in health checks
- Integrated Nginx reverse proxy
- Multi-provider support (Hetzner, DigitalOcean, Linode, custom servers)
- Fetch and stream logs from deployed services

## Installation

1. **Via Homebrew (macOS and Linux)**

   ```bash
   brew tap yarlson/ftl
   brew install ftl
   ```

2. **Download from GitHub releases**

   ```bash
   curl -L https://github.com/yarlson/ftl/releases/latest/download/ftl_$(uname -s)_$(uname -m).tar.gz | tar xz
   sudo mv ftl /usr/local/bin/
   ```

3. **Build from source**

   ```bash
   go install github.com/yarlson/ftl@latest
   ```

## Usage

### 1. Create Configuration File

Create an `ftl.yaml` configuration file in your project directory:

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

- **Required**: `${VAR_NAME}` - Must be set in the environment
- **Optional with default**: `${VAR_NAME:-default_value}` - Uses default if not set

### 2. Initialize Server

Set up your server with the required dependencies:

```bash
ftl setup
```

This command will:

- Install Docker and other necessary packages on your server
- Configure firewall rules
- Set up user permissions
- Initialize Docker networks

### 3. Deploy Application

Deploy your application to the configured servers:

```bash
ftl deploy
```

This command will:

- Connect to your servers via SSH
- Pull Docker images specified in your configuration
- Start new containers with health checks
- Configure the Nginx reverse proxy
- Manage SSL/TLS certificates via ACME
- Perform zero-downtime container replacement
- Clean up unused resources

### 4. Fetch Logs

Retrieve logs from your deployed services:

```bash
ftl logs [service] [flags]
```

- **service**: (Optional) Name of the service to fetch logs from. If omitted, logs from all services are fetched.
- **flags**:
  - `-f`, `--follow`: Stream logs in real-time.
  - `-n`, `--tail`: Number of lines to show from the end of the logs.

#### Examples

- Fetch logs from all services:

  ```bash
  ftl logs
  ```

- Stream logs from a specific service:

  ```bash
  ftl logs my-app -f
  ```

- Fetch the last 50 lines of logs from all services:

  ```bash
  ftl logs -n 50
  ```

## How It Works

FTL manages deployments and log retrieval through these main components:

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

### Logs Retrieval (`ftl logs`)

- Fetches logs from specified services
- Supports real-time streaming with the `-f` flag
- Allows limiting the number of log lines with the `-n` flag

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
    path: string # Service path (default: "./")
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
    env: # Environment variables
      - KEY=value

dependencies:
  - name: string # Dependency name (required)
    image: string # Docker image (required)
    volumes: [string] # Volume mappings (format: "volume:path")
    env: # Environment variables
      - KEY=value

volumes: [string] # Named volumes list
```

### Environment Variable Substitution

FTL supports two forms of environment variable substitution in the configuration:

1. **Required Variables**: `${VAR_NAME}`

   - Must be present in the environment
   - Deployment fails if variable is not set

2. **Variables with Defaults**: `${VAR_NAME:-default_value}`

   - Uses the environment variable if set
   - Falls back to the default value if not set

### Advanced Options

- **Health Checks**: Customize health check endpoints, intervals, timeouts, and retries for each service.
- **Volume Management**: Define named volumes for persistent data storage.
- **Environment Variables**: Set environment variables for services and dependencies, with support for environment variable substitution.
- **Service Dependencies**: Specify dependent services and their configurations.
- **Routing Rules**: Define custom routing paths and whether to strip prefixes.

## Example Projects

The [ftl-examples](https://github.com/yarlson/ftl-examples) repository contains reference implementations:

- [Flask](https://github.com/yarlson/ftl-examples/tree/main/flask) - Python Flask application with PostgreSQL
- More examples coming soon

Each example provides a complete project structure with configuration files and deployment instructions.

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

- Code follows project style guidelines
- Tests pass and new tests are added for new features
- Documentation is updated accordingly

## License

[MIT License](LICENSE)
