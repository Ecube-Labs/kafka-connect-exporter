package main

import (
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/config"
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/server"
	"github.com/Ecube-Labs/kafka-connect-exporter/pkg/logger"
)

func main() {
	server := server.New(config.Port)
	logger.Log("info", "Starting Kafka Connect Exporter")
	server.Run()
}
