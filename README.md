# unix-user-exporter

A lightweight Prometheus exporter that monitors currently logged-in users on Unix/Linux systems by directly parsing the utmp file. Perfect for tracking active sessions on servers, workstations, and shared systems.

## Features

- **Direct utmp file parsing** - No dependency on system commands, works reliably in containers
- **Comprehensive user session metrics** including:
  - Username and session count
  - TTY/terminal information
  - Origin IP addresses and hostnames
  - Login timestamps
  - Session duration tracking
- **Container-friendly** - Works perfectly in Docker without special privileges
- **Multi-architecture support** - Available for amd64, arm64, and armv7
- **Real-time monitoring** - Updates metrics every 15 seconds
- **Debug mode** - Detailed logging for troubleshooting

## Metrics

The exporter provides the following Prometheus metrics:

- `unix_users_logged_in_total`: Total number of users currently logged in
- `unix_user_session_info`: Detailed information about each user session with labels:
  - `username`: The logged-in user's name
  - `from`: Origin IP address or hostname (or "-" if local)
  - `tty`: Terminal/TTY identifier
  - `login_time`: When the session started
- `unix_user_session_count`: Number of active sessions per user
- `unix_user_session_by_ip`: Number of sessions per origin IP/hostname

## Installation

### Using Docker (Recommended)

```bash
# Run with Docker
docker run -d \
  --name unix-user-exporter \
  -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest

# Or build locally
docker build -t unix-user-exporter .
docker run -d \
  --name unix-user-exporter \
  -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  unix-user-exporter
```

### Using Go

```bash
# Install from source
go install github.com/zerepl/unix-user-exporter@latest

# Or build locally
git clone https://github.com/zerepl/unix-user-exporter.git
cd unix-user-exporter
go build .
./unix-user-exporter
```

### Multi-Architecture Docker Images

The Docker images support multiple architectures:
- `linux/amd64` - x86-64 compatible CPUs
- `linux/arm64` - 64-bit ARM CPUs (e.g., Raspberry Pi 4 with 64-bit OS, Apple Silicon)
- `linux/arm/v7` - 32-bit ARM CPUs (e.g., Raspberry Pi 3 and earlier)

## Usage

### Basic Usage

```bash
# Run with default settings (port 32142)
unix-user-exporter

# Run with custom port
unix-user-exporter --web.listen-address=:8080

# Run with custom metrics path
unix-user-exporter --web.telemetry-path=/custom-metrics

# Run with custom utmp file location
unix-user-exporter --utmp-path=/custom/path/to/utmp
```

### Docker Usage

```bash
# Basic usage
docker run -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest

# With custom configuration
docker run -p 8080:8080 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest \
  --web.listen-address=:8080 \
  --web.telemetry-path=/metrics/users

# With debug logging
docker run -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest \
  --debug
```

> **Important**: The `/var/run/utmp` file must be mounted read-only into the container for the exporter to detect host system users.

## Prometheus Configuration

Add the following job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'unix-users'
    static_configs:
      - targets: ['localhost:32142']
    scrape_interval: 30s
    metrics_path: /metrics
```

## Example Metrics Output

```
# HELP unix_users_logged_in_total Total number of users currently logged in
# TYPE unix_users_logged_in_total gauge
unix_users_logged_in_total 3

# HELP unix_user_session_info Information about user sessions
# TYPE unix_user_session_info gauge
unix_user_session_info{from="192.168.1.100",login_time="2025-07-23 14:16",tty="pts/0",username="alice"} 1
unix_user_session_info{from="-",login_time="2025-07-23 12:30",tty="tty1",username="bob"} 1
unix_user_session_info{from="10.0.0.50",login_time="2025-07-23 15:45",tty="pts/1",username="alice"} 1

# HELP unix_user_session_count Number of sessions per user
# TYPE unix_user_session_count gauge
unix_user_session_count{username="alice"} 2
unix_user_session_count{username="bob"} 1

# HELP unix_user_session_by_ip Number of sessions per origin IP
# TYPE unix_user_session_by_ip gauge
unix_user_session_by_ip{ip="192.168.1.100"} 1
unix_user_session_by_ip{ip="10.0.0.50"} 1
```

## Grafana Dashboard

A sample Grafana dashboard is available in the `dashboards` directory. Import it to visualize:
- Total logged-in users over time
- User session details and duration
- Login activity by IP address
- Most active users and terminals

## Troubleshooting

### Docker Version Shows No Users

If the Docker version shows `unix_users_logged_in_total 0` even when users are logged in:

1. **Ensure proper volume mount**: Mount the utmp file correctly:
   ```bash
   -v /var/run/utmp:/var/run/utmp:ro
   ```

2. **Check file permissions**: Verify the utmp file is readable:
   ```bash
   ls -la /var/run/utmp
   ```

3. **Verify utmp file exists**: The utmp file should exist and contain data:
   ```bash
   file /var/run/utmp
   hexdump -C /var/run/utmp | head -5
   ```

### Debug Mode

Enable debug logging to see detailed parsing information:

```bash
# Direct execution
unix-user-exporter --debug

# Docker
docker run -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  unix-user-exporter --debug
```

Debug output shows:
- utmp file size and number of entries
- Each parsed entry with type, user, TTY, and host information
- Which entries are considered active user sessions
- Parsing errors or data inconsistencies

### Common Issues

**Q: Why don't I see GUI users?**  
A: Some desktop environments may not write to utmp. The exporter shows users with active terminal sessions.

**Q: Can I monitor remote systems?**  
A: Yes, mount the remote system's utmp file via NFS, SSHFS, or copy it periodically.

**Q: Does this work on macOS/BSD?**  
A: This is designed for Linux systems. BSD variants have different utmp formats.

## Performance

- **Memory usage**: ~5-10MB
- **CPU usage**: Minimal, only during 15-second collection intervals
- **Disk I/O**: Single small file read every 15 seconds
- **Network**: HTTP metrics endpoint only

## Security Considerations

- The exporter only reads the utmp file (read-only access)
- No sensitive user data is exposed beyond usernames and login times
- Runs as non-root user in Docker containers
- No network connections except for serving metrics

## Development

### Building from Source

```bash
git clone https://github.com/zerepl/unix-user-exporter.git
cd unix-user-exporter
go mod download
go build .
```

### Running Tests

```bash
go test ./...
```

### Building Multi-Architecture Docker Images

```bash
# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t zerepl/unix-user-exporter:latest .
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Changelog

### v2.0.0
- **Breaking**: Replaced system command dependency with direct utmp parsing
- **New**: Full Docker container support without special privileges
- **New**: Multi-architecture Docker images
- **Improved**: More reliable user session detection
- **Improved**: Better error handling and debug output
- **Fixed**: Container isolation issues

### v1.0.0
- Initial release with system command-based parsing
