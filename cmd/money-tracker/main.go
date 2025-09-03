package main

import (
	"log/slog"
	"net/http"
	"os"

	"money-tracker/internal/config"
	"money-tracker/internal/http-server/handlers/auth"
	"money-tracker/internal/http-server/handlers/transactions/remove"
	"money-tracker/internal/http-server/handlers/transactions/save"
	mwAuth "money-tracker/internal/http-server/middleware/authorization"
	mwLogger "money-tracker/internal/http-server/middleware/logger"
	"money-tracker/internal/lib/logger/handlers/slogpretty"
	"money-tracker/internal/lib/logger/sl"
	"money-tracker/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init log
	log := setupLogger(cfg.Env)
	log.Info("starting app", slog.String("env", cfg.Env))
	log.Debug("debug message are enabled")

	// init storage
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
	}

	// init router
	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)

	// Public routes
	router.Post("/register", auth.Register(log, storage, cfg.JWTSecret))
	router.Post("/login", auth.Login(log, storage, cfg.JWTSecret))

	// Protected routes
	router.Route("/transactions", func(r chi.Router) {
		r.Use(mwAuth.Auth(log, cfg.JWTSecret))

		r.Post("/", save.New(log, storage))
		r.Delete("/{id}", remove.Remove(log, storage))
	})

	// run server
	log.Info("starting server", slog.String("address", cfg.Adress))

	srv := &http.Server{
		Addr:         cfg.Adress,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
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
