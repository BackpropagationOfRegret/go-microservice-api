package service

import (
	"context"
	"fmt"

	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/services/order/internal/client"
	"github.com/kostayne/go-microservice/services/order/internal/model"
	"github.com/kostayne/go-microservice/services/order/internal/repository"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

type Service struct {
	repo       *repository.Repository
	users      *client.UserClient
	restaurant *client.RestaurantClient
	producer   EventPublisher
}

func New(repo *repository.Repository, users *client.UserClient, restaurant *client.RestaurantClient, producer EventPublisher) *Service {
	return &Service{repo: repo, users: users, restaurant: restaurant, producer: producer}
}

type CreateOrderItem struct {
	MenuItemID string `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
}

type CreateOrderRequest struct {
	UserID       string            `json:"user_id"`
	RestaurantID string            `json:"restaurant_id"`
	Items        []CreateOrderItem `json:"items"`
}

func (s *Service) CreateOrder(ctx context.Context, req CreateOrderRequest) (*model.Order, error) {
	if err := s.users.ValidateUser(ctx, req.UserID); err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	itemIDs := make([]string, len(req.Items))
	qtyMap := make(map[string]int)
	for i, item := range req.Items {
		itemIDs[i] = item.MenuItemID
		qtyMap[item.MenuItemID] = item.Quantity
	}

	menuItems, err := s.restaurant.ValidateMenu(ctx, req.RestaurantID, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("invalid menu: %w", err)
	}

	var orderItems []model.OrderItem
	var total float64
	for _, mi := range menuItems {
		qty := qtyMap[mi.ID]
		if qty <= 0 {
			return nil, fmt.Errorf("invalid quantity for %s", mi.ID)
		}
		orderItems = append(orderItems, model.OrderItem{
			MenuItemID: mi.ID,
			Name:       mi.Name,
			Quantity:   qty,
			Price:      mi.Price,
		})
		total += mi.Price * float64(qty)
	}

	order := &model.Order{
		UserID:       req.UserID,
		RestaurantID: req.RestaurantID,
		TotalAmount:  total,
		Items:        orderItems,
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	eventItems := make([]events.OrderItem, len(orderItems))
	for i, item := range orderItems {
		eventItems[i] = events.OrderItem{
			MenuItemID: item.MenuItemID,
			Name:       item.Name,
			Quantity:   item.Quantity,
			Price:      item.Price,
		}
	}

	if err := s.producer.Publish(ctx, events.TopicOrderCreated, events.OrderCreated{
		OrderID:      order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
		TotalAmount:  order.TotalAmount,
		Items:        eventItems,
	}); err != nil {
		return nil, fmt.Errorf("publish order created: %w", err)
	}

	return order, nil
}

func (s *Service) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListUserOrders(ctx context.Context, userID string) ([]model.Order, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) HandlePaymentProcessed(ctx context.Context, payload events.PaymentProcessed) error {
	order, err := s.repo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return err
	}
	if isTerminal(order.Status) || order.Status == model.StatusPaid {
		return nil
	}

	if err := s.repo.UpdateStatus(ctx, payload.OrderID, model.StatusPaid); err != nil {
		return err
	}

	order, err = s.repo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return err
	}

	if err := s.producer.Publish(ctx, events.TopicOrderPaid, events.OrderPaid{
		OrderID:      order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
	}); err != nil {
		return err
	}

	return s.publishStatus(ctx, order.ID, model.StatusPaid)
}

func (s *Service) HandlePaymentFailed(ctx context.Context, payload events.PaymentFailed) error {
	return s.cancelOrder(ctx, payload.OrderID, payload.Reason, false)
}

func (s *Service) HandlePreparationFailed(ctx context.Context, payload events.OrderPreparationFailed) error {
	return s.cancelWithRefund(ctx, payload.OrderID, payload.UserID, payload.Reason)
}

func (s *Service) HandleDeliveryFailed(ctx context.Context, payload events.DeliveryFailed) error {
	return s.cancelWithRefund(ctx, payload.OrderID, payload.UserID, payload.Reason)
}

func (s *Service) HandlePaymentRefunded(ctx context.Context, payload events.PaymentRefunded) error {
	order, err := s.repo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return err
	}
	if order.Status == model.StatusRefunded {
		return nil
	}
	if err := s.repo.UpdateStatus(ctx, payload.OrderID, model.StatusRefunded); err != nil {
		return err
	}
	return s.publishStatus(ctx, payload.OrderID, model.StatusRefunded)
}

func (s *Service) HandleCourierAssigned(ctx context.Context, payload events.CourierAssigned) error {
	order, err := s.repo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return err
	}
	if order.Status == model.StatusDelivering || order.Status == model.StatusDelivered || isTerminal(order.Status) {
		return nil
	}
	return s.repo.AssignCourier(ctx, payload.OrderID, payload.CourierID)
}

func (s *Service) HandleOrderDelivered(ctx context.Context, payload events.OrderDelivered) error {
	order, err := s.repo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return err
	}
	if order.Status == model.StatusDelivered {
		return nil
	}
	if err := s.repo.UpdateStatus(ctx, payload.OrderID, model.StatusDelivered); err != nil {
		return err
	}
	return s.publishStatus(ctx, payload.OrderID, model.StatusDelivered)
}

func (s *Service) cancelWithRefund(ctx context.Context, orderID, userID, reason string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if isTerminal(order.Status) {
		return nil
	}

	refundRequired := order.Status == model.StatusPaid || order.Status == model.StatusDelivering

	if err := s.cancelOrder(ctx, orderID, reason, refundRequired); err != nil {
		return err
	}

	if !refundRequired {
		return nil
	}

	return s.producer.Publish(ctx, events.TopicPaymentRefundRequested, events.PaymentRefundRequested{
		OrderID: orderID,
		UserID:  userID,
		Amount:  order.TotalAmount,
		Reason:  reason,
	})
}

func (s *Service) cancelOrder(ctx context.Context, orderID, reason string, refundRequired bool) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if isTerminal(order.Status) {
		return nil
	}

	if err := s.repo.UpdateStatus(ctx, orderID, model.StatusCancelled); err != nil {
		return err
	}

	if err := s.publishStatus(ctx, orderID, model.StatusCancelled); err != nil {
		return err
	}

	return s.producer.Publish(ctx, events.TopicOrderCancelled, events.OrderCancelled{
		OrderID:        orderID,
		UserID:         order.UserID,
		Reason:         reason,
		RefundRequired: refundRequired,
	})
}

func (s *Service) publishStatus(ctx context.Context, orderID, status string) error {
	return s.producer.Publish(ctx, events.TopicOrderStatusChanged, events.OrderStatusChanged{
		OrderID: orderID,
		Status:  status,
	})
}

func isTerminal(status string) bool {
	return status == model.StatusCancelled || status == model.StatusRefunded || status == model.StatusDelivered
}
