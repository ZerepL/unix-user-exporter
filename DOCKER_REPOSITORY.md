# Unix User Exporter - Docker Repository

This repository contains the Docker images for the Unix User Exporter, a lightweight Prometheus exporter that monitors currently logged-in users on Unix/Linux systems.

## Features

- **Direct utmp file parsing** - Reliable operation in containers without system command dependencies
- **Multi-architecture support** - Images available for amd64, arm64, and armv7
- **Container-optimized** - Small image size, minimal resource usage
- **Production-ready** - Stable, secure, and well-tested

## Supported Architectures

- `linux/amd64` - Intel/AMD 64-bit processors
- `linux/arm64` - 64-bit ARM processors (Raspberry Pi 4, Apple Silicon, AWS Graviton)
- `linux/arm/v7` - 32-bit ARM processors (Raspberry Pi 3 and earlier)

## Quick Start

```bash
docker run -d \
  --name unix-user-exporter \
  -p 32142:32142 \
  -v /var/run/utmp:/var/run/utmp:ro \
  zerepl/unix-user-exporter:latest
```

## Tags

- `latest` - Latest stable release
- `v2.x.x` - Specific version tags
- `main` - Development builds (not recommended for production)

## Image Details

- **Base Image**: Alpine Linux (minimal footprint)
- **Size**: ~15MB compressed
- **User**: Runs as non-root user
- **Exposed Port**: 32142
- **Volume**: `/var/run/utmp` (read-only)

## Usage in Production

### Docker Compose
```yaml
version: '3.8'
services:
  unix-user-exporter:
    image: zerepl/unix-user-exporter:latest
    ports:
      - "32142:32142"
    volumes:
      - /var/run/utmp:/var/run/utmp:ro
    restart: unless-stopped
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: unix-user-exporter
spec:
  selector:
    matchLabels:
      app: unix-user-exporter
  template:
    metadata:
      labels:
        app: unix-user-exporter
    spec:
      containers:
      - name: unix-user-exporter
        image: zerepl/unix-user-exporter:latest
        ports:
        - containerPort: 32142
        volumeMounts:
        - name: utmp
          mountPath: /var/run/utmp
          readOnly: true
      volumes:
      - name: utmp
        hostPath:
          path: /var/run/utmp
```

## Security

- Images are built from source with reproducible builds
- No privileged access required
- Read-only file system access
- Regular security updates

## Support

- **Documentation**: https://github.com/zerepl/unix-user-exporter
- **Issues**: https://github.com/zerepl/unix-user-exporter/issues
- **License**: MIT

## Build Information

Images are automatically built and published using GitHub Actions with multi-architecture support. Each release is tagged and includes security scanning.

**Source Repository**: https://github.com/zerepl/unix-user-exporter
