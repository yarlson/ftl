---
title: Configuration File Reference
description: Complete specification of the FTL configuration file format and options
---

# Configuration File Reference

FTL uses a YAML configuration file (`ftl.yaml`) to define project settings, server configurations, services, dependencies, and volumes.

## File Structure

```yaml
project: # Project-level configuration
server: # Server definition
services: # Application services
dependencies: # Supporting services
volumes: # Persistent storage definitions
```

## Project Configuration

Top-level project settings that define basic information about your deployment.

```yaml
project:
  name: my-project # Required: Project identifier
  domain: my-project.example.com # Required: Primary domain for the project
  email: my-project@example.com # Required: Contact email for SSL certificates
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
  host: my-project.example.com # Required: Server hostname or IP
  port: 22 # Optional: SSH port (default: 22)
  user: my-project # Required: SSH user
  ssh_key: ~/.ssh/id_rsa # Required: Path to SSH private key
```

| Field     | Type    | Required | Default | Description                     |
| --------- | ------- | -------- | ------- | ------------------------------- |
| `host`    | string  | Yes      | -       | Server hostname or IP address   |
| `port`    | integer | No       | 22      | SSH port number                 |
| `user`    | string  | Yes      | -       | SSH username for authentication |
| `ssh_key` | string  | Yes      | -       | Path to SSH private key file    |

## Services

Defines the application services to be deployed.

```yaml
services:
  - name: my-app # Required: Service identifier
    image: my-app:latest # Optional: Docker image name
    build: # Optional: Build configuration
      context: . # Required for build: Build context path
      dockerfile: Dockerfile # Optional: Dockerfile path
    port: 80 # Required: Container port to expose
    health_check: # Optional: Health check configuration
      path: / # Required for health check: HTTP path
      interval: 10s # Optional: Check interval
      timeout: 5s # Optional: Check timeout
      retries: 3 # Optional: Number of retries
    routes: # Required: Routing configuration
      - path: / # Required: URL path to match
        strip_prefix: false # Optional: Strip path prefix
```

| Field          | Type    | Required | Default | Description                                |
| -------------- | ------- | -------- | ------- | ------------------------------------------ |
| `name`         | string  | Yes      | -       | Unique service identifier                  |
| `image`        | string  | No       | -       | Docker image for registry-based deployment |
| `build`        | object  | No       | -       | Build configuration for direct transfer    |
| `port`         | integer | Yes      | -       | Container port to expose                   |
| `health_check` | object  | No       | -       | Health check configuration                 |
| `routes`       | array   | Yes      | -       | Nginx routing configuration                |

## Dependencies

Defines supporting services required by the application.

```yaml
dependencies:
  - name: postgres # Required: Dependency identifier
    image: postgres:16 # Required: Docker image
    volumes: # Optional: Volume mounts
      - postgres_data:/var/lib/postgresql/data
    env: # Optional: Environment variables
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_USER=${POSTGRES_USER:-postgres}
      - POSTGRES_DB=${POSTGRES_DB:-app}
```

| Field     | Type   | Required | Description                      |
| --------- | ------ | -------- | -------------------------------- |
| `name`    | string | Yes      | Unique dependency identifier     |
| `image`   | string | Yes      | Docker image to use              |
| `volumes` | array  | No       | Volume mount definitions         |
| `env`     | array  | No       | Environment variable definitions |

## Volumes

Defines persistent storage volumes.

```yaml
volumes:
  - postgres_data # Volume name
```

Each entry in the `volumes` array is a string representing the volume name. These volumes can be referenced in service and dependency configurations.

## Environment Variables

The configuration file supports environment variable substitution in two formats:

1. Required variables:

   ```yaml
   ${VARIABLE_NAME}
   ```

2. Optional variables with defaults:
   ```yaml
   ${VARIABLE_NAME:-default_value}
   ```

## Complete Example

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com

server:
  host: my-project.example.com
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
