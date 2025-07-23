# Unix User Exporter

Prometheus exporter that exposes metrics about currently logged-in users on a Unix system using the `w` command. Ideal for tracking active sessions on shared servers.

## Supported Architectures

- `amd64` - x86-64 compatible CPUs
- `arm64` - 64-bit ARM CPUs (e.g., Raspberry Pi 4 with 64-bit OS)
- `armv7` - 32-bit ARM CPUs (e.g., Raspberry Pi 3 and earlier)

## Features

- Exports total number of logged-in users
- Provides detailed information about each user session including:
  - Username
  - Origin IP address
  - TTY
  - Login time
  - Idle time
  - CPU usage
  - Running commands
- Counts sessions per user
- Counts sessions per origin IP

## Metrics

The exporter provides the following metrics:

- `unix_users_logged_in_total`: Total number of users currently logged in
- `unix_user_session_info`: Information about each user session with labels for username, IP, TTY, etc.
- `unix_user_session_count`: Number of sessions per user
- `unix_user_session_by_ip`: Number of sessions per origin IP

## Usage

```bash
# Pull the image
docker pull zerepl/unix-user-exporter:latest

# Run with default settings (port 32142)
docker run -p 32142:32142 -v /var/run/utmp:/var/run/utmp:ro zerepl/unix-user-exporter:latest

# Run with custom port mapping
docker run -p 8080:32142 -v /var/run/utmp:/var/run/utmp:ro zerepl/unix-user-exporter:latest

# Run with custom metrics path
docker run -p 32142:32142 -v /var/run/utmp:/var/run/utmp:ro zerepl/unix-user-exporter:latest --web.telemetry-path=/metrics/users
```

> **Important**: To monitor host system users (not container users), you must mount the host's `/var/run/utmp` file into the container as shown above.

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'unix-users'
    static_configs:
      - targets: ['localhost:32142']
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
```

## Source Code

The source code for this image is available on GitHub: [https://github.com/zerepl/unix-user-exporter](https://github.com/zerepl/unix-user-exporter)

## License

This image is distributed under the MIT License.
