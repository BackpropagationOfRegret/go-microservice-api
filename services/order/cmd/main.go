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
	"github.com/kostayne/go-microservice/services/order/internal/client"
	"github.com/kostayne/go-microservice/services/order/internal/handler"
	"github.com/kostayne/go-microservice/services/order/internal/repository"
	"github.com/kostayne/go-microservice/services/order/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	dsn := env("DATABASE_URL", "postgres://food:food@localhost:5432/order_db?sslmode=disable")
	port := env("PORT", "8083")
	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")
	userURL := env("USER_SVC_URL", "http://localhost:8081")
	restaurantURL := env("RESTAURANT_SVC_URL", "http://localhost:8082")

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

	orderCreatedProducer := kafka.NewProducer(brokers, events.TopicOrderCreated)
	orderPaidProducer := kafka.NewProducer(brokers, events.TopicOrderPaid)
	statusProducer := kafka.NewProducer(brokers, events.TopicOrderStatusChanged)
	defer orderCreatedProducer.Close()
	defer orderPaidProducer.Close()
	defer statusProducer.Close()

	// Multi-topic producer wrapper
	multiProducer := &multiTopicProducer{
		created: orderCreatedProducer,
		paid:    orderPaidProducer,
		status:  statusProducer,
	}

	svc := service.New(
		repo,
		client.NewUserClient(userURL),
		client.NewRestaurantClient(restaurantURL),
		multiProducer,
	)

	startConsumers(brokers, svc)

	h := handler.New(svc)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      h.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("order-svc listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}

type multiTopicProducer struct {
	created *kafka.Producer
	paid    *kafka.Producer
	status  *kafka.Producer
}

func (m *multiTopicProducer) Publish(ctx context.Context, topic string, payload any) error {
	switch topic {
	case events.TopicOrderCreated:
		return m.created.Publish(ctx, topic, payload)
	case events.TopicOrderPaid:
		return m.paid.Publish(ctx, topic, payload)
	case events.TopicOrderStatusChanged:
		return m.status.Publish(ctx, topic, payload)
	default:
		return m.status.Publish(ctx, topic, payload)
	}
}

func startConsumers(brokers []string, svc *service.Service) {
	paymentProcessed := kafka.NewConsumer(brokers, events.TopicPaymentProcessed, "order-svc-payment")
	paymentFailed := kafka.NewConsumer(brokers, events.TopicPaymentFailed, "order-svc-payment-failed")
	courierAssigned := kafka.NewConsumer(brokers, events.TopicCourierAssigned, "order-svc-courier")
	orderDelivered := kafka.NewConsumer(brokers, events.TopicOrderDelivered, "order-svc-delivered")

	go runConsumer(paymentProcessed, func(ctx context.Context, env events.Envelope) error {
		p, err := kafka.DecodePayload[events.PaymentProcessed](env)
		if err != nil {
			return err
		}
		return svc.HandlePaymentProcessed(ctx, p)
	})

	go runConsumer(paymentFailed, func(ctx context.Context, env events.Envelope) error {
		p, err := kafka.DecodePayload[events.PaymentFailed](env)
		if err != nil {
			return err
		}
		return svc.HandlePaymentFailed(ctx, p)
	})

	go runConsumer(courierAssigned, func(ctx context.Context, env events.Envelope) error {
		p, err := kafka.DecodePayload[events.CourierAssigned](env)
		if err != nil {
			return err
		}
		return svc.HandleCourierAssigned(ctx, p)
	})

	go runConsumer(orderDelivered, func(ctx context.Context, env events.Envelope) error {
		p, err := kafka.DecodePayload[events.OrderDelivered](env)
		if err != nil {
			return err
		}
		return svc.HandleOrderDelivered(ctx, p)
	})
}

func runConsumer(c *kafka.Consumer, handler kafka.Handler) {
	if err := c.Run(context.Background(), handler); err != nil {
		log.Printf("consumer stopped: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
