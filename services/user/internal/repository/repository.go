package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/user/internal/model"
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
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			phone TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS addresses (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id),
			label TEXT NOT NULL,
			street TEXT NOT NULL,
			city TEXT NOT NULL,
			latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
			longitude DOUBLE PRECISION NOT NULL DEFAULT 0,
			is_default BOOLEAN NOT NULL DEFAULT FALSE
		);
	`)
	return err
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, name, phone string) (*model.User, error) {
	u := &model.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Phone:        phone,
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, name, phone) VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Phone,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, phone, created_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, phone, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) AddAddress(ctx context.Context, addr *model.Address) error {
	if addr.ID == "" {
		addr.ID = uuid.New().String()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO addresses (id, user_id, label, street, city, latitude, longitude, is_default)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		addr.ID, addr.UserID, addr.Label, addr.Street, addr.City, addr.Latitude, addr.Longitude, addr.IsDefault,
	)
	return err
}

func (r *Repository) ListAddresses(ctx context.Context, userID string) ([]model.Address, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, label, street, city, latitude, longitude, is_default FROM addresses WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []model.Address
	for rows.Next() {
		var a model.Address
		if err := rows.Scan(&a.ID, &a.UserID, &a.Label, &a.Street, &a.City, &a.Latitude, &a.Longitude, &a.IsDefault); err != nil {
			return nil, err
		}
		addrs = append(addrs, a)
	}
	return addrs, rows.Err()
}
