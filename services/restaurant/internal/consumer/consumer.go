package consumer

import (
	"context"
	"log"
	"time"

	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
)

type Kitchen struct{}

func (k *Kitchen) HandleOrderPaid(ctx context.Context, producer *kafka.Producer, env events.Envelope) error {
	paid, err := kafka.DecodePayload[events.OrderPaid](env)
	if err != nil {
		return err
	}

	log.Printf("[restaurant-svc] order %s received in kitchen, preparing...", paid.OrderID)

	// Simulate cooking time
	go func() {
		time.Sleep(3 * time.Second)
		ready := events.OrderReady{
			OrderID:      paid.OrderID,
			RestaurantID: paid.RestaurantID,
		}
		if err := producer.Publish(context.Background(), events.TopicOrderReady, ready); err != nil {
			log.Printf("[restaurant-svc] publish order ready: %v", err)
			return
		}
		log.Printf("[restaurant-svc] order %s is ready", paid.OrderID)
	}()

	return nil
}
