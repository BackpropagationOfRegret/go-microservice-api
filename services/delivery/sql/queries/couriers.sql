-- name: CountCouriers :one
SELECT COUNT(*)::bigint AS count
FROM couriers;

-- name: CreateCourier :exec
INSERT INTO couriers (id, name, status, latitude, longitude)
VALUES ($1, $2, $3, $4, $5);

-- name: FindAvailableCourier :one
SELECT id, name, status, latitude, longitude
FROM couriers
WHERE status = $1
LIMIT 1;

-- name: SetCourierStatus :exec
UPDATE couriers
SET status = $1
WHERE id = $2;

-- name: ListCouriers :many
SELECT id, name, status, latitude, longitude
FROM couriers;
