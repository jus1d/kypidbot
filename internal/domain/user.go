package domain

import "context"

type User struct {
	ID           int64
	TelegramID   int64
	Username     string
	FirstName    string
	LastName     string
	IsBot        bool
	LanguageCode string
	IsPremium    bool
	Sex          string
	About        string
	State        string
	TimeRanges   string
	IsAdmin      bool
}

type UserRepository interface {
	SaveUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, telegramID int64) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserState(ctx context.Context, telegramID int64) (string, error)
	SetUserState(ctx context.Context, telegramID int64, state string) error
	SetUserSex(ctx context.Context, telegramID int64, sex string) error
	SetUserAbout(ctx context.Context, telegramID int64, about string) error
	GetTimeRanges(ctx context.Context, telegramID int64) (string, error)
	SaveTimeRanges(ctx context.Context, telegramID int64, timeRanges string) error
	IsAdmin(ctx context.Context, telegramID int64) (bool, error)
	SetAdmin(ctx context.Context, telegramID int64, isAdmin bool) error
	GetVerifiedUsers(ctx context.Context) ([]User, error)
	GetUserUsername(ctx context.Context, telegramID int64) (string, error)
}
