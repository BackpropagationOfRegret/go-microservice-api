package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
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
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			order_id TEXT UNIQUE NOT NULL,
			amount DOUBLE PRECISION NOT NULL,
			method TEXT NOT NULL DEFAULT 'card',
			status TEXT NOT NULL,
			transaction_id TEXT,
			refund_transaction_id TEXT,
			refund_reason TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_transaction_id TEXT;
		ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_reason TEXT;
	`)
	return err
}

func (r *Repository) Create(ctx context.Context, p *Payment) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	p.CreatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payments (id, order_id, amount, method, status, transaction_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		p.ID, p.OrderID, p.Amount, p.Method, p.Status, p.TransactionID, p.CreatedAt,
	)
	return err
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	p := &Payment{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, amount, method, status,
		        COALESCE(transaction_id,''), COALESCE(refund_transaction_id,''), COALESCE(refund_reason,''), created_at
		 FROM payments WHERE order_id = $1`,
		orderID,
	).Scan(&p.ID, &p.OrderID, &p.Amount, &p.Method, &p.Status, &p.TransactionID, &p.RefundTransactionID, &p.RefundReason, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *Repository) MarkRefunded(ctx context.Context, orderID, refundTxID, reason string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE payments SET status = $1, refund_transaction_id = $2, refund_reason = $3 WHERE order_id = $4 AND status = $5`,
		StatusRefunded, refundTxID, reason, orderID, StatusSuccess,
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
