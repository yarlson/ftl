---
title: Environment Variables
description: Reference documentation for environment variable handling in FTL
---

# Environment Variables

FTL provides a flexible environment variable substitution system that allows you to configure your services without hardcoding sensitive or environment-specific values.

## Variable Syntax

FTL supports two types of environment variable substitution:

### Required Variables

Required variables must be set in the environment when running FTL commands. If a required variable is not set, FTL will return an error.

```yaml
${VARIABLE_NAME}
```

Example usage:

```yaml
env:
  - DATABASE_PASSWORD=${DB_PASSWORD}
  - API_KEY=${API_KEY}
```

### Optional Variables with Defaults

Optional variables can include a default value that will be used if the variable is not set in the environment.

```yaml
${VARIABLE_NAME:-default_value}
```

Example usage:

```yaml
env:
  - POSTGRES_USER=${POSTGRES_USER:-postgres}
  - POSTGRES_DB=${POSTGRES_DB:-app}
  - LOG_LEVEL=${LOG_LEVEL:-info}
```

## Variable Scope

Environment variables can be used in several sections of the `ftl.yaml` configuration:

- Service environment variables
- Dependency environment variables
- Build arguments
- Container configurations

## Usage Examples

### In Dependencies

```yaml
dependencies:
  - name: postgres
    image: postgres:16
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_USER=${POSTGRES_USER:-postgres}
      - POSTGRES_DB=${POSTGRES_DB:-app}
```

### In Services

```yaml
services:
  - name: web-app
    image: my-app:latest
    env:
      - DATABASE_URL=${DATABASE_URL}
      - API_KEY=${API_KEY}
      - NODE_ENV=${NODE_ENV:-production}
```

## Best Practices

1. **Security**
   - Never commit sensitive values directly in configuration files
   - Use required variables (`${VAR_NAME}`) for sensitive information
   - Use optional variables with defaults for non-sensitive configuration

2. **Defaults**
   - Provide sensible defaults for optional variables
   - Document the implications of default values
   - Use defaults for development-friendly configurations

3. **Naming**
   - Use uppercase letters and underscores for variable names
   - Choose descriptive names that indicate the variable's purpose
   - Prefix variables with the service name when appropriate

## Environment File

While FTL doesn't directly load `.env` files, you can source them before running FTL commands:

```bash
source .env && ftl deploy
```

This pattern allows you to:

- Keep environment variables organized
- Share non-sensitive defaults with your team
- Override variables per environment

## Common Variables

These are some commonly used variables in FTL deployments:

| Variable            | Purpose                  | Example                      |
| ------------------- | ------------------------ | ---------------------------- |
| `POSTGRES_PASSWORD` | PostgreSQL root password | `${POSTGRES_PASSWORD}`       |
| `POSTGRES_USER`     | PostgreSQL username      | `${POSTGRES_USER:-postgres}` |
| `POSTGRES_DB`       | PostgreSQL database name | `${POSTGRES_DB:-app}`        |

## Error Handling

When a required environment variable is missing, FTL will:

1. Stop the current operation
2. Display an error message indicating which variable is missing
3. Exit with a non-zero status code

Example error:

```bash
Error: Required environment variable 'POSTGRES_PASSWORD' is not set
```
