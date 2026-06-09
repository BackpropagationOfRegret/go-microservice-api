package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/restaurant/internal/model"
	"github.com/lib/pq"
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
		CREATE TABLE IF NOT EXISTS restaurants (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			address TEXT NOT NULL,
			is_open BOOLEAN NOT NULL DEFAULT TRUE,
			open_from TEXT NOT NULL DEFAULT '09:00',
			open_to TEXT NOT NULL DEFAULT '22:00'
		);
		CREATE TABLE IF NOT EXISTS menu_items (
			id TEXT PRIMARY KEY,
			restaurant_id TEXT NOT NULL REFERENCES restaurants(id),
			name TEXT NOT NULL,
			description TEXT,
			price DOUBLE PRECISION NOT NULL,
			available BOOLEAN NOT NULL DEFAULT TRUE
		);
	`)
	return err
}

func (r *Repository) Seed(ctx context.Context) error {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM restaurants`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	restID := uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO restaurants (id, name, address, is_open) VALUES ($1, $2, $3, TRUE)`,
		restID, "Pizza Palace", "123 Main St",
	)
	if err != nil {
		return err
	}

	items := []struct{ name, desc string; price float64 }{
		{"Margherita", "Classic tomato and mozzarella", 12.99},
		{"Pepperoni", "Spicy pepperoni pizza", 14.99},
		{"Caesar Salad", "Fresh romaine with parmesan", 8.99},
	}
	for _, item := range items {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO menu_items (id, restaurant_id, name, description, price) VALUES ($1, $2, $3, $4, $5)`,
			uuid.New().String(), restID, item.name, item.desc, item.price,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListRestaurants(ctx context.Context) ([]model.Restaurant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, address, is_open, open_from, open_to FROM restaurants`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Restaurant
	for rows.Next() {
		var rest model.Restaurant
		if err := rows.Scan(&rest.ID, &rest.Name, &rest.Address, &rest.IsOpen, &rest.OpenFrom, &rest.OpenTo); err != nil {
			return nil, err
		}
		list = append(list, rest)
	}
	return list, rows.Err()
}

func (r *Repository) GetRestaurant(ctx context.Context, id string) (*model.Restaurant, error) {
	rest := &model.Restaurant{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, address, is_open, open_from, open_to FROM restaurants WHERE id = $1`, id,
	).Scan(&rest.ID, &rest.Name, &rest.Address, &rest.IsOpen, &rest.OpenFrom, &rest.OpenTo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return rest, err
}

func (r *Repository) ListMenu(ctx context.Context, restaurantID string) ([]model.MenuItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, restaurant_id, name, description, price, available FROM menu_items WHERE restaurant_id = $1`,
		restaurantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.MenuItem
	for rows.Next() {
		var m model.MenuItem
		if err := rows.Scan(&m.ID, &m.RestaurantID, &m.Name, &m.Description, &m.Price, &m.Available); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *Repository) GetMenuItems(ctx context.Context, restaurantID string, ids []string) ([]model.MenuItem, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no menu items requested")
	}
	query := `SELECT id, restaurant_id, name, description, price, available FROM menu_items
		WHERE restaurant_id = $1 AND id = ANY($2) AND available = TRUE`
	rows, err := r.db.QueryContext(ctx, query, restaurantID, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.MenuItem
	for rows.Next() {
		var m model.MenuItem
		if err := rows.Scan(&m.ID, &m.RestaurantID, &m.Name, &m.Description, &m.Price, &m.Available); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}
