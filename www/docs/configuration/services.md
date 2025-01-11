---
title: Services Configuration
description: Configure application services in your FTL deployment
---

# Services Configuration

The `services` section defines your application services that will be deployed and managed by FTL.

## Required Fields

```yaml
services:
  - name: web
    build:
      context: .
      dockerfile: Dockerfile
    port: 3000
    routes:
      - path: /
```

| Field    | Description                                          |
| -------- | ---------------------------------------------------- |
| `name`   | Unique identifier for the service                    |
| `build`  | Docker build configuration for the service           |
| `port`   | Port that the service listens on                     |
| `routes` | HTTP route configuration for the Nginx reverse proxy |

## Build Configuration

You can specify how to build your service's Docker image in two ways:

### Direct SSH Transfer

When no `image` field is specified, FTL will build and transfer the image directly:

```yaml
services:
  - name: web
    build:
      context: .
      dockerfile: Dockerfile
```

### Registry-based Deployment

When using a Docker registry:

```yaml
services:
  - name: web
    image: registry.example.com/my-app:latest
    build:
      context: .
      dockerfile: Dockerfile
```

## Health Checks

Configure health checks to ensure reliable deployments:

```yaml
services:
  - name: web
    health_check:
      path: /
      interval: 10s
      timeout: 5s
      retries: 3
```

| Field      | Description                                         |
| ---------- | --------------------------------------------------- |
| `path`     | HTTP endpoint to check                              |
| `interval` | Time between checks                                 |
| `timeout`  | Maximum time to wait for response                   |
| `retries`  | Number of failed checks before marking as unhealthy |

## Routes Configuration

Define how HTTP traffic is routed to your service:

```yaml
services:
  - name: web
    routes:
      - path: /
        strip_prefix: false
```

| Field          | Description                                     |
| -------------- | ----------------------------------------------- |
| `path`         | URL path to match                               |
| `strip_prefix` | Whether to remove the path prefix when proxying |

## Environment Variables

Services support environment variable substitution:

```yaml
services:
  - name: web
    port: ${PORT}
    image: ${IMAGE_NAME}
```

All environment variables must be set in the environment before running FTL commands.

## Complete Example

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
    routes:
      - path: /
        strip_prefix: false
```
