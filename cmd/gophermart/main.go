package main

import (
	"flag"
	"log"
)

type config struct {
	Addr                 string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func main() {
	cfg := new(config)
	flag.StringVar(&cfg.Addr, "a", ":8080", "Service run address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Postgres URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual system address")
	flag.Parse()

	log.Printf("[DEBUG] Receive config: %#v\n", cfg)
}
