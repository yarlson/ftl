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

The minimal health check configuration in your `ftl.yaml`:

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

### Configuration Parameters

Each health check has four key parameters:

- `path`: The HTTP endpoint to check

  - Must return 2xx or 3xx status code
  - Should be a lightweight endpoint
  - Typically `/health`, `/`, or similar

- `interval`: Time between checks

  - Default: 10 seconds
  - Should be short enough to detect issues quickly
  - But not so frequent as to overload the service

- `timeout`: Maximum time to wait for response

  - Default: 5 seconds
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
      interval: 10s
      timeout: 5s
      retries: 3
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
      interval: 10s
      timeout: 5s
      retries: 3
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
      interval: 15s
      timeout: 10s
      retries: 3
```

## Best Practices

### Health Check Endpoint Design

1. **Keep It Light**

   - Avoid expensive operations
   - Don't query large datasets
   - Minimize dependency checks

2. **Response Time**

   - Should respond within 1-2 seconds
   - Set timeout accordingly
   - Consider background processing

3. **Dependency Checks**
   - Include critical dependencies only
   - Use timeouts for external services
   - Have fallback mechanisms

### Common Patterns

1. **Basic Availability**

```yaml
health_check:
  path: /
  interval: 10s
  timeout: 5s
  retries: 2
```

2. **Quick Response**

```yaml
health_check:
  path: /ping
  interval: 5s
  timeout: 2s
  retries: 3
```

3. **Thorough Check**

```yaml
health_check:
  path: /health/deep
  interval: 30s
  timeout: 10s
  retries: 2
```

## Monitoring Health Checks

Track health check status using FTL's logging:

```bash
ftl logs -f
```

The logs will show:

- Health check attempts
- Success/failure status
- Response times
- Error messages

## Troubleshooting

### 1. Failed Health Checks

**Problem**: Health checks consistently fail during deployment.

**Solution**:

- Verify the health check endpoint exists
- Check endpoint permissions
- Review service logs:

```bash
ftl logs service-name
```

### 2. Slow Health Checks

**Problem**: Health checks timeout frequently.

**Solution**:

- Optimize the health check endpoint
- Increase the timeout value
- Reduce dependency checks

### 3. Flaky Health Checks

**Problem**: Health checks pass intermittently.

**Solution**:

- Add retry logic in the health check
- Increase the retry count
- Check for resource constraints

## Example Implementations

### HTTP Service

```yaml
services:
  - name: http-service
    image: http-service:latest
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
        strip_prefix: true
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
