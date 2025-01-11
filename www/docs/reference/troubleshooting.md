---
title: Troubleshooting
description: Common issues and their solutions when working with FTL
---

# Troubleshooting

This guide covers common issues you might encounter when using FTL and their solutions.

## Build Issues

### Registry Authentication Failures

**Problem**: Unable to push images to registry

```bash
Error: authentication required
```

**Solution**:

- Ensure you're using username/password authentication (token-based auth is not supported)
- Verify your registry credentials are correct
- Try logging in manually with `docker login`

### Image Build Failures

**Problem**: Direct SSH transfer fails

```bash
Error: failed to transfer image layers
```

**Solution**:

- Check SSH connectivity to your server
- Verify SSH key permissions (should be 600)
- Ensure sufficient disk space on both local and remote machines

## Deployment Issues

### Health Check Failures

**Problem**: Service fails to start due to failed health checks

```bash
Error: health check failed after 3 retries
```

**Solution**:

- Verify the health check path is correct in your `ftl.yaml`
- Ensure your application is properly handling health check requests
- Check the service logs for application-specific errors
- Adjust health check timing parameters if needed

### SSL/TLS Certificate Issues

**Problem**: Unable to obtain SSL certificate

```bash
Error: failed to obtain SSL certificate
```

**Solution**:

- Verify DNS records are properly configured
- Ensure the domain points to your server's IP
- Check that port 80 is accessible (required for ACME challenges)
- Verify the email address in your project configuration

## Networking Issues

### SSH Connection Problems

**Problem**: Unable to connect to server

```bash
Error: ssh: connect to host example.com port 22: Connection refused
```

**Solution**:

- Verify server SSH configuration
- Check firewall rules
- Ensure correct SSH key path in `ftl.yaml`
- Verify server hostname/IP and port

### Reverse Proxy Issues

**Problem**: Services not accessible through domain

```bash
Error: 502 Bad Gateway
```

**Solution**:

- Verify service is running (`ftl logs <service>`)
- Check service port configuration
- Ensure routes are properly configured in `ftl.yaml`
- Verify Nginx configuration was properly generated

## Volume Issues

### Permission Problems

**Problem**: Container can't write to volume

```bash
Error: permission denied
```

**Solution**:

- Check volume ownership and permissions
- Verify volume path in configuration
- Ensure volume is properly mounted

### Missing Data

**Problem**: Data not persisting between deployments

```bash
Error: volume not found
```

**Solution**:

- Verify volume is defined in `volumes` section
- Check volume mount configuration
- Ensure volumes aren't being pruned accidentally

## Environment Variables

### Missing Required Variables

**Problem**: Deployment fails due to missing variable

```bash
Error: Required environment variable 'POSTGRES_PASSWORD' is not set
```

**Solution**:

- Set required environment variables before running FTL commands
- Use environment file: `source .env && ftl deploy`
- Verify variable names match your configuration

### Default Value Issues

**Problem**: Unexpected default values being used

```bash
Warning: using default value for POSTGRES_USER
```

**Solution**:

- Check environment variable syntax in `ftl.yaml`
- Verify variable export in your shell
- Review default values in configuration

## Common Commands for Troubleshooting

```bash
# Check service logs
ftl logs <service>

# Verify server connectivity
ftl setup

# Rebuild and redeploy service
ftl build && ftl deploy
```

## Getting Help

If you encounter an issue not covered here:

1. Check the service logs using `ftl logs`
2. Verify your configuration against the [Configuration File Reference](./configuration-file.md)
3. Ensure all [environment variables](./environment.md) are properly set
4. Review the [CLI Commands Reference](./cli-commands.md) for correct usage
