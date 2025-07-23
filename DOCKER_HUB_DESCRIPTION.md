# Unix User Exporter

A lightweight Prometheus exporter that monitors currently logged-in users on Unix/Linux systems by directly parsing the utmp file. Perfect for tracking active sessions on servers, workstations, and shared systems.

## Key Features

- **Direct utmp file parsing** - No dependency on system commands, works reliably in containers
- **Container-friendly** - Works perfectly in Docker without special privileges
- **Multi-architecture support** - Available for amd64, arm64, and armv7
- **Real-time monitoring** - Updates metrics every 15 seconds
- **Comprehensive metrics** - Username, TTY, origin IP, login time, session counts

## Quick Start

```bash
# Run with Docker
docker run -d \
  --name unix-user-exporter \
  -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest
```

## Metrics Provided

- `unix_users_logged_in_total`: Total number of users currently logged in
- `unix_user_session_info`: Detailed session information with labels (username, tty, origin, login_time)
- `unix_user_session_count`: Number of active sessions per user
- `unix_user_session_by_ip`: Number of sessions per origin IP/hostname

## Usage Examples

### Basic Usage
```bash
docker run -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest
```

### With Custom Port
```bash
docker run -p 8080:8080 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest \
  --web.listen-address=:8080
```

### With Debug Logging
```bash
docker run -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest \
  --debug
```

## Architecture Support

- `linux/amd64` - x86-64 compatible CPUs
- `linux/arm64` - 64-bit ARM CPUs (Raspberry Pi 4, Apple Silicon)
- `linux/arm/v7` - 32-bit ARM CPUs (Raspberry Pi 3 and earlier)

## Requirements

- Linux system with `/var/run/utmp` file
- Docker or container runtime
- Read access to the utmp file

## Configuration Options

- `--web.listen-address`: Address to listen on (default: `:32142`)
- `--web.telemetry-path`: Metrics endpoint path (default: `/metrics`)
- `--utmp-path`: Path to utmp file (default: `/var/run/utmp`)
- `--debug`: Enable debug logging

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'unix-users'
    static_configs:
      - targets: ['localhost:32142']
    scrape_interval: 30s
```

## Performance

- **Memory**: ~5-10MB
- **CPU**: Minimal usage
- **I/O**: Single file read every 15 seconds
- **Security**: Read-only access, no sensitive data exposure

Perfect for monitoring user activity on servers, workstations, and shared systems in containerized environments.

**Source**: https://github.com/zerepl/unix-user-exporter
**License**: MIT
