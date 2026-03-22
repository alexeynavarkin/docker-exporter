# Grafana Dashboard

## Import

1. In Grafana, go to **Dashboards > Import**
2. Upload `dashboard.json` or paste its contents
3. Select your Prometheus datasource when prompted

## Variables

| Variable | Description |
|----------|-------------|
| `Target` | Prometheus scrape target (`instance` label). Filter by Docker host when running multiple exporters. |
| `Container` | Filter by container name. Cascades from the selected target. |
| `Service` | Filter by Docker Swarm service name. Cascades from the selected target. |

## Panels

| Section | Panel | Description |
|---------|-------|-------------|
| Overview | Running Containers | Count of containers currently reporting metrics |
| Overview | OOM Events | Total OOM kills in the selected time range |
| CPU | CPU Usage | Usermode + kernelmode CPU time per container |
| CPU | CPU Throttled Time | Time the container was throttled by the CPU scheduler |
| Memory | Memory Usage | Current RSS memory usage per container |
| Network | Network RX / TX | Cumulative bytes received and transmitted per container |
| Network | Network Drops / Errors | Dropped and errored packets per container |
| Events | OOM Events | OOM kill events over time, displayed as a bar chart |
