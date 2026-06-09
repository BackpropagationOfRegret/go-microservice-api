package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kostayne/go-microservice/pkg/config"
	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/pkg/telemetry"
	"github.com/kostayne/go-microservice/services/order/internal/client"
	"github.com/kostayne/go-microservice/services/order/internal/handler"
	"github.com/kostayne/go-microservice/services/order/internal/repository"
	"github.com/kostayne/go-microservice/services/order/internal/service"
	_ "github.com/lib/pq"
)

const serviceName = "order-svc"

func main() {
	log.Printf("order-svc starting (APP_ENV=%s)", config.AppEnv())

	shutdown, err := telemetry.Init(context.Background(), serviceName)
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	dsn := config.String("DATABASE_URL", "postgres://food:food@localhost:5432/order_db?sslmode=disable")
	port := config.String("PORT", "8083")
	brokers := config.KafkaBrokers()
	userURL := config.String("USER_SVC_URL", "http://localhost:8081")
	restaurantURL := config.String("RESTAURANT_SVC_URL", "http://localhost:8082")

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

	producer := newTopicProducer(brokers,
		events.TopicOrderCreated,
		events.TopicOrderPaid,
		events.TopicOrderStatusChanged,
		events.TopicPaymentRefundRequested,
		events.TopicOrderCancelled,
	)
	defer producer.Close()

	svc := service.New(
		repo,
		client.NewUserClient(userURL),
		client.NewRestaurantClient(restaurantURL),
		producer,
	)

	startConsumers(brokers, svc)

	h := handler.New(svc)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      telemetry.WrapHTTP(serviceName, h.Routes()),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("order-svc listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}

type topicProducer struct {
	writers map[string]*kafka.Producer
}

func newTopicProducer(brokers []string, topics ...string) *topicProducer {
	tp := &topicProducer{writers: make(map[string]*kafka.Producer, len(topics))}
	for _, topic := range topics {
		tp.writers[topic] = kafka.NewProducer(brokers, topic)
	}
	return tp
}

func (t *topicProducer) Publish(ctx context.Context, topic string, payload any) error {
	w, ok := t.writers[topic]
	if !ok {
		return fmt.Errorf("unknown topic: %s", topic)
	}
	return w.Publish(ctx, topic, payload)
}

func (t *topicProducer) Close() {
	for _, w := range t.writers {
		_ = w.Close()
	}
}

func startConsumers(brokers []string, svc *service.Service) {
	consumers := []struct {
		topic   string
		group   string
		handler kafka.Handler
	}{
		{events.TopicPaymentProcessed, "order-svc-payment", func(ctx context.Context, env events.Envelope) error {
			p, err := kafka.DecodePayload[events.PaymentProcessed](env)
			if err != nil {
				return err
			}
			return svc.HandlePaymentProcessed(ctx, p)
		}},
		{events.TopicPaymentFailed, "order-svc-payment-failed", func(ctx context.Context, env events.Envelope) error {
			p, err := kafka.DecodePayload[events.PaymentFailed](env)
			if err != nil {
				return err
			}
			return svc.HandlePaymentFailed(ctx, p)
		}},
		{events.TopicOrderPreparationFailed, "order-svc-prep-failed", func(ctx context.Context, env events.Envelope) error {
			p, err := kafka.DecodePayload[events.OrderPreparationFailed](env)
			if err != nil {
				return err
			}
			return svc.HandlePreparationFailed(ctx, p)
		}},
		{events.TopicDeliveryFailed, "order-svc-delivery-failed", func(ctx context.Context, env events.Envelope) error {
			p, err := kafka.DecodePayload[events.DeliveryFailed](env)
			if err != nil {
				return err
			}
			return svc.HandleDeliveryFailed(ctx, p)
		}},
		{events.TopicPaymentRefunded, "order-svc-refunded", func(ctx context.Context, env events.Envelope) error {
			p, err := kafka.DecodePayload[events.PaymentRefunded](env)
			if err != nil {
				return err
			}
			return svc.HandlePaymentRefunded(ctx, p)
		}},
		{events.TopicCourierAssigned, "order-svc-courier", func(ctx context.Context, env events.Envelope) error {
			p, err := kafka.DecodePayload[events.CourierAssigned](env)
			if err != nil {
				return err
			}
			return svc.HandleCourierAssigned(ctx, p)
		}},
		{events.TopicOrderDelivered, "order-svc-delivered", func(ctx context.Context, env events.Envelope) error {
			p, err := kafka.DecodePayload[events.OrderDelivered](env)
			if err != nil {
				return err
			}
			return svc.HandleOrderDelivered(ctx, p)
		}},
	}

	for _, c := range consumers {
		topic := c.topic
		consumer := kafka.NewConsumer(brokers, topic, c.group)
		go func(cons *kafka.Consumer, h kafka.Handler, topicName string) {
			defer cons.Close()
			if err := cons.Run(context.Background(), h); err != nil {
				log.Printf("consumer %s stopped: %v", topicName, err)
			}
		}(consumer, c.handler, topic)
	}
}
