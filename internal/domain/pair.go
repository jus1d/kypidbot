package domain

import "context"

type Pair struct {
	ID               int64
	DillID           int64
	DoeID            int64
	Score            float64
	TimeIntersection string
	IsFullmatch      bool
}

type PairRepository interface {
	SavePair(ctx context.Context, p *Pair) error
	GetPairByID(ctx context.Context, id int64) (*Pair, error)
	GetRegularPairs(ctx context.Context) ([]Pair, error)
	GetFullPairs(ctx context.Context) ([]Pair, error)
	ClearPairs(ctx context.Context) error
}
