---
title: Technical Reference
description: Comprehensive technical reference documentation for FTL CLI tool
---

# Technical Reference

This section provides detailed technical documentation for FTL's core components, configuration specifications, and CLI commands. It serves as the authoritative reference for experienced engineers working with FTL deployments.

## Contents

- [CLI Commands](./cli-commands.md) - Comprehensive documentation of all FTL command-line interface commands, flags, and usage patterns
- [Configuration File](./configuration-file.md) - Complete specification of the `ftl.yaml` configuration file format, including all available options and their effects
- [Environment Variables](./environment.md) - Details about environment variable handling, substitution patterns, and runtime configuration
- [Troubleshooting](./troubleshooting.md) - Common issues, error messages, and their resolutions

## Key Technical Concepts

### Image Management

FTL supports two distinct approaches for managing Docker images:

1. **Direct SSH Transfer**

   - Default method when no `image` field is specified
   - Builds images locally
   - Implements custom layer caching
   - Transfers only modified layers via SSH
   - Optimized for single-server deployments

2. **Registry-based Deployment**
   - Activated when `image` field is specified
   - Requires registry authentication (username/password only)
   - Follows standard Docker registry workflow
   - Suitable for multi-server deployments

### Deployment Process

The deployment system implements:

- Zero-downtime container replacement
- Health check verification
- Automatic SSL/TLS certificate management via ACME
- Integrated Nginx reverse proxy configuration
- Resource cleanup post-deployment

### Networking

FTL manages several networking aspects:

- Reverse proxy routing
- SSL/TLS termination
- SSH tunneling for remote dependencies
- Docker network isolation

### Security Considerations

- SSH key-based authentication for server access
- ACME protocol for SSL/TLS certificate management
- Docker network isolation between services
- Environment variable substitution for sensitive data

For detailed information about specific aspects of FTL, please refer to the relevant sections in this reference documentation.
