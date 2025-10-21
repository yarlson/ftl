---
title: Installing FTL
description: Learn how to install the FTL CLI tool on your system using various methods
---

# Installing FTL

FTL can be installed in several ways depending on your operating system and preferences. Choose the method that works best for your environment.

## Installation Methods

### Via Homebrew (macOS and Linux)

The recommended way to install FTL on macOS and Linux is through Homebrew:

```bash
# Add the FTL tap
brew tap yarlson/ftl

# Install FTL
brew install ftl
```

### Direct Download (All Platforms)

Download and install the latest release directly from GitHub:

```bash
# Download and extract the binary
curl -L https://github.com/yarlson/ftl/releases/latest/download/ftl_$(uname -s)_$(uname -m).tar.gz | tar xz

# Move the binary to your PATH
sudo mv ftl /usr/local/bin/
```

### Build from Source

For developers who want to build from source or contribute to FTL:

```bash
# Install using Go
go install github.com/yarlson/ftl@latest
```

::: tip
Make sure you have Go 1.21 or later installed before building from source.
:::

## Verifying the Installation

After installation, verify that FTL is properly installed:

```bash
ftl version
```

You should see output showing the current version of FTL.

## System Requirements

Before installing FTL, ensure your system meets these requirements:

### Local Development Machine

- **Operating System:**
  - macOS 10.15 or later
  - Linux (major distributions)
  - Windows 10/11 with WSL2

- **Required Software:**
  - Git
  - Docker Desktop or Docker Engine
  - SSH client

- **Hardware:**
  - 4GB RAM (minimum)
  - 2 CPU cores (recommended)
  - 1GB free disk space

### Target Deployment Server

- **Operating System:**
  - Ubuntu 20.04 LTS or newer (recommended)
  - Debian 11 or newer
  - Other Linux distributions with Docker support

- **Hardware:**
  - 1GB RAM (minimum)
  - 1 CPU core (minimum)
  - 20GB disk space (recommended)

- **Network:**
  - SSH access (port 22)
  - Sudo privileges
  - Ports 80 and 443 available for web traffic

## Post-Installation Setup

After installing FTL, you should:

1. Create an SSH key if you haven't already:

   ```bash
   ssh-keygen -t ed25519 -C "your_email@example.com"
   ```

2. Verify Docker is installed and running:

   ```bash
   docker --version
   docker ps
   ```

3. Verify FTL installation:
   ```bash
   ftl version
   ```

::: tip Registry Authentication
If you plan to use a Docker registry, ensure you have:

- Registry URL
- Username/password credentials (token-based auth not supported)
- Network access to the registry from both local machine and server
  :::

## Troubleshooting

### Common Installation Issues

1. **Permission Denied When Running FTL**

   ```bash
   sudo chmod +x /usr/local/bin/ftl
   ```

2. **Binary Not Found in PATH**

   ```bash
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Homebrew Installation Fails**
   ```bash
   brew update && brew doctor
   ```

For more detailed troubleshooting information, see the [Troubleshooting Guide](/reference/troubleshooting).

## Next Steps

Once FTL is installed, you can:

1. Move on to [Configuration](./configuration.md) to set up your first project
2. Read the [CLI Commands Reference](/reference/cli-commands) to learn about available commands
3. Check out the [Guides](/guides/) section for detailed tutorials

::: warning Note
Remember to keep FTL updated to receive the latest features and security updates. If using Homebrew, run `brew upgrade ftl` periodically.
:::
