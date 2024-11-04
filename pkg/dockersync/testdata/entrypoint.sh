#!/bin/bash

# Start Docker daemon
dockerd &

# Wait for Docker daemon to be ready
timeout=30
while ! docker info >/dev/null 2>&1; do
    timeout=$((timeout - 1))
    if [ $timeout -le 0 ]; then
        echo "Failed to start Docker daemon"
        exit 1
    fi
    sleep 1
done

# Start SSH server
/usr/sbin/sshd -D
