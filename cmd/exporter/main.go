package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Ecube-Labs/kafka-connect-exporter/internal/collector"
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/config"
	"github.com/Ecube-Labs/kafka-connect-exporter/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	prometheus.MustRegister(collector.New())
	http.Handle(config.MetricsEndpoint, promhttp.Handler())
	http.HandleFunc(config.HealthCheckEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Log("info", "Starting Kafka Connect Exporter")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil))
}
