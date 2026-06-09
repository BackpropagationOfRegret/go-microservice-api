package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/order/internal/model"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			restaurant_id TEXT NOT NULL,
			status TEXT NOT NULL,
			total_amount DOUBLE PRECISION NOT NULL,
			items JSONB NOT NULL,
			courier_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
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

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO orders (id, user_id, restaurant_id, status, total_amount, items, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.ID, order.UserID, order.RestaurantID, order.Status, order.TotalAmount, itemsJSON, order.CreatedAt, order.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var itemsJSON []byte
	o := &model.Order{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, restaurant_id, status, total_amount, items, COALESCE(courier_id,''), created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.TotalAmount, &itemsJSON, &o.CourierID, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, err
	}
	return o, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]model.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, restaurant_id, status, total_amount, items, COALESCE(courier_id,''), created_at, updated_at
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		var itemsJSON []byte
		if err := rows.Scan(&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.TotalAmount, &itemsJSON, &o.CourierID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) AssignCourier(ctx context.Context, id, courierID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET courier_id = $1, status = $2, updated_at = NOW() WHERE id = $3`,
		courierID, model.StatusDelivering, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}
