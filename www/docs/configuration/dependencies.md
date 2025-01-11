---
title: Dependencies Configuration
description: Configure supporting services like databases for your FTL deployment
---

# Dependencies Configuration

The `dependencies` section defines supporting services like databases that your application requires.

## Configuration Fields

```yaml
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

| Field     | Description                              |
| --------- | ---------------------------------------- |
| `name`    | Unique identifier for the dependency     |
| `image`   | Docker image to use for this dependency  |
| `volumes` | Volume mounts for persistent storage     |
| `env`     | Environment variables for the dependency |

## Environment Variables

Dependencies support environment variable substitution with required variables and optional variables with defaults:

```yaml
dependencies:
  - name: postgres
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD} # Required
      - POSTGRES_USER=${POSTGRES_USER:-postgres} # Optional with default
      - POSTGRES_DB=${POSTGRES_DB:-app} # Optional with default
```

Required variables must be set in the environment before running FTL commands. Optional variables will use their default values if not set.

## Example

```yaml
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

This configuration:

- Creates a PostgreSQL database service
- Uses the official PostgreSQL 16 image
- Mounts persistent storage for the database
- Sets required and optional environment variables
