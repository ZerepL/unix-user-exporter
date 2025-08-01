FROM golang:1.21-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o unix-user-exporter .

FROM alpine:latest

# Install debugging tools and wget for health check
RUN apk --no-cache add ca-certificates file wget

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/unix-user-exporter .

# Make sure it's executable
RUN chmod +x ./unix-user-exporter

# Create a non-root user for security
RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 32142

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:32142/metrics || exit 1

ENTRYPOINT ["./unix-user-exporter"]
