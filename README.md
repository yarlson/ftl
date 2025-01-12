# FTL (Faster Than Light) Deployment

FTL is a lightweight deployment tool designed to simplify cloud deployments without the complexity of traditional CI/CD pipelines or container orchestration platforms. It provides automated, zero-downtime deployments to various cloud providers through a single YAML configuration file.

For comprehensive documentation, visit [https://ftl-deploy.org](https://ftl-deploy.org)

## Features

- Zero-downtime deployments with automated health checks
- Single YAML configuration with environment variable support
- Built-in Nginx reverse proxy with automatic SSL/TLS certificate management
- Docker-based deployment with layer-optimized transfers
- Real-time log streaming and monitoring
- Secure SSH tunneling for remote dependencies

## Requirements

- Docker installed locally for building images
- SSH access to target deployment servers
- Git for version control
- Go 1.16+ (only if building from source)

## Installation

Choose one of the following installation methods:

### Via Homebrew (macOS and Linux)

```bash
brew tap yarlson/ftl
brew install ftl
```

### Direct Download

```bash
curl -L https://github.com/yarlson/ftl/releases/latest/download/ftl_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv ftl /usr/local/bin/
```

### Build from Source

```bash
go install github.com/yarlson/ftl@latest
```

### Verify Installation

After installing FTL, verify it's working correctly by checking the version:

```bash
ftl version
```

## Configuration

Create an `ftl.yaml` file in your project root:

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

### Environment Variables

- Required variables: Use `${VAR_NAME}`
- Optional variables with defaults: Use `${VAR_NAME:-default_value}`

## Usage

### Server Setup

```bash
ftl setup
```

### Building Applications

FTL supports two deployment modes:

1. Direct SSH Transfer (Default):

```yaml
services:
  - name: web
    build:
      context: .
      dockerfile: Dockerfile
```

2. Registry-based Deployment:

```yaml
services:
  - name: web
    image: registry.example.com/my-app:latest
    build:
      context: .
      dockerfile: Dockerfile
```

Build command:

```bash
ftl build [--skip-push]
```

### Deployment

```bash
ftl deploy
```

### Log Management

```bash
# Stream all logs
ftl logs -f

# View specific service logs
ftl logs my-app -n 150
```

### SSH Tunnels

```bash
# Create tunnels for all dependencies
ftl tunnels

# Connect to specific server
ftl tunnels --server my-project.example.com
```

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

## Example Projects

Visit our [ftl-examples](https://github.com/yarlson/ftl-examples) repository for complete implementation examples:

- [Flask Application with PostgreSQL](https://github.com/yarlson/ftl-examples/tree/main/flask)
- Additional examples coming soon

## Troubleshooting

### Common Issues

1. Registry Authentication Failures

   - FTL currently supports only username/password authentication
   - Token-based authentication is not supported

2. SSH Connection Issues
   - Verify SSH key permissions
   - Ensure server firewall allows connections
   - Check user permissions on target server

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

Please ensure:

- Code follows project style guidelines
- All tests pass
- Documentation is updated
- Commit messages are clear and descriptive

## Security

Report security vulnerabilities by opening an issue with the "security" label. We take all security reports seriously and will respond promptly.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
