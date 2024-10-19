# Aerie

Aerie is a powerful deployment tool that simplifies the process of setting up servers and deploying applications. It's designed to make deployment easy and reliable, even for developers who aren't experts in server management or advanced deployment techniques.

## Key Features

- **Simple Configuration**: Define your entire infrastructure in a single YAML file.
- **Automated Server Setup**: Quickly provision new servers with all necessary software.
- **Zero-Downtime Deployments**: Update your applications without any service interruption.
- **Docker-based**: Leverage the power and simplicity of Docker for consistent deployments.
- **Automatic HTTPS**: Built-in Nginx proxy handles SSL/TLS certificate management.

## Installation

To install Aerie, you need to have Go installed on your system. Then, you can use the following command:

```bash
go install github.com/yarlson/aerie@latest
```

## Configuration

Create an `aerie.yaml` file in your project root. Here's an example:

```yaml
project:
  name: my-project
  domain: my-project.example.com
  email: my-project@example.com

servers:
  - host: my-project.example.com
    port: 22
    user: my-project
    ssh_key: ~/.ssh/id_rsa

services:
  - name: my-app
    image: my-app:latest
    port: 80
    health_check:
      path: /
      interval: 10s
      timeout: 5s
      retries: 3
    routes:
      - path: /
        strip_prefix: false

storages:
  - name: my-app-storage
    image: my-app-storage:latest
    volumes:
      - my-app-storage:/var/www/html/storage

volumes:
  - name: my-app-storage
    path: /var/www/html/storage
```

This configuration defines your project, servers, services, and storage requirements.

## Usage

Aerie provides two main commands: `setup` and `deploy`.

### Setup

The `setup` command prepares your servers for deployment:

```bash
aerie setup
```

This command:

1. Installs essential software and Docker
2. Configures the server firewall
3. Creates a new user and adds them to the Docker group
4. Sets up SSH access for the new user

Run this command once for each new server before deploying.

### Deploy

The `deploy` command is where the magic happens. It deploys your application to all configured servers:

```bash
aerie deploy
```

## How Aerie Deploys Your Application

Aerie uses a sophisticated deployment process to ensure your application is always available, even during updates. Here's what happens when you run `aerie deploy`:

1. **Configuration Parsing**: Aerie reads your `aerie.yaml` file to understand your infrastructure.

2. **Server Connection**: It securely connects to each server using SSH.

3. **Docker Network Creation**: A dedicated Docker network is created for your project, ensuring proper isolation.

4. **Image Pulling**: The latest versions of your Docker images are pulled to ensure you're deploying the most recent code.

5. **Zero-Downtime Deployment**: For each service:

   - A new container is started with the updated image and configuration.
   - Health checks are performed to ensure the new container is ready.
   - Once healthy, traffic is instantly switched to the new container.
   - The old container is gracefully stopped and removed.

   This process ensures that your application remains available throughout the update.

6. **Proxy Configuration**: An Nginx proxy is automatically configured to route traffic to your services, handle SSL/TLS, and provide automatic HTTPS.

7. **Cleanup**: Any unused resources are cleaned up to keep your server tidy.

The entire process is automatic and requires no manual intervention. You can deploy updates as frequently as needed without worrying about downtime or complex deployment procedures.

## Benefits of Aerie's Deployment Process

- **No Downtime**: Your application remains available during updates.
- **Automatic Rollback**: If a new version fails health checks, the old version continues to run.
- **Consistency**: Every deployment follows the same process, reducing the chance of errors.
- **Simplicity**: Complex deployment logic is handled for you, so you can focus on developing your application.
- **Scalability**: Easily deploy to multiple servers or add new services as your project grows.

## Development

To contribute to Aerie, clone the repository and install the dependencies:

```bash
git clone https://github.com/yarlson/aerie.git
cd aerie
go mod download
```

Run the tests:

```bash
go test ./...
```

## License

[MIT License](LICENSE)
