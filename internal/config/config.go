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
	Options := Options{}
	flag.StringVar(&Options.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&Options.DataBaseURI, "d", "user=postgres password=adm dbname=postgres host=localhost port=5432 sslmode=disable", "dsn for acees to DB")
	flag.StringVar(&Options.AccrualSystemAddr, "r", "", "address of accrural system")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		Options.RunAddr = envRunAddr
	}
	if envDataBaseURI := os.Getenv("DATABASE_URI"); envDataBaseURI != "" {
		Options.DataBaseURI = envDataBaseURI
	}
	if envAccrualSystemAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRES"); envAccrualSystemAddr != "" {
		Options.AccrualSystemAddr = envAccrualSystemAddr
	}
	return &Options
}

//dsn := "user=postgres password=adm dbname=postgres host=localhost port=5432 sslmode=disable"
