package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/kostayne/go-microservice/pkg/config"
	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/services/restaurant/internal/consumer"
	"github.com/kostayne/go-microservice/services/restaurant/internal/handler"
	"github.com/kostayne/go-microservice/services/restaurant/internal/repository"
	_ "github.com/lib/pq"
)

func main() {
	log.Printf("restaurant-svc starting (APP_ENV=%s)", config.AppEnv())

	dsn := config.String("DATABASE_URL", "postgres://food:food@localhost:5432/restaurant_db?sslmode=disable")
	port := config.String("PORT", "8082")
	brokers := config.KafkaBrokers()
	cookDuration := config.Duration("KITCHEN_COOK_DURATION", 3*time.Second)
	prepFailRate := config.Float("KITCHEN_FAIL_RATE", 0.0)
	seedData := config.Bool("SEED_DATA", config.IsDev())

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
	if seedData {
		if err := repo.Seed(ctx); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	readyProducer := kafka.NewProducer(brokers, events.TopicOrderReady)
	failedProducer := kafka.NewProducer(brokers, events.TopicOrderPreparationFailed)
	defer readyProducer.Close()
	defer failedProducer.Close()

	kitchen := consumer.NewKitchen(repo, cookDuration, prepFailRate, readyProducer, failedProducer)
	orderPaidConsumer := kafka.NewConsumer(brokers, events.TopicOrderPaid, "restaurant-svc")
	defer orderPaidConsumer.Close()

	go func() {
		if err := orderPaidConsumer.Run(context.Background(), kitchen.HandleOrderPaid); err != nil {
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
