---
title: Tunneling
description: Create SSH tunnels to access remote services during development
---

# Tunneling

FTL can establish SSH tunnels to your remote dependencies, allowing local access to services running on your servers.

## Basic Usage

Create tunnels to all dependency ports:

```bash
ftl tunnels
```

## Command Options

### Server Selection

When multiple servers are configured, specify the target server:

```bash
ftl tunnels --server my-project.example.com
```

### Command Flags

- `-s`, `--server <server>`: Specify the server name or index to connect to

## Purpose

SSH tunneling is useful for:

1. **Local Development**

   - Connect to remote databases
   - Access remote services
   - Test against production-like environments

2. **Debugging**

   - Inspect database contents
   - Monitor service metrics
   - Troubleshoot issues

3. **Database Management**
   - Run migrations
   - Perform backups
   - Import/export data

## Example Usage

### Basic Configuration

```yaml
dependencies:
  - name: postgres
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
```

### Accessing Services

After establishing tunnels:

```bash
# Connect to PostgreSQL database
psql -h localhost -p <mapped_port> -U postgres
```

## Best Practices

1. **Security**

   - Use tunnels only during development
   - Close tunnels when not needed
   - Keep credentials secure

2. **Usage**

   - Document mapped ports
   - Use consistent port mappings
   - Monitor tunnel status

## Common Issues

### Connection Failed

If tunnels fail to establish:

- Verify SSH access to server
- Check server configuration
- Ensure ports are available locally

### Service Unreachable

If services are inaccessible through tunnel:

- Verify service is running
- Check port mappings
- Ensure dependencies are healthy

## Next Steps

1. Configure [Health Checks](../guides/health-checks.md)
2. Learn about [SSL Management](../guides/ssl-management.md)

::: tip
Use tunnels during local development to work with your remote services seamlessly.
:::

## Reference

- [CLI Commands Reference](../reference/cli-commands.md)
- [Troubleshooting Guide](../reference/troubleshooting.md)
