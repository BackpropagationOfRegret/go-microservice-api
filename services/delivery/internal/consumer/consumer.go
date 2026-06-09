package consumer

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/services/delivery/internal/model"
	"github.com/kostayne/go-microservice/services/delivery/internal/repository"
)

type Dispatcher struct {
	repo              *repository.Repository
	orderSvcURL       string
	assignedProducer  *kafka.Producer
	deliveredProducer *kafka.Producer
	failedProducer    *kafka.Producer
	deliveryDuration  time.Duration
	http              *http.Client
}

func NewDispatcher(
	repo *repository.Repository,
	orderSvcURL string,
	assigned, delivered, failed *kafka.Producer,
	deliveryDuration time.Duration,
) *Dispatcher {
	return &Dispatcher{
		repo:              repo,
		orderSvcURL:       orderSvcURL,
		assignedProducer:  assigned,
		deliveredProducer: delivered,
		failedProducer:    failed,
		deliveryDuration:  deliveryDuration,
		http:              &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *Dispatcher) HandleOrderReady(ctx context.Context, env events.Envelope) error {
	ready, err := kafka.DecodePayload[events.OrderReady](env)
	if err != nil {
		return err
	}

	courier, err := d.repo.FindAvailableCourier(ctx)
	if err != nil {
		userID, _ := d.fetchOrderUserID(ready.OrderID)
		log.Printf("[delivery-svc] no courier for order %s", ready.OrderID)
		return d.failedProducer.Publish(ctx, events.TopicDeliveryFailed, events.DeliveryFailed{
			OrderID: ready.OrderID,
			UserID:  userID,
			Reason:  "no available courier",
		})
	}

	if err := d.repo.SetCourierStatus(ctx, courier.ID, model.CourierBusy); err != nil {
		return err
	}

	userID, err := d.fetchOrderUserID(ready.OrderID)
	if err != nil {
		userID = ""
	}

	log.Printf("[delivery-svc] courier %s assigned to order %s", courier.Name, ready.OrderID)

	if err := d.assignedProducer.Publish(ctx, events.TopicCourierAssigned, events.CourierAssigned{
		OrderID:   ready.OrderID,
		CourierID: courier.ID,
		UserID:    userID,
	}); err != nil {
		_ = d.repo.SetCourierStatus(ctx, courier.ID, model.CourierAvailable)
		return err
	}

	go d.simulateDelivery(courier.ID, ready.OrderID, userID)
	return nil
}

func (d *Dispatcher) fetchOrderUserID(orderID string) (string, error) {
	resp, err := d.http.Get(d.orderSvcURL + "/orders/" + orderID)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var order struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(body, &order); err != nil {
		return "", err
	}
	return order.UserID, nil
}

func (d *Dispatcher) simulateDelivery(courierID, orderID, userID string) {
	time.Sleep(d.deliveryDuration)

	ctx := context.Background()
	if err := d.repo.SetCourierStatus(ctx, courierID, model.CourierAvailable); err != nil {
		log.Printf("[delivery-svc] release courier: %v", err)
	}

	log.Printf("[delivery-svc] order %s delivered", orderID)
	if err := d.deliveredProducer.Publish(ctx, events.TopicOrderDelivered, events.OrderDelivered{
		OrderID: orderID,
		UserID:  userID,
	}); err != nil {
		log.Printf("[delivery-svc] publish delivered: %v", err)
	}
}
