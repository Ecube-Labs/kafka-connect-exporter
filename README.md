# kafka-connect-exporter

An Exporter that provider [Kafka Connect](https://docs.confluent.io/platform/current/connect/index.html) Metrics

## Overview

kafka-connect-exporter collects connector and task status information from Kafka Connect clusters and exposes them as Prometheus metrics. This allows you to monitor the health and performance of your Kafka Connect clusters and quickly react to any issues.

## Features

### Prometheus Metrics Exposure

Exposes metrics based on the collected data:

- `kafka_connect_total`: Total number of connectors.
- `kafka_connect_running`: Number of tasks in the RUNNING state.
- `kafka_connect_failed`: Number of tasks in the FAILED state.
- `kafka_connect_paused`: Number of tasks in the PAUSED state.
- `kafka_connect_unassigned`: Number of tasks with an undefined state (UNKNOWN).
- `kafka_connect_task_total`: Total number of tasks.
- `Multi-Host Support`: Configure multiple Kafka Connect instances for parallel metric collection.
- `Robust Error Handling`: Provides proper logging and error reporting for HTTP errors and JSON parsing issues.
