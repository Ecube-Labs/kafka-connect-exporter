package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func new() zerolog.Logger {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return l
}

var logger = new()

func Log(level string, message string) {
	if strings.ToLower(level) == "error" {
		logger.Error().Msg(message)
	}
	if strings.ToLower(level) == "info" {
		logger.Info().Msg(message)
	}
}
