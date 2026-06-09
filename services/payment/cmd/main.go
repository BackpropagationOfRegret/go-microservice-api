package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/services/payment/internal/consumer"
	"github.com/kostayne/go-microservice/services/payment/internal/handler"
	"github.com/kostayne/go-microservice/services/payment/internal/repository"
	_ "github.com/lib/pq"
)

func main() {
	dsn := env("DATABASE_URL", "postgres://food:food@localhost:5432/payment_db?sslmode=disable")
	port := env("PORT", "8084")
	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")
	failRate, _ := strconv.ParseFloat(env("PAYMENT_FAIL_RATE", "0.05"), 64)

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

	processedProducer := kafka.NewProducer(brokers, events.TopicPaymentProcessed)
	failedProducer := kafka.NewProducer(brokers, events.TopicPaymentFailed)
	defer processedProducer.Close()
	defer failedProducer.Close()

	processor := consumer.NewProcessor(repo, processedProducer, failedProducer, failRate)
	orderCreatedConsumer := kafka.NewConsumer(brokers, events.TopicOrderCreated, "payment-svc")
	defer orderCreatedConsumer.Close()

	go func() {
		if err := orderCreatedConsumer.Run(context.Background(), processor.HandleOrderCreated); err != nil {
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

	log.Printf("payment-svc listening on :%s", port)
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
