package postgres

import (
	"context"
	"database/sql"
)

type FeedbackRepo struct {
	db *sql.DB
}

func NewFeedbackRepo(d *DB) *FeedbackRepo {
	return &FeedbackRepo{db: d.db}
}

func (r *FeedbackRepo) Save(ctx context.Context, telegramID int64, text string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO feedback (telegram_id, text) VALUES ($1, $2)`,
		telegramID, text)
	return err
}
