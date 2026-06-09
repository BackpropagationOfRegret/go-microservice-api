package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/delivery/internal/model"
)

var ErrNoCourier = errors.New("no available courier")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS couriers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'AVAILABLE',
			latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
			longitude DOUBLE PRECISION NOT NULL DEFAULT 0
		);
	`)
	return err
}

func (r *Repository) Seed(ctx context.Context) error {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM couriers`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	couriers := []struct{ name string; lat, lng float64 }{
		{"Alex Rider", 55.7558, 37.6173},
		{"Maria Swift", 55.7512, 37.6184},
		{"John Express", 55.7601, 37.6200},
	}
	for _, c := range couriers {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO couriers (id, name, status, latitude, longitude) VALUES ($1, $2, $3, $4, $5)`,
			uuid.New().String(), c.name, model.CourierAvailable, c.lat, c.lng,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) FindAvailableCourier(ctx context.Context) (*model.Courier, error) {
	c := &model.Courier{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, status, latitude, longitude FROM couriers WHERE status = $1 LIMIT 1`,
		model.CourierAvailable,
	).Scan(&c.ID, &c.Name, &c.Status, &c.Latitude, &c.Longitude)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoCourier
	}
	return c, err
}

func (r *Repository) SetCourierStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE couriers SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *Repository) ListCouriers(ctx context.Context) ([]model.Courier, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, status, latitude, longitude FROM couriers`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Courier
	for rows.Next() {
		var c model.Courier
		if err := rows.Scan(&c.ID, &c.Name, &c.Status, &c.Latitude, &c.Longitude); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}
