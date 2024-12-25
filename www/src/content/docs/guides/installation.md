---
title: Installing FTL
description: Complete guide to installing the FTL deployment tool on different platforms and environments
---

This guide covers everything you need to know about installing FTL on your system, including prerequisites, installation methods, and verifying your installation.

## Prerequisites

Before installing FTL, ensure your system meets these requirements:

- **Operating System**:

  - macOS 10.15 or later
  - Linux (Ubuntu 18.04+, Debian 10+, CentOS 7+)
  - Windows 10/11 with WSL2 (Windows Subsystem for Linux)

- **Required Software**:

  - SSH client installed and configured
  - Git (optional, but recommended)

- **Hardware**:
  - Minimum 4GB RAM
  - 1GB free disk space

## Installation Methods

Choose the installation method that best suits your environment:

### 1. Using Homebrew (Recommended for macOS and Linux)

This is the simplest method for macOS users and Linux users with Homebrew installed.

```bash
# Add the FTL tap
brew tap yarlson/ftl

# Install FTL
brew install ftl
```

If you don't have Homebrew installed, you can install it first:

```bash
# On macOS/Linux
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### 2. Direct Download

You can download pre-compiled binaries for your platform:

1. Open your terminal
2. Run the following command:

```bash
# Download and extract FTL
curl -L https://github.com/yarlson/ftl/releases/latest/download/ftl_$(uname -s)_$(uname -m).tar.gz | tar xz

# Move FTL to your bin directory
sudo mv ftl /usr/local/bin/

# Make it executable (if needed)
sudo chmod +x /usr/local/bin/ftl
```

> 💡 **Tip**: If you don't have sudo access or prefer to install locally, you can place the FTL binary in `~/bin/` instead. Make sure this directory is in your PATH.

### 3. Building from Source

For developers who want to build from source or contribute to FTL:

1. Ensure you have Go 1.19 or later installed:

```bash
go version
```

2. Install FTL using Go:

```bash
go install github.com/yarlson/ftl@latest
```

The binary will be installed to your `$GOPATH/bin` directory. Make sure this is in your PATH.

## Verifying Your Installation

After installing FTL, verify it's working correctly:

1. Check the FTL version:

```bash
ftl version
```

2. View the help information:

```bash
ftl help
```

Expected output should look similar to:

```
FTL - Faster Than Light Deployment Tool
Version: x.y.z

Usage:
  ftl [command]

Available Commands:
  build       Build service images
  deploy      Deploy services to servers
  help        Help about any command
  ...
```

## Post-Installation Setup

After installing FTL, you should:

1. **Configure SSH Keys**:
   Make sure you have SSH keys generated:

   ```bash
   # Generate new SSH key if needed
   ssh-keygen -t ed25519 -C "your_email@example.com"
   ```

2. **Set up environment variables**:
   Create a `.env` file in your project directory (add to `.gitignore`):
   ```bash
   # .env example
   FTL_PROJECT_NAME=my-project
   FTL_DOMAIN=example.com
   ```

## Troubleshooting Installation

### Common Issues

1. **Command Not Found**

   ```bash
   # Add to PATH (for Bash)
   echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

2. **Permission Denied**
   ```bash
   # Fix permissions
   sudo chown $USER /usr/local/bin/ftl
   chmod +x /usr/local/bin/ftl
   ```

## Updating FTL

To update FTL to the latest version:

```bash
# If installed via Homebrew
brew upgrade ftl

# If installed via Go
go install github.com/yarlson/ftl@latest

```

## Platform-Specific Notes

### macOS

- If you're using macOS Catalina or later, you might need to approve the binary in System Preferences > Security & Privacy after first run

### Windows (WSL2)

1. Install WSL2 if not already installed:

   ```powershell
   wsl --install
   ```

2. Install Ubuntu or your preferred Linux distribution from the Microsoft Store

3. Follow the Linux installation instructions within your WSL2 environment
