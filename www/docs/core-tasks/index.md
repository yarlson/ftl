---
title: Core Tasks
description: Overview of essential operational tasks in FTL
---

# Core Tasks

This section covers the essential operational tasks you'll perform with FTL. Each task is fundamental to managing your deployments effectively.

## Available Tasks

### [Building](./building.md)

Learn how to build and manage Docker images for your services:

- Building images locally
- Using build arguments and environment variables
- Pushing to registries
- Multi-stage builds

### [Deployment](./deployment.md)

Understand the deployment process and options:

- Zero-downtime deployments
- Rolling updates
- Deployment strategies
- Rollback procedures

### [Server Setup](./server-setup.md)

Configure your server for FTL deployments:

- System requirements
- Docker installation
- Network configuration
- Security settings

### [Logging](./logging.md)

Monitor your applications through logs:

- Viewing service logs
- Log aggregation
- Real-time log streaming
- Log retention policies

### [Tunneling](./tunneling.md)

Establish secure connections to your services:

- SSH tunneling to remote services
- Local development with remote dependencies
- Database access
- Debugging tools

## Common Workflows

Here are some typical workflows combining multiple core tasks:

1. **Initial Application Setup**

   ```mermaid
   graph LR
     A[Server Setup] --> B[Building]
     B --> C[Deployment]
     C --> D[Logging]
   ```

2. **Development Cycle**

   ```mermaid
   graph LR
     A[Code Changes] --> B[Building]
     B --> C[Deployment]
     C --> D[Verification]
     D --> A
   ```

3. **Maintenance Operations**
   ```mermaid
   graph LR
     A[Logging] --> B[Diagnostics]
     B --> C[Tunneling]
     C --> D[Debug]
   ```

## Task Organization

Each task in this section follows a consistent structure:

- **Overview**: Brief description of the task
- **Usage**: Common use cases and examples
- **Options**: Available configuration options
- **Best Practices**: Recommended approaches
- **Troubleshooting**: Common issues and solutions

## Quick Reference

Here's a quick reference for the most commonly used commands in core tasks:

```bash
# Building
ftl build [--skip-push]

# Deployment
ftl deploy

# Logging
ftl logs [service] [-f]

# Tunneling
ftl tunnels
```

## Next Steps

After familiarizing yourself with these core tasks, you might want to explore:

- [Configuration Options](/configuration/) for fine-tuning your setup
- [Advanced Guides](/guides/) for specific use cases
- [CLI Reference](/reference/cli-commands) for detailed command information

::: tip
Each core task can be performed independently, but they're designed to work together seamlessly in your deployment workflow.
:::

## See Also

- [Getting Started Guide](/getting-started/)
- [Configuration Reference](/reference/configuration-file)
- [Troubleshooting Guide](/reference/troubleshooting)
