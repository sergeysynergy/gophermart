package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/hardtest/internal/api/handlers"
	"github.com/sergeysynergy/hardtest/internal/api/server"
	"github.com/sergeysynergy/hardtest/internal/db"
	"log"

	"github.com/sergeysynergy/hardtest/internal/gophermart"
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

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("[DEBUG] Receive config: %#v\n", cfg)

	//st := basicstorage.New()
	st, err := db.New(cfg.DatabaseURI)
	if err != nil {
		log.Fatalln("[FATAL] Postgres initialization failed - ", err)
	}

	gm := gophermart.New(st)

	// подключим обработчики запросов
	h := handlers.New(gm)

	// проиницилизируем сервер с использованием ранее объявленных обработчиков и файлового хранилища
	s := server.New(h.GetRouter(),
		server.WithAddress(cfg.Addr),
	)
	// запустим сервер
	go s.Serve()

	queue := gophermart.NewQueue(st, cfg.AccrualSystemAddress)
	queue.Start()
}
