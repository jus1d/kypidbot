-- +goose Up
ALTER TYPE user_state ADD VALUE IF NOT EXISTS 'awaiting_feedback';

CREATE TABLE IF NOT EXISTS feedback (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL REFERENCES users(telegram_id),
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS feedback;
-- Note: cannot remove enum value in PostgreSQL
