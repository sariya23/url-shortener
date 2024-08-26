package main

import (
	"context"
	"log/slog"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/lib/logger/xslog"
	"url-shortener/internal/storage/postgres"

	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load("config/.env")
	if err != nil {
		panic(err)
	}
	config := config.MustLoad()

	log := setUpLogger(config.Env)

	log.Info("starting url-shortener", slog.String("env", config.Env))
	log.Debug("debug messages are enabled")

	storage, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(storage.Connection)

	if err != nil {
		log.Error("failed to init storage", xslog.Err(err))
		os.Exit(1)
	}
	_ = storage

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
