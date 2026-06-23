-- name: CreateUser :exec
INSERT INTO users (id, email, password_hash, name, phone)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, phone, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, name, phone, created_at
FROM users
WHERE id = $1;
