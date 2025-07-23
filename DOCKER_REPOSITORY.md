# Docker Repository Overview: unix-user-exporter

## Repository Information

- **Repository Name**: zerepl/unix-user-exporter
- **Docker Hub URL**: [https://hub.docker.com/r/zerepl/unix-user-exporter](https://hub.docker.com/r/zerepl/unix-user-exporter)
- **Description**: Prometheus exporter that exposes metrics about currently logged-in users on a Unix system using the `w` command.

## Available Tags

| Tag | Description | Architectures | Size | Last Updated |
|-----|-------------|--------------|------|-------------|
| `latest` | Latest stable build | amd64, arm64, armv7 | ~12MB | July 22, 2025 |
| `v1.0.0` | Initial release | amd64, arm64, armv7 | ~12MB | July 22, 2025 |

## Usage

Pull the image:
```bash
docker pull zerepl/unix-user-exporter:latest
```

Run the container:
```bash
docker run -p 32142:32142 zerepl/unix-user-exporter:latest
```

With custom port:
```bash
docker run -p 8080:32142 zerepl/unix-user-exporter:latest
```

With custom metrics path:
```bash
docker run -p 32142:32142 zerepl/unix-user-exporter:latest --web.telemetry-path=/metrics/users
```

## Environment Variables

This image doesn't currently use environment variables for configuration. All configuration is done via command-line flags.

## Exposed Ports

- **32142**: Default HTTP port for metrics endpoint

## Volumes

No volumes are required for this container.

## Health Check

You can use the following health check in your Docker Compose or Docker run commands:

```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "--spider", "http://localhost:32142/"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 5s
```

## Docker Compose Example

```yaml
version: '3'

services:
  unix-user-exporter:
    image: zerepl/unix-user-exporter:latest
    ports:
      - "32142:32142"
    volumes:
      - /var/run/utmp:/var/run/utmp:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:32142/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 5s
```

## Building the Image Locally

```bash
git clone https://github.com/zerepl/unix-user-exporter.git
cd unix-user-exporter
docker build -t unix-user-exporter .
```

## Security Considerations

- The container runs as root by default
- The exporter only provides read-only metrics
- No sensitive information is exposed beyond what the `w` command shows

## Maintenance and Updates

- Images are automatically rebuilt when new versions are released
- Security patches are applied regularly
- Follow the repository on Docker Hub for update notifications

## License

This image is distributed under the MIT License.
