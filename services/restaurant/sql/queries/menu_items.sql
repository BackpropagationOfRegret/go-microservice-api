-- name: CreateMenuItem :exec
INSERT INTO menu_items (id, restaurant_id, name, description, price)
VALUES ($1, $2, $3, $4, $5);

-- name: ListMenuByRestaurant :many
SELECT id, restaurant_id, name, description, price, available
FROM menu_items
WHERE restaurant_id = $1;

-- name: GetMenuItems :many
SELECT id, restaurant_id, name, description, price, available
FROM menu_items
WHERE restaurant_id = sqlc.arg(restaurant_id)
  AND id = ANY (sqlc.arg(ids)::text[])
  AND available = TRUE;
