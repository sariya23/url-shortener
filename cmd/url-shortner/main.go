package main

import (
	"log/slog"
	"os"
	"url-shortener/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	config := config.MustLoad()

	log := setUpLogger(config.Env)

	log.Info("starting url-shortener", slog.String("env", config.Env))
	log.Debug("debuf messages are enabled")

	// TODO: init logger: slog

	// TODO: init storage: sqlite

	// TODO: init router: chi

	// TODO: run server
}

func setUpLogger(env string) *slog.Logger {
	var logger *slog.Logger
	handlerOptions := &slog.HandlerOptions{Level: slog.LevelDebug}

	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	}

	return logger
}
