---
title: CLI Commands Reference
description: Complete reference for all FTL CLI commands, flags, and usage patterns
---

# CLI Commands Reference

This page documents all available FTL CLI commands, their flags, and usage patterns.

## Commands Overview

- [`ftl setup`](#setup) - Initialize server with required dependencies
- [`ftl build`](#build) - Build and prepare application images
- [`ftl deploy`](#deploy) - Deploy application to configured server
- [`ftl logs`](#logs) - Retrieve and stream logs from services
- [`ftl tunnels`](#tunnels) - Create SSH tunnels to remote dependencies

## Setup

Initializes a server with required dependencies and configurations.

```bash
ftl setup
```

### Description

The setup command performs the following operations:

- Installs Docker and required system packages
- Configures firewall rules
- Sets up user permissions
- Initializes Docker networks
- Configures registry authentication if using registry-based deployment

### Example

```bash
ftl setup
```

## Build

Builds and prepares Docker images for deployment.

```bash
ftl build [flags]
```

### Flags

| Flag          | Description                                                                 |
| ------------- | --------------------------------------------------------------------------- |
| `--skip-push` | Skip pushing images to registry (only applies to registry-based deployment) |

### Description

The build command handles image preparation based on your configuration:

**For Direct SSH Transfer (Default)**

- Builds images locally
- Uses custom layer caching
- Prepares images for direct transfer

**For Registry-based Deployment**

- Builds images locally
- Tags images according to configuration
- Pushes to specified registry (unless `--skip-push` is used)

### Examples

```bash
# Build all services (using direct SSH transfer)
ftl build

# Build all services but skip registry push
ftl build --skip-push
```

## Deploy

Deploys the application to configured server.

```bash
ftl deploy
```

### Description

The deploy command performs these operations:

- Connects to configured server via SSH
- Pulls/transfers required Docker images
- Performs zero-downtime container replacement
- Configures Nginx reverse proxy
- Manages SSL/TLS certificates via ACME
- Runs health checks
- Cleans up unused resources

### Example

```bash
ftl deploy
```

## Logs

Retrieves logs from deployed services.

```bash
ftl logs [service] [flags]
```

### Arguments

| Argument  | Description                                       |
| --------- | ------------------------------------------------- |
| `service` | (Optional) Name of the service to fetch logs from |

### Flags

| Flag                   | Description                          | Default                 |
| ---------------------- | ------------------------------------ | ----------------------- |
| `-f`, `--follow`       | Stream logs in real-time             | `false`                 |
| `-n`, `--tail <lines>` | Number of lines to show from the end | `100` (if `-f` is used) |

### Examples

```bash
# Fetch logs from all services
ftl logs

# Stream logs from a specific service
ftl logs my-app -f

# Fetch last 50 lines from all services
ftl logs -n 50

# Fetch logs from specific service with custom tail size
ftl logs my-app -n 150
```

## Tunnels

Creates SSH tunnels to remote dependencies.

```bash
ftl tunnels
```

### Description

The tunnels command:

- Establishes SSH tunnels to dependency services
- Enables local access to remote services
- Maintains concurrent tunnel connections

### Examples

```bash
# Establish tunnels to all dependency ports
ftl tunnels
```

## Environment Variables

All commands respect environment variables defined in your `ftl.yaml` configuration. Variables can be:

- Required: `${VAR_NAME}`
- Optional with default: `${VAR_NAME:-default_value}`

For detailed information about environment variable handling, see the [Environment Variables](./environment.md) reference.
