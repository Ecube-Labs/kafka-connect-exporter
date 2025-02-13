package config

import (
	"os"
	"strings"
)

func getEnvWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	Port                = getEnvWithDefault("PORT", "9113")
	MetricsEndpoint     = getEnvWithDefault("METRICS_ENDPOINT", "/metrics")
	KafkaConnectHosts   = strings.Split(getEnvWithDefault("KAFKA_CONNECT_HOSTS", "http://localhost:4444"), ",")
	HealthCheckEndpoint = getEnvWithDefault("HEALTH_CHECK_ENDPOINT", "/health")
)
