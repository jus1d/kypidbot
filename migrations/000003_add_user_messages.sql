-- +goose Up
CREATE TABLE user_messages (
    meeting_id INTEGER NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    telegram_id BIGINT NOT NULL REFERENCES users(telegram_id),
    key TEXT NOT NULL,
    message_id INTEGER NOT NULL,
    PRIMARY KEY (meeting_id, telegram_id, key)
);

-- +goose Down
DROP TABLE IF EXISTS user_messages;
