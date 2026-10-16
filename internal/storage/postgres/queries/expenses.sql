-- name: CreateExpense :one
INSERT INTO expenses (id, user_id, amount, category, spent_at, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, amount, category, spent_at, comment, created_at, updated_at;

-- name: GetExpense :one
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE id = $1 AND user_id = $2;

-- name: GetAllExpenses :many
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category')::text)
  AND (sqlc.narg('date_from')::date IS NULL OR spent_at >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR spent_at < sqlc.narg('date_to')::date + INTERVAL '1 day')
ORDER BY spent_at DESC, created_at DESC;

-- name: ListExpenses :many
SELECT id, user_id, amount, category, spent_at, comment, created_at, updated_at
FROM expenses
WHERE user_id = sqlc.arg('user_id')::uuid
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category')::text)
  AND (sqlc.narg('date_from')::date IS NULL OR spent_at >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR spent_at < sqlc.narg('date_to')::date + INTERVAL '1 day')
ORDER BY spent_at DESC, created_at DESC;

-- name: UpdateExpense :one
UPDATE expenses
SET amount = $3,
    category = $4,
    spent_at = $5,
    comment = $6,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, amount, category, spent_at, comment, created_at, updated_at;

-- name: DeleteExpense :execrows
DELETE FROM expenses
WHERE id = $1 AND user_id = $2;


-- name: GetExpenseWithUser :one
SELECT sqlc.embed(e), sqlc.embed(u)
FROM expenses e
JOIN users u ON u.id = e.user_id
WHERE e.id = $1 AND e.user_id = $2;

-- name: GetAllExpensesWithUser :many
SELECT sqlc.embed(e), sqlc.embed(u)
FROM expenses e
JOIN users u ON u.id = e.user_id
WHERE (sqlc.narg('user_id')::uuid IS NULL OR e.user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('category')::text IS NULL OR e.category = sqlc.narg('category')::text)
  AND (sqlc.narg('date_from')::date IS NULL OR e.spent_at >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR e.spent_at < sqlc.narg('date_to')::date + INTERVAL '1 day')
ORDER BY e.spent_at DESC, e.created_at DESC;

-- name: ListExpensesWithUser :many
SELECT sqlc.embed(e), sqlc.embed(u)
FROM expenses e
JOIN users u ON u.id = e.user_id
WHERE e.user_id = sqlc.arg('user_id')::uuid
  AND (sqlc.narg('category')::text IS NULL OR e.category = sqlc.narg('category')::text)
  AND (sqlc.narg('date_from')::date IS NULL OR e.spent_at >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR e.spent_at < sqlc.narg('date_to')::date + INTERVAL '1 day')
ORDER BY e.spent_at DESC, e.created_at DESC;
