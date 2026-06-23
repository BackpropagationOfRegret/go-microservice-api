-- name: CreateAddress :exec
INSERT INTO addresses (id, user_id, label, street, city, latitude, longitude, is_default)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListAddressesByUser :many
SELECT id, user_id, label, street, city, latitude, longitude, is_default
FROM addresses
WHERE user_id = $1;
