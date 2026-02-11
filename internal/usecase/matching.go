package usecase

import (
	"context"
	"fmt"

	"github.com/jus1d/kypidbot/internal/infrastructure/gemini"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/matcher"
)

type MatchResult struct {
	PairsCount     int
	FullMatchCount int
	UsersCount     int
	UnmatchedIDs   []int64
}

type DryPair struct {
	DillTelegramID int64
	DillFirstName  string
	DillUsername   string
	DoeTelegramID  int64
	DoeFirstName   string
	DoeUsername    string
}

type Matching struct {
	users    domain.UserRepository
	meetings domain.MeetingRepository
	gemini   *gemini.Client
}

func NewMatching(users domain.UserRepository, meetings domain.MeetingRepository, c *gemini.Client) *Matching {
	return &Matching{
		users:    users,
		meetings: meetings,
		gemini:   c,
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

	pairs, fullMatches, err := matcher.Match(matchUsers, m.gemini)
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

	matched := make(map[int]bool)
	for _, p := range pairs {
		matched[p.I] = true
		matched[p.J] = true
	}
	for _, fm := range fullMatches {
		matched[fm.I] = true
		matched[fm.J] = true
	}

	var unmatchedIDs []int64
	for i, u := range users {
		if !matched[i] {
			unmatchedIDs = append(unmatchedIDs, u.TelegramID)
		}
	}

	return &MatchResult{
		PairsCount:     len(pairs),
		FullMatchCount: len(fullMatches),
		UsersCount:     len(users),
		UnmatchedIDs:   unmatchedIDs,
	}, nil
}

func (m *Matching) DryMatch(ctx context.Context) ([]DryPair, error) {
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

	pairs, fullMatches, err := matcher.Match(matchUsers, m.gemini)
	if err != nil {
		return nil, fmt.Errorf("match: %w", err)
	}

	result := make([]DryPair, 0, len(pairs)+len(fullMatches))
	for _, p := range pairs {
		result = append(result, DryPair{
			DillTelegramID: users[p.I].TelegramID,
			DillFirstName:  users[p.I].FirstName,
			DillUsername:   users[p.I].Username,
			DoeTelegramID:  users[p.J].TelegramID,
			DoeFirstName:   users[p.J].FirstName,
			DoeUsername:    users[p.J].Username,
		})
	}
	for _, fm := range fullMatches {
		result = append(result, DryPair{
			DillTelegramID: users[fm.I].TelegramID,
			DillFirstName:  users[fm.I].FirstName,
			DillUsername:   users[fm.I].Username,
			DoeTelegramID:  users[fm.J].TelegramID,
			DoeFirstName:   users[fm.J].FirstName,
			DoeUsername:    users[fm.J].Username,
		})
	}

	return result, nil
}
