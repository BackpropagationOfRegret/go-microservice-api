package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/kostayne/go-microservice/pkg/config"
	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/services/notification/internal/handler"
	"github.com/kostayne/go-microservice/services/notification/internal/notifier"
)

func main() {
	log.Printf("notification-svc starting (APP_ENV=%s)", config.AppEnv())

	port := config.String("PORT", "8086")
	brokers := config.KafkaBrokers()

	svc := notifier.New()
	topics := []string{
		events.TopicOrderCreated,
		events.TopicPaymentProcessed,
		events.TopicPaymentFailed,
		events.TopicPaymentRefundRequested,
		events.TopicPaymentRefunded,
		events.TopicOrderPreparationFailed,
		events.TopicDeliveryFailed,
		events.TopicOrderCancelled,
		events.TopicOrderReady,
		events.TopicCourierAssigned,
		events.TopicOrderDelivered,
	}

	for _, topic := range topics {
		topic := topic
		c := kafka.NewConsumer(brokers, topic, "notification-svc-"+topic)
		go func() {
			defer c.Close()
			if err := c.Run(context.Background(), svc.HandleEvent); err != nil {
				log.Printf("consumer %s stopped: %v", topic, err)
			}
		}()
	}

	h := handler.New(svc)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      h.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("notification-svc listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
