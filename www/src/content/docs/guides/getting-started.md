---
title: Getting Started with FTL
description: Introduction to FTL deployment tool and its key benefits
---

FTL (Faster Than Light) is a deployment tool designed to simplify application deployment without the complexity of full container orchestration platforms. This guide will help you understand why FTL might be the right choice for your projects.

## Why FTL?

Modern deployment tools often fall into two extremes:

- Heavy orchestration platforms (Kubernetes, ECS) - powerful but complex
- Basic deployment scripts - simple but limited functionality

FTL fills the gap between these extremes by providing:

- Zero-downtime deployments without complex infrastructure
- Built-in SSL/TLS certificate management
- Integrated reverse proxy configuration
- Multi-provider support
- Simple YAML configuration
- No additional infrastructure requirements

## When to Use FTL

FTL is ideal for:

- Small to medium-sized applications
- Teams wanting simple but robust deployments
- Projects that don't require complex orchestration
- Applications with straightforward scaling needs
- Quick deployment to multiple providers

FTL might not be the best choice if you need:

- Complex micro-service architectures (100s of services)
- Advanced service mesh features
- Provider-specific optimizations
- Custom orchestration requirements

## Deployment Philosophy

FTL follows these core principles:

1. **Simplicity First**

   - Minimal configuration
   - Sensible defaults
   - Clear, predictable behavior

2. **Convention Over Configuration**

   - Standard patterns for common scenarios
   - Reduced decision fatigue
   - Consistent deployments

3. **Just Enough Automation**
   - Automate common tasks
   - Maintain visibility
   - Keep control when needed

## Next Steps

Ready to get started with FTL? Follow these links to continue:

- [Installation Guide](/guides/installation/) - Install FTL on your system
- [Configuration Guide](/guides/basic-setup/configuration/) - Learn the basics of FTL configuration
- [Basic Web App Example](/examples/simple-webapp/) - Deploy your first application
