---
title: Server Configuration
description: Configure deployment server settings in your FTL configuration
---

# Server Configuration

The `server` section in your `ftl.yaml` defines the deployment target settings. All fields in this section are optional and come with smart defaults.

## Configuration Fields

```yaml
server:
  host: my-project.example.com
  port: 22
  user: deployer
  ssh_key: ~/.ssh/id_rsa
```

| Field     | Description                                     | Default Value                         |
| --------- | ----------------------------------------------- | ------------------------------------- |
| `host`    | Hostname or IP address of the deployment server | Value from `project.domain`           |
| `port`    | SSH port for connecting to the server           | `22`                                  |
| `user`    | SSH user for deployment                         | Current system user                   |
| `ssh_key` | Path to SSH private key                         | Auto-detected from standard locations |

## Smart Defaults

FTL implements intelligent defaults to minimize configuration:

1. `host` defaults to your project's domain
2. `user` defaults to your current system username
3. `ssh_key` is auto-detected from standard SSH key locations
4. `port` defaults to the standard SSH port 22

## Environment Variables

Server settings support environment variable substitution:

```yaml
server:
  host: ${SERVER_HOST}
  user: ${SERVER_USER}
  ssh_key: ${SSH_KEY_PATH}
```

All environment variables must be set in the environment before running FTL commands.

## Examples

### Minimal Configuration

When you're happy with the defaults, you can omit the entire `server` section:

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com
```

### Custom Server Settings

When you need to customize server settings:

```yaml
server:
  host: 192.168.1.100
  user: custom-user
  ssh_key: ~/.ssh/custom-key
```

### Mixed Defaults and Custom Settings

You can specify only the fields you want to customize:

```yaml
server:
  host: custom-host.example.com
  user: deployer
```
