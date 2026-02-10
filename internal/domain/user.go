package domain

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

type UserState string

const (
	UserStateStart              UserState = "start"
	UserStateAwaitingSex        UserState = "awaiting_sex"
	UserStateAwaitingAbout      UserState = "awaiting_about"
	UserStateAwaitingTime       UserState = "awaiting_time"
	UserStateAwaitingSupport    UserState = "awaiting_support"
	UserStateAwaitingAppearance UserState = "awaiting_appearance"
	UserStateCompleted          UserState = "completed"
)

type User struct {
	TelegramID           int64
	Username             string
	FirstName            string
	LastName             string
	IsBot                bool
	LanguageCode         string
	IsPremium            bool
	Sex                  string
	About                string
	State                UserState
	RegistrationNotified bool
	InviteNotified       bool
	TimeRanges           string
	IsAdmin              bool
	OptedOut             bool
	ReferralCode         string
	ReferrerID           *int64
	CreatedAt            time.Time
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
	GetUserState(ctx context.Context, telegramID int64) (UserState, error)
	SetUserState(ctx context.Context, telegramID int64, state UserState) error
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
	GetReferralLeaderboard(ctx context.Context) ([]ReferralLeaderboardEntry, error)
	MarkNotified(ctx context.Context, telegramID int64) error
	GetNotCompleted(ctx context.Context, interval time.Duration) ([]User, error)
	GetForInviteReminder(ctx context.Context, interval time.Duration) ([]User, error)
	MarkInviteNotified(ctx context.Context, telegramID int64) error
	SetOptedOut(ctx context.Context, telegramID int64, optedOut bool) error
	GetLastRegisteredCount(ctx context.Context) (daily uint, weekly uint, err error)
	GetSexCounts(ctx context.Context) (males uint, females uint, err error)
}

const (
	referralCodeLen     = 8
	referralCodeCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func GenerateReferralCode() (string, error) {
	b := make([]byte, referralCodeLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(referralCodeCharset))))
		if err != nil {
			return "", err
		}
		b[i] = referralCodeCharset[n.Int64()]
	}
	return string(b), nil
}
