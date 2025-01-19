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

Health checks are crucial for zero-downtime deployments. FTL uses them to verify that new containers are ready before routing traffic. The minimal required configuration only needs the path to your health check endpoint:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
```

You can further customize the health check behavior with additional timing parameters. The `interval`, `timeout`, and `retries` fields are all optional, with defaults of 15s, 10s, and 3 respectively:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
      interval: 15s
      timeout: 10s
      retries: 3
```

### 2. Configure Service Routes

For route configuration, only the path is required. Here's a minimal route setup:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    routes:
      - path: /
```

Routes can be customized with the `strip_prefix` parameter, which defaults to false if not specified:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    routes:
      - path: /
        strip_prefix: false
```

Here's a complete example showing project configuration with minimal service settings:

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
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /health
    routes:
      - path: /
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

Here's a minimal recommended configuration that includes only the required parameters:

```yaml
services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /health
    routes:
      - path: /
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
