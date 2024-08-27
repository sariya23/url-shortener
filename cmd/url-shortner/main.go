package main

import (
	"context"
	"log/slog"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/lib/logger/xslog"
	"url-shortener/internal/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	defer cancel(*storage)

	if err != nil {
		log.Error("failed to init storage", xslog.Err(err))
		os.Exit(1)
	}
	log.Info("Storage init success. Create table and index")

	router := chi.NewRouter()

	// Добавляет request id к каждому запросу
	router.Use(middleware.RequestID)
	// Добавляет ip пользователя
	router.Use(middleware.RealIP)
	// Логирует входящие запросы
	router.Use(middleware.Logger)

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
