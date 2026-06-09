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
	"github.com/kostayne/go-microservice/services/payment/internal/consumer"
	"github.com/kostayne/go-microservice/services/payment/internal/handler"
	"github.com/kostayne/go-microservice/services/payment/internal/repository"
	_ "github.com/lib/pq"
)

const serviceName = "payment-svc"

func main() {
	log.Printf("payment-svc starting (APP_ENV=%s)", config.AppEnv())

	shutdown, err := telemetry.Init(context.Background(), serviceName)
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	dsn := config.String("DATABASE_URL", "postgres://food:food@localhost:5432/payment_db?sslmode=disable")
	port := config.String("PORT", "8084")
	brokers := config.KafkaBrokers()
	failRate := config.Float("PAYMENT_FAIL_RATE", 0.0)
	processLatency := config.Duration("PAYMENT_PROCESS_LATENCY", 500*time.Millisecond)
	refundLatency := config.Duration("PAYMENT_REFUND_LATENCY", 300*time.Millisecond)
	paymentMethod := config.String("PAYMENT_METHOD", "card")

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
	refundedProducer := kafka.NewProducer(brokers, events.TopicPaymentRefunded)
	defer processedProducer.Close()
	defer failedProducer.Close()
	defer refundedProducer.Close()

	processor := consumer.NewProcessor(
		repo, processedProducer, failedProducer, refundedProducer,
		failRate, processLatency, refundLatency, paymentMethod,
	)

	orderCreatedConsumer := kafka.NewConsumer(brokers, events.TopicOrderCreated, "payment-svc")
	refundConsumer := kafka.NewConsumer(brokers, events.TopicPaymentRefundRequested, "payment-svc-refund")
	defer orderCreatedConsumer.Close()
	defer refundConsumer.Close()

	go func() {
		if err := orderCreatedConsumer.Run(context.Background(), processor.HandleOrderCreated); err != nil {
			log.Printf("kafka consumer stopped: %v", err)
		}
	}()
	go func() {
		if err := refundConsumer.Run(context.Background(), processor.HandleRefundRequested); err != nil {
			log.Printf("refund consumer stopped: %v", err)
		}
	}()

	h := handler.New(repo)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      telemetry.WrapHTTP(serviceName, h.Routes()),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("payment-svc listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
