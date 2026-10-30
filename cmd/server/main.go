package main

import (
	"log/slog"
	"os"

	"github.com/arthur-stone/order-service/internal/app"
	"github.com/arthur-stone/order-service/internal/config"
	"github.com/arthur-stone/order-service/internal/handler"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	cfg := config.Load()

	var logHandler slog.Handler
	if cfg.Env == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(logHandler)

	buildInfo := handler.BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	}

	application := app.New(cfg, logger, buildInfo)
	if err := application.Run(); err != nil {
		logger.Error("application terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
