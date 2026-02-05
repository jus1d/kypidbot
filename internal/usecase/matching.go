package usecase

import (
	"context"
	"fmt"

	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/matcher"
)

type MatchResult struct {
	PairsCount      int
	FullMatchCount  int
	UsersCount      int
}

type Matching struct {
	users domain.UserRepository
	pairs domain.PairRepository
	url   string
}

func NewMatching(users domain.UserRepository, pairs domain.PairRepository, ollamaURL string) *Matching {
	return &Matching{
		users: users,
		pairs: pairs,
		url:   ollamaURL,
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

	pairs, fullMatches, err := matcher.Match(matchUsers, m.url)
	if err != nil {
		return nil, fmt.Errorf("match: %w", err)
	}

	if err := m.pairs.ClearPairs(ctx); err != nil {
		return nil, fmt.Errorf("clear pairs: %w", err)
	}

	for _, p := range pairs {
		dill := users[p.I]
		doe := users[p.J]

		if err := m.pairs.SavePair(ctx, &domain.Pair{
			DillID:           dill.ID,
			DoeID:            doe.ID,
			Score:            p.Score,
			TimeIntersection: p.TimeIntersection,
			IsFullmatch:      false,
		}); err != nil {
			return nil, fmt.Errorf("save pair: %w", err)
		}
	}

	for _, fm := range fullMatches {
		dill := users[fm.I]
		doe := users[fm.J]

		if err := m.pairs.SavePair(ctx, &domain.Pair{
			DillID:           dill.ID,
			DoeID:            doe.ID,
			Score:            fm.Score,
			TimeIntersection: "000000",
			IsFullmatch:      true,
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
