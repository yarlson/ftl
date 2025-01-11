---
title: Project Settings
description: Configure core project settings in your FTL configuration
---

# Project Settings

The `project` section in your `ftl.yaml` defines core settings that identify your application and configure global deployment behavior.

## Required Fields

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com
```

| Field    | Description                                                          |
| -------- | -------------------------------------------------------------------- |
| `name`   | Project identifier used for container naming and internal references |
| `domain` | Primary domain for your application                                  |
| `email`  | Contact email used for SSL certificate registration                  |

## Environment Variables

Project settings support environment variable substitution:

```yaml
project:
  name: ${PROJECT_NAME}
  domain: ${DOMAIN}
  email: ${EMAIL}
```

All environment variables must be set in the environment before running FTL commands.

## Example

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com
```

This configuration:

- Sets the project name for container identification
- Configures the domain for the application
- Provides an email for SSL certificate management
