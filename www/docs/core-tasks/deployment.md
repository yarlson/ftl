---
title: Deployment
description: Learn how to deploy applications with FTL
---

# Deployment

FTL automates the deployment of your applications, handling everything from image transfers to SSL certificate management.

## Basic Usage

To deploy your application:

```bash
ftl deploy
```

## Deployment Process

The deployment process follows these steps:

1. **SSH Connection**

   - Connects to your configured server(s) via SSH
   - Verifies server access and permissions

2. **Image Handling**

   - For direct SSH transfer (no `image` field):
     - Transfers Docker images directly to servers
     - Uses FTL's layer caching for optimization
   - For registry-based deployment (`image` field specified):
     - Pulls images from configured registry
     - Requires registry authentication (username/password only)

3. **Environment Setup**

   - Sets up volumes and networks
   - Configures environment variables
   - Starts dependencies if defined

4. **Service Deployment**

   - Launches services with health checks
   - Performs zero-downtime container replacement
   - Configures Nginx reverse proxy

5. **SSL/TLS Setup**

   - Manages SSL/TLS certificates via ACME
   - Configures HTTPS endpoints

6. **Cleanup**
   - Removes unused containers
   - Cleans up temporary resources

## Configuration

### Basic Service Configuration

```yaml
services:
  - name: web
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

### Dependencies Configuration

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

### Volume Configuration

```yaml
volumes:
  - postgres_data
```

## Environment Variables

FTL supports two types of environment variables:

1. **Required Variables**

   ```yaml
   env:
     - DATABASE_URL=${DATABASE_URL}
   ```

2. **Optional Variables with Defaults**
   ```yaml
   env:
     - POSTGRES_USER=${POSTGRES_USER:-postgres}
   ```

## Health Checks

Health checks ensure your services are running correctly:

```yaml
services:
  - name: api
    health_check:
      path: /health
      interval: 10s
      timeout: 5s
      retries: 3
```

## Best Practices

1. **Configuration Management**

   - Use environment variables for sensitive data
   - Keep configuration DRY
   - Document all required variables

2. **Deployment Strategy**

   - Start with dependencies first
   - Use health checks for all services
   - Monitor logs during deployment

3. **Security**

   - Use HTTPS for all services
   - Keep SSL certificates up to date
   - Secure sensitive environment variables

4. **Monitoring**
   - Check service health after deployment
   - Monitor resource usage
   - Keep track of logs

## Common Issues

### Health Check Failures

If services fail health checks:

- Verify the health check endpoint
- Check service logs
- Ensure correct port configuration
- Verify service startup time

### SSL Certificate Issues

If SSL setup fails:

- Verify domain DNS configuration
- Check email configuration
- Ensure ports 80/443 are accessible

### Network Issues

If services can't communicate:

- Check network configuration
- Verify port mappings
- Ensure dependencies are running

## Next Steps

After successful deployment:

1. Learn about [Logging](./logging.md)
2. Configure [SSL Management](../guides/ssl-management.md)
3. Explore [Zero-downtime Deployments](../guides/zero-downtime.md)

::: warning
Always verify your application's functionality after deployment.
:::

## Reference

- [Configuration Reference](../reference/configuration-file.md)
- [CLI Commands Reference](../reference/cli-commands.md)
- [Troubleshooting Guide](../reference/troubleshooting.md)
