---
title: Your First Deployment
description: Learn how to deploy your first application using FTL
---

# Your First Deployment

This guide will walk you through deploying your first application with FTL. We'll cover the basic deployment workflow and essential commands.

## Prerequisites

Before starting your first deployment, ensure you have:

1. [Installed FTL](./installation.md) on your local machine
2. [Created your configuration file](./configuration.md) (`ftl.yaml`)
3. A Docker image of your application
4. SSH access to your target server
5. A domain name pointing to your server's IP address

## Deployment Steps

### 1. Server Setup

First, prepare your server with the required dependencies:

```bash
ftl setup
```

This command will:

- Install Docker and other required packages
- Configure system settings
- Set up Docker networks
- Configure firewall rules

### 2. Building Your Application

Build your application's Docker image. FTL provides two ways to handle images:

#### Direct SSH Transfer (Default)

If you don't specify an `image` field in your service configuration, FTL will:

- Build the image locally
- Transfer it directly to your server via SSH
- Optimize transfers using layer caching

The `path` field in your service configuration specifies the directory containing your application's source code and Dockerfile. This path is resolved relative to the location of your `ftl.yaml` file. For example:

```yaml
services:
  - name: web
    path: ./src
```

In this configuration, if your `ftl.yaml` is in `/project/ftl.yaml`, FTL will look for the Dockerfile in `/project/src/Dockerfile`.

#### Registry-based Deployment

If you specify an `image` field, FTL will use a Docker registry:

```bash
ftl build
```

::: warning Registry Authentication

- Only username/password authentication is supported
- Token-based authentication will fail
- You'll be prompted for registry credentials during server setup
  :::

::: tip
When using registry-based deployment, you can skip pushing to registry with:

```bash
ftl build --skip-push
```

:::

### 3. Deploying Your Application

Deploy your application to the configured server:

```bash
ftl deploy
```

The deployment process:

1. Connects to your server via SSH
2. Transfers images directly or pulls from registry (depending on configuration)
3. Sets up volumes and networks
4. Starts dependencies (if any)
5. Launches your services with health checks
6. Configures the Nginx reverse proxy
7. Sets up SSL certificates (if domain is configured)

### 4. Verifying the Deployment

After deployment, verify your application is running:

1. View application logs:

   ```bash
   ftl logs
   ```

2. Access your application through the configured domain

## Common First Deployment Issues

### SSL Certificate Generation

If SSL certificate generation fails:

- Ensure your domain points to the server's IP address
- Check that ports 80 and 443 are accessible
- Verify the email address in your configuration

### Container Health Checks

If containers fail health checks:

- Verify your health check endpoint is responding
- Check the application logs for errors
- Ensure the health check path is correct in `ftl.yaml`

### Network Issues

If you can't access your application:

- Verify your domain's DNS settings
- Check server firewall configuration
- Ensure your application is listening on the configured port

## Example Deployment

Here's a complete example of deploying a simple web application:

1. Configuration (`ftl.yaml`):

   ```yaml
   project:
     name: my-web-app
     domain: myapp.example.com
     email: admin@example.com

   server:
     host: myapp.example.com
     user: deploy
     ssh_key: ~/.ssh/id_rsa

   services:
     - name: web
       path: ./src
       port: 3000
       health_check:
         path: /health
       routes:
         - path: /
   ```

   This example uses direct SSH transfer. For registry-based deployment, add the `image` field:

   ```yaml
   services:
     - name: web
       image: registry.example.com/my-web-app:latest
       build:
         path: ./src # Path to directory containing Dockerfile
       port: 3000
       health_check:
         path: /health
       routes:
         - path: /
   ```

2. Deployment commands:

   ```bash
   # Setup server
   ftl setup

   # Build and deploy
   ftl build
   ftl deploy

   # Check logs
   ftl logs
   ```

## Deployment Workflow Tips

1. **Start Small**
   - Begin with a simple service configuration
   - Add dependencies and complexity gradually
   - Test each addition separately

2. **Use Environment Variables**

   ```bash
   # Create a .env file for local development
   echo "API_KEY=development-key" > .env

   # Deploy with production values
   API_KEY=production-key ftl deploy
   ```

3. **Monitor Deployments**
   - Watch the logs during deployment
   - Check health status after deployment
   - Verify all routes are working

## Next Steps

After your first successful deployment:

1. Learn about [Zero-downtime Deployments](/guides/zero-downtime)
2. Explore [Health Checks](/guides/health-checks) configuration
3. Set up [SSL Management](/guides/ssl-management)
4. Configure [Logging and Monitoring](/core-tasks/logging)

::: warning
Remember to always test deployments in a staging environment before deploying to production.
:::

## Reference

For more detailed information, see:

- [CLI Commands Reference](/reference/cli-commands)
- [Troubleshooting Guide](/reference/troubleshooting)
