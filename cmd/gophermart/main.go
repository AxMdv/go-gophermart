package main

import (
	"log"

	"github.com/AxMdv/go-gophermart/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Panic(err)
	}

	err = a.Run()
	if err != nil {
		log.Panic(err)
	}
}
