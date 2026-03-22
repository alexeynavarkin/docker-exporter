# docker-exporter

A Prometheus exporter for Docker container metrics. Collects CPU, memory, network stats, and container events (OOM kills) from the Docker API and exposes them at `/metrics` on port `8080`.

Supports Docker Swarm — containers are labeled with `serviceName` and `serviceID` when Swarm labels are present.

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `docker_cpu_stat_ns` | Gauge | `containerName`, `serviceName`, `serviceID`, `type` | CPU usage in nanoseconds. `type`: `usermode`, `kernelmode`, `throttled` |
| `docker_mem_stat_bytes` | Gauge | `containerName`, `serviceName`, `serviceID`, `type` | Memory usage in bytes. `type`: `used` |
| `docker_net_stat_bytes` | Gauge | `containerName`, `serviceName`, `serviceID`, `type` | Network I/O in bytes. `type`: `rx`, `tx`, `drop`, `error` |
| `docker_event` | Counter | `containerName`, `serviceName`, `serviceID`, `eventType` | Docker container events. Currently tracks: `oom` |

## Usage

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  alexnav/docker-exporter
```

### Docker Compose

```yaml
services:
  docker-exporter:
    image: alexnav/docker-exporter
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

Metrics are available at `http://localhost:8080/metrics`.

## Building

```bash
go build -o docker-exporter ./cmd/main.go
./docker-exporter
```

Or with Docker:

```bash
docker build -t docker-exporter .
```

## Requirements

- Docker socket access (`/var/run/docker.sock`)
- Docker API compatible with the running daemon (version negotiation is automatic)
