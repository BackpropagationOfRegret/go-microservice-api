package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/services/payment/internal/db"
	paymentsql "github.com/kostayne/go-microservice/services/payment/sql"
)

var ErrNotFound = errors.New("not found")

const (
	StatusSuccess  = "SUCCESS"
	StatusFailed   = "FAILED"
	StatusRefunded = "REFUNDED"
)

type Payment struct {
	ID                  string    `json:"id"`
	OrderID             string    `json:"order_id"`
	Amount              float64   `json:"amount"`
	Method              string    `json:"method"`
	Status              string    `json:"status"`
	TransactionID       string    `json:"transaction_id"`
	RefundTransactionID string    `json:"refund_transaction_id,omitempty"`
	RefundReason        string    `json:"refund_reason,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type Repository struct {
	dbConn *sql.DB
	q      *db.Queries
}

func New(dbConn *sql.DB) *Repository {
	return &Repository{dbConn: dbConn, q: db.New(dbConn)}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.dbConn.ExecContext(ctx, paymentsql.Schema)
	return err
}

func (r *Repository) Create(ctx context.Context, p *Payment) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	p.CreatedAt = time.Now().UTC()
	return r.q.CreatePayment(ctx, db.CreatePaymentParams{
		ID:            p.ID,
		OrderID:       p.OrderID,
		Amount:        p.Amount,
		Method:        p.Method,
		Status:        p.Status,
		TransactionID: toNullString(p.TransactionID),
		CreatedAt:     p.CreatedAt,
	})
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	p, err := r.q.GetPaymentByOrderID(ctx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toPaymentModel(p), nil
}

func (r *Repository) MarkRefunded(ctx context.Context, orderID, refundTxID, reason string) error {
	n, err := r.q.MarkPaymentRefunded(ctx, db.MarkPaymentRefundedParams{
		Status:              StatusRefunded,
		RefundTransactionID: toNullString(refundTxID),
		RefundReason:        toNullString(reason),
		OrderID:             orderID,
		Status_2:            StatusSuccess,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func toPaymentModel(p db.Payment) *Payment {
	return &Payment{
		ID:                  p.ID,
		OrderID:             p.OrderID,
		Amount:              p.Amount,
		Method:              p.Method,
		Status:              p.Status,
		TransactionID:       fromNullString(p.TransactionID),
		RefundTransactionID: fromNullString(p.RefundTransactionID),
		RefundReason:        fromNullString(p.RefundReason),
		CreatedAt:           p.CreatedAt,
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
