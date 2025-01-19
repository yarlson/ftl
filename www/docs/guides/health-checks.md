---
title: Health Checks
description: Configure and implement effective health checks for reliable FTL deployments
---

# Health Checks

Health checks are a critical component of reliable deployments in FTL. They ensure your services are running correctly and enable zero-downtime deployments by verifying container health before routing traffic.

## Overview

FTL uses health checks to:

- Verify service availability during deployments
- Ensure containers are ready to receive traffic
- Monitor ongoing service health
- Enable safe container replacement

## Configuration

### Basic Health Check

The minimal health check configuration in your `ftl.yaml` only requires the path:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
```

You can optionally configure timing parameters:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
      interval: 15s # optional
      timeout: 10s # optional
      retries: 3 # optional
```

### Configuration Parameters

Health check has one required and three optional parameters:

- `path`: (Required) The HTTP endpoint to check
  - Must return 2xx or 3xx status code
  - Should be a lightweight endpoint
  - Typically `/health`, `/`, or similar

Optional timing parameters:

- `interval`: Time between checks

  - Default: 15 seconds
  - Should be short enough to detect issues quickly
  - But not so frequent as to overload the service

- `timeout`: Maximum time to wait for response

  - Default: 10 seconds
  - Should be less than the interval
  - Consider your service's normal response time

- `retries`: Required successful checks
  - Default: 3
  - Higher values increase reliability
  - But also increase deployment time

## Implementation Patterns

### 1. Basic Health Endpoint

Simplest implementation that checks if the service is responding:

```yaml
services:
  - name: web-service
    image: web-service:latest
    port: 80
    health_check:
      path: /
```

### 2. Dedicated Health Endpoint

Recommended approach with a specific health check endpoint:

```yaml
services:
  - name: api-service
    image: api-service:latest
    port: 3000
    health_check:
      path: /health
```

### 3. Deep Health Check

Comprehensive check including dependencies:

```yaml
services:
  - name: backend-service
    image: backend-service:latest
    port: 8080
    health_check:
      path: /health/deep
      interval: 30s # customized for deeper checks
```

### API Service

```yaml
services:
  - name: api
    image: api:latest
    port: 3000
    health_check:
      path: /health/live
      interval: 15s
      timeout: 5s
      retries: 3
    routes:
      - path: /api
```

::: tip
Start with conservative timeouts and adjust based on your service's performance characteristics.
:::

::: warning
Health checks should be representative of service health but not impact performance.
:::

## Conclusion

Effective health checks are essential for reliable deployments. Key takeaways:

- Use dedicated health check endpoints
- Configure appropriate timeouts and intervals
- Monitor health check logs
- Follow best practices for endpoint implementation

By following these guidelines, you can ensure your services deploy reliably and maintain high availability.
