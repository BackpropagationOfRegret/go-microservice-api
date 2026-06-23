package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/restaurant/internal/db"
	"github.com/kostayne/go-microservice/services/restaurant/internal/model"
	restaurantsql "github.com/kostayne/go-microservice/services/restaurant/sql"
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
	_, err := r.dbConn.ExecContext(ctx, restaurantsql.Schema)
	return err
}

func (r *Repository) Seed(ctx context.Context) error {
	count, err := r.q.CountRestaurants(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	restID := uuid.New().String()
	if err := r.q.CreateRestaurant(ctx, db.CreateRestaurantParams{
		ID:      restID,
		Name:    "Pizza Palace",
		Address: "123 Main St",
		IsOpen:  true,
	}); err != nil {
		return err
	}

	items := []struct {
		name, desc string
		price      float64
	}{
		{"Margherita", "Classic tomato and mozzarella", 12.99},
		{"Pepperoni", "Spicy pepperoni pizza", 14.99},
		{"Caesar Salad", "Fresh romaine with parmesan", 8.99},
	}
	for _, item := range items {
		if err := r.q.CreateMenuItem(ctx, db.CreateMenuItemParams{
			ID:           uuid.New().String(),
			RestaurantID: restID,
			Name:         item.name,
			Description:  toNullString(item.desc),
			Price:        item.price,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListRestaurants(ctx context.Context) ([]model.Restaurant, error) {
	rows, err := r.q.ListRestaurants(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]model.Restaurant, 0, len(rows))
	for _, row := range rows {
		list = append(list, toRestaurantModel(row))
	}
	return list, nil
}

func (r *Repository) GetRestaurant(ctx context.Context, id string) (*model.Restaurant, error) {
	rest, err := r.q.GetRestaurant(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	m := toRestaurantModel(rest)
	return &m, nil
}

func (r *Repository) ListMenu(ctx context.Context, restaurantID string) ([]model.MenuItem, error) {
	rows, err := r.q.ListMenuByRestaurant(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	items := make([]model.MenuItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMenuItemModel(row))
	}
	return items, nil
}

func (r *Repository) GetMenuItems(ctx context.Context, restaurantID string, ids []string) ([]model.MenuItem, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no menu items requested")
	}
	rows, err := r.q.GetMenuItems(ctx, db.GetMenuItemsParams{
		RestaurantID: restaurantID,
		Ids:          ids,
	})
	if err != nil {
		return nil, err
	}
	items := make([]model.MenuItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMenuItemModel(row))
	}
	return items, nil
}

func toRestaurantModel(r db.Restaurant) model.Restaurant {
	return model.Restaurant{
		ID:       r.ID,
		Name:     r.Name,
		Address:  r.Address,
		IsOpen:   r.IsOpen,
		OpenFrom: r.OpenFrom,
		OpenTo:   r.OpenTo,
	}
}

func toMenuItemModel(m db.MenuItem) model.MenuItem {
	return model.MenuItem{
		ID:           m.ID,
		RestaurantID: m.RestaurantID,
		Name:         m.Name,
		Description:  fromNullString(m.Description),
		Price:        m.Price,
		Available:    m.Available,
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
