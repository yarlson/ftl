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
- Establish SSH tunnels to remote dependencies

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

### 3. Build Application Images

Build Docker images for application services:

```bash
ftl build [flags]
```

#### Flags

- `--skip-push`: Skip pushing images to the registry after building.

#### Examples

- Build and push all services:
  ```bash
  ftl build
  ```
- Build all services but skip pushing to the registry:
  ```bash
  ftl build --skip-push
  ```

### 4. Deploy Application

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

### 5. Fetch Logs

Retrieve logs from your deployed services:

```bash
ftl logs [service] [flags]
```

#### Flags

- `-f`, `--follow`: Stream logs in real-time.
- `-n`, `--tail <lines>`: Number of lines to show from the end of the logs (default is 100 if `-f` is used).

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
- Fetch logs from a specific service with a custom tail size:
  ```bash
  ftl logs my-app -n 150
  ```

### 6. Create SSH Tunnels

Establish SSH tunnels for your dependencies, allowing local access to services running on your server:

```bash
ftl tunnels [flags]
```

#### Flags

- `-s`, `--server <server>`: (Optional) Specify the server name or index to connect to, if multiple servers are defined.

#### Examples

- Establish tunnels to all dependency ports:
  ```bash
  ftl tunnels
  ```
- Specify a server to connect to (if multiple servers are configured):
  ```bash
  ftl tunnels --server my-project.example.com
  ```

#### Purpose

The `ftl tunnels` command is useful for:

- Accessing dependency services (e.g., databases) running on your server from your local machine
- Simplifying local development by connecting to remote services without modifying your code
- Testing and debugging your application against live dependencies

## Additional Notes

### Error Handling

All commands include detailed error reporting and user feedback through spinners and console messages. Examples:

- Commands gracefully handle configuration file parsing issues.
- Detailed error messages are provided for server connection or dependency issues.

### Concurrency

Commands like `build` and `tunnels` leverage concurrent operations to improve performance. For example:

- `ftl build` builds and optionally pushes images for all services concurrently.
- `ftl tunnels` establishes SSH tunnels for multiple dependencies simultaneously.

### Configuration Highlights

To ensure optimal usage:

- Ensure all dependencies have `ports` specified in `ftl.yaml` for `ftl tunnels` to function.
- Use health checks in service definitions to ensure reliability during deployment and build processes.

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
