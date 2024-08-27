package main

import (
	"context"
	"log/slog"
	"os"
	"url-shortener/internal/config"
	mwLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
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
	router.Use(mwLogger.New(log))
	// При панике, чтобы не падало все приложение из-за одного запроса
	router.Use(middleware.Recoverer)
	// Фишка chi. Позволяет писать такие роуты: /articles/{id} и потом
	// получать этот id в хендлере
	router.Use(middleware.URLFormat)
	// TODO: run server
}

func setUpLogger(env string) *slog.Logger {
	var logger *slog.Logger
	handlerOptions := &slog.HandlerOptions{Level: slog.LevelDebug}

	switch env {
	case envLocal:
		logger = setupPrettySlog()
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	}

	return logger
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
