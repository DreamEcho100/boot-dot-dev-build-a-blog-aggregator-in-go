-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING id, created_at, updated_at, name;

-- name: GetUser :one
SELECT
  id,
  name,
  created_at,
  updated_at
FROM users
WHERE name = $1;

-- name: GetUsers :many
SELECT
  id,
  created_at,
  updated_at,
  name
FROM users;

-- name: ResetUsers :exec
DELETE FROM users;
