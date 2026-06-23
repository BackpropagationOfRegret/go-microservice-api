-- name: CountRestaurants :one
SELECT COUNT(*)::bigint AS count
FROM restaurants;

-- name: CreateRestaurant :exec
INSERT INTO restaurants (id, name, address, is_open)
VALUES ($1, $2, $3, $4);

-- name: ListRestaurants :many
SELECT id, name, address, is_open, open_from, open_to
FROM restaurants;

-- name: GetRestaurant :one
SELECT id, name, address, is_open, open_from, open_to
FROM restaurants
WHERE id = $1;
