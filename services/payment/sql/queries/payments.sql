-- name: CreatePayment :exec
INSERT INTO payments (id, order_id, amount, method, status, transaction_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetPaymentByOrderID :one
SELECT id, order_id, amount, method, status, transaction_id, refund_transaction_id, refund_reason, created_at
FROM payments
WHERE order_id = $1;

-- name: MarkPaymentRefunded :execrows
UPDATE payments
SET status = $1, refund_transaction_id = $2, refund_reason = $3
WHERE order_id = $4
  AND status = $5;
