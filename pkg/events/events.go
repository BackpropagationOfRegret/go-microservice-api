package events

import "time"

const (
	TopicOrderCreated           = "order.created"
	TopicPaymentProcessed       = "payment.processed"
	TopicPaymentFailed          = "payment.failed"
	TopicPaymentRefundRequested = "payment.refund.requested"
	TopicPaymentRefunded        = "payment.refunded"
	TopicOrderPaid              = "order.paid"
	TopicOrderPreparationFailed = "order.preparation.failed"
	TopicOrderReady             = "order.ready"
	TopicDeliveryFailed         = "delivery.failed"
	TopicCourierAssigned        = "courier.assigned"
	TopicOrderDelivered         = "order.delivered"
	TopicOrderCancelled         = "order.cancelled"
	TopicOrderStatusChanged     = "order.status.changed"
)

// AllTopics returns every Kafka topic used by the platform (for init scripts).
func AllTopics() []string {
	return []string{
		TopicOrderCreated,
		TopicPaymentProcessed,
		TopicPaymentFailed,
		TopicPaymentRefundRequested,
		TopicPaymentRefunded,
		TopicOrderPaid,
		TopicOrderPreparationFailed,
		TopicOrderReady,
		TopicDeliveryFailed,
		TopicCourierAssigned,
		TopicOrderDelivered,
		TopicOrderCancelled,
		TopicOrderStatusChanged,
	}
}

type Envelope struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

type OrderItem struct {
	MenuItemID string  `json:"menu_item_id"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

type OrderCreated struct {
	OrderID      string      `json:"order_id"`
	UserID       string      `json:"user_id"`
	RestaurantID string      `json:"restaurant_id"`
	TotalAmount  float64     `json:"total_amount"`
	Items        []OrderItem `json:"items"`
}

type PaymentProcessed struct {
	OrderID       string  `json:"order_id"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
}

type PaymentFailed struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type OrderPaid struct {
	OrderID      string `json:"order_id"`
	UserID       string `json:"user_id"`
	RestaurantID string `json:"restaurant_id"`
}

type OrderReady struct {
	OrderID      string `json:"order_id"`
	RestaurantID string `json:"restaurant_id"`
}

type CourierAssigned struct {
	OrderID   string `json:"order_id"`
	CourierID string `json:"courier_id"`
	UserID    string `json:"user_id"`
}

type OrderDelivered struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

type OrderStatusChanged struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type PaymentRefundRequested struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
	Reason  string  `json:"reason"`
}

type PaymentRefunded struct {
	OrderID              string  `json:"order_id"`
	RefundTransactionID  string  `json:"refund_transaction_id"`
	Amount               float64 `json:"amount"`
	Reason               string  `json:"reason"`
}

type OrderPreparationFailed struct {
	OrderID      string `json:"order_id"`
	RestaurantID string `json:"restaurant_id"`
	UserID       string `json:"user_id"`
	Reason       string `json:"reason"`
}

type DeliveryFailed struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Reason  string `json:"reason"`
}

type OrderCancelled struct {
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	Reason         string `json:"reason"`
	RefundRequired bool   `json:"refund_required"`
}
