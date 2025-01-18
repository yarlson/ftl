---
title: Server Setup
description: Learn how to prepare your server for FTL deployment
---

# Server Setup

Before deploying applications with FTL, your server needs to be properly configured. This guide covers the server setup process and requirements.

## Overview

The `ftl setup` command prepares your server by:

- Installing required system packages
- Setting up Docker
- Configuring networking
- Setting up security measures

## Prerequisites

Before running server setup, ensure you have:

- Root or sudo access to the target server
- SSH access configured
- A server running a supported operating system:
  - Ubuntu 20.04 LTS or newer (recommended)
  - Debian 11 or newer
  - Other Linux distributions with Docker support

## Running Setup

Execute the setup command:

```bash
ftl setup
```

This command will connect to the server defined in your `ftl.yaml` configuration and perform the necessary setup steps.

## Setup Process

### 1. System Updates

The setup process begins by updating the system:

- Updates package lists
- Upgrades existing packages
- Installs required dependencies

### 2. Docker Installation

FTL installs and configures Docker:

- Adds Docker repository
- Installs Docker Engine
- Starts Docker service
- Configures Docker daemon settings
- Sets up Docker network for FTL

### 3. Network Configuration

Network setup includes:

- Configuring firewall rules
- Opening required ports:
  - 22 (SSH)
  - 80 (HTTP)
  - 443 (HTTPS)
- Setting up Docker networks

### 4. Security Configuration

Security measures implemented:

- Creating limited-privilege user for deployments
- Configuring SSH access
- Setting up firewall rules
- Applying basic security hardening

## Server Requirements

### Minimum Hardware Requirements

- **CPU**: 1 core
- **RAM**: 1GB
- **Storage**: 20GB
- **Network**: 100Mbps

### Recommended Hardware

- **CPU**: 2+ cores
- **RAM**: 2GB+
- **Storage**: 40GB+
- **Network**: 1Gbps

### Software Requirements

- **Operating System**: Ubuntu 20.04 LTS or newer
- **SSH Server**: OpenSSH

## Configuration Options

### Custom Docker Configuration

You can customize Docker daemon settings by creating a `docker.json` file:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "default-address-pools": [
    {
      "base": "172.17.0.0/16",
      "size": 24
    }
  ]
}
```

## Verification

After setup completes, verify the installation:

1. Check Docker status:

   ```bash
   ssh user@server "docker info"
   ```

2. Verify network configuration:

   ```bash
   ssh user@server "docker network ls"
   ```

3. Test Docker functionality:
   ```bash
   ssh user@server "docker run hello-world"
   ```

## Common Issues

### Permission Denied

If you encounter permission issues:

```bash
# Add your user to the docker group
sudo usermod -aG docker $USER

# Apply changes
newgrp docker
```

### Docker Network Conflicts

If you experience network conflicts:

1. Check existing networks:

   ```bash
   docker network ls
   ```

2. Remove conflicting networks:
   ```bash
   docker network rm conflicting-network
   ```

### Firewall Issues

If services are inaccessible:

1. Verify firewall rules:

   ```bash
   sudo ufw status
   ```

2. Allow required ports:
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   ```

## Best Practices

1. **Regular Updates**

   - Keep system packages updated
   - Regularly update Docker
   - Monitor security advisories

2. **Security**

   - Use strong SSH keys
   - Implement fail2ban
   - Regular security audits
   - Keep ports minimal

3. **Monitoring**

   - Set up disk space monitoring
   - Monitor system resources
   - Configure log rotation

4. **Backup**
   - Regular system backups
   - Docker volume backups
   - Configuration backups

## Next Steps

After setting up your server:

1. [Configure your application](../getting-started/configuration.md)
2. [Deploy your first application](../getting-started/first-deployment.md)
3. Learn about [Zero-downtime Deployments](../guides/zero-downtime.md)

::: warning
Always test the setup process on a staging server before setting up production environments.
:::

## Reference

- [Configuration Reference](../reference/configuration-file.md)
- [Troubleshooting Guide](../reference/troubleshooting.md)
- [Docker Documentation](https://docs.docker.com/)
