---
layout: home
title: FTL - Faster Than Light Deployment Tool
description: Simple Docker deployment tool with zero-downtime updates, SSL automation, and database provisioning. No registry required.
head:
  - - meta
    - name: keywords
      content: deployment tool, docker deployment, zero-downtime deployment, SSL automation, database provisioning
  - - meta
    - name: og:title
      content: FTL - Faster Than Light Deployment Tool
  - - meta
    - name: og:description
      content: Simple Docker deployment tool with zero-downtime updates, SSL automation, and database provisioning. No registry required.
  - - meta
    - name: og:type
      content: website
  - - meta
    - name: twitter:title
      content: FTL - Faster Than Light Deployment Tool
  - - meta
    - name: twitter:description
      content: Simple Docker deployment tool with zero-downtime updates, SSL automation, and database provisioning. No registry required.
---

<style>
.home-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
  align-items: start;
  margin-top: 2rem;
}

.feature-list {
  list-style: none;
  padding: 0;
}

.feature-list li {
  margin-bottom: 1rem;
  font-size: 1.1em;
}

.quick-links {
  margin-top: 2rem;
}

.quick-links-list {
  list-style: none;
  padding: 0;
}

.quick-links-list li {
  margin-bottom: 1rem;
}

.quick-links-list a {
  font-size: 1.1em;
  font-weight: 500;
}

.quick-links-list span {
  display: block;
  margin-top: 0.25rem;
  color: var(--vp-c-text-2);
}

.home-content {
    margin: 16px;
}
</style>

<div class="home-container">
<div class="home-content">

# FTL: Faster Than Light Deployment

Simple, zero-downtime deployments without the complexity of traditional CI/CD pipelines.

<div class="feature-list">

📦 Single binary installation

✨ Simple YAML configuration

🔐 Automated SSL/TLS management

🐳 Registry-optional Docker deployment

🚀 Zero-downtime updates

🗄️ Database provisioning included

</div>

<div class="quick-links">

## Quick Start

<ul class="quick-links-list">
  <li>
    <a href="/getting-started/installation">Installation</a>
    <span>Install via package manager or download binary</span>
  </li>
  <li>
    <a href="/getting-started/first-deployment">First Deployment</a>
    <span>Basic setup and deployment walkthrough</span>
  </li>
  <li>
    <a href="/configuration/">Configuration Guide</a>
    <span>YAML configuration reference</span>
  </li>
  <li>
    <a href="/reference/cli-commands">CLI Reference</a>
    <span>Command syntax and options</span>
  </li>
</ul>

</div>

</div>
<div class="code">

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com

servers:
  - host: my-project.example.com
    port: 22
    user: my-project
    ssh_key: ~/.ssh/id_rsa

services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
      interval: 10s
      timeout: 5s
      retries: 3

routes:
  - path: /
    strip_prefix: false

dependencies:
  - name: postgres
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_USER=${POSTGRES_USER:-postgres}
      - POSTGRES_DB=${POSTGRES_DB:-app}
```

</div>
</div>

## About FTL

FTL is a deployment tool that uses SSH to manage Docker containers on remote servers. It's a single binary that handles deployment without requiring a container registry or CI/CD pipeline.

### How It Works

FTL executes Docker commands over SSH, handling image builds, container deployment, and infrastructure configuration. It manages Nginx routing, Let's Encrypt certificates, health checks, and database containers with persistent storage.

### When to Use

FTL is optimal for deployments to 1-5 servers where direct SSH access is available. Not suitable for distributed systems or multi-region deployments.
