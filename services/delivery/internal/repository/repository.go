package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/delivery/internal/db"
	"github.com/kostayne/go-microservice/services/delivery/internal/model"
	deliverysql "github.com/kostayne/go-microservice/services/delivery/sql"
)

var ErrNoCourier = errors.New("no available courier")

type Repository struct {
	dbConn *sql.DB
	q      *db.Queries
}

func New(dbConn *sql.DB) *Repository {
	return &Repository{dbConn: dbConn, q: db.New(dbConn)}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.dbConn.ExecContext(ctx, deliverysql.Schema)
	return err
}

func (r *Repository) Seed(ctx context.Context) error {
	count, err := r.q.CountCouriers(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	couriers := []struct {
		name      string
		lat, lng  float64
	}{
		{"Alex Rider", 55.7558, 37.6173},
		{"Maria Swift", 55.7512, 37.6184},
		{"John Express", 55.7601, 37.6200},
	}
	for _, c := range couriers {
		if err := r.q.CreateCourier(ctx, db.CreateCourierParams{
			ID:        uuid.New().String(),
			Name:      c.name,
			Status:    model.CourierAvailable,
			Latitude:  c.lat,
			Longitude: c.lng,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) FindAvailableCourier(ctx context.Context) (*model.Courier, error) {
	c, err := r.q.FindAvailableCourier(ctx, model.CourierAvailable)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoCourier
	}
	if err != nil {
		return nil, err
	}
	return toCourierModel(c), nil
}

func (r *Repository) SetCourierStatus(ctx context.Context, id, status string) error {
	return r.q.SetCourierStatus(ctx, db.SetCourierStatusParams{
		Status: status,
		ID:     id,
	})
}

func (r *Repository) ListCouriers(ctx context.Context) ([]model.Courier, error) {
	rows, err := r.q.ListCouriers(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]model.Courier, 0, len(rows))
	for _, row := range rows {
		list = append(list, *toCourierModel(row))
	}
	return list, nil
}

func toCourierModel(c db.Courier) *model.Courier {
	return &model.Courier{
		ID:        c.ID,
		Name:      c.Name,
		Status:    c.Status,
		Latitude:  c.Latitude,
		Longitude: c.Longitude,
	}
}
