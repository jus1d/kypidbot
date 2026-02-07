package usecase

import (
	"context"

	"github.com/jus1d/kypidbot/internal/domain"
)

type Registration struct {
	users domain.UserRepository
}

func NewRegistration(users domain.UserRepository) *Registration {
	return &Registration{users: users}
}

func (r *Registration) SaveUser(ctx context.Context, u *domain.User) error {
	return r.users.SaveUser(ctx, u)
}

func (r *Registration) GetUser(ctx context.Context, telegramID int64) (*domain.User, error) {
	return r.users.GetUser(ctx, telegramID)
}

func (r *Registration) SetState(ctx context.Context, telegramID int64, state string) error {
	return r.users.SetUserState(ctx, telegramID, state)
}

func (r *Registration) GetState(ctx context.Context, telegramID int64) (string, error) {
	return r.users.GetUserState(ctx, telegramID)
}

func (r *Registration) SetSex(ctx context.Context, telegramID int64, sex string) error {
	return r.users.SetUserSex(ctx, telegramID, sex)
}

func (r *Registration) SetAbout(ctx context.Context, telegramID int64, about string) error {
	return r.users.SetUserAbout(ctx, telegramID, about)
}

func (r *Registration) GetTimeRanges(ctx context.Context, telegramID int64) (string, error) {
	return r.users.GetTimeRanges(ctx, telegramID)
}

func (r *Registration) SaveTimeRanges(ctx context.Context, telegramID int64, timeRanges string) error {
	return r.users.SaveTimeRanges(ctx, telegramID, timeRanges)
}

func (r *Registration) GetUserByReferralCode(ctx context.Context, code string) (*domain.User, error) {
	return r.users.GetUserByReferralCode(ctx, code)
}

func (r *Registration) SetReferralCode(ctx context.Context, telegramID int64, code string) error {
	return r.users.SetReferralCode(ctx, telegramID, code)
}

func (r *Registration) SetReferrer(ctx context.Context, telegramID int64, referrerID int64) error {
	return r.users.SetReferrer(ctx, telegramID, referrerID)
}
