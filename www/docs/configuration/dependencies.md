---
title: Dependencies Configuration
description: Configure supporting services like databases for your FTL deployment
---

# Dependencies Configuration

The `dependencies` section defines supporting services (such as databases, caches, or message queues) that your application requires. You can declare each dependency in one of two ways:

- **Short Notation** – Simply specify the service name and optional version (for example, `"mysql:8"`, `"postgres:16"`, or `"redis"`). When using short notation, FTL automatically applies default settings such as typical ports, volumes, and environment variables.
- **Detailed Definition** – Provide a full configuration (with fields like `name`, `image`, `volumes`, and `env`) for additional customization or for services where you want to override the defaults.

## Default Configurations

For many popular services, FTL supplies default settings including:

- **Image Tag**: The default Docker image (e.g. `"postgres:latest"`) is modified by specifying a version. For example, `"postgres:16"` sets the image to `postgres:16`.
- **Ports**: Typical ports are preconfigured (e.g. `5432` for PostgreSQL, `3306` for MySQL).
- **Volumes**: Default named volumes are provided (e.g. `postgres_data:/var/lib/postgresql/data`).
- **Environment Variables**: Common environment variables include placeholders that support expansion. For example:
  ```yaml
  env:
    - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-production-secret}
    - POSTGRES_USER=${POSTGRES_USER:-postgres}
    - POSTGRES_DB=${POSTGRES_DB:-app}
  ```

### Using Short Notation

When you declare a dependency using short notation, FTL will merge the dependency with default values. For example:

```yaml
dependencies:
  - "mysql:8"
  - "redis"
```

- `"mysql:8"`: FTL sets the dependency name to `mysql`, uses the image `mysql:8`, and automatically applies defaults for ports (e.g. `3306`), a named volume (e.g. `mysql_data:/var/lib/mysql`), and environment variables (e.g. `MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-secret}`).
- `"redis"`: FTL uses the image `redis:latest` and applies defaults for ports, volumes, and environment variables as defined by its default configuration.

### Providing a Detailed Definition

If you need to customize settings beyond the defaults, or if you are using a service that does not have a preset configuration, you can provide a detailed definition:

```yaml
dependencies:
  - name: "postgres"
    image: "postgres:16"
    volumes:
      - my_pg_data:/custom/postgres/path
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-secret}
      - POSTGRES_USER=${POSTGRES_USER:-myuser}
```

This approach gives you full control over the dependency configuration, allowing you to override default ports, volume mappings, or environment variables.

## Environment Variable Expansion

FTL supports environment variable expansion in dependency configurations. You can use the following patterns:

- **Optional with default**:

  ```yaml
  SOME_VAR=${SOME_VAR:-defaultValue}
  ```

  If `SOME_VAR` is not set, `defaultValue` is used.

- **Required**:

  ```yaml
  SOME_VAR=${SOME_VAR:?error message}
  ```

  FTL will produce an error if `SOME_VAR` is not set.

- **Plain reference**:
  ```yaml
  SOME_VAR=${SOME_VAR}
  ```
  This inserts the current value of `SOME_VAR` (or an empty string if it’s not set, depending on configuration).

## Examples

### Short Notation (Using Defaults)

```yaml
dependencies:
  - "redis:7"
  - "postgres:16"
```

This configuration:

- Sets the dependency for Redis to use the image `redis:7` and applies default settings for ports, volumes, and environment variables.
- Sets the PostgreSQL dependency to use the image `postgres:16` with default named volumes and environment variable expansion (for example, `POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-production-secret}`).

### Detailed Definition for Customization

```yaml
dependencies:
  - name: "postgres"
    image: "postgres:16"
    volumes:
      - my_pg_data:/custom/postgres/path
    env:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-secret}
      - POSTGRES_USER=${POSTGRES_USER:-myuser}
```

In this case, you explicitly override the default volume mapping and environment variable settings for PostgreSQL.

---

By using short notation, you can quickly set up dependencies with sensible defaults, while a detailed definition gives you the power to fine-tune settings when needed. Environment variable expansion ensures that sensitive values and deployment-specific settings are managed dynamically.
