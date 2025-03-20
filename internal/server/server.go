package server

import (
	"context"
	"fmt"
	"net/http"

	applicationError "github.com/Ecube-Labs/kafka-connect-exporter/pkg/application-error"
	"github.com/Ecube-Labs/kafka-connect-exporter/pkg/logger"
)

type Server struct {
	httpServer *http.Server
}

func New(port string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
	}
}

func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return applicationError.New(http.StatusInternalServerError, fmt.Sprintf("Failed to start server: %s", err.Error()), "")
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return applicationError.New(http.StatusInternalServerError, fmt.Sprintf("Failed to shutdown server: %s", err.Error()), "")
	}
	logger.Log("info", "Server has been safely shutdown")
	return nil
}
