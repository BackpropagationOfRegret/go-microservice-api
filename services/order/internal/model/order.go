package model

import "time"

const (
	StatusPending   = "PENDING"
	StatusPaid      = "PAID"
	StatusPreparing = "PREPARING"
	StatusReady     = "READY"
	StatusDelivering = "DELIVERING"
	StatusDelivered = "DELIVERED"
	StatusCancelled = "CANCELLED"
	StatusRefunded  = "REFUNDED"
)

type OrderItem struct {
	MenuItemID string  `json:"menu_item_id"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

type Order struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	RestaurantID string      `json:"restaurant_id"`
	Status       string      `json:"status"`
	TotalAmount  float64     `json:"total_amount"`
	Items        []OrderItem `json:"items"`
	CourierID    string      `json:"courier_id,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
