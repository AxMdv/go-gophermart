package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/AxMdv/go-gophermart/internal/config"
	"github.com/AxMdv/go-gophermart/internal/handlers"
	"github.com/AxMdv/go-gophermart/internal/router"
	"github.com/AxMdv/go-gophermart/internal/service/accrual"
	"github.com/AxMdv/go-gophermart/internal/service/gophermart"

	"github.com/AxMdv/go-gophermart/internal/storage"
	"github.com/go-chi/chi"
)

type App struct {
	cfg    *config.Config
	router *chi.Mux
}

func New() (*App, error) {
	app := &App{}
	cfg := config.ParseOptions()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repository, err := storage.NewRepository(ctx, cfg)
	if err != nil {
		return app, err
	}

	_ = accrual.NewService(cfg.AccrualSystemAddr, repository)

	gophermartService := gophermart.New(repository)

	handlers := handlers.New(gophermartService, cfg)

	router := router.New(handlers)

	app = &App{
		cfg:    cfg,
		router: router,
	}
	return app, nil
}

func (a *App) Run() error {
	log.Println("http server is running on", a.cfg.RunAddr)
	err := http.ListenAndServe(a.cfg.RunAddr, a.router)
	return err
}
