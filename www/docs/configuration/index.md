---
title: Configuration Overview
description: Learn about FTL's configuration system and how to structure your ftl.yaml file
---

# Configuration Overview

FTL uses a single YAML configuration file (`ftl.yaml`) to define your entire deployment setup. This file describes your project settings, server configurations, services, dependencies, and volumes.

## Configuration File Structure

The `ftl.yaml` file is organized into these main sections:

- **Project**: Basic project information like name, domain, and contact details
- **Server**: Target server specifications and SSH connection details
- **Services**: Your application services that will be deployed
- **Dependencies**: Supporting services like databases and caches
- **Volumes**: Persistent storage definitions

## Basic Example

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: admin@example.com

server:
  host: my-project.example.com
  port: 22
  user: deploy
  ssh_key: ~/.ssh/id_rsa

services:
  - name: web
    path: ./src
    port: 3000
    routes:
      - path: /
```

## Environment Variables

FTL supports environment variable substitution in your configuration:

- Required variables: `${VAR_NAME}`
- Optional variables with defaults: `${VAR_NAME:-default_value}`

## Detailed Configuration

Each section of the configuration has its own detailed documentation:

- [Project Settings](./project-settings.md) - Core project configuration
- [Server Configuration](./server.md) - Server and SSH settings
- [Services](./services.md) - Application service definitions
- [Dependencies](./dependencies.md) - Supporting service configuration
- [Volumes](./volumes.md) - Persistent storage management

## Configuration Validation

FTL validates your configuration file before executing any commands. Common validation checks include:

- Required fields presence
- Port number validity
- File path existence
- Environment variable resolution
- Service name uniqueness
- Volume reference validity

## Best Practices

1. **Version Control**: Always keep your `ftl.yaml` in version control
2. **Environment Variables**: Use environment variables for sensitive data
3. **Health Checks**: Define health checks for all services
4. **Documentation**: Comment complex configurations
5. **Validation**: Run `ftl validate` before deployments

## Next Steps

- Learn about [Project Settings](./project-settings.md)
- Configure your [Services](./services.md)
- Set up [Dependencies](./dependencies.md)
- Manage [Volumes](./volumes.md)
