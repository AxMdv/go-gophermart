package config

import (
	"flag"
	"os"
)

type Options struct {
	RunAddr           string
	DataBaseURI       string
	AccrualSystemAddr string
}

func ParseOptions() *Options {
	options := Options{}
	flag.StringVar(&options.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&options.DataBaseURI, "d", "user=postgres password=adm dbname=postgres host=localhost port=5432 sslmode=disable", "dsn for acees to DB")
	flag.StringVar(&options.AccrualSystemAddr, "r", "http://localhost:8081", "address of accrural system")

	flag.Parse()

	if envRunAddr, found := os.LookupEnv("RUN_ADDRESS"); envRunAddr != "" && found {
		options.RunAddr = envRunAddr
	}
	if envDataBaseURI := os.Getenv("DATABASE_URI"); envDataBaseURI != "" {
		options.DataBaseURI = envDataBaseURI
	}
	if envAccrualSystemAddr, found := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); (envAccrualSystemAddr != "") && found {
		options.AccrualSystemAddr = envAccrualSystemAddr
	}
	return &options
}

//dsn := "user=postgres password=adm dbname=postgres host=localhost port=5432 sslmode=disable"
