package consumer

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/services/restaurant/internal/repository"
)

type Kitchen struct {
	repo           *repository.Repository
	cookDuration   time.Duration
	failRate       float64
	readyProducer  *kafka.Producer
	failedProducer *kafka.Producer
}

func NewKitchen(
	repo *repository.Repository,
	cookDuration time.Duration,
	failRate float64,
	readyProducer, failedProducer *kafka.Producer,
) *Kitchen {
	return &Kitchen{
		repo:           repo,
		cookDuration:   cookDuration,
		failRate:       failRate,
		readyProducer:  readyProducer,
		failedProducer: failedProducer,
	}
}

func (k *Kitchen) HandleOrderPaid(ctx context.Context, env events.Envelope) error {
	paid, err := kafka.DecodePayload[events.OrderPaid](env)
	if err != nil {
		return err
	}

	rest, err := k.repo.GetRestaurant(ctx, paid.RestaurantID)
	if err != nil {
		return k.publishPreparationFailed(paid, "restaurant not found")
	}
	if !rest.IsOpen {
		return k.publishPreparationFailed(paid, "restaurant closed")
	}
	if k.failRate > 0 && rand.Float64() < k.failRate {
		return k.publishPreparationFailed(paid, "kitchen overloaded")
	}

	log.Printf("[restaurant-svc] order %s received in kitchen, preparing...", paid.OrderID)

	go func() {
		time.Sleep(k.cookDuration)
		ready := events.OrderReady{
			OrderID:      paid.OrderID,
			RestaurantID: paid.RestaurantID,
		}
		if err := k.readyProducer.Publish(context.Background(), events.TopicOrderReady, ready); err != nil {
			log.Printf("[restaurant-svc] publish order ready: %v", err)
			return
		}
		log.Printf("[restaurant-svc] order %s is ready", paid.OrderID)
	}()

	return nil
}

func (k *Kitchen) publishPreparationFailed(paid events.OrderPaid, reason string) error {
	log.Printf("[restaurant-svc] preparation failed for order %s: %s", paid.OrderID, reason)
	return k.failedProducer.Publish(context.Background(), events.TopicOrderPreparationFailed, events.OrderPreparationFailed{
		OrderID:      paid.OrderID,
		RestaurantID: paid.RestaurantID,
		UserID:       paid.UserID,
		Reason:       reason,
	})
}
