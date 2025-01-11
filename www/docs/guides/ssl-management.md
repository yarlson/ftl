---
title: SSL Management
description: Learn how to manage SSL/TLS certificates with FTL's automatic certificate provisioning
---

# SSL Management

FTL provides built-in SSL/TLS certificate management with automatic provisioning and renewal through ACME (Automated Certificate Management Environment) using ZeroSSL. This guide explains how to configure and manage SSL certificates for your services.

## Overview

FTL handles SSL/TLS certificates by:

- Automatically provisioning certificates through ZeroSSL
- Managing certificate renewals
- Configuring Nginx for SSL termination
- Ensuring secure defaults

## Configuration

### Basic SSL Setup

The minimal configuration in your `ftl.yaml` to enable SSL:

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com

services:
  - name: web-app
    image: web-app:latest
    port: 80
    routes:
      - path: /
        strip_prefix: false
```

The essential fields for SSL management are:

- `domain`: Your application's domain name
- `email`: Contact email for ZeroSSL notifications

## How It Works

When you deploy your application, FTL:

1. Verifies domain ownership
2. Requests certificates from ZeroSSL
3. Configures Nginx with the certificates
4. Sets up automatic renewal

## Implementation Patterns

### 1. Single Domain

Basic configuration for a single domain:

```yaml
project:
  name: my-app
  domain: app.example.com
  email: admin@example.com
```

### 2. Multiple Services

Configuration for multiple services under one domain:

```yaml
project:
  name: my-platform
  domain: platform.example.com
  email: admin@example.com

services:
  - name: frontend
    image: frontend:latest
    port: 80
    routes:
      - path: /
        strip_prefix: false

  - name: api
    image: api:latest
    port: 3000
    routes:
      - path: /api
        strip_prefix: true
```

## Best Practices

### 1. Email Configuration

- Use a monitored email address
- Ensure email is valid for certificate expiry notifications
- Consider using a role-based email address

### 2. Domain Configuration

- Verify DNS records before deployment
- Ensure domain points to the correct server IP
- Allow time for DNS propagation

### 3. Security Considerations

- Keep email address up to date
- Monitor certificate expiration
- Maintain secure DNS configuration

## Monitoring Certificates

Track certificate status through FTL logs:

```bash
ftl logs
```

The logs will show:

- Certificate request status
- Renewal attempts
- Any SSL-related errors

## Troubleshooting

### 1. Certificate Provisioning Failures

**Problem**: Certificate provisioning fails during deployment.

**Solution**:

- Verify domain DNS configuration
- Check email address validity
- Review logs for specific errors:

```bash
ftl logs
```

### 2. Certificate Renewal Issues

**Problem**: Certificates fail to renew automatically.

**Solution**:

- Check server connectivity
- Verify domain still points to correct IP
- Ensure ports 80/443 are accessible

### 3. DNS Configuration Problems

**Problem**: Domain validation fails.

**Solution**:

- Verify A/CNAME records
- Allow time for DNS propagation
- Check domain ownership

## Example Configurations

### Basic Website

```yaml
project:
  name: company-website
  domain: www.example.com
  email: webmaster@example.com

services:
  - name: website
    image: website:latest
    port: 80
    routes:
      - path: /
        strip_prefix: false
```

### API Platform

```yaml
project:
  name: api-platform
  domain: api.example.com
  email: api-admin@example.com

services:
  - name: api-gateway
    image: api-gateway:latest
    port: 3000
    routes:
      - path: /
        strip_prefix: false

  - name: docs
    image: api-docs:latest
    port: 80
    routes:
      - path: /docs
        strip_prefix: true
```

::: tip
FTL automatically handles certificate renewal, but it's good practice to monitor logs periodically to ensure everything is working correctly.
:::

::: warning
Make sure your domain's DNS is properly configured before deploying. FTL needs to validate domain ownership to provision certificates.
:::

## Conclusion

SSL management in FTL is designed to be automatic and hassle-free. Key points to remember:

- Configure domain and email in project settings
- Ensure proper DNS configuration
- Monitor logs for certificate-related events
- Follow security best practices

With FTL's built-in SSL management, you can focus on your application while maintaining secure HTTPS connections for your users.
