package postgres

import (
	"context"
	"database/sql"
	"errors"
)

type UserMessageRepo struct {
	db *sql.DB
}

func NewUserMessageRepo(d *DB) *UserMessageRepo {
	return &UserMessageRepo{db: d.db}
}

func (r *UserMessageRepo) StoreMessageID(ctx context.Context, meetingID int64, telegramID int64, key string, messageID int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_messages (meeting_id, telegram_id, key, message_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (meeting_id, telegram_id, key) DO UPDATE SET message_id = EXCLUDED.message_id`,
		meetingID, telegramID, key, messageID)
	return err
}

func (r *UserMessageRepo) GetMessageID(ctx context.Context, meetingID int64, telegramID int64, key string) (int, error) {
	var msgID int
	err := r.db.QueryRowContext(ctx, `
		SELECT message_id FROM user_messages
		WHERE meeting_id = $1 AND telegram_id = $2 AND key = $3`,
		meetingID, telegramID, key).Scan(&msgID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return msgID, err
}
