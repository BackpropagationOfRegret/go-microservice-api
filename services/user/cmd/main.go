package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/kostayne/go-microservice/pkg/config"
	"github.com/kostayne/go-microservice/services/user/internal/handler"
	"github.com/kostayne/go-microservice/services/user/internal/repository"
	_ "github.com/lib/pq"
)

func main() {
	log.Printf("user-svc starting (APP_ENV=%s)", config.AppEnv())

	dsn := config.String("DATABASE_URL", "postgres://food:food@localhost:5432/user_db?sslmode=disable")
	port := config.String("PORT", "8081")
	jwtSecret := config.String("JWT_SECRET", "dev-secret-change-me")
	jwtTTL := config.Duration("JWT_TTL", 24*time.Hour)

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

	h := handler.New(repo, jwtSecret, jwtTTL)
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
