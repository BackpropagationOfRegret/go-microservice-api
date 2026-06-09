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
	"github.com/kostayne/go-microservice/pkg/telemetry"
	"github.com/kostayne/go-microservice/services/delivery/internal/consumer"
	"github.com/kostayne/go-microservice/services/delivery/internal/handler"
	"github.com/kostayne/go-microservice/services/delivery/internal/repository"
	_ "github.com/lib/pq"
)

const serviceName = "delivery-svc"

func main() {
	log.Printf("delivery-svc starting (APP_ENV=%s)", config.AppEnv())

	shutdown, err := telemetry.Init(context.Background(), serviceName)
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	dsn := config.String("DATABASE_URL", "postgres://food:food@localhost:5432/delivery_db?sslmode=disable")
	port := config.String("PORT", "8085")
	brokers := config.KafkaBrokers()
	orderURL := config.String("ORDER_SVC_URL", "http://localhost:8083")
	deliveryDuration := config.Duration("DELIVERY_DURATION", 5*time.Second)
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

	assignedProducer := kafka.NewProducer(brokers, events.TopicCourierAssigned)
	deliveredProducer := kafka.NewProducer(brokers, events.TopicOrderDelivered)
	failedProducer := kafka.NewProducer(brokers, events.TopicDeliveryFailed)
	defer assignedProducer.Close()
	defer deliveredProducer.Close()
	defer failedProducer.Close()

	dispatcher := consumer.NewDispatcher(repo, orderURL, assignedProducer, deliveredProducer, failedProducer, deliveryDuration)
	orderReadyConsumer := kafka.NewConsumer(brokers, events.TopicOrderReady, "delivery-svc")
	defer orderReadyConsumer.Close()

	go func() {
		if err := orderReadyConsumer.Run(context.Background(), dispatcher.HandleOrderReady); err != nil {
			log.Printf("kafka consumer stopped: %v", err)
		}
	}()

	h := handler.New(repo)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      telemetry.WrapHTTP(serviceName, h.Routes()),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("delivery-svc listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
