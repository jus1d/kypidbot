package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jus1d/kypidbot/internal/domain"
)

type PairRepo struct {
	db *sql.DB
}

func NewPairRepo(d *DB) *PairRepo {
	return &PairRepo{db: d.db}
}

func (r *PairRepo) SavePair(ctx context.Context, p *domain.Pair) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pairs (dill_id, doe_id, score, time_intersection, is_fullmatch)
		VALUES ($1, $2, $3, $4, $5)`,
		p.DillID, p.DoeID, p.Score, p.TimeIntersection, p.IsFullmatch,
	)
	return err
}

func (r *PairRepo) GetPairByID(ctx context.Context, id int64) (*domain.Pair, error) {
	var p domain.Pair

	err := r.db.QueryRowContext(ctx, `
		SELECT id, dill_id, doe_id, score, time_intersection, is_fullmatch
		FROM pairs WHERE id = $1`, id).Scan(
		&p.ID, &p.DillID, &p.DoeID, &p.Score, &p.TimeIntersection, &p.IsFullmatch,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PairRepo) GetRegularPairs(ctx context.Context) ([]domain.Pair, error) {
	return r.getPairsByFullmatch(ctx, false)
}

func (r *PairRepo) GetFullPairs(ctx context.Context) ([]domain.Pair, error) {
	return r.getPairsByFullmatch(ctx, true)
}

func (r *PairRepo) getPairsByFullmatch(ctx context.Context, fullmatch bool) ([]domain.Pair, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, dill_id, doe_id, score, time_intersection, is_fullmatch
		FROM pairs WHERE is_fullmatch = $1`, fullmatch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pairs []domain.Pair
	for rows.Next() {
		var p domain.Pair
		if err := rows.Scan(&p.ID, &p.DillID, &p.DoeID, &p.Score, &p.TimeIntersection, &p.IsFullmatch); err != nil {
			return nil, err
		}
		pairs = append(pairs, p)
	}
	return pairs, rows.Err()
}

func (r *PairRepo) ClearPairs(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM pairs`)
	return err
}
