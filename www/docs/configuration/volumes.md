---
title: Volumes Configuration
description: Configure persistent storage volumes for your FTL deployment
---

# Volumes Configuration

The `volumes` section defines persistent storage volumes that can be mounted into your services and dependencies.

## Configuration

```yaml
volumes:
  - postgres_data
```

Volumes are defined as a simple list of volume names. These volumes can then be referenced in service and dependency configurations.

## Usage in Dependencies

Volumes are commonly used with dependencies to persist data:

```yaml
dependencies:
  - name: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

The format is `volume_name:container_path`, where:

- `volume_name` is the name defined in the volumes section
- `container_path` is the mount point inside the container

## Example

```yaml
# Define volumes
volumes:
  - postgres_data

# Use volumes in dependencies
dependencies:
  - name: postgres
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

This configuration:

- Creates a named volume called `postgres_data`
- Mounts the volume into the PostgreSQL container for persistent storage
