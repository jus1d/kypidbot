package usecase

import (
	"context"
	"errors"

	"github.com/jus1d/kypidbot/internal/domain"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrAlreadyAdmin = errors.New("user is already admin")
	ErrNotAdmin     = errors.New("user is not admin")
)

type Admin struct {
	users domain.UserRepository
}

func NewAdmin(users domain.UserRepository) *Admin {
	return &Admin{users: users}
}

func (a *Admin) Promote(ctx context.Context, username string) error {
	user, err := a.users.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.IsAdmin {
		return ErrAlreadyAdmin
	}
	return a.users.SetAdmin(ctx, user.TelegramID, true)
}

func (a *Admin) Demote(ctx context.Context, username string) error {
	user, err := a.users.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if !user.IsAdmin {
		return ErrNotAdmin
	}
	return a.users.SetAdmin(ctx, user.TelegramID, false)
}

func (a *Admin) GetStatistics(ctx context.Context) (domain.Statistics, error) {
	daily, weekly, err := a.users.GetLastRegisteredCount(ctx)
	if err != nil {
		return domain.Statistics{}, err
	}

	males, females, err := a.users.GetSexCounts(ctx)
	if err != nil {
		return domain.Statistics{}, err
	}

	return domain.Statistics{
		RegisteredDaily:  daily,
		RegisteredWeekly: weekly,
		MaleCount:        males,
		FemaleCount:      females,
	}, nil
}
