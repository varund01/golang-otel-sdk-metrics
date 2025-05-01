# golang-otel-sdk-metrics

A sample Golang application that demonstrates how to send metrics, traces, and logs using the OpenTelemetry SDK.

## Overview

This project showcases how to instrument a Golang application with OpenTelemetry to collect and visualize:
- **Traces**: Track request flows through your application
- **Metrics**: Measure application performance and behavior
- **Logs**: Capture application events

## Prerequisites

- Go 1.19 or newer
- Docker (optional, for containerized deployments)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/golang-otel-sdk-metrics.git
cd golang-otel-sdk-metrics
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Set up observability backends

#### Jaeger (for traces)

```bash
mkdir -p ~/Downloads/jaeger
cd ~/Downloads/jaeger
curl -L https://github.com/jaegertracing/jaeger/releases/download/v1.48.0/jaeger-1.48.0-darwin-amd64.tar.gz -o jaeger.tar.gz
tar -xzf jaeger.tar.gz
cd jaeger-1.48.0-darwin-amd64
./jaeger-all-in-one
```

Jaeger UI will be available at http://localhost:16686

#### Prometheus (for metrics)

```bash
mkdir -p ~/Downloads/prometheus
cd ~/Downloads/prometheus
curl -L https://github.com/prometheus/prometheus/releases/download/v2.49.0/prometheus-2.49.0.darwin-amd64.tar.gz -o prometheus.tar.gz
tar -xzf prometheus.tar.gz
cd prometheus-2.49.0.darwin-amd64
```

Create or edit `prometheus.yml` and add the following configuration:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "sample-otel-app"
    scrape_interval: 10s
    static_configs:
      - targets: ['localhost:2222']
```

Start Prometheus:

```bash
./prometheus --config.file=prometheus.yml
```

Prometheus UI will be available at http://localhost:9090

### 4. Run the application

```bash
go run main.go
```

### 5. Test the application

Send a request to the sample endpoint:

```bash
curl localhost:8080
```

## Viewing Telemetry Data

### Traces in Jaeger

1. Open http://localhost:16686
2. Select "sample-otel-app" from the Service dropdown
3. Click "Find Traces"
4. Explore trace details by clicking on any trace in the results

### Metrics in Prometheus

1. Open http://localhost:9090
2. Type metric names like "request_counter" in the query box
3. Click "Execute" to see your metrics data
4. Use the Graph tab to visualize metrics over time

## Project Structure

```
├── main.go               # Application entry point and main instrumentation
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── README.md             # This file
└── ...                   # Additional project files
```

## Key Features

- HTTP server with instrumented endpoints
- Custom metrics collection
- Distributed tracing with context propagation
- Structured logging integrated with traces

## Configuration

The application uses environment variables for configuration:

- `OTEL_SERVICE_NAME`: Service name (default: "sample-otel-app")
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP exporter endpoint (optional)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.