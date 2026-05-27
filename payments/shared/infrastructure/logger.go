package infrastructure

import (
	"log"
	"log/slog"
	"os"
	"strings"
)

func ConfigureLogger(environment string) *slog.Logger {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)

	options := &slog.HandlerOptions{Level: slog.LevelInfo}
	var handler slog.Handler
	if strings.EqualFold(environment, "production") {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
