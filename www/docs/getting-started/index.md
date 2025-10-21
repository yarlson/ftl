---
title: Getting Started with FTL
description: Learn how to get started with FTL - the simpler way to deploy applications
---

# Getting Started with FTL

FTL (Faster Than Light) is a deployment tool designed to simplify application deployment without the complexity of extensive orchestration infrastructure. This getting started guide will help you begin using FTL for your projects.

## What is FTL?

FTL provides automated deployment to cloud providers like Hetzner, DigitalOcean, Linode, and custom servers. It eliminates the need for complex CI/CD pipelines or container orchestration platforms while still offering essential features like:

- Zero-downtime deployments
- Automatic SSL/TLS certificate management
- Docker-based deployment with health checks
- Integrated Nginx reverse proxy
- Multi-provider support
- Log streaming
- SSH tunneling for remote dependencies

## Quick Start Guide

1. [Installation](./installation.md)  
   Learn how to install FTL on your system using Homebrew, direct download, or by building from source.

2. [Configuration](./configuration.md)  
   Set up your project's `ftl.yaml` file and understand the basic configuration options.

3. [First Deployment](./first-deployment.md)  
   Deploy your first application with FTL and learn the basic deployment workflow.

## Prerequisites

Before getting started with FTL, ensure you have:

- Basic knowledge of Docker and containerization
- SSH access to your target deployment server(s)
- Docker installed on your local machine
- A domain name (for SSL/TLS certificate management)

## System Requirements

- **Local Machine:**
  - macOS, Linux, or Windows with WSL2
  - Docker Desktop or Docker Engine
  - SSH client
  - 4GB RAM (minimum)
  - 2 CPU cores (recommended)

- **Target Server:**
  - Ubuntu 20.04 or newer (recommended)
  - 1GB RAM (minimum)
  - 1 CPU core (minimum)
  - SSH access with sudo privileges

## Next Steps

After completing the getting started guide, explore these topics to learn more:

- [Core Tasks](/core-tasks/) - Learn about building, deployment, logging, and other essential operations
- [Configuration](/configuration/) - Detailed configuration options and best practices
- [Guides](/guides/) - In-depth guides for specific features and use cases
- [Reference](/reference/) - Complete reference for CLI commands, configuration, and troubleshooting

::: tip
For the smoothest experience, we recommend following the guides in order, starting with the [Installation](./installation.md) page.
:::
