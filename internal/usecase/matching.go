package usecase

import (
	"context"
	"fmt"

	"github.com/jus1d/kypidbot/internal/config"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/matcher"
)

type MatchResult struct {
	PairsCount     int
	FullMatchCount int
	UsersCount     int
}

type Matching struct {
	users    domain.UserRepository
	meetings domain.MeetingRepository
	ollama   *config.Ollama
}

func NewMatching(users domain.UserRepository, meetings domain.MeetingRepository, c *config.Ollama) *Matching {
	return &Matching{
		users:    users,
		meetings: meetings,
		ollama:   c,
	}
}

func (m *Matching) RunMatch(ctx context.Context) (*MatchResult, error) {
	users, err := m.users.GetVerifiedUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get verified users: %w", err)
	}

	if len(users) < 2 {
		return nil, fmt.Errorf("not enough users")
	}

	matchUsers := make([]matcher.MatchUser, len(users))
	for i, u := range users {
		matchUsers[i] = matcher.MatchUser{
			Index:      i,
			Username:   u.Username,
			Sex:        u.Sex,
			About:      u.About,
			TimeRanges: u.TimeRanges,
		}
	}

	pairs, fullMatches, err := matcher.Match(matchUsers, m.ollama)
	if err != nil {
		return nil, fmt.Errorf("match: %w", err)
	}

	if err := m.meetings.ClearMeetings(ctx); err != nil {
		return nil, fmt.Errorf("clear meetings: %w", err)
	}

	for _, p := range pairs {
		dill := users[p.I]
		doe := users[p.J]

		if err := m.meetings.SaveMeeting(ctx, &domain.Meeting{
			DillID:      dill.TelegramID,
			DoeID:       doe.TelegramID,
			PairScore:   p.Score,
			IsFullmatch: false,
		}); err != nil {
			return nil, fmt.Errorf("save meeting: %w", err)
		}
	}

	for _, fm := range fullMatches {
		dill := users[fm.I]
		doe := users[fm.J]

		if err := m.meetings.SaveMeeting(ctx, &domain.Meeting{
			DillID:      dill.TelegramID,
			DoeID:       doe.TelegramID,
			PairScore:   fm.Score,
			IsFullmatch: true,
		}); err != nil {
			return nil, fmt.Errorf("save full match: %w", err)
		}
	}

	return &MatchResult{
		PairsCount:     len(pairs),
		FullMatchCount: len(fullMatches),
		UsersCount:     len(users),
	}, nil
}
