-- name: CreateOrder :exec
INSERT INTO orders (id, user_id, restaurant_id, status, total_amount, items, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetOrderByID :one
SELECT id, user_id, restaurant_id, status, total_amount, items, courier_id, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: ListOrdersByUser :many
SELECT id, user_id, restaurant_id, status, total_amount, items, courier_id, created_at, updated_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateOrderStatus :execrows
UPDATE orders
SET status = $1, updated_at = NOW()
WHERE id = $2;

-- name: AssignCourier :execrows
UPDATE orders
SET courier_id = $1, status = $2, updated_at = NOW()
WHERE id = $3;
