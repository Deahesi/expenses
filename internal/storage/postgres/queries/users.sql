-- name: CreateUser :one
INSERT INTO users (id, telegram_id, first_name, last_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpsertTelegramUser :one
INSERT INTO users (id, telegram_id, first_name, last_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (telegram_id)
DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    updated_at = NOW()
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByTelegramID :one
SELECT * FROM users
WHERE telegram_id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at, id;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = $2,
    last_name = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;