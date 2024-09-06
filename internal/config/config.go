package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr           string
	DataBaseURI       string
	AccrualSystemAddr string
}

func ParseOptions() *Config {
	cfg := Config{}
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.DataBaseURI, "d", "user=postgres password=adm dbname=postgres host=localhost port=5432 sslmode=disable", "dsn for acees to DB")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", "http://localhost:8081", "address of accrural system")

	flag.Parse()

	if envRunAddr, found := os.LookupEnv("RUN_ADDRESS"); envRunAddr != "" && found {
		cfg.RunAddr = envRunAddr
	}
	if envDataBaseURI := os.Getenv("DATABASE_URI"); envDataBaseURI != "" {
		cfg.DataBaseURI = envDataBaseURI
	}
	if envAccrualSystemAddr, found := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); (envAccrualSystemAddr != "") && found {
		cfg.AccrualSystemAddr = envAccrualSystemAddr
	}
	return &cfg
}

//dsn := "user=postgres password=adm dbname=postgres host=localhost port=5432 sslmode=disable"
