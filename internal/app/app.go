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
	"github.com/AxMdv/go-gophermart/internal/service/reward"
	"github.com/AxMdv/go-gophermart/internal/storage"
	"github.com/go-chi/chi"
)

type App struct {
	cfg            *config.Options
	repository     *storage.DBRepository
	queue          *reward.Queue
	accrualService *accrual.AccrualService
	handlers       *handlers.Handlers
	router         *chi.Mux
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

	queue, err := reward.NewRewardCollectionProcess(cfg.AccrualSystemAddr, repository)
	if err != nil {
		return app, err

	}

	accrualService := accrual.New(repository, queue)

	handlers := handlers.New(accrualService, cfg)

	router := router.New(handlers)

	app = &App{
		cfg:            cfg,
		repository:     repository,
		queue:          queue,
		accrualService: accrualService,
		handlers:       handlers,
		router:         router,
	}
	return app, nil
}

func (a *App) Run() error {
	log.Println("http server is running on", a.cfg.RunAddr)
	err := http.ListenAndServe(a.cfg.RunAddr, a.router)
	return err
}
