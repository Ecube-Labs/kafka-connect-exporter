package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ecube-Labs/kafka-connect-exporter/internal/collector"
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/config"
	exporter "github.com/Ecube-Labs/kafka-connect-exporter/internal/exporter/prometheus"
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/server"
	applicationError "github.com/Ecube-Labs/kafka-connect-exporter/pkg/application-error"
	"github.com/Ecube-Labs/kafka-connect-exporter/pkg/logger"
)

func main() {
	collector := collector.New(&http.Client{Timeout: 10 * time.Second})
	exporter := exporter.New(collector)

	mux := http.NewServeMux()
	mux.Handle(config.MetricsEndpoint, exporter.Handler())
	mux.HandleFunc(config.HealthCheckEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := server.New(config.Port, mux)
	logger.Log("info", "Starting Kafka Connect Exporter")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.Run(); err != nil {
			logger.Log("error", applicationError.UnWrap(err).Stack)
		}
	}()

	<-ctx.Done()
	logger.Log("info", "Server is shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log("error", applicationError.UnWrap(err).Stack)
	}
}
