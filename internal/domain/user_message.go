package domain

import "context"

type UserMessageRepository interface {
	StoreMessageID(ctx context.Context, meetingID int64, telegramID int64, key string, messageID int) error
	GetMessageID(ctx context.Context, meetingID int64, telegramID int64, key string) (int, error)
}
