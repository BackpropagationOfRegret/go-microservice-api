package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/order/internal/db"
	"github.com/kostayne/go-microservice/services/order/internal/model"
	ordersql "github.com/kostayne/go-microservice/services/order/sql"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	dbConn *sql.DB
	q      *db.Queries
}

func New(dbConn *sql.DB) *Repository {
	return &Repository{dbConn: dbConn, q: db.New(dbConn)}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.dbConn.ExecContext(ctx, ordersql.Schema)
	return err
}

func (r *Repository) Create(ctx context.Context, order *model.Order) error {
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	order.Status = model.StatusPending
	order.CreatedAt = time.Now().UTC()
	order.UpdatedAt = order.CreatedAt

	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	return r.q.CreateOrder(ctx, db.CreateOrderParams{
		ID:           order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
		Status:       order.Status,
		TotalAmount:  order.TotalAmount,
		Items:        itemsJSON,
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
	})
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	o, err := r.q.GetOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toOrderModel(o)
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]model.Order, error) {
	rows, err := r.q.ListOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	orders := make([]model.Order, 0, len(rows))
	for _, row := range rows {
		o, err := toOrderModel(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}
	return orders, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	n, err := r.q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		Status: status,
		ID:     id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) AssignCourier(ctx context.Context, id, courierID string) error {
	n, err := r.q.AssignCourier(ctx, db.AssignCourierParams{
		CourierID: sql.NullString{String: courierID, Valid: true},
		Status:    model.StatusDelivering,
		ID:        id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}

func toOrderModel(o db.Order) (*model.Order, error) {
	var items []model.OrderItem
	if err := json.Unmarshal(o.Items, &items); err != nil {
		return nil, err
	}
	return &model.Order{
		ID:           o.ID,
		UserID:       o.UserID,
		RestaurantID: o.RestaurantID,
		Status:       o.Status,
		TotalAmount:  o.TotalAmount,
		Items:        items,
		CourierID:    fromNullString(o.CourierID),
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}, nil
}

func fromNullString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
