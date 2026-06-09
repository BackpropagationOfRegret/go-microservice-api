package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kostayne/go-microservice/services/user/internal/handler"
	"github.com/kostayne/go-microservice/services/user/internal/repository"
	_ "github.com/lib/pq"
	"database/sql"
)

func main() {
	dsn := env("DATABASE_URL", "postgres://food:food@localhost:5432/user_db?sslmode=disable")
	port := env("PORT", "8081")
	jwtSecret := env("JWT_SECRET", "dev-secret-change-me")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := repository.New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := repo.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	h := handler.New(repo, jwtSecret)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      h.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("user-svc listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
