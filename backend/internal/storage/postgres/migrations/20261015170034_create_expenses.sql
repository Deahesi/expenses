-- +goose Up

CREATE TABLE expenses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL CHECK (amount > 0),
    category TEXT NOT NULL,
    spent_at TIMESTAMPTZ NOT NULL,
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expenses_user_id ON expenses(user_id);

-- +goose Down

DROP TABLE expenses;