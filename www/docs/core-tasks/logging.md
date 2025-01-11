---
title: Logging
description: Access and monitor logs from your deployed services
---

# Logging

FTL provides access to logs from your deployed services, allowing you to monitor and troubleshoot your applications.

## Basic Usage

View logs from all services:

```bash
ftl logs
```

## Command Options

### View Specific Service Logs

```bash
ftl logs [service]
```

### Command Flags

- `-f`, `--follow`: Stream logs in real-time
- `-n`, `--tail <lines>`: Number of lines to show from the end of the logs (default is 100 if `-f` is used)

## Examples

### View All Service Logs

```bash
# Show logs from all services
ftl logs
```

### Stream Specific Service Logs

```bash
# Stream logs from a specific service
ftl logs my-app -f
```

### Customize Log Output

```bash
# Show last 50 lines from all services
ftl logs -n 50

# Show last 150 lines from a specific service
ftl logs my-app -n 150
```

## Log Sources

FTL collects logs from:

1. **Application Services**

   - Main application containers
   - Custom service containers

2. **Dependencies**

   - Database containers
   - Cache services
   - Other supporting services

3. **System Services**
   - Nginx reverse proxy
   - SSL certificate management

## Best Practices

1. **Log Monitoring**

   - Use `-f` flag during deployments
   - Monitor application startup
   - Track dependency initialization

2. **Log Analysis**

   - Check logs after deployments
   - Monitor for error patterns
   - Review performance issues

3. **Troubleshooting**
   - Start with recent logs
   - Focus on specific services
   - Use appropriate line counts

## Common Issues

### No Logs Available

If no logs are displayed:

- Verify the service is running
- Check service name spelling
- Ensure deployment was successful

### Log Access Issues

If unable to access logs:

- Verify SSH connection
- Check server permissions
- Ensure service exists

## Next Steps

1. Learn about [Tunneling](./tunneling.md)
2. Explore [Health Checks](../guides/health-checks.md)
3. Review [Zero-downtime Deployments](../guides/zero-downtime.md)

::: tip
Use `ftl logs -f` during deployments to monitor the process in real-time.
:::

## Reference

- [CLI Commands Reference](../reference/cli-commands.md)
- [Troubleshooting Guide](../reference/troubleshooting.md)
