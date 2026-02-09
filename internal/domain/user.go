package domain

import (
	"context"
	"time"
)

type User struct {
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
	ReferralCode string
	ReferrerID   *int64
	CreatedAt    time.Time
}

type ReferralLeaderboardEntry struct {
    ReferrerID    int64  `json:"referrer_id"`
    ReferralCount int    `json:"referral_count"`
    Username      string `json:"username"`
    FirstName     string `json:"first_name"`
}

type UserRepository interface {
	SaveUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, telegramID int64) (*User, error)
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
	GetAdmins(ctx context.Context) ([]User, error)
	GetUserByReferralCode(ctx context.Context, code string) (*User, error)
	SetReferralCode(ctx context.Context, telegramID int64, code string) error
	SetReferrer(ctx context.Context, telegramID int64, referrerID int64) error
	GetReferralLeaderboard(ctx context.Context, limit int) ([]ReferralLeaderboardEntry, error)
	GetUserReferralCount(ctx context.Context, referrerID int64) (int, error)
    GetUserLeaderboardPosition(ctx context.Context, userID int64) (int, error)
}
