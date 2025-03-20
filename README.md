# kafka-connect-exporter

An Exporter that provider [Kafka Connect](https://docs.confluent.io/platform/current/connect/index.html) Metrics

## Overview

The kafka connect exporter fetches connector and task status information from configured Kafka Connect clusters.

## Features

### Prometheus Metrics Exposure

#### Connector Metrics

These metrics represent the state of tasks for each connector:

- **`kafka_connect_connector_tasks_running_total`:**
  - Total number of tasks in the `RUNNING` state.
  - **Labels:** `connector`, `host`
- **`kafka_connect_connector_tasks_failed_total`:**
  - Total number of tasks in the `FAILED` state (e.g., due to exceptions reported in status).
  - **Labels:** `connector`, `host`
- **`kafka_connect_connector_tasks_paused_total`:**
  - Total number of tasks in the `PAUSED` state (i.e., administratively paused).
  - **Labels:** `connector`, `host`
- **`kafka_connect_connector_tasks_unassigned_total`:**
  - Total number of tasks in the `Unassigned` state (i.e., not assigned to any worker.)
  - **Labels:** `connector`, `host`
- **`kafka_connect_connector_tasks_total`:**
  - Total number of tasks for the connector.
  - **Labels:** `host`

##### Example

```
# HELP kafka_connect_connector_tasks_running_total Total number of tasks in the `RUNNING` state.
# TYPE kafka_connect_connector_tasks_running_total gauge
kafka_connect_connector_tasks_running_total{connector="example-connector", host="http://example-connect:8083"} 3

# HELP kafka_connect_connector_tasks_failed_total Total number of tasks in the `FAILED` state (e.g., due to exceptions reported in status).
# TYPE kafka_connect_connector_tasks_failed_total gauge
kafka_connect_connector_tasks_failed_total{connector="example-connector", host="http://example-connect:8083"} 1
```

#### Cluster-Level Connector Count

Additionally, the exporter reports the total number of connectors discovered per connect:

- **`kafka_connect_connector_total` (Gauge):**  
  Total number of tasks for the connector.  
  **Labels:** `host`

#### Example

```
# HELP kafka_connect_connector_total Total number of tasks for the connector.
# TYPE kafka_connect_connector_total gauge
kafka_connect_connector_total{host="http://example-connect:8083"} 1
```

### Multi-Host Support

- Configure multiple Kafka Connect instances for parallel metric collection.

### Robust Error Handling

- Provides proper logging and error reporting for HTTP errors and JSON parsing issues.

## Getting Started

### Docker

1. Pull the latest image from Docker Hub:

```bash
docker pull ecubelabs/kafka-connect-exporter:latest
```

2. Run the container with the following command:

```bash
docker run --rm -p 9113:9113 \
  -e PORT=9113 \
  -e METRICS_ENDPOINT="/metrics" \
  -e KAFKA_CONNECT_HOSTS="http://<kafka-connect-host>:8083" \
  -e HEALTH_CHECK_ENDPOINT="/health" \
  ecubelabs/kafka-connect-exporter:latest
```

### Docker compose

You can also run the exporter using Docker Compose.

1. Create a docker-compose.yml file with the following content:

```yml
services:
  kafka-connect-exporter:
    image: ecubelabs/kafka-connect-exporter:latest
    ports:
      - "9113:9113"
    environment:
      - PORT=9113
      - METRICS_ENDPOINT=/metrics
      - KAFKA_CONNECT_HOSTS=http://<kafka-connect-api-host>:8083
      - HEALTH_CHECK_ENDPOINT=/health
```

2. Start kafka-connect-exporter with following command:

```bash
docker compose up -d
```

### Kubernetes

1. Create a deployment file (e.g., `deployment.yaml`) with the content below:

```yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kafka-connect-exporter
  namespace: your-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kafka-connect-exporter
  template:
    metadata:
      labels:
        app: kafka-connect-exporter
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/path: "/metrics"
        prometheus.io/port: "9113"
    spec:
      containers:
        - name: kafka-connect-exporter
          image: ecubelabs/kafka-connect-exporter:latest
          ports:
            - containerPort: 9113
          env:
            - name: PORT
              value: "9113"
            - name: METRICS_ENDPOINT
              value: "/metrics"
            - name: KAFKA_CONNECT_HOSTS
              value: "http://<kafka-connect-host1>:8083,http://<kafka-connect-host2>:8083"
            - name: HEALTH_CHECK_ENDPOINT
              value: "/health"
```

2. Then apply it with:

```bash
kubectl apply -f deployment.yaml
```

## Contributing

Refer to our [contribution guidelines](./CONTRIBUTING.md) and [Code of Conduct for contributors](./CODE_OF_CONDUCT.md).
