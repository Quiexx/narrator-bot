package main

import (
	"log"

	"github.com/Quiexx/narrator-bot/internal/app"
	"github.com/Quiexx/narrator-bot/internal/config"
	"github.com/caarlos0/env/v10"
)

func main() {
	cfg := &config.Config{}

	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to retrieve env variables, %v", err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("failed to start app, %v", err)
	}
}
