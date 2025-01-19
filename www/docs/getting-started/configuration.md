---
title: Configuring FTL
description: Learn how to configure FTL for your project using the ftl.yaml configuration file
---

# Configuring FTL

FTL uses a single YAML configuration file (`ftl.yaml`) to define your project's deployment settings. This guide walks you through creating and configuring the file to leverage defaults for common settings and dynamic environment variable expansion.

## Basic Configuration Structure

Create an `ftl.yaml` file in your project's root directory:

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com

server:
  host: example.com
  port: 22
  user: deploy
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
  name: my-project # Required: used for naming Docker networks and containers
  domain: example.com # Required: Primary domain for your application
  email: admin@example.com # Required: Contact email for notifications and SSL certificates
```

### Server Configuration

The `server` section specifies your deployment target:

```yaml
server:
  host: example.com
  port: 22
  user: deploy
  ssh_key: ~/.ssh/id_rsa
```

### Services

The `services` section defines the application services. For many services, you can simply specify the image (or even use environment variable substitution to choose the image) while relying on default settings for aspects like ports, health checks, and routes.

```yaml
services:
  - name: web # Required: Service name
    image: ${DOCKER_IMAGE:-my-project:latest} # Uses an environment variable with a default
    port: 80 # Required: The port exposed by the container
    health_check: # Optional but recommended
      path: / # Health check endpoint
      interval: 10s # Frequency of health checks
      timeout: 5s # Timeout for each check
      retries: 3 # Number of retries before considering the service unhealthy
    routes: # Required: HTTP routing rules
      - path: / # URL path to match
        strip_prefix: false # Whether to remove the prefix before forwarding requests
```

### Dependencies

Use the `dependencies` section to define supporting services—such as databases or caches—that your application requires. When you reference a service by a short identifier (for example, `"postgres:16"`, `"mysql:8"`, or `"redis"`), FTL automatically applies default configurations (including typical ports, named volumes, and environment variable substitutions). You can also override these defaults by specifying fields explicitly.

**Example using defaults:**

```yaml
dependencies:
  - "postgres:16"
```

This sets up a PostgreSQL dependency with:

- The image set to `postgres:16`
- A default named volume (e.g. `postgres_data:/var/lib/postgresql/data`)
- Preconfigured environment variables such as:
  ```yaml
  env:
    - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    - POSTGRES_USER=${POSTGRES_USER:-postgres}
    - POSTGRES_DB=${POSTGRES_DB:-app}
  ```

**Example with custom settings:**

```yaml
dependencies:
  - name: postgres
    image: postgres:16
    volumes:
      - my_pg_data:/custom/postgres/path
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-secret}
      - POSTGRES_USER=${POSTGRES_USER:-customuser}
```

### Volumes

Declare persistent storage volumes explicitly. FTL collects volume names from both services and dependencies.

```yaml
volumes:
  - postgres_data
```

## Environment Variables

FTL supports dynamic environment variable substitution throughout your configuration file. You can use the following patterns:

- **Required variables**:

  ```yaml
  env:
    - DATABASE_URL=${DATABASE_URL}
  ```

  FTL will fail if `DATABASE_URL` is not set.

- **Optional variables with defaults**:

  ```yaml
  env:
    - POSTGRES_USER=${POSTGRES_USER:-postgres}
  ```

  If `POSTGRES_USER` is not set, it uses `postgres`.

- **Custom required with message**:
  ```yaml
  env:
    - API_KEY=${API_KEY:?API_KEY must be set}
  ```

## Configuration Best Practices

1. **Health Checks**  
   Define health checks for all services to ensure their proper functioning:

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

   - Use environment variables to manage secrets.
   - Never commit actual secrets to version control.

3. **Organize Services**

   - Group related services together.
   - Use clear, descriptive names.
   - Document service dependencies and required environment variables.

4. **Volume Management**
   - Always declare volumes for persistent data.
   - Use meaningful volume names.
   - Plan for backups of persistent volumes.

## Validation

FTL validates your configuration during execution. Common checks include:

- Presence of required fields
- Valid port numbers and formats
- Proper URL and email formats
- Existence of the SSH key file
- Resolution of environment variable placeholders

## Next Steps

After configuring your project, you can:

1. Proceed to [First Deployment](./first-deployment.md) to deploy your application.
2. Learn about [Zero-downtime Deployments](/guides/zero-downtime).
3. Explore [SSL Management](/guides/ssl-management).

## Reference

For complete configuration options, see:

- [Configuration File Reference](/reference/configuration-file)
- [Environment Variables Reference](/reference/environment)
