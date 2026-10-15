-- name: CreateExpense :one
INSERT INTO expenses (id, user_id, amount, category, spent_at, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, amount, category, spent_at, comment, created_at, updated_at;

-- name: GetExpense :one
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE id = $1;

-- name: ListExpenses :many
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category')::text)
  AND (sqlc.narg('date_from')::date IS NULL OR spent_at >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR spent_at <= sqlc.narg('date_to')::date)
ORDER BY spent_at DESC, created_at DESC;

-- name: UpdateExpense :one
UPDATE expenses
SET amount = $2,
    category = $3,
    spent_at = $4,
    comment = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, amount, category, spent_at, comment, created_at, updated_at;

-- name: DeleteExpense :execrows
DELETE FROM expenses
WHERE id = $1;
