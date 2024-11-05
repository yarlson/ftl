FROM ubuntu:22.04

# Install necessary packages
RUN apt-get update && apt-get install -y \
    openssh-server \
    curl \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Set up SSH server
RUN mkdir /var/run/sshd
RUN echo 'root:testpassword' | chpasswd
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config
RUN mkdir -p /root/.ssh

# Expose SSH port
EXPOSE 22

# Start SSH server and Docker daemon
ENTRYPOINT ["/usr/sbin/sshd", "-D"]
