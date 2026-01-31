package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"warehouse/internal/config"
	httpx "warehouse/internal/http"
	"warehouse/internal/repo"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger

	db     *repo.DB
	server *http.Server
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	db, err := repo.NewDB(cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	router := httpx.NewRouter(httpx.Deps{
		Logger: logger,
		DB:     db,
		Cfg:    cfg,
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		db:     db,
		server: srv,
	}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.db != nil {
		a.db.Close()
	}
	return a.server.Shutdown(ctx)
}
