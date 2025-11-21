package main

import (
	"context"
	"log"

	"github.com/tronget/pr-management-service/internal/config"
	"github.com/tronget/pr-management-service/internal/server"
	"github.com/tronget/pr-management-service/pkg/db"
)

func main() {
	cfg := config.MustLoad()

	db := db.New(cfg)
	defer db.Close()
	dbPool, err := db.Connect(context.Background())
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	srv := server.New(cfg, dbPool)
	log.Println("Starting server on", cfg.HTTP.Address)
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
