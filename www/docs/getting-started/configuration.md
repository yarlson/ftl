---
title: Configuring FTL
description: Learn how to configure FTL for your project using the ftl.yaml configuration file
---

# Configuring FTL

FTL uses a single YAML configuration file (`ftl.yaml`) to define your project's deployment settings. This guide will walk you through creating and configuring this file.

## Basic Configuration Structure

Create an `ftl.yaml` file in your project's root directory:

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
  - name: web
    image: my-project:latest
    port: 80
    health_check:
      path: /
      interval: 10s
      timeout: 5s
      retries: 3
    routes:
      - path: /
        strip_prefix: false
```

## Essential Configuration Sections

### Project Settings

The `project` section defines your project's basic information:

```yaml
project:
  name: my-project # Required: Project name (used for Docker networks and containers)
  domain: example.com # Required: Primary domain for your application
  email: admin@example.com # Required: Email for SSL certificate notifications
```

### Server Configuration

The `servers` section specifies your deployment targets:

```yaml
servers:
  - host: example.com # Required: Server hostname or IP
    port: 22 # Optional: SSH port (defaults to 22)
    user: deploy # Required: SSH user
    ssh_key: ~/.ssh/id_rsa # Required: Path to SSH private key
```

::: tip Multiple Servers
You can define multiple servers for load balancing or high availability:

```yaml
servers:
  - host: server1.example.com
    user: deploy
    ssh_key: ~/.ssh/id_rsa
  - host: server2.example.com
    user: deploy
    ssh_key: ~/.ssh/id_rsa
```

:::

### Services

The `services` section defines your application services:

```yaml
services:
  - name: web # Required: Service name
    image: my-app:latest # Required: Docker image
    port: 80 # Required: Container port to expose
    health_check: # Optional but recommended
      path: /health # Health check endpoint
      interval: 10s # Check interval
      timeout: 5s # Check timeout
      retries: 3 # Retries before marking unhealthy
    routes: # Required: HTTP routing rules
      - path: / # URL path to match
        strip_prefix: false # Whether to remove prefix before forwarding
```

### Dependencies

Define service dependencies like databases:

```yaml
dependencies:
  - name: postgres
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_USER=${POSTGRES_USER:-postgres}
      - POSTGRES_DB=${POSTGRES_DB:-app}
```

### Volumes

Declare persistent volumes:

```yaml
volumes:
  - postgres_data # Volume name
```

## Environment Variables

FTL supports environment variable substitution in your configuration:

- **Required variables**: `${VAR_NAME}`
- **Optional variables with defaults**: `${VAR_NAME:-default_value}`

Example usage:

```yaml
services:
  - name: web
    image: ${DOCKER_IMAGE:-my-app:latest}
    env:
      - DATABASE_URL=${DATABASE_URL}
      - API_KEY=${API_KEY:-development-key}
```

## Configuration Best Practices

1. **Use Health Checks**

   ```yaml
   services:
     - name: api
       health_check:
         path: /health
         interval: 10s
         timeout: 5s
         retries: 3
   ```

2. **Secure Sensitive Data**

   - Use environment variables for sensitive information
   - Never commit actual secrets to version control

3. **Organize Services**

   - Group related services together
   - Use clear, descriptive names
   - Document service dependencies

4. **Volume Management**
   - Always declare volumes for persistent data
   - Use meaningful volume names
   - Consider backup strategies

## Validation

FTL validates your configuration when running commands. Common validation checks include:

- Required fields presence
- Valid port numbers
- Proper URL formats
- SSH key file existence
- Environment variable resolution

## Next Steps

After configuring your project:

1. Move on to [First Deployment](./first-deployment.md) to deploy your application
2. Learn about [Zero-downtime Deployments](/guides/zero-downtime)
3. Explore [SSL Management](/guides/ssl-management)

## Reference

For complete configuration options, see:

- [Configuration File Reference](/reference/configuration-file)
- [Environment Variables Reference](/reference/environment)
