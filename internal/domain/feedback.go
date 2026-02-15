package domain

import (
	"context"
	"time"
)

type Feedback struct {
	ID         int64
	TelegramID int64
	Text       string
	CreatedAt  time.Time
}

type FeedbackRepository interface {
	Save(ctx context.Context, telegramID int64, text string) error
}
