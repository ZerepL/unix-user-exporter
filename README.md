# unix-user-exporter

Prometheus exporter that exposes metrics about currently logged-in users on a Unix system using the `w` command. Ideal for tracking active sessions on shared servers.

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

## Installation

### Using Go

```bash
go get github.com/zerepl/unix-user-exporter
go install github.com/zerepl/unix-user-exporter
```

### Using Docker

```bash
# Build locally
docker build -t unix-user-exporter .
docker run -p 32142:32142 -v /var/run/utmp:/var/run/utmp:ro unix-user-exporter

# Or pull from Docker Hub
docker pull zerepl/unix-user-exporter:latest
docker run -p 32142:32142 -v /var/run/utmp:/var/run/utmp:ro zerepl/unix-user-exporter:latest
```

The Docker image supports the following architectures:
- `amd64` - x86-64 compatible CPUs
- `arm64` - 64-bit ARM CPUs (e.g., Raspberry Pi 4 with 64-bit OS)
- `armv7` - 32-bit ARM CPUs (e.g., Raspberry Pi 3 and earlier)

> **Important**: To monitor host system users (not container users), you must mount the host's `/var/run/utmp` file into the container as shown above.

## Usage

```bash
# Run with default settings (port 32142)
unix-user-exporter

# Run with custom port
unix-user-exporter --web.listen-address=:8080

# Run with custom metrics path
unix-user-exporter --web.telemetry-path=/metrics/users
```

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'unix-users'
    static_configs:
      - targets: ['localhost:32142']
```

## Grafana Dashboard

A sample Grafana dashboard is available in the `dashboards` directory.

## License

MIT
