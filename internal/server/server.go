package server

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ecube-Labs/kafka-connect-exporter/internal/collector"
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/config"
	exporter "github.com/Ecube-Labs/kafka-connect-exporter/internal/exporter/prometheus"
	"github.com/Ecube-Labs/kafka-connect-exporter/pkg/logger"
)

type Server struct {
	server *http.Server
}

func New(port string) *Server {
	return &Server{
		server: &http.Server{Addr: ":" + port},
	}
}

func (s *Server) Run() {
	collector := collector.New(&http.Client{Timeout: 10 * time.Second})
	exporter := exporter.New(collector)

	http.Handle(config.MetricsEndpoint, exporter.Handler())
	http.HandleFunc(config.HealthCheckEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log("error", fmt.Sprintf("HTTP Server Error: %s", err))
		}
	}()

	<-ctx.Done()
	logger.Log("info", "Server is shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		logger.Log("error", fmt.Sprintf("Server Shutdown Error: %s", err))
	} else {
		logger.Log("info", "Server has been safely shutdown")
	}
}
