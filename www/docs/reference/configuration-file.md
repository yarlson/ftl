---
title: Configuration File Reference
description: Complete specification of the FTL configuration file format and options
---

# Configuration File Reference

FTL uses a YAML configuration file (`ftl.yaml`) to define project settings, server configurations, services, dependencies, and volumes. FTL automatically applies default settings—such as standard ports, volume mappings, and environment variables—for common services when you provide a short image reference. You can override these defaults by explicitly specifying settings.

## File Structure

```yaml
project: # Project-level configuration
server: # Server definition
services: # Application services
dependencies: # Supporting services
volumes: # Persistent storage definitions
```

## Project Configuration

Top-level project settings define basic information about your deployment.

```yaml
project:
  name: my-project # Required: Project identifier used for resource naming
  domain: my-project.example.com # Required: Primary domain for the deployment
  email: my-project@example.com # Required: Contact email for SSL certificate notifications
```

| Field    | Type   | Required | Description                                       |
| -------- | ------ | -------- | ------------------------------------------------- |
| `name`   | string | Yes      | Project identifier used for resource naming       |
| `domain` | string | Yes      | Primary domain for the deployment                 |
| `email`  | string | Yes      | Contact email used for SSL certificate management |

## Server Configuration

Defines the target server for deployment.

```yaml
server:
  host: my-project.example.com # Required: Server hostname or IP address
  port: 22 # Optional: SSH port (default: 22)
  user: my-project # Optional: SSH username (default: current user)
  ssh_key: ~/.ssh/id_rsa # Required: Path to SSH private key file
```

| Field     | Type    | Required | Default | Description                      |
| --------- | ------- | -------- | ------- | -------------------------------- |
| `host`    | string  | Yes      | -       | Server hostname or IP address    |
| `port`    | integer | No       | 22      | SSH port number                  |
| `user`    | string  | No       | Current user | SSH username for authentication  |
| `ssh_key` | string  | Yes      | -       | Path to the SSH private key file |

## Services

Defines the application services to be deployed. Each service must have either a path to the source code or a Docker image reference.

```yaml
services:
  - name: my-app # Required: Unique service identifier
    path: ./src # Required if no image: Path to source code and Dockerfile
    image: my-app:latest # Required if no path: Docker image used for deployment
    port: 80 # Required: Container port to expose
    health_check: # Optional: Health check settings
      path: / # Required for health check: HTTP path to check
      interval: 15s # Optional: Health check interval (default: 15s)
      timeout: 10s # Optional: Health check timeout (default: 10s)
      retries: 3 # Optional: Number of health check retries (default: 3)
    routes: # Required: HTTP routing configuration
      - path: / # Required: URL path to match
        strip_prefix: false # Optional: Strip path prefix when proxying (default: false)
```

| Field          | Type    | Required | Default | Description                                                                |
| -------------- | ------- | -------- | ------- | -------------------------------------------------------------------------- |
| `name`         | string  | Yes      | -       | Unique service identifier                                                  |
| `path`         | string  | Yes\*    | -       | Path to source code directory containing Dockerfile (relative to ftl.yaml) |
| `image`        | string  | Yes\*    | -       | Docker image for deployment (can include environment substitutions)        |
| `port`         | integer | Yes      | -       | Container port to expose                                                   |
| `health_check` | object  | No       | -       | Health check configuration                                                 |
| `routes`       | array   | Yes      | -       | Routing configuration for the reverse proxy                                |

\*Either `path` or `image` must be specified, but not both.

## Dependencies

Defines supporting services (such as databases, caches, or message queues) that your application requires. Dependencies can be declared in two ways:

### 1. Short Notation

Simply specify the service name and optional version. FTL will automatically apply default settings:

```yaml
dependencies:
  - "postgres:16" # Uses default settings for PostgreSQL 16
  - "redis:7" # Uses default settings for Redis 7
  - "mysql:8" # Uses default settings for MySQL 8
```

When using short notation, FTL automatically configures:

- Default image tag (modified by the specified version)
- Standard ports (e.g., 5432 for PostgreSQL)
- Default volume mappings
- Common environment variables with defaults

### 2. Detailed Definition

For custom configurations or overriding defaults:

```yaml
dependencies:
  - name: postgres # Required: Unique identifier for the dependency
    image: postgres:16 # Required: Docker image reference
    volumes: # Optional: Override default volume mappings
      - my_pg_data:/custom/postgres/path
    env: # Optional: Override default environment variables
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-secret}
      - POSTGRES_USER=${POSTGRES_USER:-myuser}
      - POSTGRES_DB=${POSTGRES_DB:-app}
```

| Field     | Type   | Required | Description                                             |
| --------- | ------ | -------- | ------------------------------------------------------- |
| `name`    | string | Yes\*    | Unique dependency identifier                            |
| `image`   | string | Yes\*    | Docker image used for the dependency                    |
| `volumes` | array  | No       | Volume mount definitions                                |
| `env`     | array  | No       | Environment variable definitions (supporting expansion) |

\*Only required when using detailed definition. For short notation, these are derived from the service string.

## Volumes

Defines persistent storage volumes for your deployment. Each entry in the `volumes` array is a string representing the volume name.

```yaml
volumes:
  - postgres_data # Volume name that can be referenced elsewhere
```

## Environment Variables

FTL supports environment variable substitution throughout the configuration. You can use the following formats:

1. **Required variables:**

   ```yaml
   ${VARIABLE_NAME}
   ```

2. **Optional variables with default values:**
   ```yaml
   ${VARIABLE_NAME:-default_value}
   ```

Example usage:

```yaml
services:
  - name: my-app
    image: ${DOCKER_IMAGE:-my-app:latest}
    env:
      - DATABASE_URL=${DATABASE_URL}
      - API_KEY=${API_KEY:-development-key}
```

## Complete Example

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com

server:
  host: my-project.example.com
  user: my-project
  ssh_key: ~/.ssh/id_rsa

services:
  - name: web-app
    path: ./src
    port: 80
    health_check:
      path: /
    routes:
      - path: /

  - name: admin-app
    image: ${DOCKER_IMAGE:-admin-app:latest}
    port: 80
    health_check:
      path: /health
      interval: 15s
      timeout: 10s
      retries: 3
    routes:
      - path: /admin
        strip_prefix: false

dependencies:
  - "postgres:16" # Using short notation with defaults
  - name: redis # Using detailed definition for customization
    image: redis:7
    volumes:
      - redis_data:/custom/redis/path
    env:
      - REDIS_PASSWORD=${REDIS_PASSWORD:-secret}

volumes:
  - postgres_data
```

## Summary

- **Default Settings:**  
  FTL automatically applies default settings like ports, volumes, and environment variable substitutions when you specify a service or dependency with a short Docker image reference.

- **Environment Variable Expansion:**  
  Configuration supports dynamic substitution for both required and optional variables using the `${...}` syntax.

- **Customization:**  
  Override any default setting by explicitly specifying the corresponding configuration fields.

For additional details, refer to:

- [Environment Variables Reference](/reference/environment)
- [CLI Commands Reference](/reference/cli-commands)
