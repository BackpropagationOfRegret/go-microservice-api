package consumer

import (
	"context"
	"encoding/json"
	"fmt"
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
	http              *http.Client
}

func NewDispatcher(repo *repository.Repository, orderSvcURL string, assigned, delivered *kafka.Producer) *Dispatcher {
	return &Dispatcher{
		repo:              repo,
		orderSvcURL:       orderSvcURL,
		assignedProducer:  assigned,
		deliveredProducer: delivered,
		http:              &http.Client{Timeout: 5 * time.Second},
	}
}

func (d *Dispatcher) HandleOrderReady(ctx context.Context, env events.Envelope) error {
	ready, err := kafka.DecodePayload[events.OrderReady](env)
	if err != nil {
		return err
	}

	courier, err := d.repo.FindAvailableCourier(ctx)
	if err != nil {
		return fmt.Errorf("assign courier: %w", err)
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
	time.Sleep(5 * time.Second)

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
