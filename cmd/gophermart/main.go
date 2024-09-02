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
	log.Println("options parsed", cfg, cfg.AccrualSystemAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repository, err := storage.NewRepository(ctx, cfg)
	if err != nil {
		log.Panic("Failed to init repository ", err)
	}
	log.Println("repo created", repository)
	queue, err := reward.NewRewardCollectionProcess(cfg.AccrualSystemAddr, repository)
	if err != nil {
		log.Panic("Failed to init repository ", err)
	}
	log.Println("Reward collection process is running..")
	accrualService := accrual.New(repository, queue)
	log.Println("Accrual service is created")
	handlers := handlers.New(accrualService, cfg)
	log.Println("Handlers created")
	router := router.New(handlers)
	log.Println("Router created")
	log.Println("http server is running on", cfg.RunAddr)
	err = http.ListenAndServe(cfg.RunAddr, router)

	if err != nil {
		log.Panic(err)
	}

}
