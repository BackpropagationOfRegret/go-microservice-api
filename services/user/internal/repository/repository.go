package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/user/internal/db"
	"github.com/kostayne/go-microservice/services/user/internal/model"
	usersql "github.com/kostayne/go-microservice/services/user/sql"
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
	_, err := r.dbConn.ExecContext(ctx, usersql.Schema)
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
	err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		Phone:        toNullString(u.Phone),
	})
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUserModel(u), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUserModel(u), nil
}

func (r *Repository) AddAddress(ctx context.Context, addr *model.Address) error {
	if addr.ID == "" {
		addr.ID = uuid.New().String()
	}
	return r.q.CreateAddress(ctx, db.CreateAddressParams{
		ID:        addr.ID,
		UserID:    addr.UserID,
		Label:     addr.Label,
		Street:    addr.Street,
		City:      addr.City,
		Latitude:  addr.Latitude,
		Longitude: addr.Longitude,
		IsDefault: addr.IsDefault,
	})
}

func (r *Repository) ListAddresses(ctx context.Context, userID string) ([]model.Address, error) {
	rows, err := r.q.ListAddressesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	addrs := make([]model.Address, 0, len(rows))
	for _, row := range rows {
		addrs = append(addrs, toAddressModel(row))
	}
	return addrs, nil
}

func toUserModel(u db.User) *model.User {
	return &model.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		Phone:        fromNullString(u.Phone),
		CreatedAt:    u.CreatedAt,
	}
}

func toAddressModel(a db.Address) model.Address {
	return model.Address{
		ID:        a.ID,
		UserID:    a.UserID,
		Label:     a.Label,
		Street:    a.Street,
		City:      a.City,
		Latitude:  a.Latitude,
		Longitude: a.Longitude,
		IsDefault: a.IsDefault,
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
