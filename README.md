<p>&nbsp;</p>
<svg viewBox="0 0 392 200" xmlns="http://www.w3.org/2000/svg" style="height: 3em;">
<path d="m8 0c-4.4183 0-8 3.5817-8 8v184c0 4.418 3.5817 8 8 8h184c4.418 0 8-3.582 8-8v-184c0-4.4183-3.582-8-8-8h-184zm81.286 155.36 15.014 14.743 70.7-70.1-70.7-70.1-15.014 14.815 26.346 26.089h-90.632v20.958h111.8l8.32 8.2386-8.309 8.239h-111.81v20.957h90.671l-26.385 26.161z" clip-rule="evenodd" fill="currentColor" fill-rule="evenodd"/>
<path d="m292.47 70.927v15.909h-51.591v-15.909h51.591zm-38.693 87.273v-95.511c0-5.8713 1.212-10.758 3.636-14.659 2.462-3.9015 5.758-6.8182 9.886-8.75 4.129-1.9319 8.712-2.8978 13.75-2.8978 3.561 0 6.724 0.2841 9.489 0.8523s4.811 1.0796 6.136 1.5341l-4.091 15.909c-0.871-0.2652-1.969-0.5303-3.295-0.7955-1.326-0.303-2.803-0.4545-4.432-0.4545-3.826 0-6.534 0.928-8.125 2.7841-1.553 1.8182-2.329 4.4318-2.329 7.8409v94.148h-20.625z" fill="currentColor"/>
<path d="m352.16 70.927v15.909h-50.17v-15.909h50.17zm-37.784-20.909h20.568v81.932c0 2.765 0.417 4.886 1.25 6.364 0.872 1.439 2.008 2.424 3.409 2.954 1.402 0.531 2.955 0.796 4.66 0.796 1.287 0 2.462-0.095 3.522-0.284 1.099-0.19 1.932-0.36 2.5-0.512l3.466 16.08c-1.098 0.379-2.67 0.795-4.716 1.25-2.007 0.454-4.469 0.719-7.386 0.795-5.152 0.152-9.792-0.625-13.921-2.329-4.128-1.743-7.405-4.432-9.829-8.069-2.386-3.636-3.561-8.181-3.523-13.636v-85.341z" fill="currentColor"/>
<path d="m391.8 41.836v116.36h-20.568v-116.36h20.568z" fill="currentColor" />
</svg>

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

Build and deploy Docker images for your services. FTL offers two ways to handle images:

#### Direct SSH Transfer (Default)

When no `image` field is specified in your service configuration, FTL will:
- Build the image locally
- Transfer it directly to your server via SSH
- Use its own layer caching algorithm to optimize transfers
- Only transfer layers that haven't been previously sent to the server

```yaml
services:
  - name: web
    build:
      context: .
      dockerfile: Dockerfile
```

#### Registry-based Deployment

When you specify the `image` field, FTL will use a Docker registry:
- Build and tag the image locally
- Push it to the specified registry
- Pull the image on the server during deployment
- Require registry authentication during server setup (username/password only)

```yaml
services:
  - name: web
    image: registry.example.com/my-app:latest
    build:
      context: .
      dockerfile: Dockerfile
```

::: warning
Currently, FTL only supports registries with username/password authentication. Token-based authentication will fail.
:::

#### Build Command

```bash
ftl build [flags]
```

#### Flags

- `--skip-push`: Skip pushing images to the registry (only applies when using registry-based deployment)

#### Examples

- Build all services (using direct SSH transfer):
  ```bash
  ftl build
  ```
- Build all services but skip pushing to registry (when using registry-based deployment):
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
