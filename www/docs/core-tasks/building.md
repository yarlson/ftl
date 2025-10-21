---
title: Building Images
description: Learn how to build and manage Docker images with FTL
---

# Building Images

FTL provides a streamlined way to build Docker images for your services. This guide covers the build process and available options.

## Basic Usage

To build images for all services defined in your `ftl.yaml`:

```bash
ftl build
```

## Image Handling in FTL

FTL offers two ways to handle Docker images:

### 1. Direct SSH Transfer (Default)

When no `image` field is specified in your service configuration, FTL will:

- Build the image locally
- Transfer it directly to your server via SSH
- Use its own layer caching algorithm to optimize transfers
- Only transfer layers that haven't been previously sent to the server

```yaml
services:
  - name: web
    path: ./src
```

The path is relative to your `ftl.yaml` file location and should contain your application's source code and Dockerfile.

::: tip
This method is simpler as it doesn't require registry configuration and credentials management.
:::

### 2. Registry-based Deployment

When you specify the `image` field, FTL will:

- Build and tag the image locally
- Push it to the specified registry
- Pull the image on the server during deployment
- Require registry authentication during server setup
- Currently support only username/password authentication

```yaml
services:
  - name: web
    image: registry.example.com/my-app:latest
    path: ./src
```

::: warning
Currently, FTL only supports registries with username/password authentication. Token-based authentication will fail.
:::

## Build Options

### Command Line Flags

```bash
# Skip pushing images to the registry (only applies when using registry-based deployment)
ftl build --skip-push
```

## Understanding Docker Builds

### Source Code Location

The source code and Dockerfile should be placed in the directory specified by the `path` field. This path is resolved relative to your `ftl.yaml` file location. For example:

If your project structure is:

```
/project
  ├── ftl.yaml
  └── src
      ├── Dockerfile
      └── app files...
```

Your configuration would be:

```yaml
services:
  - name: web
    path: ./src
```

### Layer Caching

Docker uses a layer cache to speed up builds. Understanding how it works can significantly improve build times:

1. **Layer Order Matters**
   - Put infrequently changed commands early in Dockerfile
   - Place frequently changed commands (like copying source code) later

2. **Cache Busting**
   - Adding/modifying files invalidates cache for that layer and all following layers
   - Changing a command invalidates cache for that layer and all following layers

### Multi-stage Builds

Multi-stage builds help create smaller production images:

```dockerfile
# Build stage
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Production stage
FROM node:18-slim
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY package*.json ./
RUN npm install --production
CMD ["npm", "start"]
```

Benefits:

- Smaller final image size
- Separation of build and runtime dependencies
- Better security by excluding build tools

## Best Practices

1. **Choose the Right Deployment Method**
   - Use direct SSH transfer for simpler setups
   - Use registry-based deployment when:
     - You need image versioning
     - You require image scanning/signing

2. **Registry Configuration**
   - Use username/password authentication
   - Avoid token-based registries (currently unsupported)
   - Consider registry proximity to your server

3. **Optimize Dockerfiles**
   - Use multi-stage builds
   - Minimize layer count
   - Order commands by change frequency
   - Use specific base image tags

4. **Image Tags**
   - Use meaningful tags
   - Consider versioning strategy
   - Document tagging conventions

5. **Build Performance**
   - Leverage Docker layer caching
   - Use `.dockerignore` effectively
   - Keep source directory clean

6. **Security**
   - Use official base images
   - Keep base images updated
   - Scan images for vulnerabilities
   - Don't store secrets in images

## Common Docker Issues

### Build Performance

If builds are slow:

- Check source directory size
- Optimize layer caching
- Use `.dockerignore` appropriately
- Consider multi-stage builds

### Image Size

To reduce image size:

- Use multi-stage builds
- Choose appropriate base images
- Remove unnecessary files
- Combine RUN commands

### Cache Issues

If cache isn't working effectively:

- Check command ordering in Dockerfile
- Verify file changes aren't invalidating cache
- Use appropriate COPY commands

### Registry Authentication

If using registry-based deployment:

- Ensure registry supports username/password authentication
- Have credentials ready during server setup
- Verify registry URL is correct
- Check network connectivity to registry

## Next Steps

After building your images:

1. [Deploy your application](./deployment.md)
2. Learn about [Zero-downtime Deployments](../guides/zero-downtime.md)
3. Configure [Health Checks](../guides/health-checks.md)

::: warning
Always test builds in a development environment before building for production.
:::

## Reference

- [Configuration Reference](../reference/configuration-file.md)
- [Troubleshooting Guide](../reference/troubleshooting.md)
- [Docker Documentation](https://docs.docker.com/engine/reference/builder/)
