---
title: GitHub Actions Integration
description: Deploy with FTL using GitHub Actions workflows
---

# GitHub Actions Integration

Deploy your applications automatically using FTL and GitHub Actions. The official FTL GitHub Action handles installation, SSH setup, and deployment in your CI/CD pipeline.

## Quick Start

Add the FTL Deploy action to your workflow:

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy with FTL
        uses: yarlson/ftl-deploy-action@v1
        with:
          ssh-key: ${{ secrets.SSH_PRIVATE_KEY }}
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

## Setup

### 1. Prepare SSH Key

Generate a dedicated SSH key for deployments:

```bash
ssh-keygen -t ed25519 -C "github-deploy" -f ~/.ssh/ftl_github
```

Add the public key to your server's `~/.ssh/authorized_keys`.

### 2. Add SSH Key to GitHub Secrets

1. Go to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `SSH_PRIVATE_KEY`
4. Value: Contents of `~/.ssh/ftl_github` (complete file)

### 3. Create Workflow File

Create `.github/workflows/deploy.yml` in your repository with the deployment workflow.

## Common Patterns

### Deploy on Tag

Deploy when you create a release tag:

```yaml
on:
  push:
    tags:
      - "v*"
```

### Deploy Staging and Production

Use different workflows or environments:

```yaml
jobs:
  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    # ... deploy to staging

  deploy-production:
    if: github.ref == 'refs/heads/main'
    # ... deploy to production
```

### First Deployment

For initial server setup, enable `run-setup`:

```yaml
- uses: yarlson/ftl-deploy-action@v1
  with:
    ssh-key: ${{ secrets.SSH_PRIVATE_KEY }}
    run-setup: "true"
```

After the first deployment, remove or set to `'false'`.

## Action Reference

See the complete [action documentation](https://github.com/yarlson/ftl-deploy-action) for all inputs and advanced usage patterns.

## Troubleshooting

### Action Fails with SSH Error

Verify your SSH key in GitHub Secrets includes the complete key including BEGIN/END markers:

```
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

### Build Fails on Runner

Ensure your Dockerfile and build context are correctly configured in `ftl.yaml`. The action uses the same build process as local FTL.

### Environment Variables Not Working

Make sure environment variables are passed in the `env` section of the action step, not as `with` inputs.

## Next Steps

- [Configure health checks](/guides/health-checks) for zero-downtime deployments
- [Set up monitoring](/core-tasks/logging) with FTL logs
