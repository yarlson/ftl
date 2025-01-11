---
title: Advanced Guides
description: In-depth guides for advanced FTL deployment scenarios and features
---

# Advanced Guides

This section contains comprehensive guides for implementing advanced deployment scenarios and features with FTL. Each guide provides detailed, step-by-step instructions suitable for both experienced Docker users and those new to containerized deployments.

## Available Guides

### [Health Checks](./health-checks.md)

Learn how to implement robust health checks for your services to ensure reliable deployments and runtime monitoring. This guide covers configuration options, best practices, and troubleshooting common health check issues.

### [SSL Management](./ssl-management.md)

Comprehensive guide to SSL/TLS certificate management in FTL, including automatic certificate provisioning, renewal, and custom certificate configuration.

### [Zero-Downtime Deployments](./zero-downtime.md)

Implement zero-downtime deployments for your applications using FTL's built-in capabilities. Learn about deployment strategies, container replacement, and handling stateful services.

## Guide Structure

Each guide follows a consistent structure:

1. **Overview** - Introduction to the topic and its importance
2. **Prerequisites** - Required knowledge and setup
3. **Detailed Implementation** - Step-by-step instructions
4. **Configuration Examples** - Real-world configuration samples
5. **Troubleshooting** - Common issues and their solutions
6. **Best Practices** - Recommended approaches and patterns

## Before You Begin

Before diving into these advanced guides, ensure you:

- Have FTL installed and configured on your system
- Are familiar with basic FTL concepts and commands
- Have completed the [Getting Started](../getting-started/index.md) guide
- Understand Docker basics and container deployment concepts

::: tip
While these guides cover advanced topics, they include explanations of underlying concepts to help users of all experience levels understand and implement the features effectively.
:::

::: warning
The examples in these guides assume you have a working FTL configuration file (`ftl.yaml`). Make sure to adapt the examples to match your specific project structure and requirements.
:::
