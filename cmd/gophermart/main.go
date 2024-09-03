package main

import (
	"log"

	"github.com/AxMdv/go-gophermart/internal/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		log.Panic(err)
	}

	err = app.Run()
	if err != nil {
		log.Panic(err)
	}
}
