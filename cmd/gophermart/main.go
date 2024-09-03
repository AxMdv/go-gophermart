package main

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
)

func main() {
	cfg := config.ParseOptions()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repository, err := storage.NewRepository(ctx, cfg)
	if err != nil {
		log.Panic("Failed to init repository ", err)
	}

	queue, err := reward.NewRewardCollectionProcess(cfg.AccrualSystemAddr, repository)
	if err != nil {
		log.Panic("Failed to init repository ", err)
	}

	accrualService := accrual.New(repository, queue)

	handlers := handlers.New(accrualService, cfg)

	router := router.New(handlers)

	log.Println("http server is running on", cfg.RunAddr)
	err = http.ListenAndServe(cfg.RunAddr, router)

	if err != nil {
		log.Panic(err)
	}

}
