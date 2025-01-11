---
title: Zero-Downtime Deployments
description: Learn how to implement zero-downtime deployments using FTL's built-in capabilities
---

# Zero-Downtime Deployments

Zero-downtime deployment is a critical feature of FTL that ensures your applications remain available during updates. This guide explains how FTL implements zero-downtime deployments and how to configure your services to take full advantage of this capability.

## Overview

FTL achieves zero-downtime deployments through a carefully orchestrated process:

1. Starting new containers while old ones continue serving traffic
2. Health check verification before routing traffic
3. Graceful shutdown of old containers

## Prerequisites

Before implementing zero-downtime deployments, ensure you have:

- FTL installed and configured
- A working `ftl.yaml` configuration file
- Docker images for your services
- Basic understanding of health checks

## Implementation

### 1. Configure Health Checks

Health checks are crucial for zero-downtime deployments. FTL uses them to verify that new containers are ready before routing traffic.

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
      interval: 10s
      timeout: 5s
      retries: 3
```

The health check configuration parameters:

- `path`: The endpoint to check (must return 2xx or 3xx status)
- `interval`: Time between health checks
- `timeout`: Maximum time to wait for a response
- `retries`: Number of consecutive successful checks required

### 2. Configure Service Routes

Proper route configuration ensures traffic is handled correctly during deployments:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    routes:
      - path: /
        strip_prefix: false
```

### 3. Deployment Process

To perform a zero-downtime deployment:

1. Build your updated application:

```bash
ftl build
```

2. Deploy the changes:

```bash
ftl deploy
```

FTL will automatically:

1. Pull the new image on the server
2. Start new containers
3. Run health checks
4. Switch traffic to new containers
5. Gracefully stop old containers

## Best Practices

### 1. Application Design

- Implement graceful shutdown handling
- Design stateless services where possible
- Handle in-flight requests during shutdown

### 2. Health Check Design

- Use dedicated health check endpoints
- Keep health checks lightweight
- Include critical dependency checks

### 3. Configuration Recommendations

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /health
      interval: 10s
      timeout: 5s
      retries: 3
    routes:
      - path: /
        strip_prefix: false
```

## Common Issues and Solutions

### 1. Failed Health Checks

**Problem**: New containers fail health checks during deployment.

**Solution**:

- Verify health check endpoint functionality
- Increase timeout or retries if needed
- Check application logs using:

```bash
ftl logs my-app
```

### 2. Lingering Connections

**Problem**: Old containers have lingering connections.

**Solution**:

- Implement graceful shutdown in your application
- Handle SIGTERM signals properly
- Consider increasing shutdown timeout if needed

### 3. Database Migrations

**Problem**: Schema changes can break zero-downtime deployments.

**Solution**:

- Use backward-compatible database migrations
- Deploy schema changes separately from application changes
- Consider using blue-green deployment for major database changes

## Monitoring Deployments

Track deployment progress using FTL's logging capabilities:

```bash
ftl logs -f
```

This will show real-time logs during deployment, including:

- Container health check status
- Traffic routing changes
- Container lifecycle events

## Rollback Procedure

If issues occur during deployment:

1. Stop the problematic deployment:

```bash
Ctrl+C
```

2. Redeploy the previous version:

```bash
ftl deploy
```

FTL will automatically handle the rollback process with zero downtime.

::: tip
Always keep the previous working image available for quick rollbacks.
:::

::: warning
While FTL handles most scenarios automatically, complex applications might require additional configuration or architectural changes to fully support zero-downtime deployments.
:::

## Conclusion

Zero-downtime deployment in FTL is an automated process that requires minimal configuration. The key elements are:

- Proper health check configuration
- Well-designed application architecture
- Understanding of the deployment process

By following this guide and best practices, you can ensure reliable and seamless deployments for your applications.
