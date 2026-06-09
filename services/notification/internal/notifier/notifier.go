package notifier

import (
	"context"
	"log"
	"sync"

	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
)

type Notification struct {
	Channel string `json:"channel"`
	UserID  string `json:"user_id,omitempty"`
	OrderID string `json:"order_id"`
	Message string `json:"message"`
}

type Service struct {
	mu    sync.Mutex
	sent  []Notification
}

func New() *Service {
	return &Service{sent: make([]Notification, 0)}
}

func (s *Service) Record(n Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, n)
	log.Printf("[notification-svc] %s → order %s: %s", n.Channel, n.OrderID, n.Message)
}

func (s *Service) List() []Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Notification, len(s.sent))
	copy(out, s.sent)
	return out
}

func (s *Service) send(channel, userID, orderID, message string) {
	s.Record(Notification{
		Channel: channel,
		UserID:  userID,
		OrderID: orderID,
		Message: message,
	})
}

func (s *Service) HandleEvent(ctx context.Context, env events.Envelope) error {
	switch env.Type {
	case events.TopicOrderCreated:
		p, err := kafka.DecodePayload[events.OrderCreated](env)
		if err != nil {
			return err
		}
		s.send("push", p.UserID, p.OrderID, "Заказ принят")
		s.send("email", p.UserID, p.OrderID, "Ваш заказ принят в обработку")

	case events.TopicPaymentProcessed:
		p, err := kafka.DecodePayload[events.PaymentProcessed](env)
		if err != nil {
			return err
		}
		s.send("push", "", p.OrderID, "Оплата прошла успешно")
		s.send("sms", "", p.OrderID, "Оплата подтверждена")

	case events.TopicPaymentFailed:
		p, err := kafka.DecodePayload[events.PaymentFailed](env)
		if err != nil {
			return err
		}
		s.send("push", "", p.OrderID, "Ошибка оплаты: "+p.Reason)

	case events.TopicOrderReady:
		p, err := kafka.DecodePayload[events.OrderReady](env)
		if err != nil {
			return err
		}
		s.send("push", "", p.OrderID, "Заказ готов, ожидайте курьера")

	case events.TopicCourierAssigned:
		p, err := kafka.DecodePayload[events.CourierAssigned](env)
		if err != nil {
			return err
		}
		s.send("push", p.UserID, p.OrderID, "Курьер в пути")
		s.send("sms", p.UserID, p.OrderID, "Курьер назначен и уже едет к вам")

	case events.TopicOrderDelivered:
		p, err := kafka.DecodePayload[events.OrderDelivered](env)
		if err != nil {
			return err
		}
		s.send("push", p.UserID, p.OrderID, "Приятного аппетита!")
		s.send("email", p.UserID, p.OrderID, "Заказ доставлен. Спасибо!")
	}
	return nil
}
