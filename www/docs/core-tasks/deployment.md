---
title: Deployment
description: Learn how to deploy applications with FTL
---

# Deployment

FTL automates the deployment of your applications by taking care of image transfers, volume and network setups, SSL certificate management, and much more.

## Basic Usage

To deploy your application:

```bash
ftl deploy
```

## Deployment Process

FTL follows these steps when deploying an application:

1. **SSH Connection**
   - Connects to your configured server via SSH.
   - Verifies server access and permissions.

2. **Image Handling**
   - **Direct Transfer** (when no `image` field is specified):
     - Verifies server access and permissions.
     - Transfers Docker images directly to the server.
     - Uses layer caching for faster transfers.
   - **Registry Deployment** (when an `image` field is provided):
     - Pulls images from the specified registry.
     - Requires registry authentication (username/password).

3. **Environment Setup**
   - Creates and attaches named volumes.
   - Configures networks.
   - Expands and injects environment variables.
   - Starts any supporting services defined as dependencies.

4. **Service Deployment**
   - Launches application services with health checks.
   - Performs zero-downtime container replacements.
   - Configures the Nginx reverse proxy for routing.

5. **SSL/TLS Setup**
   - Manages SSL/TLS certificates via ACME.
   - Configures HTTPS endpoints for your application.

6. **Cleanup**
   - Removes unused containers.
   - Cleans up temporary resources.

## Configuration

FTL uses a YAML configuration file where you define your project, server, services, dependencies, and volumes. FTL supports default settings and dynamic expansion of environment variables. Below are examples for various sections.

### Basic Service Configuration

```yaml
project:
  name: my-project
  domain: example.com
  email: admin@example.com

server:
  host: example.com
  user: deploy
  ssh_key: ~/.ssh/id_rsa

services:
  - name: web
    port: 80
    path: ./src
    health_check:
      path: /
    routes:
      - path: /
```

### Dependencies Configuration

Dependencies are used to launch supporting services (for example, databases and caches). You can either use a short notation to take advantage of default settings or provide a full configuration when you need to override defaults.

**Using default settings:**

```yaml
dependencies:
  - "postgres:16"
```

In the above example, FTL will automatically:

- Set the dependency’s name to `postgres`
- Use the Docker image `postgres:16`
- Configure standard ports and named volumes (e.g. `postgres_data:/var/lib/postgresql/data`)
- Expand environment variables such as:
  ```yaml
  env:
    - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    - POSTGRES_USER=${POSTGRES_USER:-postgres}
    - POSTGRES_DB=${POSTGRES_DB:-app}
  ```

**Overriding defaults:**

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

When you override defaults, FTL uses your provided values rather than applying the built-in settings.

### Volume Configuration

Named volumes can be provided explicitly. FTL collects volume names from your services and dependencies.

```yaml
volumes:
  - postgres_data
```

### Environment Variables

FTL supports environment variable expansion in all configuration fields. You can define variables as follows:

- **Required Variables:**

  ```yaml
  env:
    - DATABASE_URL=${DATABASE_URL}
  ```

- **Optional Variables with Defaults:**

  ```yaml
  env:
    - POSTGRES_USER=${POSTGRES_USER:-postgres}
  ```

- **Required with Custom Message:**

  ```yaml
  env:
    - API_KEY=${API_KEY:?API_KEY must be set}
  ```

## Health Checks

Health checks are critical for ensuring that your services are running properly. For example:

```yaml
services:
  - name: api
    health_check:
      path: /health
```

## Best Practices

1. **Configuration Management**
   - Use environment variables to securely manage sensitive data.
   - Keep configuration DRY and well-documented.
   - Document all required environment variables.

2. **Deployment Strategy**
   - Start dependent services before deploying the main application.
   - Define health checks for every service.
   - Monitor logs during deployment for any issues.

3. **Security**
   - Use HTTPS for all external communications.
   - Regularly update and manage SSL certificates.
   - Secure sensitive environment variables with proper expansion patterns.

4. **Monitoring**
   - Verify service health post-deployment.
   - Monitor application resource usage.
   - Review logs to proactively address issues.

## Common Issues

### Health Check Failures

If services fail health checks:

- Verify the health check endpoint is correct.
- Inspect service logs.
- Ensure correct port and path configurations.
- Adjust startup time if necessary.

### SSL Certificate Issues

If SSL/TLS setup fails:

- Ensure DNS records are correct.
- Verify email configuration.
- Confirm ports 80 and 443 are accessible.

### Network Issues

If services cannot communicate:

- Check your network configuration.
- Verify volume mappings and port exposures.
- Ensure all required services are running.

## Next Steps

After deployment:

1. Learn about [Logging](./logging.md)
2. Configure [SSL Management](../guides/ssl-management.md)
3. Explore [Zero-Downtime Deployments](../guides/zero-downtime.md)

::: warning
Always verify your application’s functionality after deployment.
:::

## Reference

- [Configuration Reference](../reference/configuration-file.md)
- [CLI Commands Reference](../reference/cli-commands.md)
- [Troubleshooting Guide](../reference/troubleshooting.md)
