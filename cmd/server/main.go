package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"backend-work-mate/internal/config"
	"backend-work-mate/internal/server"
	"backend-work-mate/internal/storage/postgres"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	dbpool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer dbpool.Close()

	if err := postgres.RunMigrations(ctx, dbpool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	r := server.NewRouter(dbpool, cfg)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("server listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}

// end
