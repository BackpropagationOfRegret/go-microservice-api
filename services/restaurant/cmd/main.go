package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/services/restaurant/internal/consumer"
	"github.com/kostayne/go-microservice/services/restaurant/internal/handler"
	"github.com/kostayne/go-microservice/services/restaurant/internal/repository"
	_ "github.com/lib/pq"
)

func main() {
	dsn := env("DATABASE_URL", "postgres://food:food@localhost:5432/restaurant_db?sslmode=disable")
	port := env("PORT", "8082")
	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")

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
	if err := repo.Seed(ctx); err != nil {
		log.Fatalf("seed: %v", err)
	}

	readyProducer := kafka.NewProducer(brokers, events.TopicOrderReady)
	defer readyProducer.Close()

	kitchen := &consumer.Kitchen{}
	orderPaidConsumer := kafka.NewConsumer(brokers, events.TopicOrderPaid, "restaurant-svc")
	defer orderPaidConsumer.Close()

	go func() {
		err := orderPaidConsumer.Run(context.Background(), func(ctx context.Context, env events.Envelope) error {
			return kitchen.HandleOrderPaid(ctx, readyProducer, env)
		})
		if err != nil {
			log.Printf("kafka consumer stopped: %v", err)
		}
	}()

	h := handler.New(repo)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      h.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("restaurant-svc listening on :%s", port)
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
